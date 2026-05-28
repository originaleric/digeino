package webhook

import (
	"testing"

	"github.com/originaleric/digeino/config"
)

type fakeStore struct{}

func (f *fakeStore) AddStatus(executionID string, status ExecutionStatus) bool { return true }
func (f *fakeStore) SetResult(executionID string, result Message) bool         { return true }

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

func TestNewConfiguredCollectorUsesIndependentNotifyEventsPerChannel(t *testing.T) {
	orig := config.Get()
	cfg := config.Default()
	disabled := false
	enabled := true

	cfg.Status.Webhook.Enabled = &disabled
	cfg.Status.Store.Enabled = &disabled

	cfg.Feishu.Enabled = &enabled
	cfg.Feishu.SendViaAPI = &enabled
	cfg.Feishu.API.Enabled = &enabled
	cfg.Feishu.API.AppID = "app-id"
	cfg.Feishu.API.AppSecret = "app-secret"
	cfg.Feishu.NotifyOnEvents = []string{"failed"}

	cfg.WeChat.Enabled = &enabled
	cfg.WeChat.AppID = "wx-app-id"
	cfg.WeChat.AppSecret = "wx-secret"
	cfg.WeChat.OpenIDs = []string{"openid-1"}
	cfg.WeChat.NotifyOnEvents = []string{"completed"}

	cfg.WeCom.Enabled = &enabled
	cfg.WeCom.CorpID = "corp-id"
	cfg.WeCom.NotifyOnEvents = []string{"started", "failed"}
	cfg.WeCom.Applications = []config.WeComApplication{
		{
			AgentID:     1001,
			AgentSecret: "wecom-secret",
			ToUser:      "zhangsan",
		},
	}

	config.Set(cfg)
	defer config.Set(orig)

	collector := NewConfiguredCollector("exec", "app", "req", nil, nil, nil)
	if collector == nil {
		t.Fatalf("expected collector with feishu/wechat/wecom enabled")
	}
	if len(collector.notifiers) != 3 {
		t.Fatalf("expected 3 notifiers, got %d", len(collector.notifiers))
	}

	var (
		feishuNotifier *FeishuNotifier
		weChatNotifier *WeChatNotifier
		weComNotifier  *WeComNotifier
	)
	for _, n := range collector.notifiers {
		switch v := n.(type) {
		case *FeishuNotifier:
			feishuNotifier = v
		case *WeChatNotifier:
			weChatNotifier = v
		case *WeComNotifier:
			weComNotifier = v
		}
	}

	if feishuNotifier == nil || weChatNotifier == nil || weComNotifier == nil {
		t.Fatalf("expected feishu/wechat/wecom notifier all present")
	}
	if _, ok := feishuNotifier.events["failed"]; !ok {
		t.Fatalf("expected feishu notifier to include failed")
	}
	if _, ok := feishuNotifier.events["completed"]; ok {
		t.Fatalf("feishu notifier should not include completed")
	}

	if _, ok := weChatNotifier.events["completed"]; !ok {
		t.Fatalf("expected wechat notifier to include completed")
	}
	if _, ok := weChatNotifier.events["failed"]; ok {
		t.Fatalf("wechat notifier should not include failed")
	}

	if _, ok := weComNotifier.events["started"]; !ok {
		t.Fatalf("expected wecom notifier to include started")
	}
	if _, ok := weComNotifier.events["failed"]; !ok {
		t.Fatalf("expected wecom notifier to include failed")
	}
	if _, ok := weComNotifier.events["completed"]; ok {
		t.Fatalf("wecom notifier should not include completed")
	}
}
