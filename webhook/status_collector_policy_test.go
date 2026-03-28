package webhook

import (
	"testing"
)

func TestShouldSendToAsyncSinks(t *testing.T) {
	sc := NewStatusCollector("exec", "app", "req")
	sc.SetEventPolicy(0, 262144) // 非终态全部采样掉

	started := ExecutionStatus{
		Type:        "node_start",
		EventType:   string(EventTypeStarted),
		ExecutionID: "exec",
		NodeID:      "n1",
		Attempt:     1,
		Status:      "running",
	}
	if sc.shouldSendToAsyncSinks(started) {
		t.Fatalf("started event should be sampled out when sampleRate=0")
	}

	completed := ExecutionStatus{
		Type:        "complete",
		EventType:   string(EventTypeCompleted),
		ExecutionID: "exec",
		Status:      "success",
	}
	if !sc.shouldSendToAsyncSinks(completed) {
		t.Fatalf("terminal event should bypass sampling")
	}
}

func TestCompactStatusByPayloadLimit(t *testing.T) {
	sc := NewStatusCollector("exec", "app", "req")
	sc.SetEventPolicy(100, 200)

	status := ExecutionStatus{
		Type:        "node_end",
		EventType:   string(EventTypeSucceeded),
		ExecutionID: "exec",
		NodeID:      "n1",
		Attempt:     1,
		Status:      "success",
		DataFlow: &DataFlowStatus{
			InputCount: 1,
			InputData: []map[string]interface{}{
				{"huge": "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"},
			},
			OutputCount: 1,
			OutputData: []map[string]interface{}{
				{"huge": "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"},
			},
		},
	}

	compacted, changed := sc.compactStatusByPayloadLimit(status)
	if !changed {
		t.Fatalf("expected payload compaction to happen")
	}
	if statusPayloadSize(compacted) > 200 {
		t.Fatalf("compacted payload still exceeds limit: %d", statusPayloadSize(compacted))
	}
}
