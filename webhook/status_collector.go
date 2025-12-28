package webhook

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

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
}

// NewStatusCollector 创建状态收集器
func NewStatusCollector(executionID, appName, requestID string) *StatusCollector {
	return &StatusCollector{
		executionID:   executionID,
		appName:       appName,
		requestID:     requestID,
		startTime:     time.Now(),
		nodeStartTime: make(map[string]time.Time),
		tokenUsageMap: make(map[string]*Usage),
	}
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
	}

	sc.sendStatusAsync(ctx, status)
}

// OnNodeEnd 节点结束执行
func (sc *StatusCollector) OnNodeEnd(ctx context.Context, nodeKey, nodeType string, output interface{}, err error) {
	sc.mu.Lock()
	startTime, ok := sc.nodeStartTime[nodeKey]
	sc.mu.Unlock()

	status := ExecutionStatus{
		Type:        "node_end",
		Timestamp:   time.Now().UnixMilli(),
		ExecutionID: sc.executionID,
		NodeKey:     nodeKey,
		NodeType:    nodeType,
		Status:      "success",
		AppName:     sc.appName,
		RequestID:   sc.requestID,
	}

	if ok {
		// 可以计算耗时等
		_ = time.Since(startTime)
	}

	if err != nil {
		status.Status = "error"
		status.Error = err.Error()
	}

	sc.sendStatusAsync(ctx, status)
}

// OnComplete 执行完成
func (sc *StatusCollector) OnComplete(ctx context.Context, result *schema.Message, err error) {
	sc.mu.Lock()
	path := make([]string, len(sc.path))
	copy(path, sc.path)
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
	}

	if err != nil {
		status.Status = "error"
		status.Error = err.Error()
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

	if callback != nil {
		callback(status)
	}

	if store != nil {
		store.AddStatus(sc.executionID, status)
	}

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
