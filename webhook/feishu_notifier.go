package webhook

import (
	"context"
	"fmt"
	"strings"
)

type FeishuNotifier struct {
	client *FeishuClient
	events map[string]struct{}
}

func NewFeishuNotifier(client *FeishuClient, events []string) *FeishuNotifier {
	m := make(map[string]struct{}, len(events))
	for _, event := range events {
		v := strings.TrimSpace(strings.ToLower(event))
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return &FeishuNotifier{
		client: client,
		events: m,
	}
}

func (n *FeishuNotifier) SendStatus(ctx context.Context, status ExecutionStatus) error {
	if n == nil || n.client == nil {
		return nil
	}
	event := strings.ToLower(string(status.NormalizeEventType()))
	if len(n.events) > 0 {
		if _, ok := n.events[event]; !ok {
			return nil
		}
	}
	content := formatFeishuStatusMessage(status)
	return n.client.SendText(ctx, "", nil, content)
}

func formatFeishuStatusMessage(status ExecutionStatus) string {
	event := status.NormalizeEventType()
	nodeID := status.NodeID
	if nodeID == "" {
		nodeID = status.NodeKey
	}
	if status.Type == "complete" {
		return fmt.Sprintf("DigEino运行通知\n应用: %s\n执行ID: %s\n事件: %s\n状态: %s\n请求ID: %s",
			status.AppName, status.ExecutionID, event, status.Status, status.RequestID)
	}
	return fmt.Sprintf("DigEino节点通知\n应用: %s\n执行ID: %s\n节点: %s\n事件: %s\n状态: %s",
		status.AppName, status.ExecutionID, nodeID, event, status.Status)
}

