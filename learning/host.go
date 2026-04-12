package learning

import "context"

// 宿主（例如 DigFlow）实现 Host 内各接口，在 bootstrap 中调用 Init 或 SetHost + SetConfigOverride，
// 并保证审计表幂等与 RunContext 拼装；DigFlow 侧工作不在本仓库内完成。

// RunContextProvider 从宿主加载复盘所需上下文。
type RunContextProvider interface {
	GetRunContext(ctx context.Context, event LearningEvent) (*RunContext, error)
}

// MemorySink 长期记忆写入（宿主实现存储）。
type MemorySink interface {
	Exists(ctx context.Context, appName, userID, content string) (bool, error)
	Add(ctx context.Context, appName, userID, category, content, runID string, importance int) error
}

// SkillSink 技能创建/修补。
type SkillSink interface {
	FindSimilar(ctx context.Context, appName, query string) (skillID string, skillName string, found bool, err error)
	Create(ctx context.Context, appName, name, content, createdBy string) (skillID string, err error)
	Patch(ctx context.Context, skillID, patchNote, updatedBy string) error
}

// LearningAuditStore 决策审计与幂等。
type LearningAuditStore interface {
	Exists(ctx context.Context, executionID string, outcome TerminalOutcome) (bool, error)
	SaveDecision(ctx context.Context, decision *LearningDecision) error
	MarkApplied(ctx context.Context, decisionID string) error
	MarkFailed(ctx context.Context, decisionID string, reason string) error
	Rollback(ctx context.Context, decisionID string) error
}

// Host 聚合宿主能力；各字段可为 nil，Worker 按阶段跳过未实现能力。
type Host struct {
	RunContextProvider RunContextProvider
	Memory             MemorySink
	Skill              SkillSink
	Audit              LearningAuditStore
}
