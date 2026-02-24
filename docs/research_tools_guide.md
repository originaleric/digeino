# 调研员工具集 (Research Tools) 使用指南

本工具集旨在让 Agent 具备深度理解代码、文档和互联网信息的能力。

## 工具列表

### 1. 深度检索 (`research_grep`)
- **功能**：模拟 `grep` 或 `ripgrep`，在指定范围内搜索关键词。
- **参数**：
  - `query`: 搜索关键词。
  - `path`: (可选) 相对路径。
- **最佳实践**：用于发现特定功能在源码中的分布。

### 2. 精确阅读 (`research_read`)
- **功能**：读取文件的完整文本。
- **参数**：
  - `path`: 文件路径。
- **最佳实践**：在 `research_grep` 找到具体文件后，调用此工具深度阅读代码细节。

### 3. 文档转 Markdown (`research_doc_to_md`)
- **功能**：将 PDF、Word、TXT 等文件解析为纯文本 Markdown 结构。
- **参数**：
  - `path`: 文档路径。
- **最佳实践**：处理需求说明书、API 文档等非代码类文件。

### 4. 深度网页爬取 (`firecrawl_scrape`)
- **功能**：利用 Firecrawl API 深度清理网页噪音，返回结构化的 Markdown 内容。
- **参数**：
  - `url`: 目标网页地址。
- **前置条件**：需配置环境变量 `FIRECRAWL_API_KEY` 或在 `eino.yml` 中配置 `Tools.Firecrawl.ApiKey`。
- **最佳实践**：用于学习最新的第三方 API 文档。作为 L1 轻量级抓取方案，速度快、资源占用低。

### 5. 高级网页抓取 (`research_jina_reader`)
- **功能**：使用 Jina Reader (r.jina.ai) 深度读取网页并转换为 Markdown，对微信公众号、知乎等强反爬平台有较好兼容性。
- **参数**：
  - `url`: 目标网页地址。
- **前置条件**：无需 API Key（使用公共服务模式）。
- **最佳实践**：当 `firecrawl_scrape` 返回"环境异常"或"验证码"时，可作为 L2 备选方案。

### 6. 本地无头浏览器抓取 (`research_local_scraper`)
- **功能**：使用 go-rod + stealth 绕过复杂反爬（如微信公众号），抓取正文并输出 Text 与 Markdown。
- **参数**：
  - `url`: 目标网页地址（建议为微信公众号文章等复杂反爬页面）。
- **返回**：
  - `text`: 提取到的正文纯文本内容。
  - `markdown`: 基于正文 HTML 转换得到的 Markdown 内容。
- **前置条件**：需安装 Chromium 依赖（生产环境推荐使用 Docker 部署 browserless/chrome）。
- **最佳实践**：当 `firecrawl_scrape` 和 `research_jina_reader` 均失败时，可作为 L3 终极方案。注意：单次调用延迟较高（3-10秒），适合作为兜底而非首选。

### 7. 统一网页搜索 (`web_search`)
- **功能**：执行网页搜索以获取互联网信息，支持多引擎切换。
- **参数**：
  - `query`: 搜索关键词。
  - `max_results`: (可选) 最大返回结果数，默认 10。
  - `region`: (可选) 搜索地区，如 zh-CN, en-US。
- **支持的引擎**：Bocha, SerpApi, Google, Bing, DuckDuckGo, Firecrawl, **Tavily**。
- **配置方式**：在 `eino.yml` 中设置 `Tools.WebSearch.Engine` 选择引擎。
  - **Firecrawl**：复用 `Tools.Firecrawl.ApiKey` 或环境变量 `FIRECRAWL_API_KEY`。
  - **Tavily**：配置 `Tools.WebSearch.Tavily.ApiKey` 或环境变量 `TAVILY_API_KEY`；可选 `SearchDepth`（basic/fast/advanced/ultra-fast）、`Topic`（general/news/finance）。
- **最佳实践**：当所有抓取工具均失败时，使用 `web_search` 搜索目标 URL，获取 Bocha/Firecrawl/Tavily 的快照摘要作为 L4 兜底方案。英文/国际搜索场景推荐 Tavily。

---

## 如何在 Agent 编排中使用

由于 `DigFlow` 的底层加载机制会自动调用 `DigEino` 的 `BaseTools`，您只需在您的 `agent.yml` 的 `Tools` 部分直接引用名称即可：

```yaml
Tools:
  - Name: "research_grep"
    UseGlobal: true
  - Name: "research_read"
    UseGlobal: true
  - Name: "research_doc_to_md"
    UseGlobal: true
  - Name: "firecrawl_scrape"
    UseGlobal: true
  - Name: "research_jina_reader"
    UseGlobal: true
  - Name: "research_local_scraper"
    UseGlobal: true
  - Name: "web_search"
    UseGlobal: true
```

## 阶梯式抓取策略建议

在 Agent 的 SystemPrompt 中，建议遵循以下抓取优先级：

1. **L1 (轻量级)**：优先使用 `firecrawl_scrape`（速度快、资源占用低）。
2. **L2 (高级抓取)**：若 L1 返回"环境异常"、"验证码"或内容明显不正确，尝试 `research_jina_reader`。
3. **L3 (本地浏览器)**：若 L2 仍失败，可尝试 `research_local_scraper`（适合微信公众号等强反爬场景，但延迟较高）。
4. **L4 (搜索兜底)**：若所有抓取均失败，使用 `web_search` 搜索目标 URL，获取快照摘要。

## 注意事项
1. **Token 消耗**：由于这些工具往往返回大量文本，建议配合具备长上下文处理能力的模型（如 DeepSeek, Qwen）。
2. **环境权限**：`research_grep` 和 `research_read` 仅能访问服务部署环境下允许的文件系统范围。
3. **资源消耗**：`research_local_scraper` 需要启动本地 Chromium，单实例内存约 200-500MB，生产环境需控制并发数或使用浏览器池。
4. **延迟权衡**：`firecrawl_scrape` 和 `research_jina_reader` 为 HTTP 请求（秒级），`research_local_scraper` 为本地浏览器渲染（3-10秒），建议优先使用轻量级方案。
