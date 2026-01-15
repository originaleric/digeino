package wx

// SendWeChatTextMessageRequest 发送微信文字消息的请求参数
type SendWeChatTextMessageRequest struct {
	OpenID  string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Content string `json:"content" jsonschema:"required,description=必填：要发送的文字消息内容，建议不超过2048字符"`
}

// SendWeChatTextMessageResponse 发送微信文字消息的响应
type SendWeChatTextMessageResponse struct {
	Success  bool            `json:"success"`
	SentTo   []string        `json:"sent_to"`   // 成功发送的用户列表
	FailedTo []FailedMessage `json:"failed_to"` // 发送失败的用户列表及原因
	Message  string          `json:"message"`
}

// FailedMessage 发送失败的消息信息
type FailedMessage struct {
	OpenID string `json:"openid"`
	Reason string `json:"reason"` // 失败原因
}

// AccessTokenData AccessToken 存储结构
type AccessTokenData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` // 微信返回的过期时间（秒）
	CreatedAt   int64  `json:"created_at"` // 创建时间戳（Unix 秒）
	ExpiresAt   int64  `json:"expires_at"` // 过期时间戳（Unix 秒）
}

// WeChatTokenResponse 微信获取 AccessToken 的 API 响应
type WeChatTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WeChatMessageResponse 微信发送消息的 API 响应
type WeChatMessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
