# UI/UX Design System Agent

基于 ui_ux_pro_max v2.0 的设计系统生成 Agent，支持完整的工作流编排和工具调用。

## 功能特性

### 1. 核心工具

- **`ui_ux_search`**: 搜索 UI/UX 设计知识库
- **`generate_design_system`**: 生成完整设计系统（包含推理引擎）
- **`persist_design_system`**: 持久化设计系统到文件（Master + Overrides 模式）

### 2. 审查与标准化工具

- **`ui_ux_audit`**: 技术质量审查工具（可访问性、性能、主题、响应式、反模式）
- **`ui_ux_critique`**: UX 设计审查工具（AI Slop 检测、视觉层次、信息架构等）
- **`ui_ux_normalize`**: 设计系统标准化工具
- **`ui_ux_reference`**: 参考文档检索工具（7 个领域的最佳实践）

### 2. Agent 编排

- **`UIDesignSystemAgent`**: 智能编排的 Agent，自动完成设计系统生成工作流
- **`NewUIDesignSystemAgentTool`**: 将 Agent 包装成工具，供其他 Agent 使用

## 使用方法

### 方式一：直接使用工具

```go
import (
    "context"
    "github.com/originaleric/digeino/tools/ui_ux"
)

// 1. 搜索工具
searchTool, _ := ui_ux.NewUIUXSearchTool(ctx)
// 使用 searchTool...

// 2. 生成设计系统
designSystemTool, _ := ui_ux.NewGenerateDesignSystemTool(ctx)
// 使用 designSystemTool...

// 3. 持久化设计系统
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)
// 使用 persistTool...
```

### 方式二：使用 Agent（推荐）

**方式 2a：从配置创建（最简单，推荐）**

```go
import (
    "context"
    "github.com/originaleric/digeino/config"
    "github.com/originaleric/digeino/tools/ui_ux"
)

// 1. 加载配置（通常在应用启动时执行一次）
_, err := config.Load("config/config.yaml")
if err != nil {
    return err
}

// 2. 从配置创建 Agent（自动读取 config.yaml 中的 ChatModel 配置）
agent, err := ui_ux.NewUIDesignSystemAgentFromConfig(ctx)
if err != nil {
    return err
}

// 3. 调用 Agent
result, err := agent.Invoke(ctx, "生成一个美容spa的设计系统，项目名称：Serenity Spa")
if err != nil {
    return err
}

fmt.Println(result.Content)
```

**方式 2b：手动传入 ChatModel（向后兼容）**

```go
import (
    "context"
    "github.com/originaleric/digeino/tools/ui_ux"
    "github.com/cloudwego/eino/components/model"
    openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
)

// 1. 创建 ChatModel
chatModel, err := openaiModel.NewChatModel(ctx, &openaiModel.ChatModelConfig{
    BaseURL: "https://api.openai.com/v1",
    APIKey:  "your-api-key",
    Model:   "gpt-4",
})

// 2. 创建 Agent
agent, err := ui_ux.NewUIDesignSystemAgent(ctx, chatModel)
if err != nil {
    return err
}

// 3. 调用 Agent
result, err := agent.Invoke(ctx, "生成一个美容spa的设计系统，项目名称：Serenity Spa")
if err != nil {
    return err
}

fmt.Println(result.Content)
```

### 方式三：将 Agent 作为工具使用

**方式 3a：从配置创建（推荐）**

```go
import (
    "context"
    "github.com/originaleric/digeino/config"
    "github.com/originaleric/digeino/tools/ui_ux"
    "github.com/cloudwego/eino/components/tool"
)

// 1. 加载配置（通常在应用启动时执行一次）
_, err := config.Load("config/config.yaml")
if err != nil {
    return err
}

// 2. 从配置创建 Agent 工具（自动读取 config.yaml 中的 ChatModel 配置）
agentTool, err := ui_ux.NewUIDesignSystemAgentToolFromConfig(ctx)
if err != nil {
    return err
}

// 3. 在其他 Agent 中使用
tools := []tool.BaseTool{
    agentTool, // Agent 作为工具
    // 其他工具...
}

// 4. 在其他 Agent 中，LLM 会自动调用这个工具
```

**方式 3b：手动传入 ChatModel（向后兼容）**

```go
import (
    "context"
    "github.com/originaleric/digeino/tools/ui_ux"
    "github.com/cloudwego/eino/components/model"
)

// 1. 创建 ChatModel
chatModel, _ := createChatModel(ctx)

// 2. 将 Agent 包装成工具
agentTool, err := ui_ux.NewUIDesignSystemAgentTool(ctx, chatModel)
if err != nil {
    return err
}

// 3. 在其他 Agent 中使用
tools := []tool.BaseTool{
    agentTool, // Agent 作为工具
    // 其他工具...
}

// 4. 在其他 Agent 中，LLM 会自动调用这个工具
```

## Agent 工作流

Agent 内部自动执行以下工作流：

1. **分析需求** - 提取产品类型、样式关键词、行业、技术栈
2. **生成设计系统** - 调用 `generate_design_system` 工具
3. **补充搜索** - 根据需要调用 `ui_ux_search` 获取详细信息
4. **技术栈指南** - 如果指定了技术栈，获取相关指南
5. **持久化** - 如果用户请求，调用 `persist_design_system` 保存到文件

## 设计系统生成流程

参考 ui_ux_pro_max 的 "How Design System Generation Works"：

```
1. 用户请求
   ↓
2. 多领域并行搜索（5个领域）
   - Product type matching (100 categories)
   - Style recommendations (67 styles)
   - Color palette selection (96 palettes)
   - Landing page patterns (24 patterns)
   - Typography pairing (57 font combinations)
   ↓
3. 推理引擎
   - Match product → UI category rules
   - Apply style priorities (BM25 ranking)
   - Filter anti-patterns for industry
   - Process decision rules (JSON conditions)
   ↓
4. 完整设计系统输出
   - Pattern + Style + Colors + Typography + Effects
   - Anti-patterns to avoid
   - Pre-delivery checklist
```

## 持久化功能

设计系统可以持久化到文件系统，使用 Master + Overrides 模式。

### 存储结构

**默认存储路径**（未指定 AppName）：
```
{BaseDir}/design-system/{project-slug}/
├── MASTER.md           # 全局设计系统
└── pages/
    └── {page-name}.md  # 页面特定覆盖
```

**应用隔离存储路径**（指定 AppName）：
```
{BaseDir}/{app-name}/design-system/{project-slug}/
├── MASTER.md           # 全局设计系统
└── pages/
    └── {page-name}.md  # 页面特定覆盖
```

**配置说明**：
- `BaseDir` 默认从 `config.yaml` 的 `UIUX.Storage.BaseDir` 读取，默认为 `storage/app/ui_ux`
- `AppName` 用于隔离不同应用/agent 的存储，避免冲突
- 如果未指定 `AppName`，所有应用共享同一存储空间

**层次化检索逻辑**：
1. 构建页面时，首先检查 `design-system/pages/[page-name].md`
2. 如果页面文件存在，其规则**覆盖** Master 文件
3. 如果不存在，使用 `design-system/MASTER.md` 的规则

### 多应用/Agent 隔离

当多个应用或 Agent 调用 UI/UX 工具时，建议使用 `AppName` 参数进行隔离：

```go
// 应用 A（Travel Agent）
persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "travel booking platform",
    ProjectName: "TravelApp",
    AppName:     "travel",  // 隔离存储
})

// 应用 B（Stock Agent）
persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "stock analysis dashboard",
    ProjectName: "StockApp",
    AppName:     "stock",  // 隔离存储
})
```

存储结果：
- Travel: `storage/app/ui_ux/travel/design-system/travelapp/MASTER.md`
- Stock: `storage/app/ui_ux/stock/design-system/stockapp/MASTER.md`

## 推理引擎

Agent 使用 100 条行业特定推理规则（来自 `ui-reasoning.csv`），包括：

- SaaS、E-commerce、Healthcare、Fintech 等 100 个行业类别
- 每个类别包含：推荐模式、样式优先级、配色情绪、字体情绪、关键效果、反模式等

## 示例

### 示例 1：生成设计系统

```go
agent, _ := ui_ux.NewUIDesignSystemAgent(ctx, chatModel)
result, _ := agent.Invoke(ctx, "beauty spa wellness service")
// Agent 会自动：
// 1. 识别产品类型：Beauty/Spa/Wellness Service
// 2. 匹配推理规则
// 3. 生成完整设计系统
// 4. 返回格式化的结果
```

### 示例 2：持久化设计系统（默认路径）

```go
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)
result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "beauty spa wellness service",
    ProjectName: "Serenity Spa",
    PageName:    "homepage",
})
// 创建文件：
// - storage/app/ui_ux/design-system/serenity-spa/MASTER.md
// - storage/app/ui_ux/design-system/serenity-spa/pages/homepage.md
```

### 示例 3：持久化设计系统（应用隔离）

```go
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)
result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "beauty spa wellness service",
    ProjectName: "Serenity Spa",
    PageName:    "homepage",
    AppName:     "beauty_app",  // 应用隔离
})
// 创建文件：
// - storage/app/ui_ux/beauty_app/design-system/serenity-spa/MASTER.md
// - storage/app/ui_ux/beauty_app/design-system/serenity-spa/pages/homepage.md
```

### 示例 4：自定义存储路径

```go
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)
result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "beauty spa wellness service",
    ProjectName: "Serenity Spa",
    BaseDir:     "/custom/path/to/storage",  // 自定义基础目录
    AppName:     "beauty_app",
})
// 创建文件：
// - /custom/path/to/storage/beauty_app/design-system/serenity-spa/MASTER.md
```

## 配置说明

### UI/UX 临时文件与工作空间集成

除了上述「设计系统持久化」到 `storage/app/ui_ux` 之外，DigEino 提供**统一的临时存储能力**（`write_review_file` 工具 + `tempstorage.SaveForReview` API），供 `ui_ux_audit` / `ui_ux_critique` 审查一次性生成的 CSS/HTML/代码。引用方只需在调用时注入 Context，无需各自实现存储逻辑。

#### 1. Context 键与配置

**Context 键约定**（与 DigFlow 兼容）：

| 键名 | 模式 | 说明 |
|------|------|------|
| `workspace_path` | 工作空间 | DigFlow 已在使用，每次执行注入 workspace 根路径 |
| `agent_session_id` | 会话 | DigPdf 等应用在调用 Agent 前注入 sessionID |

**配置项**（`config.yaml`）：

```yaml
UIUX:
  TempStorage:
    BaseDir: "storage/temp"   # 会话模式下的临时根目录，默认 storage/temp
```

- `BaseDir` 支持相对路径（相对于进程 cwd）或绝对路径
- 会话模式：`{BaseDir}/agent/{session_id}/{filename}`

**程序化 API**（供 DigPdf 等应用在代码中调用）：

```go
import "github.com/originaleric/digeino/pkg/tempstorage"

path, err := tempstorage.SaveForReview(ctx, "designer_output.css", cssContent)
// 返回绝对路径，可传入 ui_ux_audit / ui_ux_critique
```

**LLM 工具**：`write_review_file`（已通过 `tools.BaseTools()` 注册）

- 参数：`filename`（必填）、`content`（必填）
- 行为：根据 context 注入的 `workspace_path` 或 `agent_session_id` 写入文件，返回绝对路径
- 与 `ui_ux_audit`、`ui_ux_critique` 配合：先调用 `write_review_file` 得到 path，再传入审查工具

#### 2. 上层应用的典型集成方式

- **DigPdf**：
  - 在调用 Agent 的 context 中注入：`ctx = context.WithValue(ctx, "agent_session_id", sessionID)`
  - 程序化写入：在 Designer/Assembler 节点中调用 `tempstorage.SaveForReview(ctx, filename, content)` 替代自有的 `saveAgentTempFile`
  - LLM 审查：若需模型自行写入再审查，可调用 `write_review_file` 得到 path 后传给 audit/critique
  - `ui_ux_reference` 仍然是**纯检索工具**，无需路径参数

- **DigFlow（UI Pro）**：
  - 已注入 `workspace_path`，无需改动
  - 可选择 DigEino 的 `write_review_file` 替代自有的 `write_file`（仅针对审查流程），或保留现有 `write_file`，两种方式均可工作

#### 3. 各工具在临时/持久存储下的推荐用法

| 工具 | 是否需要 path | 典型用法 |
|------|---------------|----------|
| `ui_ux_reference` | 否 | 在 DigPdf、DigFlow 等应用中直接调用，用于检索 typography/color/spatial/motion/interaction/responsive/ux_writing 等领域的参考文档 |
| `write_review_file` | 否（从 context 解析） | 将内容写入临时/工作空间，返回 path，供 audit/critique 使用 |
| `ui_ux_audit` | 是（必填） | 将生成的 CSS/HTML/代码写入临时文件或 workspace，再将路径传入，做技术质量审查（可访问性、性能、主题、响应式等） |
| `ui_ux_critique` | 是（必填） | 同上，用于 UX 设计层面的审查（视觉层次、信息架构、交互流等） |
| `ui_ux_normalize` | 是（必填） | 更适合与已经持久化的设计系统文件配合，做长期资产的标准化；对一次性临时输出的收益相对较低 |

综合来看：

- 推荐将 **设计系统本身** 通过 `persist_design_system` 持久化到 `storage/app/ui_ux(...)/design-system/**`。
- 将 **具体页面实现/组件代码/一次性输出** 通过 `write_review_file` 或 `tempstorage.SaveForReview` 写入，再传入 `ui_ux_audit` / `ui_ux_critique` 审查。

### UI/UX 存储配置

在 `config/config.yaml` 中配置存储路径：

```yaml
UIUX:
  Storage:
    BaseDir: "storage/app/ui_ux"     # 设计系统存储基础目录
    # 存储结构：
    # - 如果指定 AppName: {BaseDir}/{app-name}/design-system/{project}/MASTER.md
    # - 如果未指定: {BaseDir}/design-system/{project}/MASTER.md
    # 示例：
    # - AppName="travel": storage/app/ui_ux/travel/design-system/my-project/MASTER.md
    # - 未指定: storage/app/ui_ux/design-system/my-project/MASTER.md
  TempStorage:
    BaseDir: "storage/temp"          # 临时存储根目录（供 audit/critique 使用，会话模式）
```

**存储路径优先级**：
1. 如果调用时指定了 `BaseDir`，使用指定的路径
2. 否则，从配置文件的 `UIUX.Storage.BaseDir` 读取
3. 如果配置文件中也没有，使用默认值 `storage/app/ui_ux`

**应用隔离**：
- 如果调用时指定了 `AppName`，存储路径会包含应用名称，实现隔离
- 不同应用的设计系统不会互相干扰
- 适合多租户或多应用场景

### ChatModel 配置（参考 DigFlow 配置方式）

在 `config/config.yaml` 中配置 ChatModel：

```yaml
ChatModel:
  Type: "qwen"                      # 模型类型: qwen, openai
  Config:
    ApiKey: "${QWEN_API_KEY}"        # API Key（支持环境变量 ${VAR_NAME}）
    Model: "qwen-max"                 # 模型名称
    BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"  # API 地址
    Temperature: 0.7                 # 可选：温度参数
    MaxTokens: 2048                  # 可选：最大 token 数
```

**使用 OpenAI**：

```yaml
ChatModel:
  Type: "openai"
  Config:
    ApiKey: "${OPENAI_API_KEY}"
    Model: "gpt-4"
    BaseUrl: "https://api.openai.com/v1"
    Temperature: 0.7
```

**切换模型提供商**：只需修改 `Type` 字段，例如：
- `Type: "qwen"` - 使用 Qwen 模型
- `Type: "openai"` - 使用 OpenAI 模型

**环境变量支持**：配置中的值支持 `${VAR_NAME}` 格式，会自动从环境变量中读取，例如：
- `ApiKey: "${QWEN_API_KEY}"` - 从环境变量 `QWEN_API_KEY` 读取

## 注意事项

1. **配置加载**：使用 `NewUIDesignSystemAgentFromConfig` 前需要先调用 `config.Load()`
2. **ChatModel 必需**：Agent 功能需要 ChatModel，可以通过配置或手动传入
3. **工具注册**：基础工具已自动注册到 `tools.BaseTools()`
4. **Agent 工具**：可以使用 `NewUIDesignSystemAgentToolFromConfig` 从配置创建
5. **数据文件**：`ui-reasoning.csv` 包含 100 条推理规则，需要定期同步更新
6. **API Key 安全**：建议使用环境变量或加密存储 API Key

## 参考

- [ui_ux_pro_max 文档](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill)
- DigPdf Agent 实现参考：`/Users/dig/Documents/文稿 - XinYe的MacBook Pro (5)/Projects/go-app/DigPdf/eino/pdf_agent.go`
