package webhook

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeEventFilterNotifier struct {
	mu      sync.Mutex
	events  map[string]struct{}
	records []EventType
}

func newFakeEventFilterNotifier(events ...string) *fakeEventFilterNotifier {
	m := make(map[string]struct{}, len(events))
	for _, event := range events {
		v := strings.TrimSpace(strings.ToLower(event))
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return &fakeEventFilterNotifier{events: m}
}

func (f *fakeEventFilterNotifier) SendStatus(_ context.Context, status ExecutionStatus) error {
	event := strings.ToLower(string(status.NormalizeEventType()))
	if len(f.events) > 0 {
		if _, ok := f.events[event]; !ok {
			return nil
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records = append(f.records, status.NormalizeEventType())
	return nil
}

func (f *fakeEventFilterNotifier) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.records)
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(msg)
}

func TestStatusCollectorDispatchesByNotifierEventFilters(t *testing.T) {
	sc := NewStatusCollector("exec-1", "app", "req-1")
	sc.SetEventPolicy(100, 262144)

	completedOnly := newFakeEventFilterNotifier("completed")
	failedOnly := newFakeEventFilterNotifier("failed")
	startedAndFailed := newFakeEventFilterNotifier("started", "failed")

	sc.AddNotifier(completedOnly)
	sc.AddNotifier(failedOnly)
	sc.AddNotifier(startedAndFailed)

	ctx := context.Background()

	sc.OnComplete(ctx, nil, nil)
	waitUntil(t, time.Second, func() bool {
		return completedOnly.count() == 1 && failedOnly.count() == 0 && startedAndFailed.count() == 0
	}, "expected only completed notifier to receive completed event")

	sc.OnComplete(ctx, nil, fmt.Errorf("boom"))
	waitUntil(t, time.Second, func() bool {
		return completedOnly.count() == 1 && failedOnly.count() == 1 && startedAndFailed.count() == 1
	}, "expected failed notifiers to receive failed event")
}
