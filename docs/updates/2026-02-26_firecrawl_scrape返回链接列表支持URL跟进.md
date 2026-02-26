# 更新：firecrawl_scrape 返回链接列表支持 URL 跟进

**日期**: 2026-02-26
**类别**: 功能 / 工具

## 概述

增强 `firecrawl_scrape` 工具，在返回 Markdown 内容的基础上新增 `links` 字段，自动从爬取结果中提取相关链接。便于 Researcher Agent 等场景进行 URL 跟进爬取，实现从爬取内容中识别高价值链接并继续深度调研。

## 新功能

### links 字段

**功能特性**：
- 从 Markdown 中自动提取 `[text](url)` 格式的链接
- 过滤锚点、图片、data URL、mailto、javascript 等无效链接
- 校验为有效 HTTP(S) URL
- 去重，每个 URL 仅出现一次
- 返回 `url` 和 `text` 便于 Agent 筛选

**返回结构**：
```json
{
  "markdown": "爬取并转换后的 Markdown 内容",
  "links": [
    {"url": "https://example.com/doc", "text": "文档标题"}
  ],
  "success": true
}
```

## 修改的组件

### 更新的文件
- `tools/research/firecrawl.go`：
  - 新增 `Link` 结构体（`url`、`text`）
  - `FirecrawlResponse` 增加 `Links []Link` 字段
  - 新增 `extractLinksFromMarkdown` 函数：正则提取 + 过滤
  - 解析 API 响应时兼容 `markdown` 与 `content` 字段
  - 更新工具描述

## 实现细节

### 链接提取逻辑

1. **正则匹配**：`\[([^\]]*)\]\(([^)]+)\)` 匹配 Markdown 链接
2. **过滤规则**：
   - 排除 `#`、`data:`、`mailto:`、`javascript:` 前缀
   - 排除 `.png`、`.jpg`、`.jpeg`、`.gif`、`.webp`、`.svg` 等图片
   - 仅保留 `http` / `https` 协议
3. **去重**：同一 URL 只保留首次出现

### 代码示例

```go
// Link 结构
type Link struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

// FirecrawlResponse 新增 Links
type FirecrawlResponse struct {
	Markdown string `json:"markdown"`
	Links    []Link `json:"links"`
	Success  bool   `json:"success"`
}
```

## 使用场景

1. **Researcher Agent URL 跟进**：爬取页面后从 `links` 中选取官方文档、相关说明等 URL，继续调用 `firecrawl_scrape` 深度爬取
2. **减少 LLM 解析负担**：无需从 Markdown 文本中手动解析链接，直接使用结构化 `links` 列表
3. **方案三落地**：对应 DigFlow `docs/ideas/researcher_优化方案.md` 中的方案 C

## 配置

无需额外配置，与现有 `firecrawl_scrape` 使用方式一致。仍依赖 `eino.yml` 中 `Tools.Firecrawl.ApiKey` 或环境变量 `FIRECRAWL_API_KEY`。

## 向后兼容

- 新增 `links` 字段，不影响现有仅使用 `markdown` 的调用方
- 若 Markdown 中无有效链接，`links` 为空数组 `[]`

## 相关文件

- `tools/research/firecrawl.go` - 工具实现
- `DigFlow/docs/ideas/researcher_优化方案.md` - 方案背景
