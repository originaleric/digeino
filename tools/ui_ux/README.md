# UI/UX Design System Agent

基于 ui_ux_pro_max v2.0 的设计系统生成 Agent，支持完整的工作流编排和工具调用。

## 功能特性

### 1. 核心工具

- **`ui_ux_search`**: 搜索 UI/UX 设计知识库
- **`generate_design_system`**: 生成完整设计系统（包含推理引擎）
- **`persist_design_system`**: 持久化设计系统到文件（Master + Overrides 模式）

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
