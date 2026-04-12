# 2026-04-12 PostRun 自进化 Learning 模块落地与使用说明

## 概述

在 `DigEino` 内落地 **PostRun Learning Loop** 通用子系统：在 **run 级终态**（`OnComplete` 对应的 `status.Type == "complete"`）之后异步投递学习事件，由宿主通过 `learning.Host` 提供上下文拼装、记忆/技能写入与审计存储。主执行链路不阻塞；默认 **关闭**，需显式开启配置并注册宿主。

设计背景见 `docs/ideas/2026-04-09_Hermes与OpenClaw对比及DigFlow自进化强化方案.md` 与 `docs/ideas/2026-04-12_DigEino优先_PostRun自进化Learning落地方案.md`。

本次变更仅在 `DigEino` 库内；**DigFlow** 侧需自行实现 `Host`、审计表与 API（不在本仓库）。

## 变更文件

### 核心代码

- `learning/types.go`：事件、上下文、决策与阶段枚举
- `learning/host.go`：`RunContextProvider` / `MemorySink` / `SkillSink` / `LearningAuditStore` 与 `Host` 聚合
- `learning/engine.go`：`DecisionEngine`、`DefaultDecisionEngine`、可选 `LLMClient`
- `learning/worker.go`：`Enqueue`、队列消费、`SetHost` / `SetEngine` / `SetConfigOverride` / `Init`
- `learning/idempotency.go`：进程内去重（配合审计 `Exists`）
- `learning/config.go`：与全局 `config.Learning` 映射
- `webhook/status_collector.go`：run 完成单点 `learning.Enqueue`
- `config/config.go`、`config/config.yaml`：新增 `Learning` 配置块

### 测试

- `learning/engine_test.go`
- `learning/worker_test.go`

### 文档

- `docs/updates/2026-04-12_PostRun自进化Learning模块落地与使用说明.md`（本文档）

## 本次新增能力

### 1) 学习子系统（`learning` 包）

- 异步有界队列 + 单 worker 处理 `LearningEvent`
- 分阶段执行：`ExecutionPhase` 0=仅审计，1=+memory，2=+skill（与配置 `ExecutionPhase` 对齐）
- 规则门槛：`MinToolCallsForSkill`、`MinConfidence`、`PatchFirst` 等
- 幂等：`LearningAuditStore.Exists(execution_id, terminal_outcome)` 优先，内存 deduper 辅助

### 2) 终态触发（单点）

在 `StatusCollector.sendStatusAsync` 中，于回调执行之后：若 **`Learning.Enabled`** 且 **`status.Type == "complete"`**，则构造 `LearningEvent` 并 `learning.Enqueue`（成功/失败由 `NormalizeEventType` 与 `status` 区分终态 outcome）。不依赖是否启用 store/webhook，便于仅 webhook 或轻量宿主接入。

### 3) 配置

根配置新增 `Learning` 段，默认 **`Enabled: false`**。宿主可通过 `learning.SetConfigOverride` 覆盖 YAML，或在加载 `config` 后依赖 `config.Get().Learning`。

## 宿主接入要点（DigFlow 等）

1. 实现 `Host` 各接口：至少提供 **`LearningAuditStore`**（阶段 1）；阶段 2/3 再实现 `MemorySink`、`SkillSink`。
2. 进程启动时调用 **`learning.Init(cfg, host)`** 或 **`SetHost` + `SetConfigOverride`**。
3. 在审计存储中实现 **`Exists(execution_id, outcome)`**，保证跨进程/重启幂等。
4. **`GetRunContext`** 中从治理 Run、Invoke 快照等拼装 `RunContext`（DigEino 仅传入 `LearningEvent` 与 execution 维度字段）。

## 启用检查清单

- [ ] `config.yaml` 中 `Learning.Enabled: true`（或运行时 `SetConfigOverride`）
- [ ] 已 `SetHost`，且 `Audit` 非空
- [ ] 分阶段 rollout：先将 `ExecutionPhase` 设为 `0`（仅审计），验证无误后再升阶段

## 验收与风险说明

- 默认关闭时，`Enqueue` 路径几乎无开销（`Enabled()` 短路）。
- 未注册 `Host` 或 `Audit == nil` 时，事件入队后会被丢弃（与「未接入宿主」一致）。
- 详细风险与评审项见背景文档附录（异步、幂等、安全、回滚、泛化）。
