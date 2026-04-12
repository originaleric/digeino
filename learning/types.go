package learning

import "time"

// TerminalOutcome 执行终态（用于幂等键）。
type TerminalOutcome string

const (
	TerminalSucceeded TerminalOutcome = "succeeded"
	TerminalFailed    TerminalOutcome = "failed"
)

// LearningEvent 在 run 进入终态后投递（与 execution 观测对齐）。
type LearningEvent struct {
	EventID         string
	EventType       string // 例如 run.completed
	OccurredAt      time.Time
	AppName         string
	ExecutionID     string
	RunID           string // 默认与 ExecutionID 一致；宿主可映射为治理 Run ID
	SessionID       string
	UserID          string
	TriggerType     string // user | cron | webhook | system
	TerminalOutcome TerminalOutcome
}

// ToolCallDigest 工具调用摘要。
type ToolCallDigest struct {
	Name       string
	Success    bool
	DurationMs int64
	Error      string
}

// RunContext 宿主填充的复盘上下文。
type RunContext struct {
	AppName            string
	RunID              string
	ExecutionID      string
	SessionID          string
	UserID             string
	UserInput          string
	AssistantOutput    string
	SessionSummary     string
	ToolCalls          []ToolCallDigest
	RetryCount         int
	UserCorrectionHint bool
	TerminalOutcome    TerminalOutcome
	CreatedAt          *time.Time
}

// MemoryAction 记忆沉淀动作。
type MemoryAction struct {
	Action     string // add | skip
	Category   string // fact | preference | timeline
	Content    string
	Importance int
}

// SkillAction 技能沉淀动作。
type SkillAction struct {
	Action      string // create | patch | skip
	SkillName   string
	PatchNote   string
	FullContent string
}

// LearningDecision 决策引擎输出。
type LearningDecision struct {
	DecisionID    string
	RunID         string
	SessionID     string
	MemoryActions []MemoryAction
	SkillActions  []SkillAction
	Reason        string
	Confidence    float64
	CreatedAt     time.Time
}

// Phase 控制 Worker 执行到哪个阶段（与分阶段 rollout 对齐）。
type Phase int

const (
	PhaseAuditOnly Phase = iota // 仅审计
	PhaseMemory                 // + 执行 memory_actions
	PhaseSkill                  // + 执行 skill_actions
)
