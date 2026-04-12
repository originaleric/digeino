package learning

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeAudit struct {
	mu          sync.Mutex
	existsMap   map[string]bool
	saved       int
	applied     int
	lastApplied string
}

func (f *fakeAudit) key(eid string, o TerminalOutcome) string {
	return eid + "|" + string(o)
}

func (f *fakeAudit) Exists(_ context.Context, executionID string, outcome TerminalOutcome) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.existsMap[f.key(executionID, outcome)], nil
}

func (f *fakeAudit) SaveDecision(_ context.Context, decision *LearningDecision) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saved++
	return nil
}

func (f *fakeAudit) MarkApplied(_ context.Context, decisionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applied++
	f.lastApplied = decisionID
	return nil
}

func (f *fakeAudit) MarkFailed(_ context.Context, decisionID string, _ string) error {
	return nil
}

func (f *fakeAudit) Rollback(_ context.Context, _ string) error {
	return nil
}

func TestProcessOne_IdempotentViaAudit(t *testing.T) {
	fa := &fakeAudit{existsMap: make(map[string]bool)}
	h := &Host{
		Audit: fa,
		RunContextProvider: stubProvider{rc: &RunContext{
			RunID: "run1", SessionID: "sess", AppName: "app",
			ToolCalls: []ToolCallDigest{{Name: "t1"}, {Name: "t2"}, {Name: "t3"}},
		}},
	}
	SetHost(h)
	SetConfigOverride(&Config{
		Enabled:              true,
		MinToolCallsForSkill: 1,
		MinConfidence:        0.0,
		ExecutionPhase:       PhaseAuditOnly,
	})
	defer func() {
		SetHost(nil)
		SetConfigOverride(nil)
	}()

	ev := LearningEvent{
		ExecutionID:     "ex-1",
		RunID:           "ex-1",
		AppName:         "app",
		TerminalOutcome: TerminalSucceeded,
	}
	processOne(context.Background(), ev)
	if fa.saved != 1 || fa.applied != 1 {
		t.Fatalf("expected one save and apply, got saved=%d applied=%d", fa.saved, fa.applied)
	}

	fa.existsMap[fa.key("ex-1", TerminalSucceeded)] = true
	processOne(context.Background(), ev)
	if fa.saved != 1 {
		t.Fatalf("second run should skip, saved=%d", fa.saved)
	}
}

type stubProvider struct {
	rc *RunContext
}

func (s stubProvider) GetRunContext(_ context.Context, ev LearningEvent) (*RunContext, error) {
	r := s.rc
	if r == nil {
		r = minimalRunContext(ev)
	}
	return r, nil
}

func TestDeduper_TryClaim(t *testing.T) {
	d := newMemoryDeduper(time.Hour)
	if !d.TryClaim("a", TerminalSucceeded) {
		t.Fatal("first claim")
	}
	if d.TryClaim("a", TerminalSucceeded) {
		t.Fatal("second claim should fail")
	}
	if !d.TryClaim("a", TerminalFailed) {
		t.Fatal("different outcome is new key")
	}
}
