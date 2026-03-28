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

### 3. 预览与结构化 Patch（宿主 iframe / Sandpack）

- **`write_preview_manifest`**: 在工作区或会话目录写入 `preview-manifest.json`，描述 `entry`、`assets`、`editable_model`、初始 `revision`（需 context 注入 `workspace_path` 或 `agent_session_id`）
- **`apply_preview_patch`**: 按清单应用结构化补丁（`html_text`、`html_attr`、`html_inner`、`json_pointer`、`literal_replace`）；`base_revision` 必须与当前 manifest 一致，成功则 `revision` 自增
- **`export_preview_bundle`**: 将 manifest 涉及的文件打成 zip 并写回工作区/会话目录

设计说明见 [docs/ideas/2026-03-28_UI预览与用户编辑工具方案.md](../../docs/ideas/2026-03-28_UI预览与用户编辑工具方案.md)。

### 4. Agent 编排

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

// 4. 预览清单 / 打补丁 / 导出（需 ctx 注入 workspace_path 或 agent_session_id）
manifestTool, _ := ui_ux.NewWritePreviewManifestTool(ctx)
patchTool, _ := ui_ux.NewApplyPreviewPatchTool(ctx)
exportTool, _ := ui_ux.NewExportPreviewBundleTool(ctx)
// 使用方式见下文「预览产物工作流」
```

### 预览产物工作流（manifest / Patch / zip）

与 `write_review_file` 相同：**必须在 context 中注入** `workspace_path`（DigFlow）或 `agent_session_id`（DigPdf 等会话目录）。以下为推荐顺序。

1. **`write_review_file`**：写入页面片段或完整 HTML（建议关键区块带 `data-uiux-id`）、可选 `preview/content.json` 等。
2. **`write_preview_manifest`**：登记 `entry`、`assets`、`editable_model` 与初始 `revision`（默认从 1 开始）。
3. **宿主预览**：根据 manifest 的 `entry` 拉取文件做 iframe / Sandpack 展示；用户编辑后组装补丁。
4. **`apply_preview_patch`**：提交 `patches`；**`base_revision` 必须与当前 manifest 的 `revision` 一致**（乐观锁），成功后服务端将 `revision` 加 1。
5. **`export_preview_bundle`**（可选）：打包 entry、assets、editable_model 与 manifest 为 zip。

设计背景见 [docs/ideas/2026-03-28_UI预览与用户编辑工具方案.md](../../docs/ideas/2026-03-28_UI预览与用户编辑工具方案.md)。

#### LLM 工具一览（已随 `tools.BaseTools()` 注册）

| 工具名 | 主要参数 | 说明 |
|--------|----------|------|
| `write_preview_manifest` | `kind`（如 `static_html` / `react_sandpack`）、`entry`；可选 `manifest_filename`、`assets[]`、`editable_model`、`artifact_id`、`initial_revision` | 默认写入 `preview/preview-manifest.json`；未传 `artifact_id` 时自动生成 UUID |
| `apply_preview_patch` | `manifest_path` 或 `artifact_id`、`base_revision`、`patches[]` | 对清单内允许的文件应用结构化补丁；支持按 `artifact_id` 自动定位 manifest |
| `export_preview_bundle` | `manifest_path` 或 `artifact_id`；可选 `zip_filename`（默认 `preview/bundle.zip`） | 写出 zip 的**绝对路径**，供下载或后续处理 |

#### 补丁 `patches[].type` 说明

| type | 必填字段 | 作用 |
|------|----------|------|
| `html_text` | `selector`, `text` | 匹配 `entry` 指向的 HTML；将元素**文本内容**替换为 `text`（经转义，不含富文本标签） |
| `html_attr` | `selector`, `attr` 及 `value` 或 `text` | 设置属性（如 `src`、`alt`）；优先使用 `value` |
| `html_inner` | `selector`, `html` | 设置匹配元素的 inner HTML（仅建议使用可信内容） |
| `json_pointer` | `pointer`（如 `/hero/title`）、`value` | 要求 manifest 已配置 `editable_model`；按 JSON 路径写入（支持一般对象/数组路径） |
| `literal_replace` | `file`, `old`, `new` | `old` 在全文中须**唯一**匹配；`file` 须出现在 manifest 的 entry/assets/editable_model 中；后缀白名单由 `UIUX.Preview.AllowedExtensions` 控制（为空时使用默认值） |

#### `preview-manifest.json` 示例

```json
{
  "artifact_id": "550e8400-e29b-41d4-a716-446655440000",
  "kind": "static_html",
  "revision": 1,
  "entry": "preview/index.html",
  "assets": ["preview/logo.png"],
  "editable_model": "preview/content.json"
}
```

#### 程序化 API（宿主后端可直接调用，与工具逻辑一致）

```go
import (
    "context"

    "github.com/originaleric/digeino/pkg/tempstorage"
    "github.com/originaleric/digeino/tools/ui_ux"
)

ctx := context.WithValue(context.Background(), tempstorage.ContextKeyWorkspacePath, workspaceRoot)

res, err := ui_ux.ApplyPreviewPatches(ctx, "preview/preview-manifest.json", currentRevision, []ui_ux.PreviewPatch{
    {Type: "html_text", Selector: "[data-uiux-id='hero-title']", Text: "新标题"},
})
if err != nil { /* 处理冲突 revision mismatch 等 */ }
nextRev := res.Revision

_, err = ui_ux.WritePreviewZIPToFile(ctx, "preview/preview-manifest.json", "preview/export.zip")
// 或使用 BuildPreviewZIP 自行处理字节流
```

zip 二进制写入使用 **`tempstorage.SaveBytesForReview`**（与 `SaveForReview` 相同的路径规则）。

#### 历史快照（history）

- `apply_preview_patch` 在改写目标文件前会按当前 revision 生成快照，默认目录：
  - `preview/history/{artifact_id}/rev-000001/...`
- 控制项：
  - `UIUX.Preview.HistoryEnabled`：是否启用（未设置时默认启用）
  - `UIUX.Preview.HistoryDir`：快照根目录（默认 `preview/history`）
- manifest 文件本身也会在升级 revision 前保存快照，便于后续回滚。

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
6. **（可选）预览链路** - 由上层 Prompt 引导：生成 HTML/React 后用 `write_review_file`，再 `write_preview_manifest`；用户微调后 `apply_preview_patch`，需要交付压缩包时用 `export_preview_bundle`

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
| `write_preview_manifest` | 否 | 写入 `preview-manifest.json`，登记可预览 entry 与资源路径 |
| `apply_preview_patch` | 否 | 对 entry / editable_model 等应用结构化补丁（乐观锁 `base_revision`） |
| `export_preview_bundle` | 否 | 将 manifest 涉及文件导出为 zip |

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
  Preview:
    MaxPatchFileBytes: 0             # 预览 Patch 单文件大小上限（字节），0 为默认 5MiB
    AllowedExtensions: [".tsx", ".jsx", ".ts", ".js", ".html", ".htm", ".json", ".css", ".md"] # literal_replace 白名单
    HistoryEnabled: true             # 是否保留补丁前快照（未设置时默认启用）
    HistoryDir: "preview/history"    # 快照目录（相对 workspace/session 根目录）
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
