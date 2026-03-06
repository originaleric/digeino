package webhook

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/config"
)

// StatusLogger 日志接口，调用方可选择性实现
type StatusLogger interface {
	OnNodeStartLog(nodeKey, nodeType string, inputCount int)
	OnNodeEndLog(nodeKey, nodeType string, outputCount int, err error)
	OnCompleteLog(executionID string, duration time.Duration, usage *Usage)
}

// StatusCollector 状态收集器
type StatusCollector struct {
	executionID string
	appName     string
	requestID   string
	startTime   time.Time

	webhookClients []*WebhookClient
	statusStore    StatusStoreInterface
	statusCallback func(status ExecutionStatus)

	mu            sync.Mutex
	statusHistory []ExecutionStatus
	nodeStartTime map[string]time.Time
	path          []string

	tokenUsageMap map[string]*Usage
	totalUsage    Usage

	enableDataFlow bool
	logger         StatusLogger
}

// NewStatusCollector 创建状态收集器，DataFlow 开关从 config.Get().Status.DataFlow 读取，默认关闭
func NewStatusCollector(executionID, appName, requestID string) *StatusCollector {
	enableDataFlow := false
	if cfg := config.Get(); cfg != nil && cfg.Status.DataFlow.Enabled != nil && *cfg.Status.DataFlow.Enabled {
		enableDataFlow = true
	}
	return &StatusCollector{
		executionID:    executionID,
		appName:        appName,
		requestID:      requestID,
		startTime:      time.Now(),
		nodeStartTime:  make(map[string]time.Time),
		tokenUsageMap:  make(map[string]*Usage),
		enableDataFlow: enableDataFlow,
	}
}

// EnableDataFlow 启用/禁用 DataFlow 追踪（优先级高于 eino.yml 配置）
func (sc *StatusCollector) EnableDataFlow(enable bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.enableDataFlow = enable
}

// SetLogger 设置日志 hook
func (sc *StatusCollector) SetLogger(logger StatusLogger) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.logger = logger
}

// AddWebhookClient 添加 Webhook 客户端
func (sc *StatusCollector) AddWebhookClient(client *WebhookClient) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.webhookClients = append(sc.webhookClients, client)
}

// SetStatusStore 设置状态存储
func (sc *StatusCollector) SetStatusStore(store StatusStoreInterface) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.statusStore = store
}

// SetStatusCallback 设置状态回调
func (sc *StatusCollector) SetStatusCallback(callback func(status ExecutionStatus)) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.statusCallback = callback
}

// OnNodeStart 节点开始执行
func (sc *StatusCollector) OnNodeStart(ctx context.Context, nodeKey, nodeType string, input interface{}) {
	sc.mu.Lock()
	sc.nodeStartTime[nodeKey] = time.Now()
	sc.path = append(sc.path, nodeKey)
	logger := sc.logger
	sc.mu.Unlock()

	status := ExecutionStatus{
		Type:        "node_start",
		Timestamp:   time.Now().UnixMilli(),
		ExecutionID: sc.executionID,
		NodeKey:     nodeKey,
		NodeType:    nodeType,
		Status:      "running",
		AppName:     sc.appName,
		RequestID:   sc.requestID,
		DataFlow:    sc.marshalDataFlow(input, nil),
	}

	if logger != nil {
		inputCount := 0
		if status.DataFlow != nil {
			inputCount = status.DataFlow.InputCount
		}
		logger.OnNodeStartLog(nodeKey, nodeType, inputCount)
	}

	sc.sendStatusAsync(ctx, status)
}

// OnNodeEnd 节点结束执行
func (sc *StatusCollector) OnNodeEnd(ctx context.Context, nodeKey, nodeType string, output interface{}, err error) {
	sc.mu.Lock()
	startTime, ok := sc.nodeStartTime[nodeKey]
	logger := sc.logger
	sc.mu.Unlock()

	var nodeUsage *Usage
	if nodeType == "chat_model" {
		nodeUsage = extractUsageFromOutput(output)
		if nodeUsage != nil {
			sc.CollectTokenUsage(nodeKey, nodeUsage)
		}
	}

	status := ExecutionStatus{
		Type:        "node_end",
		Timestamp:   time.Now().UnixMilli(),
		ExecutionID: sc.executionID,
		NodeKey:     nodeKey,
		NodeType:    nodeType,
		Status:      "success",
		AppName:     sc.appName,
		RequestID:   sc.requestID,
		DataFlow:    sc.marshalDataFlow(nil, output),
		Usage:       nodeUsage,
	}

	if ok {
		_ = time.Since(startTime)
	}

	if err != nil {
		status.Status = "error"
		status.Error = err.Error()
	}

	if logger != nil {
		outputCount := 0
		if status.DataFlow != nil {
			outputCount = status.DataFlow.OutputCount
		}
		logger.OnNodeEndLog(nodeKey, nodeType, outputCount, err)
	}

	sc.sendStatusAsync(ctx, status)
}

// OnComplete 执行完成
func (sc *StatusCollector) OnComplete(ctx context.Context, result *schema.Message, err error) {
	sc.mu.Lock()
	path := make([]string, len(sc.path))
	copy(path, sc.path)
	totalUsage := sc.getTotalUsageLocked()
	logger := sc.logger
	sc.mu.Unlock()

	status := ExecutionStatus{
		Type:        "complete",
		Timestamp:   time.Now().UnixMilli(),
		ExecutionID: sc.executionID,
		Status:      "success",
		AppName:     sc.appName,
		RequestID:   sc.requestID,
		ControlFlow: &ControlFlowStatus{
			Path: path,
		},
		Usage: totalUsage,
	}

	if err != nil {
		status.Status = "error"
		status.Error = err.Error()
	}

	if logger != nil {
		logger.OnCompleteLog(sc.executionID, time.Since(sc.startTime), totalUsage)
	}

	sc.sendStatusAsync(ctx, status)
}

// sendStatusAsync 异步发送状态
func (sc *StatusCollector) sendStatusAsync(ctx context.Context, status ExecutionStatus) {
	sc.mu.Lock()
	sc.statusHistory = append(sc.statusHistory, status)
	callback := sc.statusCallback
	store := sc.statusStore
	clients := sc.webhookClients
	sc.mu.Unlock()

	// 1. 优先执行回调（SSE 等），虽然它是同步调用，但在应用端可能已实现异步
	if callback != nil {
		callback(status)
	}

	// 2. 异步存储状态，避免数据库等慢操作阻塞主流程
	if store != nil {
		go func() {
			store.AddStatus(sc.executionID, status)
		}()
	}

	// 3. 异步发送 Webhook
	if len(clients) > 0 {
		go func() {
			for _, client := range clients {
				_ = client.SendStatus(ctx, status)
			}
		}()
	}
}

// CollectTokenUsage 收集TokenUsage
func (sc *StatusCollector) CollectTokenUsage(nodeKey string, usage *Usage) {
	if usage == nil {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.tokenUsageMap[nodeKey] = usage
	sc.totalUsage.PromptTokens += usage.PromptTokens
	sc.totalUsage.CompletionTokens += usage.CompletionTokens
	sc.totalUsage.TotalTokens += usage.TotalTokens
}

// GetTotalUsage 获取汇总的TokenUsage
func (sc *StatusCollector) GetTotalUsage() *Usage {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.totalUsage.PromptTokens == 0 && sc.totalUsage.CompletionTokens == 0 && sc.totalUsage.TotalTokens == 0 {
		return nil
	}

	return &Usage{
		PromptTokens:     sc.totalUsage.PromptTokens,
		CompletionTokens: sc.totalUsage.CompletionTokens,
		TotalTokens:      sc.totalUsage.TotalTokens,
	}
}

// GetExecutionID 获取执行ID
func (sc *StatusCollector) GetExecutionID() string {
	return sc.executionID
}

// GetStatusHistory 获取状态历史
func (sc *StatusCollector) GetStatusHistory() []ExecutionStatus {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	history := make([]ExecutionStatus, len(sc.statusHistory))
	copy(history, sc.statusHistory)
	return history
}

// getTotalUsageLocked 获取汇总 Usage，调用方需已持 sc.mu
func (sc *StatusCollector) getTotalUsageLocked() *Usage {
	if sc.totalUsage.PromptTokens == 0 && sc.totalUsage.CompletionTokens == 0 && sc.totalUsage.TotalTokens == 0 {
		return nil
	}
	return &Usage{
		PromptTokens:     sc.totalUsage.PromptTokens,
		CompletionTokens: sc.totalUsage.CompletionTokens,
		TotalTokens:      sc.totalUsage.TotalTokens,
	}
}

func (sc *StatusCollector) marshalDataFlow(input, output interface{}) *DataFlowStatus {
	sc.mu.Lock()
	enable := sc.enableDataFlow
	sc.mu.Unlock()
	if !enable {
		return nil
	}
	df := &DataFlowStatus{}
	if input != nil {
		df.InputCount, df.InputData = marshalAny(input)
	}
	if output != nil {
		df.OutputCount, df.OutputData = marshalAny(output)
	}
	return df
}

func marshalAny(v interface{}) (int, []map[string]interface{}) {
	switch val := v.(type) {
	case []*schema.Message:
		data := make([]map[string]interface{}, len(val))
		for i, msg := range val {
			item := map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			}
			if len(msg.ToolCalls) > 0 {
				toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = map[string]interface{}{
						"id":        tc.ID,
						"function":  tc.Function.Name,
						"arguments": tc.Function.Arguments,
					}
				}
				item["tool_calls"] = toolCalls
			}
			data[i] = item
		}
		return len(val), data
	case *schema.Message:
		if val == nil {
			return 0, nil
		}
		n, data := marshalAny([]*schema.Message{val})
		return n, data
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return 1, nil
		}
		var generic map[string]interface{}
		if json.Unmarshal(b, &generic) == nil {
			return 1, []map[string]interface{}{generic}
		}
		return 1, nil
	}
}

func extractUsageFromOutput(output interface{}) *Usage {
	if msg, ok := output.(*schema.Message); ok && msg != nil {
		return extractUsageFromMessage(msg)
	}
	if msgs, ok := output.([]*schema.Message); ok && len(msgs) > 0 {
		return extractUsageFromMessage(msgs[0])
	}
	return nil
}

func extractUsageFromMessage(msg *schema.Message) *Usage {
	if msg == nil || msg.ResponseMeta == nil || msg.ResponseMeta.Usage == nil {
		return nil
	}
	u := msg.ResponseMeta.Usage
	usage := &Usage{
		PromptTokens:     int(u.PromptTokens),
		CompletionTokens: int(u.CompletionTokens),
		TotalTokens:      int(u.TotalTokens),
	}
	if usage.TotalTokens == 0 && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return usage
}
