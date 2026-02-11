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

	// 3. 获取 AccessToken（优先使用请求中传入的）
	accessToken := req.AccessToken
	if accessToken == "" {
		var err error
		accessToken, err = GetAccessToken(ctx)
		if err != nil {
			return SendWeChatTextMessageResponse{}, fmt.Errorf("failed to get access token: %w", err)
		}
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
	if content == "" {
		return fmt.Errorf("content is required")
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

	return sendMessageRequest(ctx, url, requestBody, openid)
}

// SendWeChatImageMessage 发送微信图片消息
func SendWeChatImageMessage(ctx context.Context, req SendWeChatImageMessageRequest) (SendWeChatImageMessageResponse, error) {
	// 1. 验证请求参数
	if req.MediaID == "" {
		return SendWeChatImageMessageResponse{}, fmt.Errorf("media_id is required")
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
			return SendWeChatImageMessageResponse{}, fmt.Errorf("no openid specified and no openids configured")
		}
	}

	// 3. 获取 AccessToken（优先使用请求中传入的）
	accessToken := req.AccessToken
	if accessToken == "" {
		var err error
		accessToken, err = GetAccessToken(ctx)
		if err != nil {
			return SendWeChatImageMessageResponse{}, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// 4. 逐个发送消息
	var sentTo []string
	var failedTo []FailedMessage

	for _, openid := range openids {
		err := sendImageMessageToUser(ctx, accessToken, openid, req.MediaID)
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
	response := SendWeChatImageMessageResponse{
		Success:  len(failedTo) == 0,
		SentTo:   sentTo,
		FailedTo: failedTo,
	}

	if len(sentTo) > 0 && len(failedTo) == 0 {
		response.Message = fmt.Sprintf("成功发送图片消息给 %d 个用户", len(sentTo))
	} else if len(sentTo) > 0 && len(failedTo) > 0 {
		response.Message = fmt.Sprintf("部分成功：成功发送给 %d 个用户，失败 %d 个", len(sentTo), len(failedTo))
	} else {
		response.Message = fmt.Sprintf("发送失败：所有 %d 个用户都发送失败", len(failedTo))
	}

	return response, nil
}

// sendImageMessageToUser 向单个用户发送图片消息
func sendImageMessageToUser(ctx context.Context, accessToken, openid, mediaID string) error {
	if openid == "" {
		return fmt.Errorf("openid is required")
	}
	if mediaID == "" {
		return fmt.Errorf("media_id is required")
	}

	// 微信客服消息接口
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

	// 构建请求体
	requestBody := map[string]interface{}{
		"touser":  openid,
		"msgtype": "image",
		"image": map[string]string{
			"media_id": mediaID,
		},
	}

	return sendMessageRequest(ctx, url, requestBody, openid)
}

// SendWeChatMiniProgramPage 发送微信小程序卡片消息
func SendWeChatMiniProgramPage(ctx context.Context, req SendWeChatMiniProgramPageRequest) (SendWeChatMiniProgramPageResponse, error) {
	// 1. 验证请求参数
	if req.Title == "" {
		return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("title is required")
	}

	cfg := config.Get()

	// 2. 确定小程序配置
	appID := req.AppID
	if appID == "" {
		appID = cfg.WeChat.MiniProgram.AppID
		if appID == "" {
			return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("appid is required (either in request or config)")
		}
	}

	pagePath := req.PagePath
	if pagePath == "" {
		pagePath = cfg.WeChat.MiniProgram.DefaultPath
		if pagePath == "" {
			return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("pagepath is required (either in request or config)")
		}
	}

	thumbMediaID := req.ThumbMediaID
	if thumbMediaID == "" {
		thumbMediaID = cfg.WeChat.MiniProgram.ThumbMediaID
		if thumbMediaID == "" {
			return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("thumb_media_id is required (either in request or config)")
		}
	}

	// 3. 确定接收者列表
	var openids []string
	if req.OpenID != "" {
		// 指定了单个 openid
		openids = []string{req.OpenID}
	} else {
		// 从配置中读取 openid 列表
		openids = cfg.WeChat.OpenIDs
		if len(openids) == 0 {
			return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("no openid specified and no openids configured")
		}
	}

	// 4. 获取 AccessToken（优先使用请求中传入的）
	accessToken := req.AccessToken
	if accessToken == "" {
		var err error
		accessToken, err = GetAccessToken(ctx)
		if err != nil {
			return SendWeChatMiniProgramPageResponse{}, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// 5. 逐个发送消息
	var sentTo []string
	var failedTo []FailedMessage

	for _, openid := range openids {
		err := sendMiniProgramPageMessageToUser(ctx, accessToken, openid, req.Title, appID, pagePath, thumbMediaID)
		if err != nil {
			failedTo = append(failedTo, FailedMessage{
				OpenID: openid,
				Reason: err.Error(),
			})
		} else {
			sentTo = append(sentTo, openid)
		}
	}

	// 6. 构建响应
	response := SendWeChatMiniProgramPageResponse{
		Success:  len(failedTo) == 0,
		SentTo:   sentTo,
		FailedTo: failedTo,
	}

	if len(sentTo) > 0 && len(failedTo) == 0 {
		response.Message = fmt.Sprintf("成功发送小程序卡片消息给 %d 个用户", len(sentTo))
	} else if len(sentTo) > 0 && len(failedTo) > 0 {
		response.Message = fmt.Sprintf("部分成功：成功发送给 %d 个用户，失败 %d 个", len(sentTo), len(failedTo))
	} else {
		response.Message = fmt.Sprintf("发送失败：所有 %d 个用户都发送失败", len(failedTo))
	}

	return response, nil
}

// sendMiniProgramPageMessageToUser 向单个用户发送小程序卡片消息
func sendMiniProgramPageMessageToUser(ctx context.Context, accessToken, openid, title, appID, pagePath, thumbMediaID string) error {
	if openid == "" {
		return fmt.Errorf("openid is required")
	}
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if appID == "" {
		return fmt.Errorf("appid is required")
	}
	if pagePath == "" {
		return fmt.Errorf("pagepath is required")
	}
	if thumbMediaID == "" {
		return fmt.Errorf("thumb_media_id is required")
	}

	// 微信客服消息接口
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

	// 构建请求体
	requestBody := map[string]interface{}{
		"touser":  openid,
		"msgtype": "miniprogrampage",
		"miniprogrampage": map[string]string{
			"title":         title,
			"appid":         appID,
			"pagepath":      pagePath,
			"thumb_media_id": thumbMediaID,
		},
	}

	return sendMessageRequest(ctx, url, requestBody, openid)
}

// SendWeChatLinkMessage 发送微信图文链接消息
func SendWeChatLinkMessage(ctx context.Context, req SendWeChatLinkMessageRequest) (SendWeChatLinkMessageResponse, error) {
	// 1. 验证请求参数
	if len(req.Articles) == 0 {
		return SendWeChatLinkMessageResponse{}, fmt.Errorf("articles is required and cannot be empty")
	}
	if len(req.Articles) > 1 {
		return SendWeChatLinkMessageResponse{}, fmt.Errorf("articles count must be 1 or less")
	}

	// 验证文章内容
	for i, article := range req.Articles {
		if article.Title == "" {
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("article[%d].title is required", i)
		}
		if article.Description == "" {
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("article[%d].description is required", i)
		}
		if article.PicURL == "" {
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("article[%d].picurl is required", i)
		}
		if article.URL == "" {
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("article[%d].url is required", i)
		}
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
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("no openid specified and no openids configured")
		}
	}

	// 3. 获取 AccessToken（优先使用请求中传入的）
	accessToken := req.AccessToken
	if accessToken == "" {
		var err error
		accessToken, err = GetAccessToken(ctx)
		if err != nil {
			return SendWeChatLinkMessageResponse{}, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// 4. 逐个发送消息
	var sentTo []string
	var failedTo []FailedMessage

	for _, openid := range openids {
		err := sendLinkMessageToUser(ctx, accessToken, openid, req.Articles)
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
	response := SendWeChatLinkMessageResponse{
		Success:  len(failedTo) == 0,
		SentTo:   sentTo,
		FailedTo: failedTo,
	}

	if len(sentTo) > 0 && len(failedTo) == 0 {
		response.Message = fmt.Sprintf("成功发送图文链接消息给 %d 个用户", len(sentTo))
	} else if len(sentTo) > 0 && len(failedTo) > 0 {
		response.Message = fmt.Sprintf("部分成功：成功发送给 %d 个用户，失败 %d 个", len(sentTo), len(failedTo))
	} else {
		response.Message = fmt.Sprintf("发送失败：所有 %d 个用户都发送失败", len(failedTo))
	}

	return response, nil
}

// sendLinkMessageToUser 向单个用户发送图文链接消息
func sendLinkMessageToUser(ctx context.Context, accessToken, openid string, articles []LinkMessageArticle) error {
	if openid == "" {
		return fmt.Errorf("openid is required")
	}
	if len(articles) == 0 {
		return fmt.Errorf("articles is required")
	}

	// 微信客服消息接口
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

	// 构建文章列表
	articlesList := make([]map[string]string, len(articles))
	for i, article := range articles {
		articlesList[i] = map[string]string{
			"title":       article.Title,
			"description": article.Description,
			"picurl":      article.PicURL,
			"url":         article.URL,
		}
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"touser":  openid,
		"msgtype": "news",
		"news": map[string]interface{}{
			"articles": articlesList,
		},
	}

	return sendMessageRequest(ctx, url, requestBody, openid)
}

// sendMessageRequest 通用的发送消息请求函数
func sendMessageRequest(ctx context.Context, url string, requestBody map[string]interface{}, openid string) error {
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
