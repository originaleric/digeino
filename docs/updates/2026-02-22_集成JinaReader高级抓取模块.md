# 2026-02-22 新增 Jina Reader 高级抓取工具

为了解决微信公众号、知乎等具有高强度反爬机制平台的抓取问题，特别是在 Firecrawl 遇到“环境异常”或“验证码”拦截时提供可靠的备选方案。

## 更新亮点

- **新增 Jina Reader 驱动**：实现在 `research_jina_reader` 工具。
- **反爬增强**：采用 `text/event-stream` 协议解析，有效提升了对极客、微信等页面的穿透力。
- **自动解析**：内置了针对 Jina Reader 事件流格式的解析逻辑，直接返回纯净的 Markdown。
- **智能策略集成**：已在 `url_ingest` 智能体中集成，支持“Firecrawl 失败 -> Jina 重试 -> 搜索摘要兜底”的阶梯式抓取逻辑。

## 使用指南

### 基础调用
该工具无需 API Key（使用公共服务模式），通过 `r.jina.ai` 代理访问。

```go
// 注册名称: research_jina_reader
args := map[string]interface{}{
    "url": "https://mp.weixin.qq.com/s/...",
}
```

### 推荐工作流 (Smart Ingest)
在 Agent 的 System Prompt 中建议遵循以下逻辑：
1. 优先尝试 `firecrawl_scrape`。
2. 若返回内容包含拦截指纹（环境异常），切换到 `research_jina_reader`。
3. 若抓取皆失败，使用 `web_search` 搜索对应 URL，获取 Bocha/Google 的快照摘要。

## 相关文件
- [jina_reader.go](file:///Users/dig/Documents/文稿 - XinYe的MacBook Pro (5)/Projects/go-app/DigEino/tools/research/jina_reader.go)
- [tools.go](file:///Users/dig/Documents/文稿 - XinYe的MacBook Pro (5)/Projects/go-app/DigEino/tools/tools.go)
- [agent.yml (url_ingest)](file:///Users/dig/Documents/文稿 - XinYe的MacBook Pro (5)/Projects/go-app/DigFlow/app/memo/ingest/config/agent.yml)
