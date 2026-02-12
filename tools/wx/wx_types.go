package wx

// SendWeChatTextMessageRequest 发送微信文字消息的请求参数
type SendWeChatTextMessageRequest struct {
	OpenID      string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Content     string `json:"content" jsonschema:"required,description=必填：要发送的文字消息内容，建议不超过2048字符"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用，否则内部自动获取"`
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
	OpenID      string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：图片的 media_id，通过素材上传接口获得"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
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
	OpenID       string `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Title        string `json:"title" jsonschema:"required,description=必填：小程序卡片标题"`
	AppID        string `json:"appid" jsonschema:"description=可选：小程序 AppID，如果不提供则使用配置中的值"`
	PagePath     string `json:"pagepath" jsonschema:"description=可选：小程序页面路径，如果不提供则使用配置中的默认路径"`
	ThumbMediaID string `json:"thumb_media_id" jsonschema:"description=可选：小程序卡片封面图片的 media_id，如果不提供则使用配置中的值"`
	AccessToken  string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
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
	OpenID      string              `json:"openid" jsonschema:"description=可选：指定单个接收者的 openid。如果不提供，将发送给配置中的所有用户"`
	Articles    []LinkMessageArticle `json:"articles" jsonschema:"required,description=必填：图文消息列表，限制在1条以内"`
	AccessToken string              `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
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

// --- 企业微信相关类型 ---

// SendWeComMessageRequest 发送企业微信文字消息的请求参数
type SendWeComMessageRequest struct {
	UserID      string `json:"user_id" jsonschema:"required,description=必填：企业成员 userID"`
	Content     string `json:"content" jsonschema:"required,description=必填：要发送的文字内容，建议不超过2048字符"`
	AgentID     int64  `json:"agent_id" jsonschema:"description=可选：应用 ID，不传则用配置中第一个应用"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
}

// SendWeComMessageResponse 发送企业微信消息的响应
type SendWeComMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComImageMessageRequest 发送企业微信图片消息的请求参数
type SendWeComImageMessageRequest struct {
	UserID      string `json:"user_id" jsonschema:"required,description=必填：企业成员 userID"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：图片的 media_id，通过临时素材上传接口获得"`
	AgentID     int64  `json:"agent_id" jsonschema:"description=可选：应用 ID"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
}

// SendWeComImageMessageResponse 发送企业微信图片消息的响应
type SendWeComImageMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComTextCardRequest 发送企业微信文本卡片消息的请求参数
type SendWeComTextCardRequest struct {
	UserID      string `json:"user_id" jsonschema:"required,description=必填：企业成员 userID"`
	Title       string `json:"title" jsonschema:"required,description=必填：卡片标题"`
	Description string `json:"description" jsonschema:"required,description=必填：卡片描述"`
	URL         string `json:"url" jsonschema:"required,description=必填：点击卡片的跳转链接"`
	AgentID     int64  `json:"agent_id" jsonschema:"description=可选：应用 ID"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，若提供则优先使用"`
}

// SendWeComTextCardResponse 发送企业微信文本卡片消息的响应
type SendWeComTextCardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WeComTokenData 企业微信 AccessToken 存储结构（按 agentID 区分）
type WeComTokenData struct {
	AgentID     int64  `json:"agent_id"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	CreatedAt   int64  `json:"created_at"`
	ExpiresAt   int64  `json:"expires_at"`
}

// WeComTokenAPIResponse 企业微信获取 token 的 API 响应
type WeComTokenAPIResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WeComMessageAPIResponse 企业微信发送消息 API 响应
type WeComMessageAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendWeComCustomerMessageRequest 发送企业微信客服消息的请求参数（发给个人微信用户）
// 使用企业微信「客户联系」的客服能力，用户需先通过扫码/链接添加企业为客服后才可收到消息
type SendWeComCustomerMessageRequest struct {
	OpenKfID    string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID，在企业微信管理后台创建客服账号后获得"`
	CustomerID  string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID（external_userid），用户添加企业为客服后产生"`
	Content     string `json:"content" jsonschema:"required,description=必填：要发送的文字内容，建议不超过2048字符"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token，需来自具备「管理所有客服会话」权限的应用"`
}

// SendWeComCustomerMessageResponse 发送企业微信客服消息的响应
type SendWeComCustomerMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
