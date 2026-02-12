package wx

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// NewWeComMessageTool 创建企业微信文字消息发送工具
func NewWeComMessageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"send_wecom_message",
		"通过企业微信应用消息发送文字给指定的企业成员。user_id 为企业成员 ID，由 DigFlow 等调用方从会话上下文传入。支持超过 850 字时自动分条发送。",
		func(ctx context.Context, req *SendWeComMessageRequest) (*SendWeComMessageResponse, error) {
			resp, err := SendWeComMessage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComImageTool 创建企业微信图片消息发送工具
func NewWeComImageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"send_wecom_image",
		"通过企业微信应用消息发送图片给指定的企业成员。需要先通过临时素材上传接口获取 media_id。",
		func(ctx context.Context, req *SendWeComImageMessageRequest) (*SendWeComImageMessageResponse, error) {
			resp, err := SendWeComImageMessage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComTextCardTool 创建企业微信文本卡片消息发送工具
func NewWeComTextCardTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"send_wecom_text_card",
		"通过企业微信应用消息发送文本卡片给指定的企业成员。文本卡片包含标题、描述和可点击的链接，适合通知类场景。",
		func(ctx context.Context, req *SendWeComTextCardRequest) (*SendWeComTextCardResponse, error) {
			resp, err := SendWeComTextCard(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerMessageTool 创建企业微信客服消息发送工具（发给个人微信用户）
func NewWeComCustomerMessageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"send_wecom_customer_message",
		"通过企业微信客服消息发送文字给个人微信用户。需使用 open_kf_id（客服账号ID）和 customer_id（外部联系人ID）。用户需先通过扫码/链接添加企业为客服后才能收到消息。支持超过 850 字时自动分条发送。",
		func(ctx context.Context, req *SendWeComCustomerMessageRequest) (*SendWeComCustomerMessageResponse, error) {
			resp, err := SendWeComCustomerMessage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}
