# browser_browse 使用与配置指南

本文说明如何在 DigEino 中启用和使用本地浏览器工具 `browser_browse`。

## 1. 工具定位

`browser_browse` 是基于 `go-rod + stealth` 的本地无头浏览器工具，适用于：

- 动态渲染页面（纯 HTTP 抓取不完整）
- 强反爬页面（例如微信公众号、Cloudflare 保护页面）
- 需要截图留档的场景
- 需要复用 Cookie 维持登录态的场景

它与 `research_local_scraper` 的区别：

- `research_local_scraper`：更偏微信公众号正文抓取，能力较聚焦。
- `browser_browse`：更通用，支持 `read/screenshot`、`wait_selector`、Cookie 域复用与并发控制。

## 2. 配置方式（config/config.yaml）

在 `Tools` 下新增/配置 `LocalBrowser`：

```yaml
Tools:
  LocalBrowser:
    Enabled: true
    MaxConcurrency: 3
    TotalTimeoutSec: 60
    NavigateTimeoutSec: 30
    WaitSelectorTimeoutSec: 20
    AllowedDomains:
      - "mp.weixin.qq.com"
      - "zhihu.com"
      - "juejin.cn"
    ChromePath: "" # 可选，留空表示由 rod 自动寻找或下载 Chromium
    Headless: true
    CookieStoreDir: "storage/app/browser_cookies"
```

### 配置项说明

- `Enabled`
  - 是否启用工具注入；`false` 时 `browser_browse` 不会注册到全局工具集。
- `MaxConcurrency`
  - 浏览器任务并发上限。建议 3~5（结合机器内存）。
- `TotalTimeoutSec`
  - 单次调用总超时。
- `NavigateTimeoutSec`
  - 页面导航超时（`Navigate` 阶段）。
- `WaitSelectorTimeoutSec`
  - 等待 `wait_selector` 或正文候选元素的超时。
- `AllowedDomains`
  - 允许访问的域名白名单。为空表示不限制；生产环境建议显式配置。
- `ChromePath`
  - 可选，自定义 Chromium 可执行文件路径。
- `Headless`
  - 是否无头运行。开发调试可设为 `false`。
- `CookieStoreDir`
  - Cookie 持久化目录，按域名存储为 `<domain>.json`。

## 3. Agent 接入

在你的 `agent.yml` 里加入全局工具引用：

```yaml
Tools:
  - Name: "browser_browse"
    UseGlobal: true
```

建议配合 Prompt 提示模型在以下情况调用：

- HTTP 抓取失败或被拦截
- 页面为前端渲染（需等待元素）
- 需要网页截图
- 需要复用指定域 Cookie

## 4. 调用参数与返回

### 输入参数

- `url` (string, 必填)：目标网页地址
- `action` (string, 可选)：`read` 或 `screenshot`，默认 `read`
- `wait_selector` (string, 可选)：等待出现的 CSS 选择器
- `use_cookie_domain` (string, 可选)：要加载/回写 Cookie 的域名

### 输出字段

- `url`：最终访问地址
- `title`：页面标题
- `text`：正文纯文本（`action=read`）
- `markdown`：正文 Markdown（`action=read`）
- `screenshot_base64`：整页截图 base64（`action=screenshot`）

## 5. 调用示例

### 5.1 读取正文（默认 action=read）

```json
{
  "url": "https://mp.weixin.qq.com/s/xxxx",
  "wait_selector": "#js_content"
}
```

### 5.2 截图

```json
{
  "url": "https://example.com",
  "action": "screenshot",
  "wait_selector": "main"
}
```

### 5.3 复用登录态 Cookie

```json
{
  "url": "https://wiki.company.com/page/123",
  "action": "read",
  "use_cookie_domain": "wiki.company.com",
  "wait_selector": ".content"
}
```

## 6. 推荐策略（与其它抓取工具配合）

建议按阶梯方式使用：

1. `firecrawl_scrape`（轻量优先）
2. `research_jina_reader`（反爬兼容）
3. `browser_browse`（动态渲染/截图/Cookie 复用）
4. `web_search`（兜底摘要）

## 7. 常见问题

### Q1: 为什么提示“域名不在允许列表”？

- 说明 `AllowedDomains` 启用了白名单且未包含目标域。
- 将主域加入白名单（例如 `zhihu.com` 会匹配 `www.zhihu.com`）。

### Q2: 为什么截图或抓取超时？

- 可适当提高 `TotalTimeoutSec` 与 `WaitSelectorTimeoutSec`。
- 动态页面建议提供 `wait_selector`，避免过早提取。

### Q3: Cookie 没生效？

- 确认 `use_cookie_domain` 与目标站点域名一致。
- 确认 `CookieStoreDir` 下存在对应 `<domain>.json`。
- 某些站点 Cookie 受 `Secure/SameSite/Path` 限制，需在同源访问下验证。

## 8. 生产环境建议

- 建议显式配置 `AllowedDomains`，避免 SSRF 风险。
- 结合机器内存设置 `MaxConcurrency`（通常 3~5）。
- Linux/CentOS 场景确保安装 Chromium 依赖和中文字体。
- 长期运行建议配合服务级监控观察内存与浏览器进程数量。
