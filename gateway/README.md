# DigEino Gateway — 宿主对接指南

DigEino 作为 **通用 Agent 插件运行时**，通过统一协议暴露工具能力。宿主项目（Knowledge、DigFlow、自定义 Agent）只需实现协议客户端，无需 fork DigEino 业务代码。

- **Wrap → Expose 新工具**（DigEino 开发者）：见 [USAGE.md](./USAGE.md)

## 连接模式

| 模式 | 命令 / 包 | 适用场景 |
|------|-----------|----------|
| HTTP Gateway | `digeino gateway` | 内网服务、云端调本地网关 |
| WebSocket Collector | `digeino collector` | 本机无公网 IP，反向连接云端 |
| MCP (stdio) | `digeino mcp` | Cursor / Claude Desktop / IDE |
| stdio JSON | `digeino stdio` | CLI、桌面应用子进程 |
| Go Client SDK | `gateway/client` | 宿主 Go 项目 import 调用 |

## 核心类型（`gateway/protocol`）

- `ToolManifest` — 工具发现
- `ToolCall` — 发起调用
- `ToolResult` — 执行结果
- `Artifact` — 大对象引用（截图等）

## HTTP API

基址示例：`http://127.0.0.1:8787`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/manifest` | 工具清单 |
| POST | `/tools/call` | 执行工具 |
| GET | `/artifacts/{id}` | 下载 Artifact |

鉴权（可选）：`Authorization: Bearer <token>` 或 `X-Digeino-Token`

### Go 宿主 SDK

```go
import gwclient "github.com/originaleric/digeino/gateway/client"

c := gwclient.New("http://127.0.0.1:8787", "your-token")
manifest, _ := c.Manifest(ctx)
result, _ := c.Call(ctx, protocol.ToolCall{
    ID: "call_1", Tool: "wechat.article.read",
    Input: json.RawMessage(`{"url":"https://mp.weixin.qq.com/s/..."}`),
})
```

## WebSocket Collector 线协议

路径默认：`/digeino/v1/collector/ws`

详见 [落地与使用说明](../docs/updates/2026-05-19_Agent插件运行时落地与使用说明.md) 第二节。

消息类型：`collector_hello`、`collector_hello_ack`、`collector_manifest`、`instance_status`、`pull_tasks`、`pull_tasks_ack`、`tool_call`、`tool_result`、`ping`/`pong`。

本地联调可用 `digeino dev-host`（仅开发参考，非生产宿主）。

## stdio JSON 协议

每行一个 JSON 对象：

```json
{"type":"get_manifest"}
```

```json
{"type":"tool_call","id":"c1","tool":"browser.browse","input":{"url":"https://example.com"}}
```

响应为单行 `ToolManifest` 或 `ToolResult`。

## MCP

在 Cursor / Claude Desktop 中配置：

```json
{
  "mcpServers": {
    "digeino": {
      "command": "digeino",
      "args": ["mcp", "--config", "/path/to/config.yaml"]
    }
  }
}
```

或 `go run ./cmd/digeino mcp`。

## 已暴露工具（网关名）

| 工具 | 说明 |
|------|------|
| `browser.browse` | 本地浏览器抓取 |
| `browser.snapshot` | 页面可交互元素快照 |
| `browser.action` | 点击/输入等操作 |
| `wechat.article.read` | 公众号文章采集 |
| `file.read` | 本地文件读取（需 `Gateway.AllowedReadPaths`） |

## 安全

- 域名白名单：`Tools.LocalBrowser.AllowedDomains` + `ToolCall.policy.allowed_domains`
- 工具白名单：`Gateway.AllowedTools` / `Collector.AllowedTools`
- 文件路径白名单：`Gateway.AllowedReadPaths`
- Cookie 仅存本地 Collector / 浏览器配置目录

## 实现宿主时的建议顺序

1. 先对接 HTTP `GET /manifest` + `POST /tools/call`（或用 `gateway/client`）
2. 需要本机穿透时实现 WebSocket 宿主端（参考 `devhost` 包）
3. IDE 场景配置 MCP

DigEino 仓库内 **不包含** Knowledge/DigFlow 业务代码；宿主在各自仓库引用本协议即可。
