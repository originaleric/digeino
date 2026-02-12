package wx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/originaleric/digeino/config"
	"github.com/whyiyhw/go-workwx"
)

const maxTextLength = 850 // 企业微信单条消息建议不超过 850 字符

// getWeComAgentID 解析 agentID，若为 0 则从配置取第一个
func getWeComAgentID(agentID int64) (int64, error) {
	if agentID != 0 {
		return agentID, nil
	}
	cfg := config.Get()
	if len(cfg.WeCom.Applications) == 0 {
		return 0, fmt.Errorf("WeCom Applications not configured")
	}
	return cfg.WeCom.Applications[0].AgentID, nil
}

// getWeComAPIHost 获取企业微信 API 地址
func getWeComAPIHost() string {
	cfg := config.Get()
	if cfg.WeCom.QYAPIHost != "" {
		return cfg.WeCom.QYAPIHost
	}
	return "https://qyapi.weixin.qq.com"
}

// sendWeComMessageAPI 直接调用企业微信 API 发送消息（用于第三方传入 token）
func sendWeComMessageAPI(ctx context.Context, accessToken string, agentID int64, userID string, body map[string]interface{}) error {
	url := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", getWeComAPIHost(), accessToken)
	body["touser"] = userID
	body["agentid"] = agentID

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call WeCom API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var apiResp WeComMessageAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if apiResp.ErrCode != 0 {
		return fmt.Errorf("WeCom API errcode=%d errmsg=%s", apiResp.ErrCode, apiResp.ErrMsg)
	}
	return nil
}

// sendWeComCustomerMessageAPI 直接调用企业微信客服 API 发送消息（用于第三方传入 token）
func sendWeComCustomerMessageAPI(ctx context.Context, accessToken string, openKfID string, customerID string, body map[string]interface{}) error {
	url := fmt.Sprintf("%s/cgi-bin/kf/send_msg?access_token=%s", getWeComAPIHost(), accessToken)
	body["touser"] = customerID
	body["open_kfid"] = openKfID

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call WeCom API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var apiResp WeComMessageAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if apiResp.ErrCode != 0 {
		return fmt.Errorf("WeCom API errcode=%d errmsg=%s", apiResp.ErrCode, apiResp.ErrMsg)
	}
	return nil
}

// getWeComCustomerApp 获取具备「管理所有客服会话」权限的企业微信应用（用于发送客服消息给个人微信）
func getWeComCustomerApp() (*workwx.WorkwxApp, error) {
	cfg := config.Get()
	if cfg.WeCom.CorpID == "" {
		return nil, fmt.Errorf("WeCom CorpID not configured")
	}
	if len(cfg.WeCom.Applications) == 0 {
		return nil, fmt.Errorf("WeCom Applications not configured")
	}

	var corpSecret string
	var targetAgentID int64
	for _, app := range cfg.WeCom.Applications {
		if app.ManageAllKFSession {
			corpSecret = app.AgentSecret
			targetAgentID = app.AgentID
			break
		}
	}
	if corpSecret == "" {
		return nil, fmt.Errorf("no WeCom application with ManageAllKFSession=true found in config")
	}

	var wx *workwx.Workwx
	if cfg.WeCom.QYAPIHost != "" {
		wx = workwx.New(cfg.WeCom.CorpID, workwx.WithQYAPIHost(cfg.WeCom.QYAPIHost))
	} else {
		wx = workwx.New(cfg.WeCom.CorpID)
	}
	return wx.WithApp(corpSecret, targetAgentID), nil
}

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

	agentID, err := getWeComAgentID(req.AgentID)
	if err != nil {
		return SendWeComMessageResponse{}, err
	}

	// 第三方传入 token 时直接调 API
	if req.AccessToken != "" {
		if utf8.RuneCountInString(req.Content) > maxTextLength {
			parts := splitMsg(req.Content, maxTextLength)
			for _, part := range parts {
				body := map[string]interface{}{
					"msgtype": "text",
					"text":    map[string]string{"content": part},
				}
				if err := sendWeComMessageAPI(ctx, req.AccessToken, agentID, req.UserID, body); err != nil {
					return SendWeComMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
				}
			}
			return SendWeComMessageResponse{Success: true, Message: fmt.Sprintf("成功发送 %d 条消息", len(parts))}, nil
		}
		body := map[string]interface{}{
			"msgtype": "text",
			"text":    map[string]string{"content": req.Content},
		}
		if err := sendWeComMessageAPI(ctx, req.AccessToken, agentID, req.UserID, body); err != nil {
			return SendWeComMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
		}
		return SendWeComMessageResponse{Success: true, Message: "消息发送成功"}, nil
	}

	// 内部 token：使用 go-workwx
	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComMessageResponse{}, err
	}
	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}

	if utf8.RuneCountInString(req.Content) > maxTextLength {
		parts := splitMsg(req.Content, maxTextLength)
		for _, part := range parts {
			if err := app.SendTextMessage(recipient, part, false); err != nil {
				return SendWeComMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
			}
		}
		return SendWeComMessageResponse{Success: true, Message: fmt.Sprintf("成功发送 %d 条消息", len(parts))}, nil
	}

	if err := app.SendTextMessage(recipient, req.Content, false); err != nil {
		return SendWeComMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
	}
	return SendWeComMessageResponse{Success: true, Message: "消息发送成功"}, nil
}

// SendWeComImageMessage 发送企业微信图片消息
func SendWeComImageMessage(ctx context.Context, req SendWeComImageMessageRequest) (SendWeComImageMessageResponse, error) {
	if req.UserID == "" {
		return SendWeComImageMessageResponse{}, fmt.Errorf("user_id is required")
	}
	if req.MediaID == "" {
		return SendWeComImageMessageResponse{}, fmt.Errorf("media_id is required")
	}

	agentID, err := getWeComAgentID(req.AgentID)
	if err != nil {
		return SendWeComImageMessageResponse{}, err
	}

	if req.AccessToken != "" {
		body := map[string]interface{}{
			"msgtype": "image",
			"image":   map[string]string{"media_id": req.MediaID},
		}
		if err := sendWeComMessageAPI(ctx, req.AccessToken, agentID, req.UserID, body); err != nil {
			return SendWeComImageMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
		}
		return SendWeComImageMessageResponse{Success: true, Message: "图片消息发送成功"}, nil
	}

	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComImageMessageResponse{}, err
	}
	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}
	if err := app.SendImageMessage(recipient, req.MediaID, false); err != nil {
		return SendWeComImageMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
	}
	return SendWeComImageMessageResponse{Success: true, Message: "图片消息发送成功"}, nil
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

	agentID, err := getWeComAgentID(req.AgentID)
	if err != nil {
		return SendWeComTextCardResponse{}, err
	}

	if req.AccessToken != "" {
		body := map[string]interface{}{
			"msgtype": "textcard",
			"textcard": map[string]string{
				"title":       req.Title,
				"description": req.Description,
				"url":         req.URL,
				"btntxt":      "详情",
			},
		}
		if err := sendWeComMessageAPI(ctx, req.AccessToken, agentID, req.UserID, body); err != nil {
			return SendWeComTextCardResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
		}
		return SendWeComTextCardResponse{Success: true, Message: "文本卡片消息发送成功"}, nil
	}

	app, err := getWeComApp(req.AgentID)
	if err != nil {
		return SendWeComTextCardResponse{}, err
	}
	recipient := &workwx.Recipient{UserIDs: []string{req.UserID}}
	if err := app.SendTextCardMessage(recipient, req.Title, req.Description, req.URL, "详情", false); err != nil {
		return SendWeComTextCardResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
	}
	return SendWeComTextCardResponse{Success: true, Message: "文本卡片消息发送成功"}, nil
}

// SendWeComCustomerMessage 发送企业微信客服消息（发给个人微信用户）
// 使用企业微信「客户联系」的客服能力，用户需先通过扫码/链接添加企业为客服后才可收到消息
func SendWeComCustomerMessage(ctx context.Context, req SendWeComCustomerMessageRequest) (SendWeComCustomerMessageResponse, error) {
	if req.OpenKfID == "" {
		return SendWeComCustomerMessageResponse{}, fmt.Errorf("open_kf_id is required")
	}
	if req.CustomerID == "" {
		return SendWeComCustomerMessageResponse{}, fmt.Errorf("customer_id is required")
	}
	if req.Content == "" {
		return SendWeComCustomerMessageResponse{}, fmt.Errorf("content is required")
	}

	// 第三方传入 token 时直接调客服 API
	if req.AccessToken != "" {
		if utf8.RuneCountInString(req.Content) > maxTextLength {
			parts := splitMsg(req.Content, maxTextLength)
			for _, part := range parts {
				body := map[string]interface{}{
					"msgtype": "text",
					"text":    map[string]string{"content": part},
				}
				if err := sendWeComCustomerMessageAPI(ctx, req.AccessToken, req.OpenKfID, req.CustomerID, body); err != nil {
					return SendWeComCustomerMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
				}
			}
			return SendWeComCustomerMessageResponse{Success: true, Message: fmt.Sprintf("成功发送 %d 条消息", len(parts))}, nil
		}
		body := map[string]interface{}{
			"msgtype": "text",
			"text":    map[string]string{"content": req.Content},
		}
		if err := sendWeComCustomerMessageAPI(ctx, req.AccessToken, req.OpenKfID, req.CustomerID, body); err != nil {
			return SendWeComCustomerMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
		}
		return SendWeComCustomerMessageResponse{Success: true, Message: "客服消息发送成功"}, nil
	}

	// 内部 token：使用 go-workwx，需配置 ManageAllKFSession
	app, err := getWeComCustomerApp()
	if err != nil {
		return SendWeComCustomerMessageResponse{}, err
	}
	recipient := &workwx.Recipient{
		UserIDs:  []string{req.CustomerID},
		OpenKfID: req.OpenKfID,
	}

	if utf8.RuneCountInString(req.Content) > maxTextLength {
		parts := splitMsg(req.Content, maxTextLength)
		for _, part := range parts {
			if err := app.SendTextMessage(recipient, part, false); err != nil {
				return SendWeComCustomerMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
			}
		}
		return SendWeComCustomerMessageResponse{Success: true, Message: fmt.Sprintf("成功发送 %d 条消息", len(parts))}, nil
	}

	if err := app.SendTextMessage(recipient, req.Content, false); err != nil {
		return SendWeComCustomerMessageResponse{Success: false, Message: fmt.Sprintf("发送失败: %v", err)}, err
	}
	return SendWeComCustomerMessageResponse{Success: true, Message: "客服消息发送成功"}, nil
}
