package learning

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LLMClient 可选：由宿主注入以实现结构化决策；未配置时走规则引擎。
type LLMClient interface {
	EvaluateLearning(ctx context.Context, rc *RunContext, cfg Config) (*LearningDecision, error)
}

// DecisionEngine 产出 LearningDecision。
type DecisionEngine interface {
	Evaluate(ctx context.Context, rc *RunContext, cfg Config) (*LearningDecision, error)
}

// DefaultDecisionEngine 规则 + 可选 LLM。
type DefaultDecisionEngine struct {
	LLM LLMClient
}

// Evaluate 应用门槛规则；若配置了 LLM 则优先使用 LLM 结果并做校验。
func (e *DefaultDecisionEngine) Evaluate(ctx context.Context, rc *RunContext, cfg Config) (*LearningDecision, error) {
	if rc == nil {
		return nil, fmt.Errorf("learning: nil RunContext")
	}
	now := time.Now()
	if rc.CreatedAt != nil {
		now = *rc.CreatedAt
	}

	dec := &LearningDecision{
		DecisionID:      uuid.New().String(),
		AppName:         rc.AppName,
		RunID:           rc.RunID,
		SessionID:       rc.SessionID,
		ExecutionID:     rc.ExecutionID,
		TerminalOutcome: rc.TerminalOutcome,
		Reason:          "rule_prefilter",
		Confidence:      0.5,
		CreatedAt:       now,
		MemoryActions:   nil,
		SkillActions:    nil,
	}

	if e != nil && e.LLM != nil {
		llmDec, err := e.LLM.EvaluateLearning(ctx, rc, cfg)
		if err != nil {
			return nil, err
		}
		if llmDec != nil {
			dec = sanitizeDecision(llmDec, rc, now)
		}
	}

	applyRuleFilters(rc, cfg, dec)
	return dec, nil
}

func sanitizeDecision(in *LearningDecision, rc *RunContext, now time.Time) *LearningDecision {
	out := *in
	if out.DecisionID == "" {
		out.DecisionID = uuid.New().String()
	}
	if out.RunID == "" {
		out.RunID = rc.RunID
	}
	if out.SessionID == "" {
		out.SessionID = rc.SessionID
	}
	if out.AppName == "" {
		out.AppName = rc.AppName
	}
	if out.ExecutionID == "" {
		out.ExecutionID = rc.ExecutionID
	}
	if out.TerminalOutcome == "" {
		out.TerminalOutcome = rc.TerminalOutcome
	}
	if out.CreatedAt.IsZero() {
		out.CreatedAt = now
	}
	return &out
}

func applyRuleFilters(rc *RunContext, cfg Config, dec *LearningDecision) {
	toolCount := len(rc.ToolCalls)
	if toolCount < cfg.MinToolCallsForSkill {
		filtered := dec.SkillActions[:0]
		for _, a := range dec.SkillActions {
			if a.Action == "skip" {
				filtered = append(filtered, a)
				continue
			}
			na := a
			na.Action = "skip"
			filtered = append(filtered, na)
		}
		dec.SkillActions = filtered
		if dec.Reason == "" {
			dec.Reason = "skill_gated_by_min_tool_calls"
		}
	}

	if dec.Confidence < cfg.MinConfidence {
		dec.MemoryActions = downgradeMemoryActions(dec.MemoryActions)
		dec.SkillActions = downgradeSkillActions(dec.SkillActions)
		if dec.Reason == "" {
			dec.Reason = "low_confidence_skip_actions"
		}
	}
}

func downgradeMemoryActions(actions []MemoryAction) []MemoryAction {
	out := make([]MemoryAction, 0, len(actions))
	for _, a := range actions {
		if a.Action == "skip" {
			out = append(out, a)
			continue
		}
		a.Action = "skip"
		out = append(out, a)
	}
	return out
}

func downgradeSkillActions(actions []SkillAction) []SkillAction {
	out := make([]SkillAction, 0, len(actions))
	for _, a := range actions {
		if a.Action == "skip" {
			out = append(out, a)
			continue
		}
		a.Action = "skip"
		out = append(out, a)
	}
	return out
}
