package status

import (
	"github.com/originaleric/digeino/webhook"
)

// StatusStoreAdapter 适配器，将 status.StatusStore 适配为 webhook.StatusStoreInterface
type StatusStoreAdapter struct {
	store StatusStore
}

// NewStatusStoreAdapter 创建适配器
func NewStatusStoreAdapter(store StatusStore) *StatusStoreAdapter {
	return &StatusStoreAdapter{store: store}
}

// CreateExecution 创建执行记录
func (a *StatusStoreAdapter) CreateExecution(executionID, appName, requestID string) interface{} {
	return a.store.CreateExecution(executionID, appName, requestID)
}

// AddStatus 添加状态更新
func (a *StatusStoreAdapter) AddStatus(executionID string, status webhook.ExecutionStatus) bool {
	return a.store.AddStatus(executionID, status)
}

// SetResult 设置执行结果
func (a *StatusStoreAdapter) SetResult(executionID string, result webhook.Message) bool {
	return a.store.SetResult(executionID, result)
}
