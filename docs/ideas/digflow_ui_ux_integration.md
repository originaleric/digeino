# 在 DigFlow 中使用 DigEino UI/UX 工具

本指南说明如何在 DigFlow 项目中使用 DigEino 的 UI/UX 工具。

## 目录

- [前置条件](#前置条件)
- [快速开始](#快速开始)
- [使用方式](#使用方式)
  - [方式一：单独注册工具](#方式一单独注册工具)
  - [方式二：批量注册所有工具](#方式二批量注册所有工具)
  - [方式三：使用 UI/UX Agent 工具（推荐）](#方式三使用-uiux-agent-工具推荐)
- [Agent 工具详解](#agent-工具详解)
- [配置说明](#配置说明)
- [使用示例](#使用示例)
- [故障排查](#故障排查)
- [参考](#参考)

## 前置条件

1. **DigFlow 项目已存在**：确保你已经有一个 DigFlow 项目
2. **DigEino 依赖已添加**：DigFlow 的 `go.mod` 中应该已经包含 DigEino 依赖

```go
require (
    github.com/originaleric/digeino v1.0.2  // 或更新版本
)
```

## 快速开始

### 最简单的使用方式

**1. 添加依赖**（如果还没有）

```bash
go get github.com/originaleric/digeino
```

**2. 注册工具**

在你的 DigFlow 应用的 `app/your_app/tools/tool_register.go` 中：

```go
package tools

import (
    "context"
    "DigFlow/internal/global/variable"
    "github.com/originaleric/digeino/tools/ui_ux"
)

func init() {
    ctx := context.Background()
    
    // 注册 UI/UX 搜索工具
    tool, _ := ui_ux.NewUIUXSearchTool(ctx)
    variable.ToolRegistry.Register("ui_ux_search", tool)
}
```

**3. 在 Agent 中使用**

在你的 Agent 配置 `app/your_app/config/agent.yml` 中：

```yaml
Tools:
  - Name: "ui_ux_search"
    Description: "检索 UI/UX 设计知识库"

Nodes:
  - Key: "designer"
    Type: "chat_model"
    BindTools:
      - "ui_ux_search"
```

## 使用方式

### 方式一：单独注册工具

适合只需要特定工具的场景。

#### 步骤 1：创建或编辑工具注册文件

在你的 DigFlow 应用中创建 `app/your_app/tools/tool_register.go`：

```go
package tools

import (
    "context"
    
    "DigFlow/internal/global/variable"
    "DigFlow/internal/utils/logger"
    
    "github.com/originaleric/digeino/tools/ui_ux"
    "go.uber.org/zap"
)

func init() {
    ctx := context.Background()
    
    // 注册 UI/UX 搜索工具
    searchTool, err := ui_ux.NewUIUXSearchTool(ctx)
    if err != nil {
        logger.Error("创建 UI/UX 搜索工具失败", zap.Error(err))
        return
    }
    if err := variable.ToolRegistry.Register("ui_ux_search", searchTool); err != nil {
        logger.Warn("UI/UX 搜索工具已存在", zap.String("name", "ui_ux_search"))
    } else {
        logger.Info("UI/UX 搜索工具注册成功", zap.String("name", "ui_ux_search"))
    }
    
    // 注册设计系统生成工具
    designSystemTool, err := ui_ux.NewGenerateDesignSystemTool(ctx)
    if err != nil {
        logger.Error("创建设计系统生成工具失败", zap.Error(err))
        return
    }
    if err := variable.ToolRegistry.Register("generate_design_system", designSystemTool); err != nil {
        logger.Warn("设计系统生成工具已存在", zap.String("name", "generate_design_system"))
    } else {
        logger.Info("设计系统生成工具注册成功", zap.String("name", "generate_design_system"))
    }
    
    // 注册设计系统持久化工具
    persistTool, err := ui_ux.NewPersistDesignSystemTool(ctx)
    if err != nil {
        logger.Error("创建设计系统持久化工具失败", zap.Error(err))
        return
    }
    if err := variable.ToolRegistry.Register("persist_design_system", persistTool); err != nil {
        logger.Warn("设计系统持久化工具已存在", zap.String("name", "persist_design_system"))
    } else {
        logger.Info("设计系统持久化工具注册成功", zap.String("name", "persist_design_system"))
    }
}
```

#### 步骤 2：在 bootstrap 中导入工具包

在 `bootstrap/your_app/init.go` 中导入工具包：

```go
package yourappbootstrap

import (
    "DigFlow/bootstrap/common"
    _ "DigFlow/app/your_app/tools"  // 导入工具包，触发 init() 注册
    // ... 其他导入
)

func init() {
    common.InitCommon()
    // ... 其他初始化逻辑
}
```

#### 步骤 3：在 agent.yml 中配置工具

在你的应用配置 `app/your_app/config/agent.yml` 中添加工具：

```yaml
Tools:
  - Name: "ui_ux_search"
    Description: "检索 UI/UX 设计知识库"
    UseGlobal: false
  
  - Name: "generate_design_system"
    Description: "生成完整设计系统"
    UseGlobal: false
  
  - Name: "persist_design_system"
    Description: "持久化设计系统到文件"
    UseGlobal: false

Nodes:
  - Key: "designer"
    Type: "chat_model"
    Model: "qwen-max"
    SystemPrompt: |
      你是一位专业的 UI/UX 设计师。
      你可以使用以下工具：
      - ui_ux_search: 搜索设计知识库
      - generate_design_system: 生成完整设计系统
      - persist_design_system: 持久化设计系统
    BindTools:
      - "ui_ux_search"
      - "generate_design_system"
      - "persist_design_system"
```

### 方式二：批量注册所有工具

如果你想要使用 DigEino 提供的所有基础工具（包括 UI/UX 工具和微信工具），可以使用 `BaseTools()` 函数：

```go
package tools

import (
    "context"
    
    "DigFlow/internal/global/variable"
    "DigFlow/internal/utils/logger"
    
    digeinoTools "github.com/originaleric/digeino/tools"
    "go.uber.org/zap"
)

func init() {
    ctx := context.Background()
    
    // 获取 DigEino 的所有基础工具
    baseTools, err := digeinoTools.BaseTools(ctx)
    if err != nil {
        logger.Error("获取 DigEino 基础工具失败", zap.Error(err))
        return
    }
    
    // 批量注册工具
    for _, tool := range baseTools {
        info, err := tool.Info(ctx)
        if err != nil {
            logger.Warn("获取工具信息失败", zap.Error(err))
            continue
        }
        
        if err := variable.ToolRegistry.Register(info.Name, tool); err != nil {
            logger.Warn("工具已存在", zap.String("name", info.Name))
        } else {
            logger.Info("工具注册成功", zap.String("name", info.Name))
        }
    }
}
```

**注意**：这种方式会注册 DigEino 的所有基础工具，包括：
- `ui_ux_search`
- `generate_design_system`
- `persist_design_system`
- `send_wechat_message`（如果配置了微信）

### 方式三：使用 UI/UX Agent 工具（推荐）⭐

**这是最推荐的方式**，Agent 工具会自动完成整个设计系统生成工作流，包括需求分析、生成、补充搜索和持久化。

#### 什么是 UI/UX Agent 工具？

`generate_ui_design_system` 是一个智能编排的 Agent 工具，它将整个设计系统生成工作流封装成一个工具。相比单独使用各个工具，Agent 工具的优势是：

- ✅ **自动工作流编排**：自动完成需求分析、生成、搜索、持久化
- ✅ **智能决策**：根据用户需求自动决定调用哪些工具
- ✅ **完整输出**：一次性返回完整的设计系统
- ✅ **简单易用**：只需一个工具调用，无需手动编排

#### 步骤 1：在 agent.yml 中配置模型

在你的应用配置 `app/your_app/config/agent.yml` 中确保有模型配置：

```yaml
Models:
  - Name: "ui_model"  # 模型名称，用于 Agent（必须与代码中的名称一致）
    Type: "qwen"      # 或 "openai"
    Config:
      api_key: "${QWEN_API_KEY}"
      model: "qwen-max"
      base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
```

#### 步骤 2：注册 Agent 工具（使用 DigFlow 的模型）⭐ 推荐

**推荐方式**：使用 DigFlow 的 ModelRegistry 获取 ChatModel

```go
package tools

import (
    "context"
    
    "DigFlow/internal/global/variable"
    "DigFlow/internal/utils/logger"
    
    "github.com/originaleric/digeino/tools/ui_ux"
    "go.uber.org/zap"
)

func init() {
    ctx := context.Background()
    
    // 从 DigFlow 的 ModelRegistry 获取 ChatModel
    // 注意：模型名称必须与 agent.yml 中配置的 Model Name 一致
    chatModel, err := variable.ModelRegistry.GetChatModel("ui_model")
    if err != nil {
        logger.Error("获取 ChatModel 失败", 
            zap.String("model_name", "ui_model"),
            zap.Error(err))
        return
    }
    
    // 创建 UI/UX Agent 工具
    agentTool, err := ui_ux.NewUIDesignSystemAgentTool(ctx, chatModel)
    if err != nil {
        logger.Error("创建 UI/UX Agent 工具失败", zap.Error(err))
        return
    }
    
    // 注册工具
    if err := variable.ToolRegistry.Register("generate_ui_design_system", agentTool); err != nil {
        logger.Warn("UI/UX Agent 工具已存在", zap.String("name", "generate_ui_design_system"))
    } else {
        logger.Info("UI/UX Agent 工具注册成功", zap.String("name", "generate_ui_design_system"))
    }
}
```

#### 步骤 3：在 agent.yml 中配置工具

```yaml
Tools:
  - Name: "generate_ui_design_system"
    Description: "生成完整的 UI/UX 设计系统（智能编排 Agent）"
    UseGlobal: false

Nodes:
  - Key: "designer"
    Type: "chat_model"
    Model: "ui_model"
    SystemPrompt: |
      你是一位专业的 UI/UX 设计师。
      当用户需要生成设计系统时，使用 generate_ui_design_system 工具。
      这个工具会自动完成：
      1. 分析需求（产品类型、样式、行业、技术栈）
      2. 生成完整设计系统
      3. 补充搜索详细信息
      4. 获取技术栈指南
      5. 持久化（如果用户请求）
    BindTools:
      - "generate_ui_design_system"

  - Key: "tools"
    Type: "tools_node"
    Tools:
      - "generate_ui_design_system"

Edges:
  - From: "START"
    To: "designer"
  - From: "tools"
    To: "designer"

Branches:
  - From: "designer"
    Condition: "has_tool_calls"
    Targets:
      - Key: "tools"
        When: true
      - Key: "END"
        When: false
```

#### 替代方式：使用 DigEino 配置创建 Agent

如果你想要使用 DigEino 的配置系统（而不是 DigFlow 的模型），可以这样做：

```go
package tools

import (
    "context"
    
    "DigFlow/internal/global/variable"
    "DigFlow/internal/utils/logger"
    
    "github.com/originaleric/digeino/config"
    "github.com/originaleric/digeino/tools/ui_ux"
    "go.uber.org/zap"
)

func init() {
    ctx := context.Background()
    
    // 加载 DigEino 配置
    // 注意：需要创建 DigEino 的配置文件 config/digeino.yaml
    _, err := config.Load("config/digeino.yaml")
    if err != nil {
        logger.Error("加载 DigEino 配置失败", zap.Error(err))
        return
    }
    
    // 从配置创建 Agent 工具（自动读取 DigEino 配置中的 ChatModel）
    agentTool, err := ui_ux.NewUIDesignSystemAgentToolFromConfig(ctx)
    if err != nil {
        logger.Error("创建 UI/UX Agent 工具失败", zap.Error(err))
        return
    }
    
    // 注册工具
    if err := variable.ToolRegistry.Register("generate_ui_design_system", agentTool); err != nil {
        logger.Warn("UI/UX Agent 工具已存在", zap.String("name", "generate_ui_design_system"))
    } else {
        logger.Info("UI/UX Agent 工具注册成功", zap.String("name", "generate_ui_design_system"))
    }
}
```

**DigEino 配置文件** (`config/digeino.yaml`):

```yaml
ChatModel:
  Type: "qwen"
  Config:
    ApiKey: "${QWEN_API_KEY}"
    Model: "qwen-max"
    BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"

UIUX:
  Storage:
    BaseDir: "storage/app/ui_ux"
```

## Agent 工具详解

### Agent 工具参数

Agent 工具接受以下参数：

```json
{
  "query": "beauty spa wellness service",      // 必需：用户需求描述
  "project_name": "Serenity Spa",              // 可选：项目名称
  "stack": "react",                            // 可选：技术栈（react, vue, tailwind 等）
  "persist": true,                             // 可选：是否持久化设计系统
  "page_name": "homepage"                      // 可选：页面名称（用于生成页面覆盖）
}
```

**参数说明**：
- **query**（必需）：用户需求描述，例如 "beauty spa wellness service"
- **project_name**（可选）：项目名称，用于生成设计系统和文件命名
- **stack**（可选）：技术栈，Agent 会自动获取该技术栈的指南
- **persist**（可选）：是否持久化设计系统到文件
- **page_name**（可选）：页面名称，用于生成页面特定的覆盖文件

### Agent 工作流

Agent 内部自动执行以下工作流：

```
1. 分析需求
   ↓
2. 生成设计系统（generate_design_system）
   ↓
3. 补充搜索（ui_ux_search，如果需要）
   - 样式选项
   - 图表推荐
   - UX 最佳实践
   - 字体选项
   ↓
4. 技术栈指南（ui_ux_search with stack，如果指定了技术栈）
   ↓
5. 持久化（persist_design_system，如果用户请求）
   ↓
6. 返回完整结果
```

### 与单独工具的区别

#### 使用单独工具（方式一、二）

```yaml
# 需要手动编排工作流
BindTools:
  - "ui_ux_search"
  - "generate_design_system"
  - "persist_design_system"

# 需要在 SystemPrompt 中详细说明调用顺序
SystemPrompt: |
  1. 先调用 generate_design_system
  2. 然后调用 ui_ux_search 补充信息
  3. 最后调用 persist_design_system
```

#### 使用 Agent 工具（方式三）⭐

```yaml
# 只需一个工具
BindTools:
  - "generate_ui_design_system"

# SystemPrompt 简单明了
SystemPrompt: |
  使用 generate_ui_design_system 生成设计系统
```

## 配置说明

### 配置存储路径（可选）

如果你想要自定义设计系统的存储路径，可以在 DigFlow 的配置中添加 DigEino 的配置：

#### 创建 DigEino 配置文件

在 DigFlow 项目中创建 `config/digeino.yaml`：

```yaml
ChatModel:
  Type: "qwen"
  Config:
    ApiKey: "${QWEN_API_KEY}"
    Model: "qwen-max"
    BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"

UIUX:
  Storage:
    BaseDir: "storage/app/ui_ux"  # 设计系统存储路径
```

#### 在工具注册时加载配置

```go
import (
    "github.com/originaleric/digeino/config"
)

func init() {
    // 加载 DigEino 配置
    _, err := config.Load("config/digeino.yaml")
    if err != nil {
        logger.Warn("加载 DigEino 配置失败，使用默认配置", zap.Error(err))
    }
    
    // ... 注册工具
}
```

## 使用示例

### 示例 1：使用 UI/UX Agent 工具（推荐）⭐

这是最简单的方式，Agent 会自动完成整个工作流：

```yaml
Nodes:
  - Key: "designer"
    Type: "chat_model"
    Model: "ui_model"
    SystemPrompt: |
      你是一位专业的 UI/UX 设计师。
      当用户需要生成设计系统时，使用 generate_ui_design_system 工具。
      这个工具会自动完成需求分析、生成设计系统、补充搜索和持久化。
    BindTools:
      - "generate_ui_design_system"
```

**用户请求示例**：
- "生成一个美容spa的设计系统，项目名称：Serenity Spa"
- "为我的电商平台生成设计系统，技术栈：react"
- "生成设计系统并持久化，项目名称：MyApp，页面：homepage"

**Agent 会自动**：
1. 分析需求：识别产品类型（Beauty/Spa）、样式关键词、行业、技术栈
2. 生成设计系统：调用 `generate_design_system` 生成完整设计系统
3. 补充搜索：根据需要调用 `ui_ux_search` 获取 React 技术栈指南
4. 返回结果：返回完整的设计系统，包括 Pattern、Style、Colors、Typography、Effects

### 示例 2：使用场景

#### 场景 1：快速生成设计系统

**用户请求**：
```
生成一个电商平台的设计系统
```

**Agent 响应**：
- 自动识别产品类型：E-commerce
- 生成完整设计系统（配色、字体、样式、布局模式）
- 返回格式化的设计系统文档

#### 场景 2：生成并持久化设计系统

**用户请求**：
```
生成一个 SaaS 产品的设计系统，项目名称：MySaaS，技术栈：react，并持久化
```

**Agent 响应**：
- 生成设计系统
- 获取 React 技术栈指南
- 持久化到 `storage/app/ui_ux/design-system/mysaas/MASTER.md`
- 返回持久化路径

#### 场景 3：生成页面特定设计系统

**用户请求**：
```
生成设计系统，项目名称：MyApp，页面：homepage，并持久化
```

**Agent 响应**：
- 生成设计系统
- 持久化 Master 文件：`MASTER.md`
- 持久化页面覆盖：`pages/homepage.md`
- 返回两个文件的路径

### 示例 3：在 Agent 中使用单独的 UI/UX 搜索工具

```yaml
Nodes:
  - Key: "designer"
    Type: "chat_model"
    Model: "ui_model"
    SystemPrompt: |
      你是一位 UI/UX 设计师。使用 ui_ux_search 工具搜索设计知识库。
    BindTools:
      - "ui_ux_search"
```

### 示例 4：手动使用设计系统生成和持久化工具

```yaml
Nodes:
  - Key: "designer"
    Type: "chat_model"
    Model: "ui_model"
    SystemPrompt: |
      你是一位 UI/UX 设计师。
      当用户请求生成设计系统时：
      1. 使用 generate_design_system 生成设计系统
      2. 使用 persist_design_system 持久化到文件
      3. 使用 AppName="your_app" 进行应用隔离
    BindTools:
      - "generate_design_system"
      - "persist_design_system"
```

## 注意事项

1. **配置隔离**：
   - DigEino 的配置和 DigFlow 的配置是分开的
   - 如果使用 DigEino 的配置驱动功能，需要单独加载 DigEino 的配置文件
   - 如果使用 DigFlow 的模型注册表，可以直接使用 DigFlow 的模型

2. **存储路径**：
   - 设计系统默认存储在 `storage/app/ui_ux/` 目录
   - 可以通过 `AppName` 参数实现应用隔离
   - 存储路径可以在 DigEino 配置中自定义

3. **依赖管理**：
   - 确保 `go.mod` 中的 DigEino 版本是最新的
   - 运行 `go mod tidy` 确保依赖正确

4. **工具命名冲突**：
   - 如果 DigFlow 中已有同名工具，注册会失败（但不会报错）
   - 建议检查日志确认工具是否注册成功

5. **模型名称一致性**：
   - 使用 Agent 工具时，`agent.yml` 中的模型名称必须与 `tool_register.go` 中的名称一致
   - 推荐使用 `ui_model` 作为模型名称

## 故障排查

### 问题 1：工具未找到

**错误信息**：`工具 xxx 未注册`

**解决方案**：
1. 检查工具是否在 `tool_register.go` 中正确注册
2. 检查 `bootstrap/init.go` 是否导入了工具包
3. 查看日志确认工具注册是否成功

### 问题 2：模型未找到（Agent 工具）

**错误**：`模型 ui_model 不存在`

**解决**：
1. 检查 `agent.yml` 中是否配置了名为 `ui_model` 的模型
2. 确保模型名称与 `tool_register.go` 中的名称一致
3. 检查模型是否已正确注册到 ModelRegistry

### 问题 3：Agent 工具创建失败

**错误**：`failed to create UI design system agent`

**解决**：
1. 检查 ChatModel 是否正确获取
2. 检查 DigEino 依赖是否正确安装
3. 查看日志获取详细错误信息

### 问题 4：工具调用失败

**错误**：`agent execution failed`

**解决**：
1. 检查模型 API Key 是否正确配置
2. 检查网络连接
3. 查看 Agent 内部日志

### 问题 5：配置加载失败

**错误信息**：`加载 DigEino 配置失败`

**解决方案**：
1. 检查配置文件路径是否正确
2. 如果不需要配置驱动功能，可以不加载 DigEino 配置
3. 使用 DigFlow 的模型注册表代替

### 问题 6：存储路径问题

**错误信息**：`failed to create directories`

**解决方案**：
1. 检查存储目录的写入权限
2. 确认 `BaseDir` 配置是否正确
3. 使用绝对路径避免路径问题

## 完整示例：在 UI Pro 应用中使用

参考 `DigFlow/app/ui_pro/config/agent.yml`，你可以看到 UI Pro 应用已经配置了 UI/UX 工具：

```yaml
Tools:
  - Name: "ui_ux_search"
    Description: "检索 UI/UX 设计知识库"
    UseGlobal: true  # 如果工具在全局注册，使用 UseGlobal: true
```

如果你的工具在应用级别注册，使用 `UseGlobal: false`。

## 参考

- [DigEino UI/UX 工具文档](../../tools/ui_ux/README.md)
- [DigFlow 工具注册文档](https://github.com/your-org/digflow/blob/main/README.md)
- [DigEino 更新日志](../../updates/2026-02-01_ui_ux_config_and_storage_improvements.md)
- [Agent 实现代码](../../tools/ui_ux/agent.go)
- [Agent 工具包装代码](../../tools/ui_ux/agent_tool.go)
