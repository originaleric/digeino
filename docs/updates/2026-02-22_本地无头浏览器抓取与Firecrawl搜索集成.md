# 2026-02-22 本地无头浏览器抓取与 Firecrawl 搜索集成

基于 [基于本地无头浏览器的防微反爬策略](../ideas/2026-02-22_基于本地无头浏览器的防微反爬策略.md)，本次更新实现本地无头浏览器抓取工具，并将 Firecrawl Search 接入 `web_search`，形成「L1 HTTP → L2 本地浏览器 → L3 搜索」的阶梯式抓取与搜索能力。

## 更新亮点

- **本地无头浏览器抓取**：新增 `research_local_scraper` 工具，使用 go-rod + stealth 绕过微信公众号等强反爬，抓取正文并输出 Text 与 Markdown。
- **HTML → Markdown**：集成 `github.com/JohannesKaufmann/html-to-markdown` v1，将正文 HTML 转为 Markdown，与 firecrawl_scrape / jina_reader 输出格式对齐。
- **Firecrawl Search 引擎**：`web_search` 支持 `Engine: "firecrawl"`，复用 `Tools.Firecrawl.ApiKey`，与 firecrawl_scrape 共用配置。
- **依赖**：新增 go-rod、go-rod/stealth、html-to-markdown v1.6.0。

## 1. 本地无头浏览器抓取（research_local_scraper）

### 功能说明

- 工具名：`research_local_scraper`
- 请求：`{ "url": "https://..." }`
- 响应：`{ "text": "正文纯文本", "markdown": "正文 Markdown" }`
- 适用场景：微信公众号、知乎等对 HTTP 爬虫有强反爬的页面；当 firecrawl_scrape / research_jina_reader 返回「环境异常」时可作为 L2 兜底。

### 实现要点

- 使用 `go-rod` 启动本地 Chromium（Headless + NoSandbox），`stealth` 插件隐藏自动化特征。
- 导航后等待 `#js_content` 出现并读取 HTML / Text，再经 html-to-markdown 转 Markdown。
- 转换失败时仍返回 `text`，`markdown` 为空，并返回 error 供上层判断。

### 相关文件

- `tools/research/local_scraper.go`：工具实现
- `tools/tools.go`：在 `BaseTools` 中注册 `research_local_scraper`

## 2. Firecrawl Search 接入 web_search

### 配置方式

在 `eino.yml`（或等价配置）中：

```yaml
Tools:
  WebSearch:
    Engine: "firecrawl"
  Firecrawl:
    ApiKey: "fc-xxx"   # 与 firecrawl_scrape 共用
```

或通过环境变量 `FIRECRAWL_API_KEY` 设置。

### 实现要点

- 新增 `tools/research/websearch/firecrawl.go`，实现 `SearchProvider` 接口，调用 `POST https://api.firecrawl.dev/v2/search`。
- 请求体：`query`、`limit`、`sources: [{ "type": "web" }]`；解析 `data.web` 映射为统一 `SearchResult` 格式。
- `web_search.go` 中增加 `case "firecrawl"`，从 `cfg.Tools.Firecrawl` 读取 ApiKey。

### 相关文件

- `tools/research/websearch/firecrawl.go`：Firecrawl 搜索 Provider
- `tools/research/websearch/web_search.go`：引擎分支与配置绑定

## 3. 依赖变更

| 模块 | 版本 | 说明 |
|------|------|------|
| github.com/JohannesKaufmann/html-to-markdown | v1.6.0 | 正文 HTML 转 Markdown（v1 API：NewConverter + ConvertString） |
| github.com/go-rod/rod | v0.113.0 | 无头浏览器控制 |
| github.com/go-rod/stealth | v0.4.9 | 隐藏自动化特征 |

请在项目根目录执行 `go mod tidy` 或 `go get github.com/originaleric/digeino/...` 以更新 go.sum。

## 4. 推荐抓取与搜索策略

在 URL 接入 / 调研类 Agent 的 System Prompt 中可建议：

1. **L1**：`firecrawl_scrape`
2. **L2**：`research_jina_reader`
3. **L3**：`research_local_scraper`（微信公众号等强反爬场景）
4. **L4**：`web_search`（Bocha / Firecrawl 等）获取摘要兜底

当检测到「环境异常」「验证码」或内容明显不完整时，可优先尝试 `research_local_scraper`，再 fallback 到 `web_search`。
