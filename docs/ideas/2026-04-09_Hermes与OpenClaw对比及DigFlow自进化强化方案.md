# 2026-04-09 Hermes与OpenClaw对比及DigFlow自进化强化方案

## 一、背景与目标

本文档整理两部分结论：

1. **Hermes Agent 与 OpenClaw 的设计对比**（重点：自我进化机制）
2. **DigFlow 借鉴 Hermes 的单点强化方案**（选择最值得落地的一项）

目标是把“已有能力”升级为“可持续自进化闭环”，在不显著增加主链路复杂度和延迟的前提下，提升长期可复用性与用户体验稳定性。

---

## 二、Hermes Agent 与 OpenClaw 设计对比（聚焦自我进化）

### 2.1 定位差异

- **OpenClaw（当前常见实践）**  
  更偏执行编排闭环：需求接入 -> 分发执行器 -> 结果回传。  
  优势在任务执行链路清晰、可控；但“经验沉淀”通常依赖外部流程或人工复盘。

- **Hermes Agent**  
  更偏“执行 + 学习并重”的 runtime：把记忆、技能沉淀、跨会话召回作为系统主路径。  
  自我进化不是口号，而是通过触发器、后台复盘、结构化存储与再注入实现闭环。

### 2.2 Hermes 的关键亮点（可借鉴点）

1. **后台复盘机制（非阻塞）**  
   主任务完成后，异步启动复盘流程，判断是否写记忆、是否创建/修补技能。  
   学习行为与主响应解耦，降低对交互体验的干扰。

2. **双层知识结构**  
   - 声明式记忆（长期事实、偏好）
   - 程序性记忆（Skills，强调“怎么做”）  
   避免把所有信息混在单一存储中导致可维护性下降。

3. **“技能修补”优先于“技能新增”**  
   对已存在技能执行 patch/update，而非无限新增，减少知识膨胀和陈旧知识堆积。

4. **跨会话检索与摘要召回**  
   通过会话检索+摘要层回忆历史任务关键点，避免直接喂全量历史文本，提高上下文效率与可解释性。

5. **学习触发器可配置**  
   按轮次/工具调用次数触发“该不该沉淀”的审查，兼顾学习率与噪音控制。

### 2.3 对 OpenClaw 体系的启示

OpenClaw 适合作为执行与编排底座；若要增强“越用越聪明”，最有效路径不是重写执行引擎，而是补齐一个**后置学习环**：

- 触发学习（何时复盘）
- 复盘判定（是否值得沉淀）
- 结构化落库（记忆/技能）
- 后续任务回注（检索/注入）

---

## 三、DigFlow 当前能力盘点（与强化点相关）

DigFlow 已具备较完整基础能力：

- **短期记忆摘要与滑窗**：`memory/manager.go`
- **长期记忆提取与召回（含重排）**：`memory/extractor.go`
- **Skill 仓库与 CRUD/导出**：`skills/service_impl.go`
- **Trigger 体系（cron/webhook）**：`trigger/trigger_manager.go`

结论：DigFlow 已有“组件能力”，下一步最值得做的是把它们连成**自动自进化闭环**。

---

## 四、DigFlow 最值得强化的一点：PostRun Learning Loop

### 4.1 核心建议

新增一个 **PostRun Learning Loop（后置复盘学习环）**：

- 主任务完成后异步触发
- 自动判断是否写入长期记忆
- 自动判断是否创建/修补技能
- 全程审计可追溯，可回滚

### 4.2 为什么优先做这一项

1. **ROI 高**：主要是编排层增强，最大化复用现有 memory/skills/trigger 模块。  
2. **风险可控**：走异步，不阻塞主链路，不改变核心执行路径。  
3. **见效快**：能较快降低“重复问题反复纠偏”的用户成本。

---

## 五、落地方案（MVP）

### 5.1 事件触发

在 Run 结束后发出 `run.completed` 事件（或复用现有治理事件总线），由 Learning Worker 消费。

### 5.2 复盘决策链

Learning Worker 输入：

- 本次会话核心上下文（用户意图、最终答复、工具调用摘要、错误重试信息）

输出一个结构化决策（JSON）：

- `memory_actions[]`：add / skip
- `skill_actions[]`：create / patch / skip
- `reason`：决策依据（用于审计）
- `confidence`：置信度

### 5.3 执行动作

- 记忆沉淀：复用 `MemoryExtractor` + `LongTermStore`
- 技能沉淀：复用 `SkillService`
- 优先策略：**有同类技能先 patch，无匹配再 create**

### 5.4 限流与门槛（防止过学习）

新增配置建议（如 `config/eino.yml`）：

- `learning.memory_nudge_interval`
- `learning.skill_nudge_interval`
- `learning.min_tool_calls_for_skill`
- `learning.min_confidence`

基础规则（建议默认）：

- 工具调用 < 3 次：默认不建 Skill
- 无失败重试/无用户纠偏：默认不建 Skill
- 低置信度：仅记录审计，不执行落库

### 5.5 审计与回滚

新增学习审计记录（可表或治理事件扩展），关键字段：

- `decision_id`
- `run_id` / `session_id`
- `decision_type`（memory/skill）
- `action`（create/patch/skip）
- `reason` / `confidence`
- `status`（applied/rolled_back）

建议补两个接口：

- 查询：`GET /learning/decisions?run_id=...`
- 回滚：`POST /learning/decisions/{id}/rollback`

---

## 六、分阶段实施计划

### 阶段 1：只判定不执行（2-3 天）

- 接入 `run.completed` -> Learning Worker
- 产出决策 JSON 与审计日志
- 不自动写 memory/skill

**成功标准**：可稳定产出可解释决策，主链路无性能回退。

### 阶段 2：自动写记忆（3-5 天）

- 启用 memory_actions 自动执行
- 增加幂等与重复检测

**成功标准**：记忆写入准确率可接受，无明显噪音。

### 阶段 3：自动技能修补/创建（3-5 天）

- 开启 skill patch/create
- 默认“patch 优先”
- 保留人工审核开关

**成功标准**：技能复用率提升，技能总量增速可控。

### 阶段 4：调参与评估

- A/B 对比（learning on/off）
- 调整阈值、置信度门槛
- 形成稳定策略

---

## 七、验收指标（建议）

1. 用户重复纠偏率（7/14 天）
2. 技能复用命中率（使用次数/总技能数）
3. 新增技能中 patch 占比（越高通常代表知识维护健康）
4. 学习任务失败率与回滚率
5. 主链路 P95 延迟变化（目标：近似不变）

---

## 八、结论

对于 DigFlow，最值得借鉴 Hermes 的不是某个单独算法，而是其“**后台学习闭环**”工程模式。  
在 DigFlow 现有能力基础上，优先实现 `PostRun Learning Loop`，可以以较低改造成本获得长期收益：

- 让系统从“会执行”进化为“会沉淀并持续变好”
- 保持主链路稳定与可控
- 建立可审计、可回滚、可调参的学习机制

---

## 九、平台化扩展：方案是否可移植到 DigEino

结论：**可以，而且建议优先做成 DigEino 平台能力，再由 DigFlow 作为首个接入方。**

原因是 DigFlow 更偏业务编排层，而 DigEino 已承担平台插件能力供给；`PostRun Learning Loop` 本质是通用的学习与沉淀编排，不依赖特定业务领域，天然适合平台化。

### 9.1 可移植性的技术依据

该方案依赖的能力在相似架构系统中普遍存在：

- 运行结束事件（如 `run.completed`）
- 会话上下文与工具调用轨迹
- 记忆写入接口
- 技能管理接口（create/patch/list）
- 异步执行框架（worker/trigger）

因此可在 DigEino 做抽象，在多个宿主系统复用。

### 9.2 推荐的 DigEino 三层插件架构

1. **learning-core**（策略核）  
   负责复盘判定逻辑：是否沉淀、沉淀到 memory 还是 skill、置信度与理由输出。

2. **learning-adapter**（宿主适配层）  
   对接不同系统的 run 数据模型、记忆接口、技能接口。  
   DigFlow 实现一个 adapter，其他系统按相同契约实现自己的 adapter。

3. **learning-runtime**（执行与治理层）  
   提供异步 worker、重试、幂等、审计、回滚与指标上报。

### 9.3 建议抽象的通用接口（平台契约）

建议在 DigEino 定义以下接口（命名可调整）：

- `RunContextProvider`：读取 run/session/tool timeline 与必要上下文
- `MemorySink`：执行记忆写入、去重与查询
- `SkillSink`：执行 skill create/patch/list
- `LearningAuditStore`：记录决策、状态与回滚信息

只要宿主系统实现这四类接口，即可复用同一套学习闭环。

### 9.4 建议实施路径（先业务验证，再平台下沉）

1. 在 DigFlow 内先落一版可运行 MVP（验证效果与风险）
2. 抽象接口并下沉到 DigEino
3. 将 DigFlow 改为调用 DigEino learning 插件
4. 接入第二个非 DigFlow 系统，验证跨系统复用

### 9.5 平台化后的直接收益

- 避免每个业务系统重复实现学习闭环
- 保持学习策略一致，便于治理与审计
- 新系统接入成本降低（实现 adapter 即可）
- 为后续统一评估体系（学习命中率、噪音率、回滚率）打基础

---

## 附录A：DigEino 接口草案（Go 伪代码）

以下为可评审的最小接口草案，目标是让不同宿主系统以统一契约接入 `PostRun Learning Loop`。

### A.1 核心数据结构

```go
package learning

import "time"

type LearningEvent struct {
    EventID      string
    EventType    string // run.completed
    OccurredAt   time.Time
    AppName      string
    RunID        string
    SessionID    string
    UserID       string
    TriggerType  string // user | cron | webhook | system
}

type ToolCallDigest struct {
    Name       string
    Success    bool
    DurationMs int64
    Error      string
}

type RunContext struct {
    AppName            string
    RunID              string
    SessionID          string
    UserID             string
    UserInput          string
    AssistantOutput    string
    SessionSummary     string
    ToolCalls          []ToolCallDigest
    RetryCount         int
    UserCorrectionHint bool
}

type MemoryAction struct {
    Action     string // add | skip
    Category   string // fact | preference | timeline
    Content    string
    Importance int
}

type SkillAction struct {
    Action      string // create | patch | skip
    SkillName   string
    PatchNote   string
    FullContent string // create 时可用
}

type LearningDecision struct {
    DecisionID     string
    RunID          string
    SessionID      string
    MemoryActions  []MemoryAction
    SkillActions   []SkillAction
    Reason         string
    Confidence     float64
    CreatedAt      time.Time
}
```

### A.2 平台契约接口

```go
package learning

import "context"

// 1) 从宿主系统获取 run 上下文
type RunContextProvider interface {
    GetRunContext(ctx context.Context, event LearningEvent) (*RunContext, error)
}

// 2) 记忆写入能力
type MemorySink interface {
    Add(ctx context.Context, appName, userID, category, content, runID string, importance int) error
    Exists(ctx context.Context, appName, userID, content string) (bool, error)
}

// 3) 技能写入能力（优先 patch）
type SkillSink interface {
    FindSimilar(ctx context.Context, appName, query string) (skillID string, skillName string, found bool, err error)
    Create(ctx context.Context, appName, name, content, createdBy string) (skillID string, err error)
    Patch(ctx context.Context, skillID, patchNote, updatedBy string) error
}

// 4) 审计与回滚
type LearningAuditStore interface {
    SaveDecision(ctx context.Context, decision *LearningDecision) error
    MarkApplied(ctx context.Context, decisionID string) error
    MarkFailed(ctx context.Context, decisionID string, reason string) error
    Rollback(ctx context.Context, decisionID string) error
}
```

### A.3 决策引擎接口与默认实现

```go
package learning

import "context"

type DecisionEngine interface {
    Evaluate(ctx context.Context, rc *RunContext, cfg Config) (*LearningDecision, error)
}

type DefaultDecisionEngine struct {
    LLM LLMClient
}

func (e *DefaultDecisionEngine) Evaluate(ctx context.Context, rc *RunContext, cfg Config) (*LearningDecision, error) {
    // 伪代码:
    // 1. 快速规则过滤: 工具调用次数、重试次数、是否有纠偏
    // 2. 调用 LLM 生成结构化 JSON 决策
    // 3. 清洗/校验动作集合
    // 4. 返回 LearningDecision
    return &LearningDecision{}, nil
}
```

### A.4 Worker 执行流程（最小可运行）

```go
package learning

import "context"

type Worker struct {
    Cfg       Config
    Provider  RunContextProvider
    Memory    MemorySink
    Skill     SkillSink
    Audit     LearningAuditStore
    Engine    DecisionEngine
}

func (w *Worker) HandleRunCompleted(ctx context.Context, event LearningEvent) error {
    rc, err := w.Provider.GetRunContext(ctx, event)
    if err != nil {
        return err
    }

    decision, err := w.Engine.Evaluate(ctx, rc, w.Cfg)
    if err != nil {
        return err
    }
    _ = w.Audit.SaveDecision(ctx, decision)

    // 记忆动作
    for _, ma := range decision.MemoryActions {
        if ma.Action != "add" {
            continue
        }
        exists, _ := w.Memory.Exists(ctx, rc.AppName, rc.UserID, ma.Content)
        if exists {
            continue
        }
        _ = w.Memory.Add(ctx, rc.AppName, rc.UserID, ma.Category, ma.Content, rc.RunID, ma.Importance)
    }

    // 技能动作（patch 优先）
    for _, sa := range decision.SkillActions {
        switch sa.Action {
        case "patch":
            id, _, found, _ := w.Skill.FindSimilar(ctx, rc.AppName, sa.SkillName)
            if found {
                _ = w.Skill.Patch(ctx, id, sa.PatchNote, "learning-worker")
            }
        case "create":
            _, _ = w.Skill.Create(ctx, rc.AppName, sa.SkillName, sa.FullContent, "learning-worker")
        }
    }

    _ = w.Audit.MarkApplied(ctx, decision.DecisionID)
    return nil
}
```

### A.5 配置结构草案（建议放 `config/eino.yml`）

```yaml
learning:
  enabled: true
  async: true
  event: run.completed

  # 触发节奏（nudge）
  memory_nudge_interval: 10
  skill_nudge_interval: 15

  # 基础门槛
  min_tool_calls_for_skill: 3
  min_confidence: 0.65
  patch_first: true

  # 失败重试
  retry:
    max_attempts: 2
    backoff_ms: 500
```

### A.6 事件 schema 草案

```json
{
  "event_id": "evt_20260409_xxx",
  "event_type": "run.completed",
  "occurred_at": "2026-04-09T10:30:00Z",
  "app_name": "coder",
  "run_id": "run_xxx",
  "session_id": "sess_xxx",
  "user_id": "u_xxx",
  "trigger_type": "user"
}
```

### A.7 与 DigFlow 的最小接入点

1. 在 Run 完成处发布 `run.completed` 事件  
2. 实现 DigFlow 的 `RunContextProvider/MemorySink/SkillSink/AuditStore`  
3. 由 DigEino 的 `learning-runtime` 订阅事件并执行  
4. 通过治理 API 暴露审计查询与回滚

---

## 附录B：评审建议清单

为避免“先编码后返工”，建议评审时重点确认：

1. **边界**：学习环是否严格异步、是否不阻塞主链路
2. **幂等**：重复事件是否会导致重复写入
3. **安全**：Skill 自动写入是否有审核/白名单/回滚
4. **治理**：是否可查、可停、可回滚、可调阈值
5. **泛化**：接口命名是否业务无关，能否支撑第二宿主接入

