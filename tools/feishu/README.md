# Feishu 工具使用说明

本文档说明 `DigEino` 中飞书通知能力的配置与使用方式。

## 1. 设计说明（对齐 OpenClaw 主路径）

- 入站接入默认：`ConnectionMode=websocket`
- 出站发送默认：`SendViaAPI=true`（通过飞书开放平台 HTTP API）
- 本实现不启用双通道并发发送，避免额外复杂度

## 2. 配置示例

在 `config/config.yaml` 中新增或修改：

```yaml
Feishu:
  Enabled: true
  ConnectionMode: "websocket"       # websocket | webhook
  EventIngest:
    WebsocketEnabled: true
    WebhookEnabled: false
    VerificationToken: ""
    WebhookPath: "/feishu/events"
    WebhookHost: "127.0.0.1"
    WebhookPort: 3000
  SendViaAPI: true
  NotifyOnEvents: ["failed", "completed"]  # 自动通知事件
  API:
    Enabled: true
    AppID: "cli_xxx"
    AppSecret: "xxx"
    BaseURL: "https://open.feishu.cn"
    Timeout: 5
    RetryCount: 2
    RetryDelayMs: 500
    ReceiveIDType: "chat_id"              # chat_id | open_id | user_id | email
    ReceiveIDs:
      - "oc_xxx"
```

## 3. 自动通知（状态事件）

当你使用 `webhook.NewConfiguredCollector(...)` 创建状态采集器时：

- 若 `Feishu.Enabled=true` 且 `SendViaAPI=true`
- 且 `API.AppID/AppSecret` 已配置
- 则会自动启用飞书通知 sink

消息触发事件由 `Feishu.NotifyOnEvents` 控制，默认推荐：

- `failed`
- `completed`

## 4. 手动工具调用

工具名：`send_feishu_message`

参数：

- `content`：消息正文（必填）
- `receive_id_type`：接收者类型（可选，默认使用配置）
- `receive_ids`：接收者 ID 列表（可选，默认使用配置 `API.ReceiveIDs`）

示例：

```json
{
  "content": "发布完成：v1.2.3",
  "receive_id_type": "chat_id",
  "receive_ids": ["oc_xxx"]
}
```

若不传 `receive_ids`，将回退到配置中的默认接收对象。

## 5. 常见问题

- `Feishu API config is invalid or disabled`
  - 检查 `Feishu.Enabled`、`SendViaAPI`、`API.Enabled` 是否为 `true`
  - 检查 `API.AppID` 与 `API.AppSecret` 是否填写

- `receive_ids is empty and no default receive_ids configured`
  - 在请求中传 `receive_ids`
  - 或在配置中设置 `Feishu.API.ReceiveIDs`

- 发送失败
  - 检查 `BaseURL`（中国区 `https://open.feishu.cn`，国际版可改 `https://open.larksuite.com`）
  - 检查应用权限与可见范围是否已发布生效

