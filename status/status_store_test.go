package status

import (
	"testing"

	"github.com/originaleric/digeino/webhook"
)

func TestMemoryStatusStoreTerminalByEventTypeCompleted(t *testing.T) {
	store := NewMemoryStatusStore(10)
	executionID := "exec-completed"
	store.CreateExecution(executionID, "app", "req")

	ok := store.AddStatus(executionID, webhook.ExecutionStatus{
		EventType: string(webhook.EventTypeCompleted),
		Status:    "success",
	})
	if !ok {
		t.Fatalf("AddStatus should return true")
	}

	record, exists := store.GetExecution(executionID)
	if !exists {
		t.Fatalf("execution record should exist")
	}
	if record.Status != "completed" {
		t.Fatalf("record.Status = %q, want completed", record.Status)
	}
	if record.EndTime == nil {
		t.Fatalf("record.EndTime should be set for terminal event")
	}
}

func TestMemoryStatusStoreTerminalByEventTypeFailed(t *testing.T) {
	store := NewMemoryStatusStore(10)
	executionID := "exec-failed"
	store.CreateExecution(executionID, "app", "req")

	ok := store.AddStatus(executionID, webhook.ExecutionStatus{
		EventType: string(webhook.EventTypeFailed),
		Status:    "error",
		Error:     "boom",
	})
	if !ok {
		t.Fatalf("AddStatus should return true")
	}

	record, exists := store.GetExecution(executionID)
	if !exists {
		t.Fatalf("execution record should exist")
	}
	if record.Status != "failed" {
		t.Fatalf("record.Status = %q, want failed", record.Status)
	}
	if record.Error != "boom" {
		t.Fatalf("record.Error = %q, want boom", record.Error)
	}
	if record.EndTime == nil {
		t.Fatalf("record.EndTime should be set for terminal event")
	}
}
