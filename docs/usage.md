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

## 4. 常见问题

- **独立性**: `DigEino` 现在不再依赖 `DigPdf` 的 internal 包，可以安全地在任何 Go 项目中 import。
- **扩展性**: 如果需要增加新的配置项，请在 `github.com/originaleric/digeino/config` 包中扩展 `Config` 结构体。
