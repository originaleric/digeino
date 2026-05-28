package webhook

import "fmt"

func formatStatusTextMessage(status ExecutionStatus) string {
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
