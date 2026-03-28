package digeino

import (
	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/webhook"

	"github.com/cloudwego/eino/schema"
)

type UserMessage struct {
	ID      string            `json:"id"`
	Query   string            `json:"query"`
	History []*schema.Message `json:"history"`
}

// ========== Webhook 和 StatusStore 相关类型定义（从 webhook 或 config 包重新导出） ==========

// Usage 使用量统计
type Usage = webhook.Usage

// EventType 统一生命周期事件类型
type EventType = webhook.EventType

const (
	EventTypeStarted   = webhook.EventTypeStarted
	EventTypeSucceeded = webhook.EventTypeSucceeded
	EventTypeFailed    = webhook.EventTypeFailed
	EventTypeRetried   = webhook.EventTypeRetried
	EventTypeSkipped   = webhook.EventTypeSkipped
	EventTypeCompleted = webhook.EventTypeCompleted
)

// ExecutionStatus 执行状态
type ExecutionStatus = webhook.ExecutionStatus

// ExecutionEvent 统一执行事件（推荐消费）
type ExecutionEvent = webhook.ExecutionEvent

// DispatchStats 分发统计信息
type DispatchStats = webhook.DispatchStats

// DataFlowStatus 数据流状态
type DataFlowStatus = webhook.DataFlowStatus

// ControlFlowStatus 控制流状态
type ControlFlowStatus = webhook.ControlFlowStatus

// WebhookConfig Webhook 配置
type WebhookConfig = config.WebhookConfig

// WebhookPayload Webhook 请求体
type WebhookPayload = webhook.WebhookPayload

// Message 消息（用于 Webhook 和 StatusStore）
type Message = webhook.Message
