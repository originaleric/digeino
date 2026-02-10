package wx

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/originaleric/digeino/config"
	"github.com/whyiyhw/go-workwx"
)

const maxTextLength = 850 // 企业微信单条消息建议不超过 850 字符

// getWeComApp 获取企业微信应用客户端
func getWeComApp(agentID int64) (*workwx.WorkwxApp, error) {
	cfg := config.Get()
	if cfg.WeCom.CorpID == "" {
		return nil, fmt.Errorf("WeCom CorpID not configured")
	}
	if len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom Applications not configured")
	}

	var corpSecret string
	var targetAgentID int64
	if agentID != 0 {
		for _, app := range cfg.WeCom.Applications {
			if app.AgentID == agentID {
				corpSecret = app.AgentSecret
				targetAgentID = app.AgentID
				break
			}
		}
		if corpSecret == "" {
			return nil, fmt.Errorf("WeCom application with AgentID %d not found", agentID)
		}
	} else {
		app := cfg.WeCom.Applications[0]
		corpSecret = app.AgentSecret
		targetAgentID = app.AgentID
	}

	if corpSecret == "" {
		return nil, fmt.Errorf("WeCom AgentSecret not configured")
	}

	var wx *workwx.Workwx
	if cfg.WeCom.QYAPIHost != "" {
		wx = workwx.New(cfg.WeCom.CorpID, workwx.WithQYAPIHost(cfg.WeCom.QYAPIHost))
	} else {
		wx = workwx.New(cfg.WeCom.CorpID)
	}
	return wx.WithApp(corpSecret, targetAgentID), nil
}

// splitMsg 按字符数切割消息（支持多字节字符）
func splitMsg(s string, maxLen int) []string {
	var result []string
	runes := []rune(s)
	for len(runes) > maxLen {
		result = append(result, string(runes[:maxLen]))
		runes = runes[maxLen:]
	}
	if len(runes) > 0 {
		result = append(result, string(runes))
	}
	return result
}

// SendWeComMessage 发送企业微信文字消息
func SendWeComMessage(ctx context.Context, req SendWeComMessageRequest) (SendWeComMessageResponse, error) {
	if req.UserID == "" {
		return SendWeComMessageResponse{}, fmt.Errorf("user_id is required")
	}
	if req.Content == "" {
		return SendWeComMessageResponse{}, fmt.Errorf("content is required")
	}

	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComMessageResponse{}, err
	}

	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}

	// 超过 850 字分条发送
	if utf8.RuneCountInString(req.Content) > maxTextLength {
		parts := splitMsg(req.Content, maxTextLength)
		for _, part := range parts {
			if err := app.SendTextMessage(recipient, part, false); err != nil {
				return SendWeComMessageResponse{
					Success: false,
					Message: fmt.Sprintf("发送失败: %v", err),
				}, err
			}
		}
		return SendWeComMessageResponse{
			Success: true,
			Message: fmt.Sprintf("成功发送 %d 条消息", len(parts)),
		}, nil
	}

	if err := app.SendTextMessage(recipient, req.Content, false); err != nil {
		return SendWeComMessageResponse{
			Success: false,
			Message: fmt.Sprintf("发送失败: %v", err),
		}, err
	}

	return SendWeComMessageResponse{
		Success: true,
		Message: "消息发送成功",
	}, nil
}

// SendWeComImageMessage 发送企业微信图片消息
func SendWeComImageMessage(ctx context.Context, req SendWeComImageMessageRequest) (SendWeComImageMessageResponse, error) {
	if req.UserID == "" {
		return SendWeComImageMessageResponse{}, fmt.Errorf("user_id is required")
	}
	if req.MediaID == "" {
		return SendWeComImageMessageResponse{}, fmt.Errorf("media_id is required")
	}

	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComImageMessageResponse{}, err
	}

	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}
	if err := app.SendImageMessage(recipient, req.MediaID, false); err != nil {
		return SendWeComImageMessageResponse{
			Success: false,
			Message: fmt.Sprintf("发送失败: %v", err),
		}, err
	}

	return SendWeComImageMessageResponse{
		Success: true,
		Message: "图片消息发送成功",
	}, nil
}

// SendWeComTextCard 发送企业微信文本卡片消息
func SendWeComTextCard(ctx context.Context, req SendWeComTextCardRequest) (SendWeComTextCardResponse, error) {
	if req.UserID == "" {
		return SendWeComTextCardResponse{}, fmt.Errorf("user_id is required")
	}
	if req.Title == "" {
		return SendWeComTextCardResponse{}, fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return SendWeComTextCardResponse{}, fmt.Errorf("description is required")
	}
	if req.URL == "" {
		return SendWeComTextCardResponse{}, fmt.Errorf("url is required")
	}

	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComTextCardResponse{}, err
	}

	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}
	btnText := "详情"
	if err := app.SendTextCardMessage(recipient, req.Title, req.Description, req.URL, btnText, false); err != nil {
		return SendWeComTextCardResponse{
			Success: false,
			Message: fmt.Sprintf("发送失败: %v", err),
		}, err
	}

	return SendWeComTextCardResponse{
		Success: true,
		Message: "文本卡片消息发送成功",
	}, nil
}
