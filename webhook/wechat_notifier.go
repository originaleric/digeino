package webhook

import (
	"context"
	"strings"

	"github.com/originaleric/digeino/tools/wx"
)

type WeChatNotifier struct {
	events map[string]struct{}
}

func NewWeChatNotifier(events []string) *WeChatNotifier {
	m := make(map[string]struct{}, len(events))
	for _, event := range events {
		v := strings.TrimSpace(strings.ToLower(event))
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return &WeChatNotifier{events: m}
}

func (n *WeChatNotifier) SendStatus(ctx context.Context, status ExecutionStatus) error {
	if n == nil {
		return nil
	}
	event := strings.ToLower(string(status.NormalizeEventType()))
	if len(n.events) > 0 {
		if _, ok := n.events[event]; !ok {
			return nil
		}
	}
	_, err := wx.SendWeChatTextMessage(ctx, wx.SendWeChatTextMessageRequest{
		Content: formatStatusTextMessage(status),
	})
	return err
}
