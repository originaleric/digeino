package wx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/originaleric/digeino/config"
)

// SendWeChatTextMessage 发送微信文字消息
func SendWeChatTextMessage(ctx context.Context, req SendWeChatTextMessageRequest) (SendWeChatTextMessageResponse, error) {
	// 1. 验证请求参数
	if req.Content == "" {
		return SendWeChatTextMessageResponse{}, fmt.Errorf("content is required")
	}

	cfg := config.Get()

	// 2. 确定接收者列表
	var openids []string
	if req.OpenID != "" {
		// 指定了单个 openid
		openids = []string{req.OpenID}
	} else {
		// 从配置中读取 openid 列表
		openids = cfg.WeChat.OpenIDs
		if len(openids) == 0 {
			return SendWeChatTextMessageResponse{}, fmt.Errorf("no openid specified and no openids configured")
		}
	}

	// 3. 获取 AccessToken
	accessToken, err := GetAccessToken(ctx)
	if err != nil {
		return SendWeChatTextMessageResponse{}, fmt.Errorf("failed to get access token: %w", err)
	}

	// 4. 逐个发送消息
	var sentTo []string
	var failedTo []FailedMessage

	for _, openid := range openids {
		err := sendTextMessageToUser(ctx, accessToken, openid, req.Content)
		if err != nil {
			failedTo = append(failedTo, FailedMessage{
				OpenID: openid,
				Reason: err.Error(),
			})
		} else {
			sentTo = append(sentTo, openid)
		}
	}

	// 5. 构建响应
	response := SendWeChatTextMessageResponse{
		Success:  len(failedTo) == 0,
		SentTo:   sentTo,
		FailedTo: failedTo,
	}

	if len(sentTo) > 0 && len(failedTo) == 0 {
		response.Message = fmt.Sprintf("成功发送消息给 %d 个用户", len(sentTo))
	} else if len(sentTo) > 0 && len(failedTo) > 0 {
		response.Message = fmt.Sprintf("部分成功：成功发送给 %d 个用户，失败 %d 个", len(sentTo), len(failedTo))
	} else {
		response.Message = fmt.Sprintf("发送失败：所有 %d 个用户都发送失败", len(failedTo))
	}

	return response, nil
}

// sendTextMessageToUser 向单个用户发送文字消息
func sendTextMessageToUser(ctx context.Context, accessToken, openid, content string) error {
	if openid == "" {
		return fmt.Errorf("openid is required")
	}

	// 微信客服消息接口
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

	// 构建请求体
	requestBody := map[string]interface{}{
		"touser":  openid,
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call WeChat API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp WeChatMessageResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查响应
	if apiResp.ErrCode != 0 {
		// 处理常见的错误码
		switch apiResp.ErrCode {
		case 45015:
			return fmt.Errorf("用户48小时内未与公众号交互，无法发送客服消息")
		case 40001:
			return fmt.Errorf("access_token 无效，请检查配置")
		case 40003:
			return fmt.Errorf("openid 无效: %s", openid)
		case 45047:
			return fmt.Errorf("发送消息过于频繁，请稍后再试")
		default:
			return fmt.Errorf("WeChat API error: errcode=%d, errmsg=%s", apiResp.ErrCode, apiResp.ErrMsg)
		}
	}

	return nil
}
