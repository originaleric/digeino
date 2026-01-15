package wx

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// NewWeChatMessageTool 创建微信推送工具
func NewWeChatMessageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeChat.Enabled == nil || !*cfg.WeChat.Enabled {
		return nil, fmt.Errorf("WeChat tool is not enabled in config")
	}

	if cfg.WeChat.AppID == "" || cfg.WeChat.AppSecret == "" {
		return nil, fmt.Errorf("WeChat AppID or AppSecret not configured")
	}

	return utils.InferTool(
		"send_wechat_message",
		"通过微信服务号推送文字消息给指定的用户openid。支持发送给单个用户或配置中的所有用户。注意：用户需要在48小时内与公众号有交互才能接收客服消息。",
		func(ctx context.Context, req *SendWeChatTextMessageRequest) (*SendWeChatTextMessageResponse, error) {
			resp, err := SendWeChatTextMessage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}
