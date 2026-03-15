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
| `browser_browse` | **本地浏览器通用访问工具**。使用 go-rod + stealth 浏览器访问动态网页，支持 read/screenshot、wait_selector 与 cookie 域复用。**新增功能**：支持快照模式、摘要模式、标签页复用。 | `url`: 目标网页地址; `action`: read/screenshot; `mode`: full/snapshot/summary（提取模式）; `tab_id`: 标签页ID（复用）; `wait_selector`: 等待选择器; `content_selector`: 内容选择器; `use_cookie_domain`: Cookie域名 |
| `browser_snapshot` | **浏览器页面快照**。获取页面结构化快照，提取可交互元素信息（按钮、链接、输入框等），返回元素引用ID（e0, e1, e2等）和属性信息。**Token高效**：每页约800 tokens（相比完整HTML降低5-13倍）。 | `url`: 目标网页地址; `filter`: 过滤类型（interactive/visible/all，默认interactive）; `max_depth`: 最大深度（-1表示不限制）; `wait_selector`: 等待选择器; `use_cookie_domain`: Cookie域名 |
| `browser_action` | **浏览器交互操作**。执行点击、输入、填充、悬停、滚动、聚焦、按键等操作。可通过元素引用（ref）或CSS选择器定位元素。**支持**：humanClick/humanType（模拟人类操作，随机延迟）。 | `url`: 目标网页地址; `ref`: 元素引用ID（优先使用）; `selector`: CSS选择器; `action`: 操作类型（click/type/fill/hover/scroll/focus/press）; `text`: 输入文本（type/fill需要）; `key`: 按键名称（press需要，如enter/tab/escape/arrowup等）; `scroll_x`/`scroll_y`: 滚动偏移; `human_like`: 是否模拟人类操作; `use_cookie_domain`: Cookie域名 |

**web_search 引擎选择**：

| 引擎 | 特点 | 配置 |
|------|------|------|
| Bocha | 国内服务，中文友好 | `Tools.WebSearch.Bocha.ApiKey` |
| Firecrawl | 与 scrape 共用 ApiKey | `Tools.Firecrawl.ApiKey` |
| **Tavily** | 面向 AI、低延迟、英文/国际搜索 | `Tools.WebSearch.Tavily.ApiKey` 或 `TAVILY_API_KEY` |
| SerpApi/Google | 依赖第三方 API | 各自 ApiKey |

**浏览器工具使用建议**：

推荐的使用优先级（阶梯式抓取策略）：
1. **L1**：`firecrawl_scrape`（轻量级 HTTP 抓取）
2. **L2**：`research_jina_reader`（高级 HTTP 抓取，反爬兼容）
3. **L3**：`browser_snapshot`（Token 高效，获取页面结构）
4. **L3.5**：`browser_browse`（完整内容提取、截图、Cookie 复用）
5. **L4**：`web_search`（搜索兜底）

**浏览器工具配合使用流程**：
1. 使用 `browser_snapshot` 获取页面元素快照和引用ID（e0, e1, e2...）
2. 使用 `browser_action` 通过引用ID或选择器执行操作（点击、输入等）
3. 使用 `browser_browse` 获取页面内容（支持快照模式、摘要模式优化 Token）

**浏览器工具详细说明**：

#### browser_browse（增强版浏览器浏览工具）

**功能**：使用本地 go-rod + stealth 浏览器访问动态网页，支持多种提取模式和标签页复用。

**新增功能**：
- `mode` 参数：支持三种提取模式
  - `full`：完整内容提取（默认），返回 Text 和 Markdown
  - `snapshot`：快照模式，仅返回可交互元素（~800 tokens）
  - `summary`：摘要模式，返回页面摘要
- `tab_id` 参数：支持标签页复用，减少浏览器启动开销
- `content_selector` 参数：指定提取内容的范围（`mode=full` 时有效）

**使用示例**：
```json
// 完整内容提取（默认）
{"url": "https://example.com/article", "action": "read"}

// 快照模式（Token 优化）
{"url": "https://example.com/article", "action": "read", "mode": "snapshot"}

// 摘要模式
{"url": "https://example.com/article", "action": "read", "mode": "summary"}

// 指定内容范围
{"url": "https://example.com/article", "mode": "full", "content_selector": ".article-content"}

// 复用标签页
{"url": "https://example.com/page2", "tab_id": "tab_1234567890"}
```

#### browser_snapshot（浏览器快照工具）

**功能**：获取页面结构化快照，提取可交互元素信息，返回元素引用ID（e0, e1, e2等）。

**特点**：
- **Token 高效**：每页约 800 tokens（相比完整 HTML 降低 5-13 倍）
- **稳定引用**：基于无障碍树生成稳定的元素引用，避免 CSS 选择器失效
- **结构化输出**：返回元素角色、名称、状态等结构化信息

**使用示例**：
```json
// 获取可交互元素（默认）
{"url": "https://example.com/login", "filter": "interactive"}

// 等待元素后获取快照
{"url": "https://example.com", "wait_selector": "#main-content"}

// 使用 Cookie 复用登录态
{"url": "https://example.com/dashboard", "use_cookie_domain": "example.com"}
```

**返回示例**：
```json
{
  "url": "https://example.com/login",
  "title": "Login Page",
  "elements": [
    {"ref": "e0", "role": "textbox", "name": "Email", "disabled": false},
    {"ref": "e1", "role": "textbox", "name": "Password", "disabled": false},
    {"ref": "e2", "role": "button", "name": "Sign In", "disabled": false}
  ]
}
```

#### browser_action（浏览器操作工具）

**功能**：执行浏览器交互操作，支持 7 种操作类型。

**支持的操作**：
- `click`：点击元素
- `type`：输入文本（需要 `text` 参数）
- `fill`：填充表单（清空后输入，需要 `text` 参数）
- `hover`：鼠标悬停
- `scroll`：滚动（可指定 `scroll_x`、`scroll_y`）
- `focus`：聚焦元素
- `press`：按键（需要 `key` 参数）

**元素定位**：
- 优先使用 `ref`（元素引用ID，如 e0, e1, e2）
- `ref` 不可用时使用 `selector`（CSS 选择器）

**使用示例**：
```json
// 点击按钮（使用 selector）
{"url": "https://example.com/login", "selector": "#submit-btn", "action": "click", "human_like": true}

// 填写表单（使用 ref）
{"url": "https://example.com/login", "ref": "e0", "action": "fill", "text": "user@example.com"}

// 按键操作
{"url": "https://example.com/search", "selector": "input[name='q']", "action": "press", "key": "enter"}

// 滚动页面
{"url": "https://example.com/article", "action": "scroll", "scroll_y": 800}
```

**按键支持**：
- `enter`、`tab`、`escape`、`backspace`、`delete`
- `arrowup`、`arrowdown`、`arrowleft`、`arrowright`

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
| `receive_wecom_customer_message` | **接收客服消息**。接收个人微信用户发送给企业微信机器人应用的消息。支持实时回调模式和主动拉取模式。 | `mode`: 模式（realtime/pull）; `open_kf_id`: 客服账号ID（可选）; `limit`: 拉取数量（可选） |

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
    
    // 浏览器工具配置（如果使用）
    cfg.Tools.LocalBrowser.Enabled = true
    cfg.Tools.LocalBrowser.MaxConcurrency = 3
    cfg.Tools.LocalBrowser.Headless = true
}
```

**浏览器工具配置示例**（`config/config.yaml`）：

```yaml
Tools:
  LocalBrowser:
    Enabled: true
    MaxConcurrency: 3
    TotalTimeoutSec: 60
    NavigateTimeoutSec: 30
    WaitSelectorTimeoutSec: 20
    AllowedDomains: []  # 为空表示不限制
    ChromePath: ""       # 留空自动查找 Chromium
    Headless: true
    CookieStoreDir: "storage/app/browser_cookies"
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
