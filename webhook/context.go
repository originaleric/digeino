package webhook

import (
	"context"
)

// statusCollectorKey 状态收集器在Context中的key
type statusCollectorKey struct{}

// WithStatusCollector 将状态收集器注入到Context中
func WithStatusCollector(ctx context.Context, collector *StatusCollector) context.Context {
	return context.WithValue(ctx, statusCollectorKey{}, collector)
}

// GetStatusCollector 从Context中获取状态收集器
func GetStatusCollector(ctx context.Context) *StatusCollector {
	if collector, ok := ctx.Value(statusCollectorKey{}).(*StatusCollector); ok {
		return collector
	}
	return nil
}
