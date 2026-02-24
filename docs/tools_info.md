# DigEino 工具集概览 (Tools Information)

本文档旨在汇总 `DigEino` 中集成的所有 Eino 工具，详细说明其功能、参数及使用场景。

---

## 1. 调研与搜索 (Researcher & Search)

这类工具旨在帮助 Agent 理解项目代码、阅读外部文档以及搜索互联网信息。

### 1.1 基础调研 (Basic Research)
| 工具名称 | 功能描述 | 主要参数 |
| :--- | :--- | :--- |
| `research_grep` | 在源码或指定目录中进行关键词检索（模拟 ripgrep）。 | `query`: 关键词; `path`: 搜索路径 |
| `research_read` | 精确读取指定文件的完整内容。 | `path`: 文件路径 |
| `research_doc_to_md` | 集成 **Unstructured.io**，将 PDF/Word/Docx 高质量转为 Markdown。 | `path`: 文档路径 |

### 1.2 高级检索 (Advanced Retrieval)
| 工具名称 | 功能描述 | 主要参数 |
| :--- | :--- | :--- |
| `research_semantic_search` | **向量语义搜索**。基于 Pinecone，通过相似度定位偏逻辑意图的内容。 | `query`: 搜索意图; `max_results`: 结果数 |
| `research_code_index` | **项目索引工具**。扫描指定目录并将代码/文档向量化存入 Pinecone。 | `path`: 目标目录 |
| `web_search` | **统一网页搜索**。支持 Bocha, SerpApi, Google, Bing, DuckDuckGo, Firecrawl 和 Tavily 引擎。 | `query`: 搜索词; `max_results`: 结果数; `region`: 地区 |
| `firecrawl_scrape` | **深度网页爬取**。调用 Firecrawl API 将 URL 转换为清洗后的 Markdown。 | `url`: 目标网页地址 |
| `research_jina_reader` | **高级网页抓取**。使用 Jina Reader (r.jina.ai) 深度读取网页并转换为 Markdown，对微信公众号、知乎等强反爬平台有较好兼容性。 | `url`: 目标网页地址 |
| `research_local_scraper` | **本地无头浏览器抓取**。使用 go-rod + stealth 绕过复杂反爬（如微信公众号），抓取正文并输出 Text 与 Markdown。 | `url`: 目标网页地址 |

**web_search 引擎选择**：

| 引擎 | 特点 | 配置 |
|------|------|------|
| Bocha | 国内服务，中文友好 | `Tools.WebSearch.Bocha.ApiKey` |
| Firecrawl | 与 scrape 共用 ApiKey | `Tools.Firecrawl.ApiKey` |
| **Tavily** | 面向 AI、低延迟、英文/国际搜索 | `Tools.WebSearch.Tavily.ApiKey` 或 `TAVILY_API_KEY` |
| SerpApi/Google | 依赖第三方 API | 各自 ApiKey |

---

## 2. 智能 UI/UX 设计 (UI/UX Design)

基于 AI 的设计知识库检索与设计系统自动生成工具。

| 工具名称 | 功能描述 | 主要参数 |
| :--- | :--- | :--- |
| `ui_ux_search` | 检索 UI/UX 知识库（样式、配色、字体、交互、落地页模式等）。 | `query`: 检索词; `stack`: 技术栈(如 tailwind) |
| `generate_design_system` | 为特定产品需求生成完整的 UI/UX 设计系统。 | `query`: 需求描述; `project_name`: 项目名 |
| `persist_design_system` | 将生成的设计系统持久化为 `MASTER.md` 及页面覆盖文件。 | `project_name`: 项目名; `page_name`: 页面名 |

---

## 3. 社交推送与通知 (Messaging & Push)

支持 WeChat 公众号及 WeCom（企业微信）的多种消息类型推送。

### 3.1 企业微信 (WeCom)
| 工具名称 | 功能描述 | 主要参数 |
| :--- | :--- | :--- |
| `send_wecom_message` | 发送企业微信应用消息（文字）。 | `user_id`: 成员ID; `content`: 文本内容 |
| `send_wecom_image` | 发送图片消息。 | `user_id`: 成员ID; `media_id`: 临时素材ID |
| `send_wecom_text_card` | 发送文本卡片（带链接和描述）。 | `user_id`: 成员ID; `title`: 标题; `url`: 链接 |
| `send_wecom_customer_message` | **客服消息**。发送给添加了客服的个人微信用户。 | `customer_id`: 外部ID; `content`: 文本内容 |

### 3.2 个人微信 (WeChat)
| 工具名称 | 功能描述 | 主要参数 |
| :--- | :--- | :--- |
| `send_wechat_message` | 通过微信服务号推送文字消息给关注用户。 | `openid`: 用户识别码; `content`: 文本内容 |

---

---

## 4. 如何在 Go 项目中集成 (Integration Guide)

`DigEino` 的工具集是基于 [CloudWeGo Eino](https://github.com/cloudwego/eino) 框架开发的标准 `tool.BaseTool`。这意味着它们可以被集成到任何使用 Eino 框架的 Go 应用中。

### 4.1 引入依赖
首先，在您的项目中安装 `DigEino`：
```bash
go get github.com/originaleric/digeino
```

### 4.2 初始化配置
`DigEino` 工具依赖于一个全局配置单例。在使用工具前，您需要初始化配置（支持从 YAML 加载或环境变量覆盖）：

```go
import "github.com/originaleric/digeino/config"

func main() {
    // 方式 A：从 YAML 文件加载
    cfg, _ := config.LoadConfig("path/to/your/config.yaml")
    
    // 方式 B：手动设置关键参数（如从环境变量读取）
    cfg := config.Get()
    cfg.Tools.WebSearch.Engine = "tavily"
    cfg.Tools.WebSearch.Tavily.ApiKey = os.Getenv("TAVILY_API_KEY")
    // 或使用 Google: cfg.Tools.WebSearch.Google.ApiKey = "..."
    cfg.Tools.Pinecone.ApiKey = os.Getenv("PINECONE_KEY")
}
```

### 4.3 获取并使用工具
您可以一次性获取全量工具，或者仅引入特定工具：

```go
import (
    "github.com/originaleric/digeino/tools"
    "github.com/originaleric/digeino/tools/research"
)

func useTools(ctx context.Context) {
    // 1. 获取所有通过 BaseTools 注册的工具
    allTools, _ := tools.BaseTools(ctx)
    
    // 2. 或者初始化特定工具
    grepTool, _ := research.NewGrepSearchTool(ctx)
    
    // 3. 在 Eino Agent (Graph) 中使用
    // graph.AddChatModelNode("model", model, compose.WithBindTools(allTools))
}
```

### 4.4 兼容性说明
所有的工具都返回标准的 `tool.BaseTool` 接口：
- **Schema 定义**：自动根据 Go 结构体生成 JSON Schema，模型可直接理解。
- **并发安全**：支持 context 控制和超时刻。
- **错误处理**：内部已进行安全封装，返回友好的错误提示供模型自我纠错。

---
**本工具库旨在为 Golang 生态提供一套开箱即用的「大模型工具箱」，助力开发者快速构建具备调研、设计与推送能力的智能体。**
