# 2026-03-09 本地浏览器工具 browser_browse 实现与接入

基于《基于本地无头浏览器的防微反爬策略》方案，已在 DigEino 完成通用本地浏览器工具 `browser_browse` 的实现与接入，形成可配置、可控并发、可复用 Cookie 的本地抓取能力。

## 本次更新内容

### 1) 新增通用工具 `browser_browse`

- 工具名：`browser_browse`
- 输入参数：
  - `url`：目标网页 URL（必填）
  - `action`：`read` / `screenshot`（默认 `read`）
  - `wait_selector`：可选，等待指定元素出现
  - `use_cookie_domain`：可选，加载并复用指定域名 Cookie
- 输出字段：
  - `url`、`title`
  - `text`、`markdown`（`action=read`）
  - `screenshot_base64`（`action=screenshot`）

### 2) 浏览器会话管理与并发控制

- 新增浏览器管理模块，采用全局单例复用浏览器实例。
- 使用并发槽位限制同时执行的浏览器任务数量（默认 3）。
- 支持浏览器断连重建与会话回收，避免长期占用资源。
- 增加轻量僵尸会话清理逻辑，防止异常任务泄漏。

### 3) 安全与稳定性能力

- URL 校验：仅允许 `http/https`。
- 域名白名单：可通过配置限制可访问域名。
- Cookie 域安全处理：限制域名字符，防止非法路径拼接。
- 超时分层：总超时、导航超时、等待选择器超时均可配置。

### 4) 配置体系扩展

在 `Tools` 下新增 `LocalBrowser` 配置项：

- `Enabled`
- `MaxConcurrency`
- `TotalTimeoutSec`
- `NavigateTimeoutSec`
- `WaitSelectorTimeoutSec`
- `AllowedDomains`
- `ChromePath`
- `Headless`
- `CookieStoreDir`

并已在默认配置与示例 YAML 中补齐默认值与注释。

### 5) 工具注册与文档更新

- 已在全局 `BaseTools` 注册 `browser_browse`（受 `Tools.LocalBrowser.Enabled` 开关控制）。
- 已更新研究工具指南，补充 `browser_browse` 的参数、返回和使用建议。

## 主要变更文件

- `tools/research/browser_browse.go`
- `tools/research/browser_pool.go`
- `tools/tools.go`
- `config/config.go`
- `config/config.yaml`
- `docs/research_tools_guide.md`

## 说明

- 原有 `research_local_scraper` 保持兼容，不替换。
- `browser_browse` 定位为更通用的本地浏览器访问能力，适用于动态渲染、复杂反爬、截图留档与 Cookie 复用场景。
