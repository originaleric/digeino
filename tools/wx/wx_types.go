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

// SendWeChatImageMessageRequest 发送微信图片消息的请求参数
type SendWeChatImageMessageRequest struct {
	OpenID  string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	MediaID string `json:"media_id" jsonschema:"required,description=必填：图片的 media_id，通过素材上传接口获得"`
}

// SendWeChatImageMessageResponse 发送微信图片消息的响应
type SendWeChatImageMessageResponse struct {
	Success  bool            `json:"success"`
	SentTo   []string        `json:"sent_to"`   // 成功发送的用户列表
	FailedTo []FailedMessage `json:"failed_to"` // 发送失败的用户列表及原因
	Message  string          `json:"message"`
}

// SendWeChatMiniProgramPageRequest 发送微信小程序卡片消息的请求参数
type SendWeChatMiniProgramPageRequest struct {
	OpenID      string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Title       string `json:"title" jsonschema:"required,description=必填：小程序卡片标题"`
	AppID       string `json:"appid" jsonschema:"description=可选：小程序 AppID，如果不提供则使用配置中的值"`
	PagePath    string `json:"pagepath" jsonschema:"description=可选：小程序页面路径，如果不提供则使用配置中的默认路径"`
	ThumbMediaID string `json:"thumb_media_id" jsonschema:"description=可选：小程序卡片封面图片的 media_id，如果不提供则使用配置中的值"`
}

// SendWeChatMiniProgramPageResponse 发送微信小程序卡片消息的响应
type SendWeChatMiniProgramPageResponse struct {
	Success  bool            `json:"success"`
	SentTo   []string        `json:"sent_to"`   // 成功发送的用户列表
	FailedTo []FailedMessage `json:"failed_to"` // 发送失败的用户列表及原因
	Message  string          `json:"message"`
}

// SendWeChatLinkMessageRequest 发送微信图文链接消息的请求参数
type SendWeChatLinkMessageRequest struct {
	OpenID   string              `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Articles []LinkMessageArticle `json:"articles" jsonschema:"required,description=必填：图文消息列表，限制在1条以内"`
}

// LinkMessageArticle 图文消息文章
type LinkMessageArticle struct {
	Title       string `json:"title" jsonschema:"required,description=消息标题"`
	Description string `json:"description" jsonschema:"required,description=消息描述"`
	PicURL      string `json:"picurl" jsonschema:"required,description=封面图片url"`
	URL         string `json:"url" jsonschema:"required,description=跳转url"`
}

// SendWeChatLinkMessageResponse 发送微信图文链接消息的响应
type SendWeChatLinkMessageResponse struct {
	Success  bool            `json:"success"`
	SentTo   []string        `json:"sent_to"`   // 成功发送的用户列表
	FailedTo []FailedMessage `json:"failed_to"` // 发送失败的用户列表及原因
	Message  string          `json:"message"`
}
