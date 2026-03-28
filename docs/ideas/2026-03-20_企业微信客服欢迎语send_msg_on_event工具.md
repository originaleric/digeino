# DigEino 新增 send_msg_on_event 工具方案

## 一、背景

企业微信客服支持「进入会话事件」场景，当用户首次进入客服会话（或 48 小时内未收过欢迎语且未发过消息）时，sync_msg 拉取的事件会包含 `welcome_code`。开发者需在收到事件后 **20 秒内** 调用「发送客服欢迎语」接口（`send_msg_on_event`），用该 code 向用户发送欢迎消息。

DigEino 已有 `send_wecom_customer_message`（使用 `send_msg` API），但该接口要求用户**先发过消息**，48 小时内才能回复。进入会话欢迎语场景下用户尚未发消息，必须使用 `send_msg_on_event`。

## 二、API 说明

**接口地址**：`POST https://qyapi.weixin.qq.com/cgi-bin/kf/send_msg_on_event?access_token=ACCESS_TOKEN`

**请求体**（文本消息）：

```json
{
  "code": "CODE",
  "msgid": "MSG_ID",
  "msgtype": "text",
  "text": {
    "content": "欢迎咨询"
  }
}
```

**参数说明**：

| 参数 | 必填 | 说明 |
|------|------|------|
| code | 是 | 事件响应消息对应的 code（即 welcome_code），通过 sync_msg 事件下发，仅可使用一次 |
| msgid | 否 | 消息 ID，不传则系统自动生成 |
| msgtype | 是 | 消息类型（text、msgmenu） |
| text | 是* | msgtype 为 text 时必填 |
| msgmenu | 是* | msgtype 为 msgmenu 时必填 |

**权限**：仅允许微信客服 Secret 所获取的 access_token 调用（与 ManageAllKFSession 应用相同）。

**限制**：收到相关事件后 20 秒内调用，且只可调用一次。

## 三、与现有 send_msg 的区别

| 维度 | send_msg（send_wecom_customer_message） | send_msg_on_event（待实现） |
|------|------------------------------------------|-----------------------------|
| API | cgi-bin/kf/send_msg | cgi-bin/kf/send_msg_on_event |
| 参数 | open_kf_id + customer_id + content | code + content |
| 适用场景 | 用户发消息后 48 小时内回复 | 用户进入会话时发送欢迎语 |
| 限制 | 48 小时 + 5 条 | 20 秒内、仅一次 |

## 四、实现方案

### 4.1 新增类型（tools/wx/wx_types.go）

```go
// SendWeComMsgOnEventRequest 发送客服欢迎语（事件响应消息）的请求参数
// 用于「进入会话」等事件，code 来自 sync_msg 的 event.welcome_code 字段
type SendWeComMsgOnEventRequest struct {
	Code        string `json:"code" jsonschema:"required,description=必填：welcome_code，来自 sync_msg 事件，仅可使用一次"`
	MsgID       string `json:"msgid" jsonschema:"description=可选：消息 ID，不传则系统自动生成"`
	Content     string `json:"content" jsonschema:"required,description=必填：文本欢迎语内容"`
	AccessToken string `json:"access_token" jsonschema:"description=可选：第三方传入的 access_token"`
}

// SendWeComMsgOnEventResponse 发送客服欢迎语的响应
type SendWeComMsgOnEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	MsgID   string `json:"msgid,omitempty"`
}
```

### 4.2 新增 API 调用（tools/wx/wecom_message.go）

```go
// sendWeComMsgOnEventAPI 调用 send_msg_on_event 接口
func sendWeComMsgOnEventAPI(ctx context.Context, accessToken string, code string, msgID string, msgtype string, body map[string]interface{}) error
```

- URL：`POST {host}/cgi-bin/kf/send_msg_on_event?access_token={token}`
- body 需包含：code、msgid（可选）、msgtype、text/msgmenu

### 4.3 新增发送函数（tools/wx/wecom_message.go）

```go
// SendWeComMsgOnEvent 发送客服欢迎语（事件响应消息）
// 见 https://kf.weixin.qq.com/api/doc/path/95123
func SendWeComMsgOnEvent(ctx context.Context, req SendWeComMsgOnEventRequest) (SendWeComMsgOnEventResponse, error)
```

- 校验 code、content 非空
- 若 req.AccessToken 非空，直接调用 API；否则使用 getWeComCustomerAccessToken
- 构造 body：`{"code": req.Code, "msgtype": "text", "text": {"content": req.Content}}`
- 支持超长文本分条（与 send_wecom_customer_message 类似，但 code 仅一次有效，需注意：欢迎语通常不建议分条，可限制单条 2048 字符）

### 4.4 新增工具（tools/wx/wecom_tool.go）

```go
// NewWeComMsgOnEventTool 创建发送客服欢迎语工具
func NewWeComMsgOnEventTool(ctx context.Context) (tool.BaseTool, error)
```

- 工具名：`send_wecom_msg_on_event`
- 描述：通过企业微信客服发送欢迎语（事件响应消息）。用于用户进入会话时，使用 sync_msg 事件中的 welcome_code 发送欢迎语。code 仅可使用一次，收到事件后 20 秒内有效。

### 4.5 注册工具（tools/tools.go）

在 `BaseTools` 中于 `wecomCustomerTool` 之后注册 `NewWeComMsgOnEventTool`。

## 五、使用场景

1. **DigUserService**：收到 kf_msg_or_event → 调用 sync_msg → 解析到 welcome_code → 可调用 Knowledge 或直接调用 DigEino 工具发送欢迎语
2. **Knowledge**：若 DigUserService 将 welcome_code 事件转发给 Knowledge，Knowledge 可通过 DigEino 的 `send_wecom_msg_on_event` 工具发送欢迎语（需先解决 kf_msg_or_event 转发到 Agent 的架构问题）

## 六、涉及文件

| 文件 | 修改内容 |
|------|----------|
| tools/wx/wx_types.go | 新增 SendWeComMsgOnEventRequest 和 SendWeComMsgOnEventResponse |
| tools/wx/wecom_message.go | 新增 sendWeComMsgOnEventAPI、SendWeComMsgOnEvent |
| tools/wx/wecom_tool.go | 新增 NewWeComMsgOnEventTool |
| tools/tools.go | 在 BaseTools 中注册新工具 |

## 七、注意事项

1. **时效**：welcome_code 收到后 20 秒内有效，调用方需尽快完成调用
2. **一次性**：code 仅可使用一次，调用失败后不可重试
3. **Token**：必须使用 ManageAllKFSession 应用的 access_token（与 send_wecom_customer_message 相同）
4. **msgmenu**：后续可扩展支持菜单消息，需增加 msgmenu 结构体及请求参数
