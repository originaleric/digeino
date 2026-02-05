# 更新：微信消息类型增强

**日期**: 2026-02-05
**类别**: 功能 / 工具

## 概述
增强了微信推送通知工具，支持除文本消息之外的多种消息类型。该工具现在支持通过公众号 API 向微信用户发送图片消息、小程序页面卡片和图文链接消息。这为 AI 智能体提供了更灵活和丰富的消息发送能力。

## 新功能

### 1. 图片消息支持
- **发送图片消息**：使用从微信素材 API 获取的 media_id 发送图片消息
- **方法**：`SendWeChatImageMessage(ctx, req)`
- **请求**：`SendWeChatImageMessageRequest`，包含 `MediaID` 字段

### 2. 小程序页面卡片支持
- **发送小程序卡片**：发送小程序页面卡片，允许用户直接跳转到小程序页面
- **方法**：`SendWeChatMiniProgramPage(ctx, req)`
- **请求**：`SendWeChatMiniProgramPageRequest`，包含 `Title`、`AppID`、`PagePath` 和 `ThumbMediaID` 字段
- **配置支持**：小程序设置可在 `config.yaml` 中配置，并可在每次请求时覆盖

### 3. 图文链接消息支持
- **发送图文链接消息**：发送包含标题、描述、封面图片和 URL 的图文链接消息
- **方法**：`SendWeChatLinkMessage(ctx, req)`
- **请求**：`SendWeChatLinkMessageRequest`，包含 `Articles` 数组（每条消息限制为 1 篇文章）

## 修改的组件

### 更新的文件
- `tools/wx/wx_types.go`：为图片、小程序和链接消息添加了新的请求/响应类型
- `tools/wx/wx_message.go`：
  - 添加了 `SendWeChatImageMessage()` 函数
  - 添加了 `SendWeChatMiniProgramPage()` 函数
  - 添加了 `SendWeChatLinkMessage()` 函数
  - 将通用消息发送逻辑重构为 `sendMessageRequest()` 辅助函数
  - 重构了 `sendTextMessageToUser()` 以使用通用辅助函数
- `config/config.go`：在 `WeChatConfig` 中添加了 `MiniProgramConfig` 结构，用于小程序设置

### 新增的类型
- `SendWeChatImageMessageRequest` / `SendWeChatImageMessageResponse`
- `SendWeChatMiniProgramPageRequest` / `SendWeChatMiniProgramPageResponse`
- `SendWeChatLinkMessageRequest` / `SendWeChatLinkMessageResponse`
- `LinkMessageArticle`：链接消息文章的结构
- `MiniProgramConfig`：小程序设置的配置结构

## 配置

配置已扩展以支持小程序设置：

```yaml
WeChat:
  Enabled: true
  AppID: "your_appid"              # 微信服务号 AppID
  AppSecret: "your_app_secret"     # 微信服务号 AppSecret
  OpenIDs:                         # 默认接收消息的用户 openid 列表
    - "openid1"
    - "openid2"
  TokenFilePath: "storage/app/wechat/access_token.json"
  MiniProgram:                     # 小程序配置（用于小程序卡片消息）
    AppID: "your_miniprogram_appid"        # 小程序 AppID
    DefaultPath: "pages/index/index"      # 默认页面路径
    ThumbMediaID: "your_thumb_media_id"    # 封面图片 media_id
```

## 使用示例

### 发送图片消息

```go
import "github.com/originaleric/digeino/tools/wx"

req := wx.SendWeChatImageMessageRequest{
    OpenID:  "user_openid",  // 可选：如果为空，则发送给配置中的所有用户
    MediaID: "media_id_from_wechat_api",
}

response, err := wx.SendWeChatImageMessage(ctx, req)
if err != nil {
    // 处理错误
}
```

### 发送小程序页面卡片

```go
req := wx.SendWeChatMiniProgramPageRequest{
    OpenID:       "user_openid",  // 可选
    Title:       "小程序标题",
    AppID:       "miniprogram_appid",  // 可选：如果为空则使用配置中的值
    PagePath:    "pages/detail/index?id=123",  // 可选：如果为空则使用配置中的默认路径
    ThumbMediaID: "thumb_media_id",  // 可选：如果为空则使用配置中的值
}

response, err := wx.SendWeChatMiniProgramPage(ctx, req)
if err != nil {
    // 处理错误
}
```

### 发送图文链接消息

```go
req := wx.SendWeChatLinkMessageRequest{
    OpenID: "user_openid",  // 可选
    Articles: []wx.LinkMessageArticle{
        {
            Title:       "文章标题",
            Description: "文章描述",
            PicURL:      "https://example.com/cover.jpg",
            URL:         "https://example.com/article",
        },
    },
}

response, err := wx.SendWeChatLinkMessage(ctx, req)
if err != nil {
    // 处理错误
}
```

## 技术细节

### 代码重构
- 将通用 HTTP 请求逻辑提取到 `sendMessageRequest()` 函数中
- 所有消息类型现在使用相同的错误处理机制
- 所有消息类型保持一致的响应格式

### 配置优先级
对于小程序消息，配置值可以通过三种方式提供（按优先级排序）：
1. 请求参数（最高优先级）
2. 配置文件（`config.yaml`）
3. 如果未提供则报错（最低优先级）

### 消息类型支持
根据微信 API 文档，支持以下消息类型：
- `text`：文本消息（已有）
- `image`：图片消息（新增）
- `news`：图文链接消息（新增，msgtype="news"）
- `miniprogrampage`：小程序页面卡片（新增）

## API 参考

所有新函数都遵循与现有 `SendWeChatTextMessage` 相同的模式：
- 支持单个用户（通过 `OpenID`）或批量发送（通过配置）
- 详细的错误报告，包含 `FailedTo` 列表
- 每个接收者的成功/失败跟踪
- 对常见微信 API 错误的一致错误处理

## 影响

此增强功能显著扩展了微信工具的消息发送能力，实现了：
- **富媒体**：向用户发送图片和视觉内容
- **深度链接**：直接将用户引导到特定的小程序页面
- **内容分享**：通过丰富的预览分享文章和链接
- **更好的用户体验**：更具吸引力和交互性的消息选项

该工具保持向后兼容性 - 现有的文本消息功能保持不变。所有新功能都遵循与原始实现相同的设计模式和错误处理机制。
