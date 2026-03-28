package webhook

import "fmt"

const (
	// ExecutionEventSchemaV1 统一事件结构版本（首版）
	ExecutionEventSchemaV1 = "1.0"
)

// EventType 统一生命周期事件类型
type EventType string

const (
	EventTypeStarted   EventType = "started"
	EventTypeSucceeded EventType = "succeeded"
	EventTypeFailed    EventType = "failed"
	EventTypeRetried   EventType = "retried"
	EventTypeSkipped   EventType = "skipped"
	EventTypeCompleted EventType = "completed"
)

// Usage 使用量统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ExecutionEvent 统一执行事件（推荐消费）
// 注意：字段设计保持与 ExecutionStatus 对齐，便于平滑迁移。
type ExecutionEvent struct {
	SchemaVersion string    `json:"schema_version"`      // 事件 schema 版本
	EventType     EventType `json:"event_type"`          // started | succeeded | failed | retried | skipped | completed
	Timestamp     int64     `json:"timestamp"`           // Unix 时间戳（毫秒）
	ExecutionID   string    `json:"execution_id"`        // 执行 ID
	NodeID        string    `json:"node_id,omitempty"`   // 节点 ID（推荐）
	NodeType      string    `json:"node_type,omitempty"` // 节点类型
	Attempt       int       `json:"attempt,omitempty"`   // 节点尝试次数，默认 1
	Source        string    `json:"source,omitempty"`    // 事件来源，例如 digeino.status_collector
	Status        string    `json:"status"`              // running | success | error
	Error         string    `json:"error,omitempty"`     // 错误信息
	DataFlow      *DataFlowStatus    `json:"data_flow,omitempty"`
	ControlFlow   *ControlFlowStatus `json:"control_flow,omitempty"`
	AppName       string             `json:"app_name"`
	RequestID     string             `json:"request_id"`
	Usage         *Usage             `json:"usage,omitempty"`

	// LegacyType 保留旧字段语义，便于历史系统兼容。
	LegacyType string `json:"type,omitempty"`
}

// ExecutionStatus 执行状态
type ExecutionStatus struct {
	// 基础信息
	Type          string `json:"type"`                     // 兼容字段："node_start" | "node_end" | "complete"
	SchemaVersion string `json:"schema_version,omitempty"` // 事件 schema 版本（推荐）
	EventType     string `json:"event_type,omitempty"`     // 推荐字段：started | succeeded | failed | retried | skipped | completed
	Timestamp     int64  `json:"timestamp"`                // Unix 时间戳（毫秒）
	ExecutionID   string `json:"execution_id"`             // 执行 ID（用于关联）

	// 节点信息
	NodeKey  string `json:"node_key,omitempty"`  // 兼容字段：节点标识
	NodeID   string `json:"node_id,omitempty"`   // 推荐字段：节点标识
	NodeType string `json:"node_type,omitempty"` // 节点类型
	Attempt  int    `json:"attempt,omitempty"`   // 尝试次数，默认 1
	Source   string `json:"source,omitempty"`    // 事件来源，例如 digeino.status_collector
	Status   string `json:"status"`              // running | success | error
	Error    string `json:"error,omitempty"`     // 错误信息

	// 数据流信息
	DataFlow *DataFlowStatus `json:"data_flow,omitempty"`

	// 控制流信息
	ControlFlow *ControlFlowStatus `json:"control_flow,omitempty"`

	// 执行上下文
	AppName   string `json:"app_name"`   // Agent 名称
	RequestID string `json:"request_id"` // 请求 ID（可选）

	// Usage 信息（可选）
	Usage *Usage `json:"usage,omitempty"` // Token 使用统计
}

// NormalizeEventType 统一事件类型，优先使用 event_type，再回退到历史字段。
func (s ExecutionStatus) NormalizeEventType() EventType {
	if s.EventType != "" {
		return EventType(s.EventType)
	}
	switch s.Type {
	case "node_start":
		return EventTypeStarted
	case "node_end":
		if s.Status == "error" {
			return EventTypeFailed
		}
		return EventTypeSucceeded
	case "complete":
		if s.Status == "error" {
			return EventTypeFailed
		}
		return EventTypeCompleted
	default:
		if s.Status == "error" {
			return EventTypeFailed
		}
		return EventTypeCompleted
	}
}

// AsExecutionEvent 将兼容结构转换为统一事件结构。
func (s ExecutionStatus) AsExecutionEvent() ExecutionEvent {
	schemaVersion := s.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = ExecutionEventSchemaV1
	}
	nodeID := s.NodeID
	if nodeID == "" {
		nodeID = s.NodeKey
	}
	return ExecutionEvent{
		SchemaVersion: schemaVersion,
		EventType:     s.NormalizeEventType(),
		Timestamp:     s.Timestamp,
		ExecutionID:   s.ExecutionID,
		NodeID:        nodeID,
		NodeType:      s.NodeType,
		Attempt:       s.Attempt,
		Source:        s.Source,
		Status:        s.Status,
		Error:         s.Error,
		DataFlow:      s.DataFlow,
		ControlFlow:   s.ControlFlow,
		AppName:       s.AppName,
		RequestID:     s.RequestID,
		Usage:         s.Usage,
		LegacyType:    s.Type,
	}
}

// DedupeKey 返回推荐去重键（execution_id + node_id + attempt + event_type）。
func (s ExecutionStatus) DedupeKey() string {
	nodeID := s.NodeID
	if nodeID == "" {
		nodeID = s.NodeKey
	}
	attempt := s.Attempt
	if attempt <= 0 {
		attempt = 1
	}
	return fmt.Sprintf("%s|%s|%d|%s", s.ExecutionID, nodeID, attempt, s.NormalizeEventType())
}

// DataFlowStatus 数据流状态
type DataFlowStatus struct {
	InputCount  int                      `json:"input_count"`
	OutputCount int                      `json:"output_count"`
	InputData   []map[string]interface{} `json:"input_data,omitempty"`  // 可选，可配置是否包含
	OutputData  []map[string]interface{} `json:"output_data,omitempty"` // 可选
}

// ControlFlowStatus 控制流状态
type ControlFlowStatus struct {
	BranchFrom string   `json:"branch_from,omitempty"`
	BranchTo   string   `json:"branch_to,omitempty"`
	Condition  string   `json:"condition,omitempty"`
	Path       []string `json:"path"` // 执行路径
}

// WebhookPayload Webhook 请求体
type WebhookPayload struct {
	Event     string          `json:"event"`               // 事件类型
	Status    ExecutionStatus `json:"status"`              // 状态信息
	Signature string          `json:"signature,omitempty"` // 签名（如果配置了 secret）
}

// Message 消息（用于 Webhook 和 StatusStore）
type Message struct {
	Role    string `json:"role"` // user | assistant | system
	Content string `json:"content"`
}

// StatusStoreInterface 状态存储接口（为了解耦）
type StatusStoreInterface interface {
	AddStatus(executionID string, status ExecutionStatus) bool
	SetResult(executionID string, result Message) bool
}
