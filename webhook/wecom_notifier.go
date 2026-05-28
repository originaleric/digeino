package webhook

import (
	"context"
	"strings"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/tools/wx"
)

type WeComNotifier struct {
	events map[string]struct{}
}

func NewWeComNotifier(events []string) *WeComNotifier {
	m := make(map[string]struct{}, len(events))
	for _, event := range events {
		v := strings.TrimSpace(strings.ToLower(event))
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return &WeComNotifier{events: m}
}

func (n *WeComNotifier) SendStatus(ctx context.Context, status ExecutionStatus) error {
	if n == nil {
		return nil
	}
	event := strings.ToLower(string(status.NormalizeEventType()))
	if len(n.events) > 0 {
		if _, ok := n.events[event]; !ok {
			return nil
		}
	}
	cfg := config.Get()
	content := formatStatusTextMessage(status)
	for _, app := range cfg.WeCom.Applications {
		if strings.TrimSpace(app.ToUser) == "" {
			continue
		}
		if _, err := wx.SendWeComMessage(ctx, wx.SendWeComMessageRequest{
			UserID:  app.ToUser,
			Content: content,
			AgentID: app.AgentID,
		}); err != nil {
			return err
		}
	}
	return nil
}
