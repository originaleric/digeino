package learning

import (
	"context"
	"testing"
)

func TestDefaultDecisionEngine_RuleMinToolCalls(t *testing.T) {
	e := &DefaultDecisionEngine{}
	rc := &RunContext{
		RunID:     "r1",
		SessionID: "s1",
		ToolCalls: []ToolCallDigest{{Name: "a"}},
	}
	cfg := Config{
		MinToolCallsForSkill: 3,
		MinConfidence:        0.1,
		ExecutionPhase:       PhaseAuditOnly,
	}
	dec, err := e.Evaluate(context.Background(), rc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if dec == nil {
		t.Fatal("nil decision")
	}
	for _, a := range dec.SkillActions {
		if a.Action != "skip" && a.Action != "" {
			t.Fatalf("expected skill actions gated, got %#v", a)
		}
	}
}

func TestDefaultDecisionEngine_LowConfidenceSkips(t *testing.T) {
	e := &DefaultDecisionEngine{}
	rc := &RunContext{
		RunID:       "r1",
		SessionID:   "s1",
		ToolCalls: []ToolCallDigest{
			{Name: "a"}, {Name: "b"}, {Name: "c"},
		},
	}
	cfg := Config{
		MinToolCallsForSkill: 1,
		MinConfidence:        0.99,
		ExecutionPhase:       PhaseAuditOnly,
	}
	dec, err := e.Evaluate(context.Background(), rc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	dec.Confidence = 0.1
	dec.MemoryActions = []MemoryAction{{Action: "add", Content: "x"}}
	dec.SkillActions = []SkillAction{{Action: "create", SkillName: "n"}}
	applyRuleFilters(rc, cfg, dec)
	for _, a := range dec.MemoryActions {
		if a.Action != "skip" {
			t.Fatalf("memory should downgrade: %#v", a)
		}
	}
	for _, a := range dec.SkillActions {
		if a.Action != "skip" {
			t.Fatalf("skill should downgrade: %#v", a)
		}
	}
}
