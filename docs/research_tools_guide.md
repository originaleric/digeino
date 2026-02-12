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
- **前置条件**：需配置环境变量 `FIRECRAWL_API_KEY`。
- **最佳实践**：用于学习最新的第三方 API 文档。

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
```

## 注意事项
1. **Token 消耗**：由于这些工具往往返回大量文本，建议配合具备长上下文处理能力的模型（如 DeepSeek, Qwen）。
2. **环境权限**：`research_grep` 和 `research_read` 仅能访问服务部署环境下允许的文件系统范围。
