package feishu

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/webhook"
)

type SendFeishuMessageRequest struct {
	Content       string   `json:"content" jsonschema:"description=消息文本内容"`
	ReceiveIDType string   `json:"receive_id_type,omitempty" jsonschema:"description=接收者类型（chat_id/open_id/user_id/email）"`
	ReceiveIDs    []string `json:"receive_ids,omitempty" jsonschema:"description=接收者ID列表（为空则使用配置默认值）"`
}

type SendFeishuMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewFeishuMessageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.Feishu.Enabled == nil || !*cfg.Feishu.Enabled {
		return nil, fmt.Errorf("Feishu tool is not enabled in config")
	}
	apiCfg := webhook.GetFeishuAPIConfig()
	if apiCfg == nil {
		return nil, fmt.Errorf("Feishu API config is invalid or disabled")
	}
	client := webhook.NewFeishuClient(*apiCfg)

	return utils.InferTool(
		"send_feishu_message",
		"发送飞书文本消息。默认读取配置中的 ReceiveIDType 与 ReceiveIDs，也可在参数中覆盖。",
		func(ctx context.Context, req *SendFeishuMessageRequest) (*SendFeishuMessageResponse, error) {
			if req == nil || req.Content == "" {
				return nil, fmt.Errorf("content is required")
			}
			if err := client.SendText(ctx, req.ReceiveIDType, req.ReceiveIDs, req.Content); err != nil {
				return nil, err
			}
			return &SendFeishuMessageResponse{
				Success: true,
				Message: "飞书消息发送成功",
			}, nil
		},
	)
}

