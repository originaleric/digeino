package webhook

// Usage 使用量统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ExecutionStatus 执行状态
type ExecutionStatus struct {
	// 基础信息
	Type        string `json:"type"`         // "node_start" | "node_end" | "data_flow" | "control_flow" | "error" | "complete"
	Timestamp   int64  `json:"timestamp"`    // Unix 时间戳（毫秒）
	ExecutionID string `json:"execution_id"` // 执行 ID（用于关联）

	// 节点信息
	NodeKey  string `json:"node_key,omitempty"`  // 节点标识
	NodeType string `json:"node_type,omitempty"` // 节点类型
	Status   string `json:"status"`              // "running" | "success" | "error"
	Error    string `json:"error,omitempty"`     // 错误信息

	// 数据流信息
	DataFlow *DataFlowStatus `json:"data_flow,omitempty"`

	// 控制流信息
	ControlFlow *ControlFlowStatus `json:"control_flow,omitempty"`

	// 执行上下文
	AppName   string `json:"app_name"`   // Agent 名称
	RequestID string `json:"request_id"` // 请求 ID（可选）
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
