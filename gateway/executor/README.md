# Gateway Executor 使用说明

本目录是 DigEino **Agent 插件运行时**的 Wrap 层：把 `tools/` 下的 Legacy 能力包装成统一的 `registry.Entry`，供 HTTP Gateway、WebSocket Collector、MCP、stdio 调用。

每个文件对应一类（或一个）**网关工具名**（点分式，如 `browser.browse`）。宿主只需发送 `ToolCall`，不必 import 本包。

---

## 1. 前置条件

| 配置项 | 说明 |
|--------|------|
| `Gateway.AllowedTools` | 工具白名单，未列入的工具会被拒绝 |
| `Collector.AllowedTools` | Collector 进程独立白名单（为空时回退 Gateway） |
| `Tools.LocalBrowser.Enabled` | 浏览器类 / 平台类工具须为 `true` |
| `Tools.LocalBrowser.AllowedDomains` | 目标 URL 与 Cookie 域名白名单；为空表示不限制（仅建议开发环境） |
| `Tools.LocalBrowser.CookieStoreDir` | 本地 Cookie 持久化目录，登录态仅保存在本机 |
| `Gateway.AllowedReadPaths` | **仅** `file.read` 需要；非空才会注册该工具 |

启动 Gateway 后可通过清单确认已暴露工具：

```bash
go run ./cmd/digeino gateway --config config/config.yaml
curl -s http://127.0.0.1:8787/manifest
```

通用协议与多出口说明见 [`../README.md`](../README.md)；Wrap 开发流程见 [`../USAGE.md`](../USAGE.md)。

---

## 2. 目录与分类

```text
gateway/executor/
  helpers.go              # decodeInput、URL/路径/Cookie 域校验
  platform_common.go      # 平台 Executor 共用模板（输入/输出/Artifact）
  browser_browse.go       # 通用：页面正文抓取
  browser_snapshot.go     # 通用：可交互元素快照
  browser_action.go       # 通用：点击/输入/滚动等
  file_read.go            # 通用：本地文件读取
  wechat_article.go       # 平台：公众号文章
  xiaohongshu_note.go     # 平台：小红书笔记
  douyin_video.go         # 平台：抖音视频页
  x_post.go               # 平台：X 帖子
```

| 分类 | 工具 | 底层实现 | 适用场景 |
|------|------|----------|----------|
| **通用浏览器** | `browser.browse` / `snapshot` / `action` | `tools/research` | 任意允许域名上的自定义抓取或自动化 |
| **平台结构化** | `wechat.article.read` 等 | `tools/platform/*` | 固定平台单篇内容，结构化标题/作者/互动等 |
| **本地文件** | `file.read` | `tools/research.ReadFile` | 读取宿主授权目录下的文件 |

**如何选择：**

- 已知平台 URL（公众号 / 小红书 / 抖音 / X）→ 优先用对应 **平台 Executor**，输出字段统一、Selector 已内置。
- 未知站点或需自定义选择器 / 多步操作 → 用 **`browser.browse`** + 可选 **`browser.action`** / **`browser.snapshot`**。
- 读本地产物（报告、缓存 JSON）→ **`file.read`**（须配置 `AllowedReadPaths`）。

---

## 3. 通用调用格式

所有工具均通过 `POST /tools/call`（或 Collector / MCP / stdio 的等价 `tool_call`）调用：

```json
{
  "type": "tool_call",
  "id": "call_unique_id",
  "tool": "<工具名>",
  "input": { },
  "policy": {
    "timeout_ms": 90000,
    "allowed_domains": ["example.com"],
    "max_output_bytes": 2000000
  }
}
```

**策略字段（`policy`）说明：**

| 字段 | 作用 |
|------|------|
| `timeout_ms` | 单次调用超时（由 `runtime` 控制） |
| `allowed_domains` | **优先于** `config.yaml` 的 `LocalBrowser.AllowedDomains`；平台工具在未配置时还会使用 Executor 内置默认域名 |
| `max_output_bytes` | 限制 `output` JSON 大小 |

**Cookie：** `use_cookie_domain` 指定从 `CookieStoreDir` 加载哪一域的 Cookie；Cookie **不会**出现在 `ToolResult` 中。

**Artifact：** 大对象（如截图）通过 `artifacts[]` 返回，`output` 内仅含 `screenshot_artifact_id`，下载：`GET /artifacts/{id}`。

---

## 4. 通用浏览器 Executor

### 4.1 `browser.browse`

**文件：** `browser_browse.go`  
**能力：** 本地 go-rod 打开 URL，提取正文或截图。

| 输入字段 | 类型 | 说明 |
|----------|------|------|
| `url` | string | 必填 |
| `action` | string | `read`（默认）或 `screenshot` |
| `mode` | string | `full`（默认）、`snapshot`、`summary` |
| `wait_selector` | string | 等待出现的 CSS 选择器 |
| `content_selector` | string | `mode=full` 时限定正文范围 |
| `use_cookie_domain` | string | Cookie 域名 |
| `tab_id` | string | 复用标签页（高级） |

**输出（read）：** `source_url`、`title`、`text`、`markdown`；截图时另有 `screenshot_artifact_id`。

```json
{
  "tool": "browser.browse",
  "input": {
    "url": "https://example.com/docs",
    "action": "read",
    "mode": "full",
    "wait_selector": "main",
    "content_selector": "article",
    "use_cookie_domain": "example.com"
  },
  "policy": { "allowed_domains": ["example.com"], "timeout_ms": 60000 }
}
```

---

### 4.2 `browser.snapshot`

**文件：** `browser_snapshot.go`  
**能力：** 返回页面可访问性树中的可交互元素列表（`ref` 供 `browser.action` 使用）。

| 输入字段 | 说明 |
|----------|------|
| `url` | 必填 |
| `filter` | 元素过滤 |
| `max_depth` | 树深度 |
| `wait_selector` / `use_cookie_domain` | 同 browse |

**输出：** `source_url`、`title`、`elements`（含 `ref`、角色、文本等）。

典型流程：`browser.snapshot` 获取 `ref` → `browser.action` 执行 `click` / `type`。

---

### 4.3 `browser.action`

**文件：** `browser_action.go`  
**能力：** 在已打开或新导航的页面上执行交互。  
**风险：** `RequiresUserApproval: true`（Manifest 中标记需用户确认）。

| 输入字段 | 说明 |
|----------|------|
| `url` | 必填，目标页 |
| `action` | 必填：`click`、`type`、`fill`、`hover`、`scroll`、`focus`、`press` |
| `ref` | 快照中的元素引用（优先） |
| `selector` | 无 `ref` 时用 CSS 选择器 |
| `text` / `key` | 输入或按键 |
| `scroll_x` / `scroll_y` | 滚动偏移 |
| `human_like` | 随机延迟，模拟人工 |
| `use_cookie_domain` | Cookie 域名 |

```json
{
  "tool": "browser.action",
  "input": {
    "url": "https://example.com/login",
    "ref": "e3",
    "action": "click",
    "use_cookie_domain": "example.com"
  }
}
```

---

### 4.4 `file.read`

**文件：** `file_read.go`  
**注册条件：** `Gateway.AllowedReadPaths` 非空，且白名单含 `file.read`。

| 输入 | 输出 |
|------|------|
| `path`（必填，绝对或相对路径） | `path`、`content` |

路径必须落在 `AllowedReadPaths` 某一前缀之下，否则返回 `TOOL_NOT_ALLOWED`。

```yaml
Gateway:
  AllowedTools:
    - file.read
  AllowedReadPaths:
    - storage/app/reports
```

---

## 5. 平台结构化 Executor

平台类工具由 `platform_common.go` 统一处理：**输入模型一致、输出模型一致**，具体采集逻辑在 `tools/platform/<平台>/`。

实现链：

```text
ToolCall → gateway/executor/*.go → tools/platform/* → research.BrowserBrowse (+ metadata_script)
```

### 5.1 统一输入

| 字段 | 说明 |
|------|------|
| `url` | 必填，单篇内容链接 |
| `format` | `["text","markdown","html"]` 子集；省略则返回全部文本字段 |
| `include_media` | 是否填充 `media`（图片 URL 等） |
| `include_screenshot` | 是否截图并写入 Artifact |
| `wait_selector` | 覆盖平台默认等待选择器 |
| `use_cookie_domain` | 本地 Cookie 域（登录态页面建议填写） |

### 5.2 统一输出（成功时）

| 字段 | 说明 |
|------|------|
| `platform` | 如 `xiaohongshu`、`douyin` |
| `content_type` | 如 `note`、`video`、`post`、`article` |
| `source_url` / `canonical_url` | 来源与规范链接 |
| `title` / `text` / `markdown` | 标题与正文 |
| `author` | `{ id, name, profile_url, avatar_url }` |
| `published_at` / `captured_at` | 发布时间与采集时间 |
| `media` | `[{ type, url, artifact_id }]` |
| `engagement` | `{ likes, comments, shares, reposts, bookmarks }` |
| `tags` | 标签列表（小红书等） |
| `platform_metadata` | 平台原始扩展字段 |
| `screenshot_artifact_id` | 仅 `include_screenshot: true` |

---

### 5.3 `wechat.article.read`

**文件：** `wechat_article.go` → `tools/platform/wechat`  
**默认域名：** `mp.weixin.qq.com`、`weixin.qq.com`  
**建议 Cookie 域：** `mp.weixin.qq.com`

```json
{
  "tool": "wechat.article.read",
  "input": {
    "url": "https://mp.weixin.qq.com/s/xxxxx",
    "format": ["text", "markdown"],
    "use_cookie_domain": "mp.weixin.qq.com"
  },
  "policy": { "allowed_domains": ["mp.weixin.qq.com"], "timeout_ms": 90000 }
}
```

说明：内部对 `#js_article` / `#js_content` 做 fallback；作者与发布时间通过页面 JS 元数据脚本提取。

---

### 5.4 `xiaohongshu.note.read`

**文件：** `xiaohongshu_note.go` → `tools/platform/xiaohongshu`  
**默认域名：** `xiaohongshu.com`、`www.xiaohongshu.com`、`xhslink.com`  
**建议 Cookie 域：** `www.xiaohongshu.com`

```json
{
  "tool": "xiaohongshu.note.read",
  "input": {
    "url": "https://www.xiaohongshu.com/explore/xxxxxxxx",
    "format": ["text", "markdown"],
    "include_media": true,
    "use_cookie_domain": "www.xiaohongshu.com"
  },
  "policy": {
    "allowed_domains": ["xiaohongshu.com", "www.xiaohongshu.com"],
    "timeout_ms": 90000
  }
}
```

Selector 维护：`tools/platform/xiaohongshu/selectors.go`。

---

### 5.5 `douyin.video.read`

**文件：** `douyin_video.go` → `tools/platform/douyin`  
**默认域名：** `douyin.com`、`www.douyin.com`、`iesdouyin.com`  
**建议 Cookie 域：** `www.douyin.com`

```json
{
  "tool": "douyin.video.read",
  "input": {
    "url": "https://www.douyin.com/video/xxxxxxxx",
    "include_media": true,
    "include_screenshot": true,
    "use_cookie_domain": "www.douyin.com"
  },
  "policy": {
    "allowed_domains": ["douyin.com", "www.douyin.com"],
    "timeout_ms": 90000
  }
}
```

说明：`include_media` 时尝试返回封面图 URL；视频本体不大块塞进 `output`。

---

### 5.6 `x.post.read`

**文件：** `x_post.go` → `tools/platform/x`  
**默认域名：** `x.com`、`twitter.com`  
**建议 Cookie 域：** `x.com`

```json
{
  "tool": "x.post.read",
  "input": {
    "url": "https://x.com/user/status/1234567890",
    "include_media": true,
    "use_cookie_domain": "x.com"
  },
  "policy": {
    "allowed_domains": ["x.com", "twitter.com"],
    "timeout_ms": 90000
  }
}
```

说明：正文优先从 `tweetText` 区域元数据提取；`twitter.com` 与 `x.com` 链接均可。

---

## 6. 平台 vs 浏览器：组合用法

| 目标 | 推荐方式 |
|------|----------|
| 读一篇公众号 | `wechat.article.read` |
| 读小红书笔记 | `xiaohongshu.note.read` |
| 登录后任意站内多步操作 | `browser.snapshot` → `browser.action`（循环） |
| 自定义正文 CSS | `browser.browse` + `content_selector` |
| 调试某站 DOM | `browser.browse` + `wait_selector`，或改 `tools/platform/*/selectors.go` |
| 采集后读本地 JSON | `file.read` |

**不推荐：** 在宿主侧重复实现平台 DOM 解析；页面变更时应只改 `tools/platform`，Executor 协议层保持稳定。

---

## 7. 错误码与排查

| 错误码 | 常见原因 |
|--------|----------|
| `TOOL_NOT_ALLOWED` | 工具未加入 `AllowedTools`，或 `file.read` 未配置 `AllowedReadPaths` |
| `DOMAIN_NOT_ALLOWED` | URL / Cookie 域不在 `policy.allowed_domains` 或 `LocalBrowser.AllowedDomains` |
| `INVALID_INPUT` | 缺少 `url`/`path`、JSON 格式错误 |
| 浏览器启动失败 | `LocalBrowser.Enabled: false` 或 Chromium 未安装 |
| 等待选择器超时 | 页面结构变化、未登录、地区限制；检查 Cookie 与 `wait_selector` |
| 正文为空 | 平台改版 → 更新对应 `selectors.go` 中的 `MetadataScript` / 选择器 |

验证码、风控、登录墙：**应返回明确错误**，Executor 不做绕过。

---

## 8. 注册与扩展

所有 Executor 在 [`../bootstrap.go`](../bootstrap.go) 的 `NewRegistry()` 中注册：

```go
reg.Register(executor.BrowserBrowseEntry(domains, store))
reg.Register(executor.WechatArticleReadEntry(domains, store))
// ...
```

**新增平台 Executor：**

1. 在 `tools/platform/<name>/` 实现 `ReadXxx(ctx, platform.ReadInput) (*platform.Content, error)`
2. 新增 `gateway/executor/<name>.go`，调用 `platformReadEntry(...)` 或仿照 `browser_browse.go` 手写 Handler
3. `bootstrap.go` 中 `Register`
4. 更新 `config.yaml` 白名单与 `AllowedDomains`

**新增通用 Executor：** 见 [`../USAGE.md`](../USAGE.md)。

---

## 9. 相关文档

| 文档 | 内容 |
|------|------|
| [平台采集 Executor 落地说明](../../docs/updates/2026-05-20_平台采集Executor落地与使用说明.md) | 变更说明与配置示例 |
| [平台采集 Executor 方案](../../docs/ideas/2026-05-19_DigEino平台采集Executor方案.md) | 设计与后续演进 |
| [Agent 插件运行时落地说明](../../docs/updates/2026-05-19_Agent插件运行时落地与使用说明.md) | Gateway / Collector / MCP |
| [`../README.md`](../README.md) | 宿主 HTTP / WS API |
| [`../USAGE.md`](../USAGE.md) | Wrap → Expose 开发清单 |
