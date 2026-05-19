# DigEino 通用 Agent 插件运行时与本地采集网关方案

## 1. 背景

DigEino 当前已经沉淀了浏览器、搜索、文件、企业微信、UI/UX 等工具能力，并且已经能够作为 DigFlow/Knowledge 的内部工具库使用。随着 Agent 宿主形态增多，DigEino 不应只作为 DigFlow 的附属工具包，而应升级为一个可被任意 Agent 调用的通用插件运行时。

目标是让 DigEino 成为一种独立的 **Agent Plugin Runtime / Local Tool Gateway**：

- 任意宿主 Agent 可以发现 DigEino 提供的工具。
- 任意宿主 Agent 可以用统一协议调用 DigEino 工具。
- DigEino 可以在本地、私有网络或云端执行工具。
- 敏感能力如浏览器 Cookie、本地文件、企业内网访问可以留在本地或私有节点。
- DigFlow、Knowledge、第三方 Agent 框架都只是 DigEino 的宿主之一，而不是协议中心。

## 2. 典型场景

### 2.1 公众号文章采集

用户在 Knowledge 或其他 Agent 应用里表示对某个公众号或某篇文章感兴趣。云端 Agent 不直接用云服务器批量抓取，而是把采集任务派发给本地 DigEino Collector。

本地 Collector 调用 `browser.browse` 或 `wechat.article.read`，用本机 Chromium 打开 `mp.weixin.qq.com` 页面，提取标题、正文、Markdown、原始链接和必要元数据，然后回传给宿主 Agent。

### 2.2 企业内网工具调用

企业用户可以在内网部署 DigEino Collector，暴露 CRM、ERP、Wiki、数据库查询等私有工具。云端 Agent 只下发工具调用请求，实际访问发生在企业内网。

### 2.3 用户本地文件和浏览器能力

桌面 Agent、IDE Agent 或云端 Agent 可以通过 DigEino 调用用户本机工具，例如读取本地文件、打开浏览器、截图、访问登录态页面。用户凭证不需要上传到云端。

## 3. 设计原则

1. **宿主无关**
   协议不绑定 DigFlow、Knowledge 或任何具体 Agent 框架。DigFlow 只需要实现一个 Adapter。

2. **工具通用**
   工具以 manifest、input schema、output schema 的形式公开，任意 Agent 框架都可以适配。

3. **本地优先**
   浏览器 Cookie、本地文件、企业内网凭证默认保留在本地 Collector，不上传到云端。

4. **策略内置**
   DigEino 运行时必须内置权限、域名白名单、工具白名单、限流、审计和脱敏。

5. **多连接形态**
   同一套工具协议应支持 HTTP、WebSocket、stdio 和 MCP Server 等连接方式。

6. **结果结构化**
   工具返回通用 `ToolResult`，包含结构化 `output`、可选 `artifacts`、错误信息和执行用量。

## 4. 总体架构

```text
Any Host Agent
  - DigFlow
  - Knowledge
  - LangChain
  - Dify
  - Coze
  - OpenAI Agents
  - Custom Agent
        |
        | Tool Manifest / Tool Call / Tool Result
        v
DigEino Gateway Protocol
        |
        v
DigEino Runtime
  - Tool Registry
  - Tool Executor
  - Permission Policy
  - Artifact Store
  - Audit Log
        |
        v
DigEino Tools
  - browser.browse
  - browser.action
  - browser.snapshot
  - wechat.article.read
  - file.read
  - internal.api.call
```

## 5. 连接模式

### 5.1 MCP Server 模式

DigEino 暴露为 MCP Server。支持 MCP 的宿主可以直接发现和调用工具。

适用场景：

- IDE Agent
- 桌面 Agent
- 支持 MCP 的第三方 Agent 平台

优点：

- 生态兼容性最好。
- 工具发现和调用模型成熟。
- 不强依赖 DigFlow。

### 5.2 HTTP Tool Gateway 模式

DigEino 作为 HTTP Server，暴露 `/manifest`、`/tools/call`、`/tools/result` 等接口。

适用场景：

- 同一内网里的宿主调用本地 DigEino。
- 云端宿主调用私有网络里的 DigEino 服务。
- 简单服务化部署。

优点：

- 实现简单。
- 易于调试。
- 方便接入非 MCP 宿主。

### 5.3 WebSocket Reverse Connector 模式

本地 DigEino Collector 主动连接云端宿主，保持长连接。云端通过该连接下发工具调用任务，本地执行后回传结果。

适用场景：

- 用户本机没有公网 IP。
- 不希望本地开放端口。
- 云端 Knowledge 或其他 SaaS 希望调用用户本地浏览器能力。

优点：

- 穿透友好。
- 更适合个人电脑和企业内网。
- Cookie、文件、内网访问能力可以留在本地。

### 5.4 stdio 模式

DigEino 作为本地子进程，由宿主通过 stdin/stdout 调用。

适用场景：

- CLI Agent
- 桌面应用
- 轻量本地集成

## 6. 通用数据结构

### 6.1 ToolManifest

工具清单用于宿主发现 DigEino 当前可用能力。

```json
{
  "type": "tool_manifest",
  "runtime": "digeino",
  "runtime_version": "1.0",
  "instance_id": "collector_local_mac_001",
  "tools": [
    {
      "name": "browser.browse",
      "description": "Open a URL in a local browser and extract rendered content.",
      "input_schema": {
        "type": "object",
        "required": ["url"],
        "properties": {
          "url": { "type": "string" },
          "wait_selector": { "type": "string" },
          "content_selector": { "type": "string" },
          "use_cookie_domain": { "type": "string" }
        }
      },
      "output_schema": {
        "type": "object",
        "properties": {
          "title": { "type": "string" },
          "text": { "type": "string" },
          "markdown": { "type": "string" },
          "source_url": { "type": "string" }
        }
      },
      "capabilities": ["browser", "web.read", "cookie.local"],
      "risk": "network",
      "requires_user_approval": false
    }
  ]
}
```

### 6.2 ToolCall

宿主向 DigEino 发起工具调用。

```json
{
  "type": "tool_call",
  "id": "call_123",
  "tool": "browser.browse",
  "input": {
    "url": "https://mp.weixin.qq.com/s/example",
    "wait_selector": "#js_article",
    "content_selector": "#js_article",
    "use_cookie_domain": "mp.weixin.qq.com"
  },
  "context": {
    "user_id": "u_1",
    "tenant_id": "t_1",
    "trace_id": "trace_1",
    "host": "knowledge"
  },
  "policy": {
    "timeout_ms": 60000,
    "allowed_domains": ["mp.weixin.qq.com"],
    "store_cookies": "local_only",
    "max_output_bytes": 2000000
  }
}
```

### 6.3 ToolResult

DigEino 返回工具执行结果。

```json
{
  "type": "tool_result",
  "id": "call_123",
  "status": "success",
  "output": {
    "source_url": "https://mp.weixin.qq.com/s/example",
    "title": "文章标题",
    "text": "正文纯文本",
    "markdown": "# 文章标题\n\n正文",
    "metadata": {
      "captured_at": "2026-05-19T10:00:00+08:00"
    }
  },
  "artifacts": [],
  "usage": {
    "duration_ms": 8300
  }
}
```

失败结果：

```json
{
  "type": "tool_result",
  "id": "call_123",
  "status": "error",
  "error": {
    "code": "DOMAIN_NOT_ALLOWED",
    "message": "target domain is not allowed"
  },
  "usage": {
    "duration_ms": 12
  }
}
```

### 6.4 Artifact

当工具产生截图、PDF、HTML 文件、图片资源等大对象时，不应直接塞进 `output`。应通过 `artifacts` 描述。

```json
{
  "id": "art_001",
  "type": "image/png",
  "name": "page.png",
  "size": 123456,
  "uri": "digeino-artifact://art_001",
  "expires_at": "2026-05-19T11:00:00+08:00"
}
```

## 7. DigEino Collector

Collector 是 DigEino 的本地/私有运行模式，用于承接远程宿主下发的工具调用。

建议提供命令：

```bash
digeino collector --server https://knowledge.example.com --token xxx
```

Collector 负责：

- 与宿主建立 HTTP/WebSocket/MCP/stdin 连接。
- 上报 ToolManifest 和实例状态。
- 拉取或接收 ToolCall。
- 执行本地工具。
- 应用权限策略和安全校验。
- 保存本地 Cookie 和工具状态。
- 回传 ToolResult 和 Artifact。
- 记录脱敏审计日志。

## 8. 公众号采集示例

### 8.1 云端宿主创建任务

```json
{
  "type": "tool_call",
  "id": "call_wechat_001",
  "tool": "wechat.article.read",
  "input": {
    "url": "https://mp.weixin.qq.com/s/example",
    "format": ["text", "markdown", "html"]
  },
  "policy": {
    "timeout_ms": 90000,
    "allowed_domains": ["mp.weixin.qq.com"],
    "store_cookies": "local_only",
    "rate_limit_key": "wechat:mp.weixin.qq.com"
  }
}
```

### 8.2 本地 Collector 执行

执行步骤：

1. 校验 URL 域名是否允许。
2. 加载 `mp.weixin.qq.com` 本地 Cookie。
3. 用 `browser.browse` 打开文章页。
4. 等待 `#js_article` 或 `#js_content`。
5. 提取正文、标题、Markdown 和可选 HTML。
6. 保存更新后的 Cookie。
7. 回传结构化结果。

### 8.3 宿主处理结果

Knowledge 或其他宿主接收结果后：

- 存储原文链接。
- 存储必要正文或摘要。
- 向量化入库。
- 生成总结。
- 通知用户。

## 9. 安全与合规策略

### 9.1 凭证本地化

Cookie、浏览器 profile、本地文件凭证默认只保存在 Collector 节点，不上传宿主。

### 9.2 域名白名单

工具调用必须经过域名白名单校验。对于公众号采集场景，应限制到：

```json
["mp.weixin.qq.com"]
```

### 9.3 工具白名单

Collector 应支持按宿主、用户、租户配置可用工具。例如某个宿主只能调用：

```json
["browser.browse", "wechat.article.read"]
```

### 9.4 限流与冷却

对高风险工具增加限流：

- 每用户每分钟调用数。
- 每域名并发数。
- 每工具每日配额。
- 失败后的冷却窗口。

公众号采集建议默认并发为 1，必要时最多 2。

### 9.5 日志脱敏

日志中不能记录：

- Cookie
- Authorization
- API Key
- pass_ticket
- 私有文件内容
- 大段 HTML 原文

### 9.6 用户授权

对本地文件、浏览器登录态、企业内网系统等敏感工具，可以增加用户确认或首次授权流程。

## 10. 与 DigFlow、Knowledge 的关系

### 10.1 DigEino

DigEino 是通用插件运行时和本地工具网关。

负责：

- Tool schema。
- Tool registry。
- Tool executor。
- Collector。
- 本地浏览器和本地工具能力。
- 权限和审计。

### 10.2 DigFlow

DigFlow 是一个宿主和编排框架，不是协议中心。

负责：

- 实现 DigEino Tool Gateway Adapter。
- 将 DigEino 工具映射成 DigFlow Tool。
- 在工作流中路由 ToolCall。
- 管理 Agent 上下文和结果消费。

### 10.3 Knowledge

Knowledge 是产品宿主。

负责：

- 用户入口。
- 设备绑定。
- Collector 在线状态。
- 公众号采集任务创建。
- 采集结果入库、摘要、通知。

## 11. 兼容与迁移策略

新方案不应破坏 DigEino 现有的 Go module 内嵌调用方式。DigEino 当前已经被 Knowledge、DigFlow 等项目直接依赖，这条路径应继续保留。

当前形态：

```text
宿主 Go 进程
  -> import github.com/originaleric/digeino
  -> 注册 DigEino tools
  -> 进程内调用工具
```

新方案增加的是一层通用协议和运行时出口：

```text
任意宿主
  -> MCP / HTTP / WebSocket / stdio
  -> DigEino Runtime
  -> 同一批 DigEino tools
```

两条路径应长期共存。

### 11.1 保留现有调用 API

现有工具构造和注册方式继续可用，例如：

```text
NewBrowserBrowseTool(ctx)
NewBrowserSnapshotTool(ctx)
NewBrowserActionTool(ctx)
BaseTools
```

现有项目如果仍然在同一个 Go 进程里使用 DigEino 工具，不需要立即适配新协议。

### 11.2 新增统一工具包装层

每个 DigEino 工具逐步补充统一元信息和执行包装：

```text
DigEino Core Tool
  - browser_browse
  - browser_action
  - web_search
  - firecrawl
  - wecom_send
        |
        +-- Legacy Go Adapter
        |     给现有 DigFlow/Knowledge Go 内嵌调用
        |
        +-- Gateway Adapter
              转成 ToolManifest / ToolCall / ToolResult
              给 MCP/HTTP/WebSocket/stdio 调用
```

### 11.3 老宿主何时需要适配

不需要适配的情况：

- 宿主和 DigEino 在同一个 Go 进程里。
- 宿主只使用现有工具注册 API。
- 不需要远程 Collector、本地浏览器节点或跨进程工具调用。

需要适配的情况：

- Knowledge 云端希望调用用户本地 DigEino Collector。
- DigFlow 希望把远程 DigEino 工具当成普通 Tool 使用。
- 第三方 Agent 希望通过 MCP/HTTP/WebSocket 调用 DigEino。
- 工具需要运行在企业内网、用户本机或私有采集节点。

### 11.4 推荐演进方式

迁移策略采用 **Keep -> Wrap -> Expose**：

1. **Keep**
   保留现有 Go API、工具注册方式和宿主内嵌调用方式。

2. **Wrap**
   给现有工具增加统一 metadata、input schema、output schema、risk、capabilities 和 executor 包装。

3. **Expose**
   通过 MCP、HTTP、WebSocket、stdio 暴露同一批工具。

这样可以避免一次性大重构。DigEino 可以从 Go 工具库平滑演进为通用 Agent 插件运行时，同时现有 Knowledge、DigFlow 接入不被打断。

## 12. 推荐落地路线

### 阶段一：DigEino 内部协议雏形

- 定义 `ToolManifest`、`ToolCall`、`ToolResult`、`Artifact`、`Policy`。
- 将现有 `browser_browse` 包装为通用 executor。
- 实现本地 HTTP 调用接口。

### 阶段二：Collector MVP

- 实现 `digeino collector` 运行模式。
- 支持 WebSocket 反向连接。
- 支持任务拉取、执行、回传。
- 支持本地 Cookie 保存和域名白名单。

### 阶段三：Knowledge 首个宿主接入

- Knowledge 增加 Collector 绑定和在线状态。
- Knowledge 增加工具任务表。
- Knowledge 可创建 `wechat.article.read` 任务。
- Collector 回传后 Knowledge 入库、摘要和通知。

### 阶段四：DigFlow Adapter

- DigFlow 将远程 DigEino 工具暴露为普通 Tool。
- 支持同步等待和异步回调两种模式。
- 支持工具调用审计、超时、取消。

### 阶段五：MCP Server

- DigEino 暴露 MCP Server。
- 外部 Agent/IDE 可以直接发现和调用 DigEino 工具。
- 文档化通用接入方式。

## 13. 关键取舍

### 不把协议绑死在 DigFlow

DigFlow 是重要宿主，但 DigEino 的长期价值在于成为通用工具运行时。协议必须保持宿主无关。

### 不把 Knowledge 变成本地应用

Knowledge 可以继续云端部署。浏览器、本地文件、企业内网访问等敏感工具由本地或私有 Collector 执行。

### 不追求一开始全量标准化

短期可以先用 HTTP/WebSocket 跑通 Knowledge + 公众号采集。等真实链路稳定后，再抽象为 MCP 和更完整的远程工具协议。

## 14. 总结

DigEino 应升级为一个通用的 Agent 插件运行时，而不是只服务于 DigFlow 的工具库。它通过统一的 Tool Manifest、Tool Call、Tool Result 和多连接模式，让任意 Agent 都能安全调用本地或私有工具能力。

在公众号采集场景下，DigEino Collector 可以作为本地浏览器采集节点，降低云端抓取的风控和凭证风险；在更广泛场景下，它也可以成为企业内网工具、本地文件工具和浏览器自动化工具的统一网关。

最终目标是：

```text
DigEino = Universal Agent Plugin Runtime + Local/Private Tool Gateway
```
