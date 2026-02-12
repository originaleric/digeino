# 更新：企业微信推送工具

**日期**: 2026-02-10
**类别**: 功能 / 工具

## 概述

新增企业微信（WeCom）推送工具，使 DigEino 作为 DigFlow Agent 插件时，能够将 Agent 运行结果通过企业微信应用消息发送给用户。企业微信通过中转到个人微信，相比公众号具有无封号风险、无需 48 小时内互动等优势。

本更新新增三个 Eino 工具：文字消息、图片消息、文本卡片消息。Token 由 `go-workwx` 库内部管理，无需额外配置。

## 新功能

### 1. 文字消息 (`send_wecom_message`)
- 发送纯文本消息给企业成员
- 超过 850 字自动分条发送（企业微信单条建议上限）
- 支持指定 `agent_id` 或使用配置中第一个应用

### 2. 图片消息 (`send_wecom_image`)
- 使用临时素材的 `media_id` 发送图片
- 需先通过企业微信临时素材上传接口获取 `media_id`

### 3. 文本卡片消息 (`send_wecom_text_card`)
- 发送标题 + 描述 + 链接的卡片，适合通知类场景
- 用户点击可跳转指定 URL

### 4. 客服消息 (`send_wecom_customer_message`) — 发给个人微信用户
- 使用企业微信「客户联系」的客服能力，将消息发给**个人微信**用户
- 需提供 `open_kf_id`（客服账号 ID）和 `customer_id`（外部联系人 ID）
- 用户需先通过扫码/链接添加企业为客服后才能收到消息
- 支持第三方传入 `access_token`（需来自具备「管理所有客服会话」权限的应用）
- 或配置 `ManageAllKFSession: true` 的应用，由 DigEino 内部获取 token

## 修改的组件

### 新增文件
- `tools/wx/wecom_message.go`：企业微信消息发送逻辑（文字、图片、文本卡片）
- `tools/wx/wecom_tool.go`：三个 Eino 工具的工厂函数

### 更新的文件
- `config/config.go`：新增 `WeComConfig`、`WeComApplication` 结构体
- `config/config.yaml`：新增 `WeCom` 配置块
- `tools/wx/wx_types.go`：新增企业微信相关请求/响应类型
- `tools/tools.go`：在 `BaseTools` 中注册三个企业微信工具

### 依赖
- `github.com/whyiyhw/go-workwx`：企业微信 Go SDK

## 配置

在 `config/config.yaml` 中新增 WeCom 配置：

```yaml
WeCom:
  Enabled: true                          # 是否启用企业微信推送
  CorpID: "wwxxxxxxxx"                   # 企业 ID（我的企业 > 企业信息）
  QYAPIHost: "https://qyapi.weixin.qq.com"  # 可选，支持自定义域名
  TokenFilePath: "storage/app/wecom/access_token.json"
  Applications:
    - AgentID: 1000002                   # 应用 ID（应用管理 > 自建应用）
      AgentSecret: "xxx"                 # 应用 Secret
    - AgentID: 1000003                   # 客服消息专用应用（需开通客户联系）
      AgentSecret: "yyy"
      ManageAllKFSession: true           # 管理所有客服会话，用于 send_wecom_customer_message
```

## 工具参数说明

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `send_wecom_message` | `user_id` (必填) | 企业成员 userID |
| | `content` (必填) | 文字内容 |
| | `agent_id` (可选) | 应用 ID，不传则用配置中第一个 |
| `send_wecom_image` | `user_id` (必填) | 企业成员 userID |
| | `media_id` (必填) | 图片 media_id（临时素材） |
| | `agent_id` (可选) | 应用 ID |
| `send_wecom_text_card` | `user_id` (必填) | 企业成员 userID |
| | `title` (必填) | 卡片标题 |
| | `description` (必填) | 卡片描述 |
| | `url` (必填) | 点击跳转链接 |
| | `agent_id` (可选) | 应用 ID |
| `send_wecom_customer_message` | `open_kf_id` (必填) | 客服账号 ID |
| | `customer_id` (必填) | 外部联系人 ID（external_userid） |
| | `content` (必填) | 文字内容 |
| | `access_token` (可选) | 第三方传入的 access_token |

**user_id 来源**：由 DigFlow Agent 在调用时从会话上下文（如企业微信消息的 FromUserID）解析并传入。

**客服消息前置**：用户需先通过企业微信客服链接/扫码添加企业为客服，系统会为每个用户分配 `external_userid`（即 `customer_id`）。`open_kf_id` 在企业微信管理后台创建客服账号后获得。

---

## 使用指南

### 1. 前置条件

- 已注册企业微信（[work.weixin.qq.com](https://work.weixin.qq.com/)）
- 创建自建应用并获取 AgentID、AgentSecret
- 在 DigFlow 项目中已引入 DigEino 依赖

### 2. 配置

在 DigEino 或 DigFlow 的 `config.yaml` 中启用并填写企业微信配置（见上方配置示例）。

### 3. 在 DigFlow 中注册工具

DigEino 的 `BaseTools()` 会自动注册企业微信工具（当 `WeCom.Enabled: true` 时）。在 DigFlow 的 `tool_register.go` 中引入 DigEino 工具：

```go
import (
    "github.com/originaleric/digeino/tools"
)

func init() {
    ctx := context.Background()
    baseTools, _ := tools.BaseTools(ctx)
    for _, t := range baseTools {
        variable.ToolRegistry.Register(t.Info(ctx).Name, t)
    }
}
```

或在 Agent 配置中显式绑定：

```yaml
Tools:
  - Name: "send_wecom_message"
    Description: "通过企业微信发送文字消息"
  - Name: "send_wecom_image"
    Description: "通过企业微信发送图片消息"
  - Name: "send_wecom_text_card"
    Description: "通过企业微信发送文本卡片消息"
```

### 4. 在 Agent 中调用

Agent 在执行过程中可调用上述工具，将结果发送给企业微信用户。`user_id` 需由调用方（如 chatgpt-wechat、DigFlow 的企业微信接入层）从收到的企业微信消息中解析 `FromUserID` 并传入工具参数。

**示例（Agent 内部逻辑）**：
当用户通过企业微信发送消息时，会话上下文中应包含 `user_id`。Agent 完成处理后调用：

```json
{
  "tool": "send_wecom_message",
  "arguments": {
    "user_id": "ZhangSan",
    "content": "您的任务已处理完成，结果如下：..."
  }
}
```

### 5. 图片消息的 media_id 获取

图片消息需要先上传临时素材获取 `media_id`。可通过企业微信 API 或 `go-workwx` 的 `UploadTempImageMedia` 实现。若需在 DigEino 中增加「通过 URL 或本地路径上传并发送图片」的能力，可在此基础上扩展。

### 6. 与 chatgpt-wechat 的配合

若使用 [chatgpt-wechat](https://github.com/whyiyhw/chatgpt-wechat) 作为企业微信消息网关，可将 DigEino 工具注册到 DigFlow Agent，由 chatgpt-wechat 将用户消息转发给 Agent，Agent 返回后通过 `send_wecom_message` 等工具将结果发回用户。

### 7. 注意事项

- 企业可信 IP：若服务器在国内，需在企业微信管理后台配置「企业可信 IP」
- 多应用：`Applications` 可配置多个应用，通过 `agent_id` 指定使用哪个应用发送
- 长消息：文字超过 850 字会自动分多条发送，无需手动拆分
