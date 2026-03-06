# 2026-03-06 StatusCollector 增强与 DataFlow 配置

## 概述

对 `webhook.StatusCollector` 进行四项增强：DataFlow 数据流追踪（支持通过 eino.yml 配置开关，默认关闭）、chat_model 节点自动提取 Token Usage、OnComplete 事件附带汇总 Usage、以及可配置的日志 Hook。现有项目无需修改配置文件即可兼容。

## 变更文件

### config/config.go

- **StatusConfig** 新增字段：`DataFlow DataFlowConfig`
- **DataFlowConfig** 新增类型：`Enabled *bool`（nil 或 false 表示关闭，默认关闭）
- **Default()** 中为 `Status.DataFlow.Enabled` 赋默认值 `false`

### webhook/types.go

- **ExecutionStatus** 新增字段：`Usage *Usage`（`json:"usage,omitempty"`），用于在 node_end / complete 事件中携带 Token 统计

### webhook/status_collector.go

- **StatusLogger** 新增接口：`OnNodeStartLog`、`OnNodeEndLog`、`OnCompleteLog`，供调用方注入 zap/slog 等实现
- **StatusCollector** 新增字段：`enableDataFlow bool`（从 config 读取）、`logger StatusLogger`
- **NewStatusCollector**：创建时从 `config.Get().Status.DataFlow.Enabled` 读取 DataFlow 开关，未配置或为 false 时默认关闭
- **EnableDataFlow(enable bool)**：手动覆盖配置（优先级高于 eino.yml）
- **SetLogger(logger StatusLogger)**：设置日志 Hook
- **OnNodeStart**：写入 `DataFlow`（当 enableDataFlow 为 true 时），并调用 `logger.OnNodeStartLog`（若已设置）
- **OnNodeEnd**：写入 `DataFlow`；当节点类型为 `chat_model` 时自动从输出提取 Usage 并调用 `CollectTokenUsage`，同时在 status 中携带 `Usage`；调用 `logger.OnNodeEndLog`
- **OnComplete**：在 status 中携带汇总 `Usage`，并调用 `logger.OnCompleteLog`
- **marshalDataFlow**：通用序列化，支持 `[]*schema.Message`、`*schema.Message` 及任意类型的 JSON 序列化
- **extractUsageFromOutput** / **extractUsageFromMessage**：从 chat_model 输出的 `ResponseMeta.Usage` 提取 Token 统计
- **getTotalUsageLocked**：在持锁下返回汇总 Usage，供 OnComplete 使用

### status/status_store.go

- **GetDefaultStore()**：当 MySQL 配置存在但连接失败时，不再返回 nil，改为回退为 `NewMemoryStatusStore(1000)`，避免调用方拿到 nil

## 配置说明

### eino.yml 可选配置

```yaml
Status:
  Webhook:
    Enabled: true
    URL: "http://localhost:20201/api/v1/webhook/status"
  Store:
    Enabled: true
    Type: "memory"
  DataFlow:          # 新增，可选
    Enabled: false   # 默认 false，需要时改为 true
```

- **不配置 DataFlow**：行为与之前一致，DataFlow 不记录，现有项目零改动
- **DataFlow.Enabled: true**：在 node_start / node_end 的 `ExecutionStatus` 中填充 `data_flow`（input_count/output_count 及可选的 input_data/output_data）

## 兼容性

| 项 | 影响 |
|----|------|
| NewStatusCollector 签名 | 不变，仍为 `(executionID, appName, requestID string)` |
| OnNodeStart / OnNodeEnd 入参 | 不变，仍为 `interface{}` |
| ExecutionStatus JSON | 向后兼容，`usage`、`data_flow` 均为 `omitempty` |
| DigFlow / Knowledge 等 | 无需修改 eino.yml 或代码即可升级 |

## 影响面

- 影响 `github.com/originaleric/digeino/config`、`webhook`、`status` 包
- Knowledge 已移除本地 webhook/status 实现并切换至本版本 DigEino
