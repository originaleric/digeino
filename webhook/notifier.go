package webhook

import "context"

// StatusNotifier 统一状态通知器接口。
type StatusNotifier interface {
	SendStatus(ctx context.Context, status ExecutionStatus) error
}
