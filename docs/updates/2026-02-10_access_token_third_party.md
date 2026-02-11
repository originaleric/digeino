# 更新：access_token 支持第三方传入

**日期**: 2026-02-10
**类别**: 功能增强 / 工具

## 概述

公众号（OA）和企业微信（WeCom）的发送工具现已支持通过请求参数传入 `access_token`。当调用方已有自己的 token 管理（如 chatgpt-wechat、DigFlow 等）时，可直接传入 token，无需 DigEino 内部获取，便于统一 token 策略、避免多端重复获取。

## 新功能

### 请求参数新增 `access_token`

所有发送接口的请求体中均新增可选字段 `access_token`：

- **公众号**：`SendWeChatTextMessageRequest`、`SendWeChatImageMessageRequest`、`SendWeChatMiniProgramPageRequest`、`SendWeChatLinkMessageRequest`
- **企业微信**：`SendWeComMessageRequest`、`SendWeComImageMessageRequest`、`SendWeComTextCardRequest`

**优先级**：若请求中提供了 `access_token`，则优先使用；否则仍由 DigEino 内部自动获取（公众号从缓存/API，企业微信由 go-workwx 管理）。

## 修改的组件

### 更新的文件
- `tools/wx/wx_types.go`：各 Request 类型新增 `AccessToken string` 字段
- `tools/wx/wx_message.go`：公众号发送逻辑中，优先使用 `req.AccessToken`
- `tools/wx/wecom_message.go`：企业微信发送逻辑中，若传入 token 则直接调 API，否则使用 go-workwx

### 新增类型
- `WeComMessageAPIResponse`：企业微信发送消息 API 的响应结构

## 使用示例

### 公众号

```go
req := wx.SendWeChatTextMessageRequest{
    OpenID:      "user_openid",
    Content:     "消息内容",
    AccessToken: "第三方获取的 token",  // 可选
}
resp, err := wx.SendWeChatTextMessage(ctx, req)
```

### 企业微信

```go
req := wx.SendWeComMessageRequest{
    UserID:      "ZhangSan",
    Content:     "消息内容",
    AccessToken: "第三方获取的 token",  // 可选
}
resp, err := wx.SendWeComMessage(ctx, req)
```

### Agent 工具调用（JSON）

```json
{
  "user_id": "ZhangSan",
  "content": "任务已完成",
  "access_token": "xxx"
}
```

## 技术说明

### 企业微信外部 token 路径

当 `access_token` 由第三方传入时，企业微信不再通过 go-workwx 的 `WithApp` 获取 token，而是直接请求企业微信 API：

- 接口：`POST {QYAPIHost}/cgi-bin/message/send?access_token={token}`
- 仍需提供 `agent_id`（可从请求或配置中获取），用于指定发送应用

### 向后兼容

未传入 `access_token` 时，行为与之前一致，内部自动获取并管理 token，不影响现有用法。
