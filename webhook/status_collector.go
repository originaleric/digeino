package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/config"
)

const (
	defaultEventSource = "digeino.status_collector"
	sinkQueueSize       = 256
	sinkEnqueueTimeout  = 50 * time.Millisecond
	sinkRetryCount      = 3
	sinkRetryDelay      = 100 * time.Millisecond
	defaultSampleRate   = 100
	defaultMaxPayload   = 262144 // 256KiB
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
	sampleRate     int
	maxPayloadSize int

	dispatchOnce sync.Once
	sinkQueue    chan sinkTask
	statsMu      sync.Mutex
	stats        DispatchStats
}

type sinkTask struct {
	ctx    context.Context
	status ExecutionStatus
	store  StatusStoreInterface
	client *WebhookClient
}

// DispatchStats 分发统计信息，用于运行期观测。
type DispatchStats struct {
	Enqueued         int64 `json:"enqueued"`
	QueueFallback    int64 `json:"queue_fallback"`
	StoreSuccess     int64 `json:"store_success"`
	StoreFailure     int64 `json:"store_failure"`
	WebhookSuccess   int64 `json:"webhook_success"`
	WebhookFailure   int64 `json:"webhook_failure"`
	RetryCount       int64 `json:"retry_count"`
	SampledOut       int64 `json:"sampled_out"`
	PayloadCompacted int64 `json:"payload_compacted"`
}

// NewStatusCollector 创建状态收集器，DataFlow 开关从 config.Get().Status.DataFlow 读取，默认关闭
func NewStatusCollector(executionID, appName, requestID string) *StatusCollector {
	enableDataFlow := false
	sampleRate := defaultSampleRate
	maxPayloadSize := defaultMaxPayload

	if cfg := config.Get(); cfg != nil {
		if cfg.Status.DataFlow.Enabled != nil && *cfg.Status.DataFlow.Enabled {
			enableDataFlow = true
		}
		if cfg.Status.Event.SampleRate >= 0 && cfg.Status.Event.SampleRate <= 100 {
			sampleRate = cfg.Status.Event.SampleRate
		}
		if cfg.Status.Event.MaxPayloadBytes > 0 {
			maxPayloadSize = cfg.Status.Event.MaxPayloadBytes
		}
	}
	return &StatusCollector{
		executionID:    executionID,
		appName:        appName,
		requestID:      requestID,
		startTime:      time.Now(),
		nodeStartTime:  make(map[string]time.Time),
		tokenUsageMap:  make(map[string]*Usage),
		enableDataFlow: enableDataFlow,
		sampleRate:     sampleRate,
		maxPayloadSize: maxPayloadSize,
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

// SetEventPolicy 设置事件分发策略（覆盖配置文件）。
func (sc *StatusCollector) SetEventPolicy(sampleRate, maxPayloadBytes int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sampleRate >= 0 && sampleRate <= 100 {
		sc.sampleRate = sampleRate
	}
	if maxPayloadBytes > 0 {
		sc.maxPayloadSize = maxPayloadBytes
	}
}

// AddWebhookClient 添加 Webhook 客户端
func (sc *StatusCollector) AddWebhookClient(client *WebhookClient) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.webhookClients = append(sc.webhookClients, client)
	sc.ensureDispatcherLocked()
}

// SetStatusStore 设置状态存储
func (sc *StatusCollector) SetStatusStore(store StatusStoreInterface) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.statusStore = store
	sc.ensureDispatcherLocked()
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
		Type:          "node_start",
		SchemaVersion: ExecutionEventSchemaV1,
		EventType:     string(EventTypeStarted),
		Timestamp:     time.Now().UnixMilli(),
		ExecutionID:   sc.executionID,
		NodeKey:       nodeKey,
		NodeID:        nodeKey,
		NodeType:      nodeType,
		Attempt:       1,
		Source:        defaultEventSource,
		Status:        "running",
		AppName:       sc.appName,
		RequestID:     sc.requestID,
		DataFlow:      sc.marshalDataFlow(input, nil),
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
		Type:          "node_end",
		SchemaVersion: ExecutionEventSchemaV1,
		EventType:     string(EventTypeSucceeded),
		Timestamp:     time.Now().UnixMilli(),
		ExecutionID:   sc.executionID,
		NodeKey:       nodeKey,
		NodeID:        nodeKey,
		NodeType:      nodeType,
		Attempt:       1,
		Source:        defaultEventSource,
		Status:        "success",
		AppName:       sc.appName,
		RequestID:     sc.requestID,
		DataFlow:      sc.marshalDataFlow(nil, output),
		Usage:         nodeUsage,
	}

	if ok {
		_ = time.Since(startTime)
	}

	if err != nil {
		status.Status = "error"
		status.EventType = string(EventTypeFailed)
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
		Type:          "complete",
		SchemaVersion: ExecutionEventSchemaV1,
		EventType:     string(EventTypeCompleted),
		Timestamp:     time.Now().UnixMilli(),
		ExecutionID:   sc.executionID,
		Attempt:       1,
		Source:        defaultEventSource,
		Status:        "success",
		AppName:       sc.appName,
		RequestID:     sc.requestID,
		ControlFlow: &ControlFlowStatus{
			Path: path,
		},
		Usage: totalUsage,
	}

	if err != nil {
		status.Status = "error"
		status.EventType = string(EventTypeFailed)
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

	asyncStatus := status
	if !sc.shouldSendToAsyncSinks(asyncStatus) {
		sc.updateStats(func(stats *DispatchStats) {
			stats.SampledOut++
		})
		return
	}
	var compacted bool
	asyncStatus, compacted = sc.compactStatusByPayloadLimit(asyncStatus)
	if compacted {
		sc.updateStats(func(stats *DispatchStats) {
			stats.PayloadCompacted++
		})
	}

	// 2. 异步存储状态，避免数据库等慢操作阻塞主流程
	if store != nil {
		sc.enqueueSinkTask(sinkTask{
			ctx:    detachContext(ctx),
			status: asyncStatus,
			store:  store,
		})
	}

	// 3. 异步发送 Webhook
	if len(clients) > 0 {
		safeCtx := detachContext(ctx)
		for _, client := range clients {
			sc.enqueueSinkTask(sinkTask{
				ctx:    safeCtx,
				status: asyncStatus,
				client: client,
			})
		}
	}
}

func detachContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func (sc *StatusCollector) ensureDispatcherLocked() {
	sc.dispatchOnce.Do(func() {
		sc.sinkQueue = make(chan sinkTask, sinkQueueSize)
		go sc.runSinkDispatcher()
	})
}

func (sc *StatusCollector) enqueueSinkTask(task sinkTask) {
	sc.mu.Lock()
	sc.ensureDispatcherLocked()
	queue := sc.sinkQueue
	sc.mu.Unlock()

	timer := time.NewTimer(sinkEnqueueTimeout)
	defer timer.Stop()

	select {
	case queue <- task:
		sc.updateStats(func(stats *DispatchStats) {
			stats.Enqueued++
		})
		return
	case <-timer.C:
		// 队列拥堵时降级为独立 goroutine，优先保证事件不丢。
		sc.updateStats(func(stats *DispatchStats) {
			stats.QueueFallback++
		})
		go sc.executeSinkTask(task)
	}
}

func (sc *StatusCollector) runSinkDispatcher() {
	for task := range sc.sinkQueue {
		sc.executeSinkTask(task)
	}
}

func (sc *StatusCollector) executeSinkTask(task sinkTask) {
	for i := 0; i <= sinkRetryCount; i++ {
		var err error
		sinkName := "unknown"
		if task.store != nil {
			sinkName = "store"
			if ok := task.store.AddStatus(sc.executionID, task.status); !ok {
				err = fmt.Errorf("status store rejected execution_id=%s", sc.executionID)
			}
		}
		if task.client != nil {
			sinkName = "webhook"
			err = task.client.SendStatus(task.ctx, task.status)
		}
		if err == nil {
			sc.updateSinkResultStats(sinkName, true)
			return
		}
		if i < sinkRetryCount {
			sc.updateStats(func(stats *DispatchStats) {
				stats.RetryCount++
			})
			time.Sleep(sinkRetryDelay)
			continue
		}
		sc.updateSinkResultStats(sinkName, false)
	}
}

func (sc *StatusCollector) shouldSendToAsyncSinks(status ExecutionStatus) bool {
	sc.mu.Lock()
	sampleRate := sc.sampleRate
	sc.mu.Unlock()

	if status.NormalizeEventType() == EventTypeCompleted || status.NormalizeEventType() == EventTypeFailed {
		return true
	}
	if sampleRate >= 100 {
		return true
	}
	if sampleRate <= 0 {
		return false
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(strings.ToLower(status.DedupeKey())))
	return int(hasher.Sum32()%100) < sampleRate
}

func (sc *StatusCollector) compactStatusByPayloadLimit(status ExecutionStatus) (ExecutionStatus, bool) {
	sc.mu.Lock()
	limit := sc.maxPayloadSize
	sc.mu.Unlock()
	if limit <= 0 {
		return status, false
	}

	if statusPayloadSize(status) <= limit {
		return status, false
	}

	compacted := status
	changed := false
	if compacted.DataFlow != nil && (len(compacted.DataFlow.InputData) > 0 || len(compacted.DataFlow.OutputData) > 0) {
		df := *compacted.DataFlow
		df.InputData = nil
		df.OutputData = nil
		compacted.DataFlow = &df
		changed = true
	}
	if statusPayloadSize(compacted) <= limit {
		return compacted, changed
	}

	if compacted.DataFlow != nil {
		compacted.DataFlow = nil
		changed = true
	}
	if statusPayloadSize(compacted) <= limit {
		return compacted, changed
	}

	if compacted.ControlFlow != nil {
		cf := *compacted.ControlFlow
		cf.Path = nil
		compacted.ControlFlow = &cf
		changed = true
	}
	if statusPayloadSize(compacted) <= limit {
		return compacted, changed
	}

	if len(compacted.Error) > 512 {
		compacted.Error = compacted.Error[:512] + "...(truncated)"
		changed = true
	}
	return compacted, changed
}

func statusPayloadSize(status ExecutionStatus) int {
	b, err := json.Marshal(status)
	if err != nil {
		return 0
	}
	return len(b)
}

func (sc *StatusCollector) updateSinkResultStats(sink string, success bool) {
	sc.updateStats(func(stats *DispatchStats) {
		switch sink {
		case "store":
			if success {
				stats.StoreSuccess++
			} else {
				stats.StoreFailure++
			}
		case "webhook":
			if success {
				stats.WebhookSuccess++
			} else {
				stats.WebhookFailure++
			}
		}
	})
}

func (sc *StatusCollector) updateStats(update func(stats *DispatchStats)) {
	sc.statsMu.Lock()
	defer sc.statsMu.Unlock()
	update(&sc.stats)
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

// GetDispatchStats 获取 sink 分发统计信息。
func (sc *StatusCollector) GetDispatchStats() DispatchStats {
	sc.statsMu.Lock()
	defer sc.statsMu.Unlock()
	return sc.stats
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
