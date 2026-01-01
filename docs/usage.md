# DigEino 使用指南

`DigEino` 是基于 `Eino` 的二次封装，旨在为 Go 项目提供统一的状态追踪、Webhook 回调及配置管理功能。

## 1. 安装

在您的 Go 项目中，通过以下命令导入：

```bash
go get github.com/originaleric/digeino
```

## 2. 配置管理

`DigEino` 提供了灵活的配置管理方式，推荐在项目启动时进行初始化。

### 2.1 依赖结构

配置主要包含两个部分：
- `HttpServer`: 用于获取本地服务端口（构建默认 Webhook URL 时使用）。
- `Status`: 包含 `Webhook` 和 `Store`（存储）的配置。

### 2.2 使用示例

#### 方式 A：从 YAML 文件加载（推荐）

首先，在您的项目中准备一个 `config.yaml` 文件（可以参考 `DigEino/config/config.yaml`）：

```yaml
HttpServer:
  Api:
    Port: ":20201"

Status:
  Webhook:
    Enabled: true
    URL: "https://your-callback-url.com/webhook"
    Events: ["complete", "error"]
  Store:
    Enabled: true
    Type: "memory" # 或 "mysql"
```

然后在代码中加载：

```go
import "github.com/originaleric/digeino/config"

func main() {
    // 加载配置文件
    _, err := config.Load("path/to/your/config.yaml")
    if err != nil {
        panic(err)
    }
}
```

#### 方式 B：在代码中手动设置

```go
import "github.com/originaleric/digeino/config"

func initConfig() {
    cfg := config.Default()
    cfg.Status.Webhook.Config.URL = "https://custom-url.com"
    config.Set(cfg)
}
```

## 3. 核心功能使用

### 3.1 状态存储 (Status Store)

`DigEino` 支持将执行状态存储在内存或 MySQL 中。

```go
import "github.com/originaleric/digeino/status"

func someProcess() {
    // 获取默认存储实例（根据配置自动选择 memory 或 mysql）
    store := status.GetDefaultStore()
    
    // 创建执行记录
    record := store.CreateExecution("exec_id_001", "my_app", "req_id_001")
    
    // 更新状态
    store.AddStatus("exec_id_001", webhook.ExecutionStatus{
        Type: "node_start",
        NodeKey: "node_1",
        Status: "running",
    })
}
```

### 3.2 Webhook 客户端

用于将状态更新通过 HTTP 回调发送给第三方系统。

```go
import (
    "github.com/originaleric/digeino/webhook"
    "github.com/originaleric/digeino/config"
)

func notifyCallback() {
    // 获取 Webhook 配置
    webhookCfg := webhook.GetWebhookConfig(nil)
    if webhookCfg == nil {
        return
    }

    client := webhook.NewWebhookClient(webhookCfg)
    err := client.SendStatus(context.Background(), statusInfo)
}
```

## 4. AgentState 扩展字段

`AgentState` 结构体提供了 `Extensions` 字段，允许各个项目存储特定的扩展数据，而无需修改核心库。

### 4.1 基本使用

```go
import "github.com/originaleric/digeino"

// 创建 AgentState
state := &digeino.AgentState{
    SessionID: "session_001",
    Query:     "用户查询",
}

// 设置扩展字段（字符串类型）
state.SetStringExtension("pdf_path", "/path/to/file.pdf")
state.SetStringExtension("custom_field", "custom_value")

// 设置扩展字段（整数类型）
state.SetIntExtension("page_count", 10)

// 设置扩展字段（布尔类型）
state.SetBoolExtension("is_processed", true)

// 设置任意类型的扩展字段
state.SetExtension("complex_data", map[string]interface{}{
    "key1": "value1",
    "key2": 123,
})
```

### 4.2 获取扩展字段

```go
// 获取字符串类型
pdfPath, ok := state.GetStringExtension("pdf_path")
if ok {
    fmt.Printf("PDF Path: %s\n", pdfPath)
}

// 获取整数类型
pageCount, ok := state.GetIntExtension("page_count")
if ok {
    fmt.Printf("Page Count: %d\n", pageCount)
}

// 获取布尔类型
isProcessed, ok := state.GetBoolExtension("is_processed")
if ok {
    fmt.Printf("Is Processed: %v\n", isProcessed)
}

// 获取任意类型
val, ok := state.GetExtension("complex_data")
if ok {
    // 进行类型断言
    if data, ok := val.(map[string]interface{}); ok {
        // 使用 data
    }
}
```

### 4.3 最佳实践

1. **定义扩展键常量**：建议在项目中使用常量定义扩展键，避免硬编码字符串：

```go
// 在项目根目录或 constants 包中定义
const (
    ExtensionKeyPdfPath = "pdf_path"
    ExtensionKeyCustomField = "custom_field"
)

// 使用
state.SetStringExtension(ExtensionKeyPdfPath, "/path/to/file.pdf")
```

2. **类型安全**：优先使用类型安全的辅助方法（`GetStringExtension`、`SetStringExtension` 等），而不是直接使用 `GetExtension`。

3. **JSON 序列化**：`Extensions` 字段会被自动序列化到 JSON，使用 `omitempty` 标签，空 map 不会出现在 JSON 中。

### 4.4 存储业务数据结构

对于复杂的业务数据结构（如文档大纲、页面列表等），可以使用 `GetBusinessData` 和 `SetBusinessData` 方法：

```go
// 在您的项目中定义业务结构体
type DocumentOutline struct {
    Title          string        `json:"title"`
    Topic          string        `json:"topic"`
    TargetAudience string        `json:"target_audience"`
    PageOutlines   []PageOutline `json:"pages"`
}

// 设置业务数据
outline := &DocumentOutline{
    Title: "我的文档",
    Topic: "技术文档",
}
state.SetBusinessData("outline", outline)

// 获取业务数据
var outline *DocumentOutline
err := state.GetBusinessData("outline", &outline)
if err != nil {
    // 处理错误
}
```

**最佳实践**：在项目包中定义业务结构体和类型安全的辅助方法：

```go
// 在项目包中（如 DigPdf/eino/types.go）
package eino_agent

const ExtensionKeyOutline = "outline"

func GetOutline(state *digeino.AgentState) (*DocumentOutline, error) {
    var outline *DocumentOutline
    err := state.GetBusinessData(ExtensionKeyOutline, &outline)
    return outline, err
}

func SetOutline(state *digeino.AgentState, outline *DocumentOutline) {
    state.SetBusinessData(ExtensionKeyOutline, outline)
}
```

## 5. AgentState 架构说明

### 5.1 核心字段

`AgentState` 只包含通用的核心字段：
- `SessionID`: 会话标识
- `Query`: 用户查询
- `Status`: 状态标识
- `ResearchSummary`: 研究总结
- `Extensions`: 扩展字段（用于存储项目特定的数据）

### 5.2 业务结构体

**重要**：`DigEino` 是一个通用库，不包含任何业务特定的结构体（如 `DocumentOutline`、`Page`、`DesignConfig` 等）。这些结构体应该在您的项目包中定义，并通过 `Extensions` 字段存储。

这样做的好处：
- ✅ 保持 `DigEino` 的通用性和可复用性
- ✅ 不同项目可以定义自己的业务结构体
- ✅ 避免业务逻辑污染核心库
- ✅ 更好的类型安全和代码组织

### 5.3 迁移指南

如果您之前使用了 `AgentState` 中的业务字段（如 `Outline`、`Pages` 等），需要：

1. 在项目包中定义这些业务结构体
2. 使用 `GetBusinessData`/`SetBusinessData` 或自定义辅助方法访问
3. 参考 `DigPdf` 项目的实现作为示例

## 6. 常见问题

- **独立性**: `DigEino` 现在不再依赖 `DigPdf` 的 internal 包，可以安全地在任何 Go 项目中 import。
- **扩展性**: 如果需要增加新的配置项，请在 `github.com/originaleric/digeino/config` 包中扩展 `Config` 结构体。
- **项目特定字段**: 使用 `AgentState.Extensions` 字段存储项目特定的数据，而不是修改核心库的 `AgentState` 结构体。
- **业务结构体**: 所有业务特定的结构体都应该在项目包中定义，通过 `Extensions` 字段存储，并使用 `GetBusinessData`/`SetBusinessData` 方法访问。
