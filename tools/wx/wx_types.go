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

// SendWeComCustomerImageRequest 发送企业微信客服图片消息的请求参数（发给个人微信用户）
type SendWeComCustomerImageRequest struct {
	OpenKfID    string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID  string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID（external_userid）"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：图片的 media_id，通过上传临时素材接口获得"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerImageResponse 发送企业微信客服图片消息的响应
type SendWeComCustomerImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComCustomerVoiceRequest 发送企业微信客服语音消息的请求参数（发给个人微信用户）
type SendWeComCustomerVoiceRequest struct {
	OpenKfID    string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID  string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：语音的 media_id，通过上传临时素材接口获得"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerVoiceResponse 发送企业微信客服语音消息的响应
type SendWeComCustomerVoiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComCustomerVideoRequest 发送企业微信客服视频消息的请求参数（发给个人微信用户）
type SendWeComCustomerVideoRequest struct {
	OpenKfID    string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID  string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：视频的 media_id，通过上传临时素材接口获得"`
	Title       string `json:"title" jsonschema:"description=可选：视频标题"`
	Description string `json:"description" jsonschema:"description=可选：视频描述"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerVideoResponse 发送企业微信客服视频消息的响应
type SendWeComCustomerVideoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComCustomerFileRequest 发送企业微信客服文件消息的请求参数（发给个人微信用户）
type SendWeComCustomerFileRequest struct {
	OpenKfID    string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID  string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID"`
	MediaID     string `json:"media_id" jsonschema:"required,description=必填：文件的 media_id，通过上传临时素材接口获得"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerFileResponse 发送企业微信客服文件消息的响应
type SendWeComCustomerFileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComCustomerLinkRequest 发送企业微信客服图文链接消息的请求参数（发给个人微信用户）
type SendWeComCustomerLinkRequest struct {
	OpenKfID      string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID    string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID"`
	Title         string `json:"title" jsonschema:"required,description=必填：链接标题"`
	Desc          string `json:"desc" jsonschema:"required,description=必填：链接描述"`
	URL           string `json:"url" jsonschema:"required,description=必填：点击跳转的 URL"`
	ThumbMediaID  string `json:"thumb_media_id" jsonschema:"required,description=必填：封面图 media_id，通过上传临时素材获得"`
	AccessToken   string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerLinkResponse 发送企业微信客服图文链接消息的响应
type SendWeComCustomerLinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendWeComCustomerMiniprogramRequest 发送企业微信客服小程序卡片的请求参数（发给个人微信用户）
type SendWeComCustomerMiniprogramRequest struct {
	OpenKfID       string `json:"open_kf_id" jsonschema:"required,description=必填：客服账号 ID"`
	CustomerID     string `json:"customer_id" jsonschema:"required,description=必填：外部联系人 ID"`
	Title          string `json:"title" jsonschema:"required,description=必填：小程序卡片标题"`
	AppID          string `json:"appid" jsonschema:"required,description=必填：小程序 appid"`
	PagePath       string `json:"pagepath" jsonschema:"description=可选：小程序页面路径"`
	ThumbMediaID   string `json:"thumb_media_id" jsonschema:"required,description=必填：封面图 media_id，通过上传临时素材获得"`
	AccessToken    string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComCustomerMiniprogramResponse 发送企业微信客服小程序卡片的响应
type SendWeComCustomerMiniprogramResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- 企业微信客服回调接收消息相关类型 ---

// WeComCallbackEvent 企业微信客服回调事件
type WeComCallbackEvent struct {
	ToUserName string `json:"ToUserName" xml:"ToUserName"` // 企业微信CorpID
	CreateTime int64  `json:"CreateTime" xml:"CreateTime"` // 消息创建时间
	MsgType    string `json:"MsgType" xml:"MsgType"`       // 固定为 "event"
	Event      string `json:"Event" xml:"Event"`           // 固定为 "kf_msg_or_event"
	Token      string `json:"Token" xml:"Token"`           // 用于拉取消息的token
	OpenKfId   string `json:"OpenKfId" xml:"OpenKfId"`     // 客服账号ID
}

// SyncMessageRequest 拉取消息请求
type SyncMessageRequest struct {
	Token  string `json:"token"`  // 回调事件中的Token
	Limit  int    `json:"limit"`  // 可选，默认1000
	Cursor string `json:"cursor"` // 可选，用于分页
}

// SyncMessageResponse 拉取消息响应
type SyncMessageResponse struct {
	ErrCode    int              `json:"errcode"`
	ErrMsg     string           `json:"errmsg"`
	NextCursor string           `json:"next_cursor"` // 下次拉取的游标
	HasMore    int              `json:"has_more"`    // 是否还有更多消息（0=没有，1=有）
	MsgList    []CustomerMessage `json:"msg_list"`   // 消息列表
}

// CustomerMessage 客服消息
type CustomerMessage struct {
	MsgID          string       `json:"msgid"`           // 消息ID
	OpenKfId       string       `json:"open_kfid"`      // 客服账号ID
	ExternalUserID string       `json:"external_userid"` // 外部联系人ID（客户）
	SendTime       int64        `json:"send_time"`       // 发送时间
	Origin         int          `json:"origin"`         // 消息来源（3=客户发送）
	ServicerUserID string       `json:"servicer_userid,omitempty"` // 客服userid（可选）
	MsgType        string       `json:"msgtype"`        // 消息类型：text, image, voice, video, file, location, link, business_card, miniprogram, msgmenu, event
	Text           *TextMessage `json:"text,omitempty"`
	Image          *MediaMessage `json:"image,omitempty"`
	Voice          *MediaMessage `json:"voice,omitempty"`
	Video          *MediaMessage `json:"video,omitempty"`
	File           *MediaMessage `json:"file,omitempty"`
	Location       *LocationMessage `json:"location,omitempty"`
	Link           *LinkMessage `json:"link,omitempty"`
	BusinessCard   *BusinessCardMessage `json:"business_card,omitempty"`
	Miniprogram    *MiniprogramMessage `json:"miniprogram,omitempty"`
	MsgMenu        *MsgMenuMessage `json:"msgmenu,omitempty"`
	Event          *EventMessage `json:"event,omitempty"`
}

// TextMessage 文本消息
type TextMessage struct {
	Content string `json:"content"` // 消息内容
	MenuID  string `json:"menu_id,omitempty"` // 菜单ID（可选）
}

// MediaMessage 媒体消息（图片、语音、视频、文件）
type MediaMessage struct {
	MediaID string `json:"media_id"` // 媒体ID
}

// LocationMessage 位置消息
type LocationMessage struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	Name      string  `json:"name"`      // 位置名称
	Address   string  `json:"address"`   // 地址
}

// LinkMessage 链接消息
type LinkMessage struct {
	Title       string `json:"title"`        // 标题
	Desc        string `json:"desc"`         // 描述
	URL         string `json:"url"`          // 链接
	ThumbMediaID string `json:"thumb_media_id"` // 封面图media_id
}

// BusinessCardMessage 名片消息
type BusinessCardMessage struct {
	UserID string `json:"userid"` // 用户ID
}

// MiniprogramMessage 小程序消息
type MiniprogramMessage struct {
	Title        string `json:"title"`         // 标题
	AppID        string `json:"appid"`         // 小程序appid
	PagePath     string `json:"pagepath"`      // 页面路径
	ThumbMediaID string `json:"thumb_media_id"` // 封面图media_id
}

// MsgMenuMessage 菜单消息
type MsgMenuMessage struct {
	HeadContent string        `json:"head_content"` // 头部内容
	List        []MsgMenuItem `json:"list"`         // 菜单项列表
	TailContent string        `json:"tail_content"` // 尾部内容
}

// MsgMenuItem 菜单项
type MsgMenuItem struct {
	Type  string `json:"type"`  // 类型：click, view, miniprogram
	Click *struct {
		ID      string `json:"id"`      // 菜单ID
		Content string `json:"content"` // 菜单内容
	} `json:"click,omitempty"`
	View *struct {
		URL     string `json:"url"`     // 跳转URL
		Content string `json:"content"` // 菜单内容
	} `json:"view,omitempty"`
	Miniprogram *struct {
		AppID    string `json:"appid"`    // 小程序appid
		PagePath string `json:"pagepath"` // 页面路径
		Content  string `json:"content"`  // 菜单内容
	} `json:"miniprogram,omitempty"`
}

// EventMessage 事件消息
type EventMessage struct {
	EventType string `json:"event_type"` // 事件类型
	// 根据事件类型不同，可能有不同的字段
}

// ReceiveWeComCustomerMessageRequest 接收消息工具请求（供外部系统调用）
type ReceiveWeComCustomerMessageRequest struct {
	OpenKfID      string `json:"open_kf_id" jsonschema:"description=可选：客服账号ID，不传则拉取所有"`
	ExternalUserID string `json:"external_userid" jsonschema:"description=可选：外部联系人ID，不传则拉取所有"`
	Limit         int    `json:"limit" jsonschema:"description=可选：拉取数量，默认1000（仅拉取模式有效）"`
	Cursor        string `json:"cursor" jsonschema:"description=可选：游标，用于分页（仅拉取模式有效）"`
	Mode          string `json:"mode" jsonschema:"description=可选：realtime（实时回调，默认）或 pull（主动拉取）"`
}

// ReceiveWeComCustomerMessageResponse 接收消息工具响应
type ReceiveWeComCustomerMessageResponse struct {
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
	Messages   []CustomerMessage `json:"messages"`   // 接收到的消息列表
	NextCursor string           `json:"next_cursor"` // 下次拉取的游标
	HasMore    bool             `json:"has_more"`    // 是否还有更多消息
}
