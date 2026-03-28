package webhook

import (
	"testing"

	"github.com/originaleric/digeino/config"
)

type fakeStore struct{}

func (f *fakeStore) AddStatus(executionID string, status ExecutionStatus) bool { return true }
func (f *fakeStore) SetResult(executionID string, result Message) bool          { return true }

func TestNewConfiguredCollectorReturnsNilWhenNoSinkEnabled(t *testing.T) {
	orig := config.Get()
	cfg := config.Default()
	disabled := false
	cfg.Status.Webhook.Enabled = &disabled
	cfg.Status.Store.Enabled = &disabled
	config.Set(cfg)
	defer config.Set(orig)

	collector := NewConfiguredCollector("exec", "app", "req", nil, nil, nil)
	if collector != nil {
		t.Fatalf("expected nil collector when no sink enabled")
	}
}

func TestNewConfiguredCollectorCreatesCollectorForCallbackSink(t *testing.T) {
	orig := config.Get()
	cfg := config.Default()
	disabled := false
	cfg.Status.Webhook.Enabled = &disabled
	cfg.Status.Store.Enabled = &disabled
	config.Set(cfg)
	defer config.Set(orig)

	collector := NewConfiguredCollector("exec", "app", "req", nil, func(status ExecutionStatus) {}, nil)
	if collector == nil {
		t.Fatalf("expected collector for callback sink")
	}
}

func TestNewConfiguredCollectorSkipsWebhookWhenURLMissing(t *testing.T) {
	orig := config.Get()
	cfg := config.Default()
	enabled := true
	disabled := false
	cfg.Status.Webhook.Enabled = &enabled
	cfg.Status.Webhook.URL = ""
	cfg.Status.Store.Enabled = &disabled
	config.Set(cfg)
	defer config.Set(orig)

	collector := NewConfiguredCollector("exec", "app", "req", nil, nil, nil)
	if collector != nil {
		t.Fatalf("expected nil collector when only webhook enabled but url missing")
	}
}

func TestNewConfiguredCollectorCreatesCollectorForStoreSink(t *testing.T) {
	orig := config.Get()
	cfg := config.Default()
	disabled := false
	enabled := true
	cfg.Status.Webhook.Enabled = &disabled
	cfg.Status.Store.Enabled = &enabled
	config.Set(cfg)
	defer config.Set(orig)

	collector := NewConfiguredCollector("exec", "app", "req", &fakeStore{}, nil, nil)
	if collector == nil {
		t.Fatalf("expected collector for store sink")
	}
}
