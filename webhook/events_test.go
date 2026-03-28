package webhook

import "testing"

func TestNormalizeEventType(t *testing.T) {
	tests := []struct {
		name   string
		status ExecutionStatus
		want   EventType
	}{
		{
			name:   "prefer explicit event_type",
			status: ExecutionStatus{Type: "node_start", EventType: "retried", Status: "running"},
			want:   EventTypeRetried,
		},
		{
			name:   "map node_start to started",
			status: ExecutionStatus{Type: "node_start", Status: "running"},
			want:   EventTypeStarted,
		},
		{
			name:   "map node_end success to succeeded",
			status: ExecutionStatus{Type: "node_end", Status: "success"},
			want:   EventTypeSucceeded,
		},
		{
			name:   "map node_end error to failed",
			status: ExecutionStatus{Type: "node_end", Status: "error"},
			want:   EventTypeFailed,
		},
		{
			name:   "map complete success to completed",
			status: ExecutionStatus{Type: "complete", Status: "success"},
			want:   EventTypeCompleted,
		},
		{
			name:   "map complete error to failed",
			status: ExecutionStatus{Type: "complete", Status: "error"},
			want:   EventTypeFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.NormalizeEventType()
			if got != tt.want {
				t.Fatalf("NormalizeEventType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsEventSubscribed(t *testing.T) {
	status := ExecutionStatus{
		Type:      "node_end",
		EventType: string(EventTypeFailed),
		Status:    "error",
	}

	if !isEventSubscribed([]string{"failed"}, status) {
		t.Fatalf("expected failed to match event_type")
	}
	if !isEventSubscribed([]string{"error"}, status) {
		t.Fatalf("expected error to match failed status")
	}
	if !isEventSubscribed([]string{"node_end"}, status) {
		t.Fatalf("expected node_end to match legacy type")
	}
	if isEventSubscribed([]string{"started"}, status) {
		t.Fatalf("did not expect started to match failed status")
	}
}

func TestDedupeKey(t *testing.T) {
	status := ExecutionStatus{
		ExecutionID: "exec-1",
		NodeKey:     "node-a",
		Status:      "success",
		Type:        "node_end",
	}
	got := status.DedupeKey()
	want := "exec-1|node-a|1|succeeded"
	if got != want {
		t.Fatalf("DedupeKey() = %q, want %q", got, want)
	}
}
