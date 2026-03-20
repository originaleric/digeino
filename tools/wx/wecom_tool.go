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

// NewWeComMsgOnEventTool 创建发送客服欢迎语工具
func NewWeComMsgOnEventTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"send_wecom_msg_on_event",
		"通过企业微信客服发送欢迎语（事件响应消息）。用于用户进入会话时，使用 sync_msg 事件中的 welcome_code 发送欢迎语。code 仅可使用一次，收到事件后 20 秒内有效。",
		func(ctx context.Context, req *SendWeComMsgOnEventRequest) (*SendWeComMsgOnEventResponse, error) {
			resp, err := SendWeComMsgOnEvent(ctx, *req)
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

// NewWeComCustomerImageTool 创建企业微信客服图片消息发送工具（发给个人微信用户）
func NewWeComCustomerImageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_image",
		"通过企业微信客服发送图片给个人微信用户。需 open_kf_id、customer_id 和 media_id（通过企业微信上传临时素材接口获得）。",
		func(ctx context.Context, req *SendWeComCustomerImageRequest) (*SendWeComCustomerImageResponse, error) {
			resp, err := SendWeComCustomerImage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerVoiceTool 创建企业微信客服语音消息发送工具（发给个人微信用户）
func NewWeComCustomerVoiceTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_voice",
		"通过企业微信客服发送语音给个人微信用户。需 open_kf_id、customer_id 和 media_id（通过企业微信上传临时素材接口获得）。",
		func(ctx context.Context, req *SendWeComCustomerVoiceRequest) (*SendWeComCustomerVoiceResponse, error) {
			resp, err := SendWeComCustomerVoice(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerVideoTool 创建企业微信客服视频消息发送工具（发给个人微信用户）
func NewWeComCustomerVideoTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_video",
		"通过企业微信客服发送视频给个人微信用户。需 open_kf_id、customer_id 和 media_id；可选 title、description。",
		func(ctx context.Context, req *SendWeComCustomerVideoRequest) (*SendWeComCustomerVideoResponse, error) {
			resp, err := SendWeComCustomerVideo(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerFileTool 创建企业微信客服文件消息发送工具（发给个人微信用户）
func NewWeComCustomerFileTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_file",
		"通过企业微信客服发送文件给个人微信用户。需 open_kf_id、customer_id 和 media_id（通过企业微信上传临时素材接口获得）。",
		func(ctx context.Context, req *SendWeComCustomerFileRequest) (*SendWeComCustomerFileResponse, error) {
			resp, err := SendWeComCustomerFile(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerLinkTool 创建企业微信客服图文链接消息发送工具（发给个人微信用户）
func NewWeComCustomerLinkTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_link",
		"通过企业微信客服发送图文链接给个人微信用户。需 open_kf_id、customer_id、title、desc、url、thumb_media_id（封面图通过上传临时素材获得）。",
		func(ctx context.Context, req *SendWeComCustomerLinkRequest) (*SendWeComCustomerLinkResponse, error) {
			resp, err := SendWeComCustomerLink(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewWeComCustomerMiniprogramTool 创建企业微信客服小程序卡片发送工具（发给个人微信用户）
func NewWeComCustomerMiniprogramTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}
	return utils.InferTool(
		"send_wecom_customer_miniprogram",
		"通过企业微信客服发送小程序卡片给个人微信用户。需 open_kf_id、customer_id、title、appid、thumb_media_id；可选 pagepath。",
		func(ctx context.Context, req *SendWeComCustomerMiniprogramRequest) (*SendWeComCustomerMiniprogramResponse, error) {
			resp, err := SendWeComCustomerMiniprogram(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// NewReceiveWeComCustomerMessageTool 创建接收企业微信客服消息工具
func NewReceiveWeComCustomerMessageTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.WeCom.Enabled == nil || !*cfg.WeCom.Enabled {
		return nil, fmt.Errorf("WeCom tool is not enabled in config")
	}
	if cfg.WeCom.CorpID == "" || len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom CorpID or Applications not configured")
	}

	return utils.InferTool(
		"receive_wecom_customer_message",
		"接收个人微信用户发送给企业微信机器人应用的消息。支持两种模式：1) 实时模式（realtime）：通过回调实时接收消息（推荐，消息即时到达）；2) 拉取模式（pull）：主动拉取最近3天内的消息（用于补拉历史消息或主动查询）。",
		func(ctx context.Context, req *ReceiveWeComCustomerMessageRequest) (*ReceiveWeComCustomerMessageResponse, error) {
			resp, err := ReceiveWeComCustomerMessage(ctx, *req)
			if err != nil {
				return nil, err
			}
			return &resp, nil
		},
	)
}

// ReceiveWeComCustomerMessage 接收企业微信客服消息
func ReceiveWeComCustomerMessage(ctx context.Context, req ReceiveWeComCustomerMessageRequest) (ReceiveWeComCustomerMessageResponse, error) {
	// 默认使用拉取模式
	mode := req.Mode
	if mode == "" {
		mode = "pull"
	}

	if mode == "pull" {
		// 拉取模式：需要先获取一个 token（通过回调事件获取，或使用最近的回调 token）
		// 这里简化处理，实际使用时需要传入 token
		// TODO: 可以考虑存储最近的回调 token，或提供其他方式获取 token
		return ReceiveWeComCustomerMessageResponse{
			Success:  false,
			Message:  "拉取模式需要 token，请通过回调事件获取或使用实时模式",
			Messages: []CustomerMessage{},
		}, fmt.Errorf("pull mode requires token from callback event")
	}

	// 实时模式：需要通过回调接收，这里返回提示信息
	return ReceiveWeComCustomerMessageResponse{
		Success:  false,
		Message:  "实时模式需要通过回调接收消息，请配置回调 URL 并设置 OnMessage 回调函数",
		Messages: []CustomerMessage{},
	}, fmt.Errorf("realtime mode requires callback configuration")
}
