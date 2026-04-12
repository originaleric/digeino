package learning

import (
	"context"
	"sync"
	"time"
)

var (
	mu sync.RWMutex

	runtimeOverride *Config
	globalHost      *Host
	globalEngine    DecisionEngine = &DefaultDecisionEngine{}

	deduper = newMemoryDeduper(24 * time.Hour)

	initOnce sync.Once
	jobCh    chan LearningEvent
)

// SetHost 注册宿主实现；可为 nil 表示不执行学习。
func SetHost(h *Host) {
	mu.Lock()
	defer mu.Unlock()
	globalHost = h
}

// SetEngine 覆盖默认决策引擎（测试或自定义策略）。
func SetEngine(e DecisionEngine) {
	mu.Lock()
	defer mu.Unlock()
	if e == nil {
		globalEngine = &DefaultDecisionEngine{}
		return
	}
	globalEngine = e
}

// SetConfigOverride 覆盖全局 YAML 配置（nil 表示仅使用 config.Get().Learning）。
func SetConfigOverride(c *Config) {
	mu.Lock()
	defer mu.Unlock()
	runtimeOverride = c
}

// Init 同时设置配置与宿主，便于宿主进程 bootstrap 一次完成注册。
func Init(cfg Config, h *Host) {
	SetConfigOverride(&cfg)
	SetHost(h)
}

func effectiveConfig() Config {
	mu.RLock()
	defer mu.RUnlock()
	if runtimeOverride != nil {
		return *runtimeOverride
	}
	return ConfigFromGlobal()
}

func getHost() *Host {
	mu.RLock()
	defer mu.RUnlock()
	return globalHost
}

func getEngine() DecisionEngine {
	mu.RLock()
	defer mu.RUnlock()
	if globalEngine == nil {
		return &DefaultDecisionEngine{}
	}
	return globalEngine
}

func ensureWorker() {
	initOnce.Do(func() {
		jobCh = make(chan LearningEvent, 256)
		go func() {
			for ev := range jobCh {
				processOne(context.Background(), ev)
			}
		}()
	})
}

// Enqueue 非阻塞投递；未启用或队列满时降级为独立 goroutine 发送。
func Enqueue(ev LearningEvent) {
	cfg := effectiveConfig()
	if !cfg.Enabled {
		return
	}
	ensureWorker()
	select {
	case jobCh <- ev:
	default:
		go func() { jobCh <- ev }()
	}
}

// Enabled 是否启用（供 webhook 等快速判断）。
func Enabled() bool {
	return effectiveConfig().Enabled
}

func processOne(ctx context.Context, ev LearningEvent) {
	cfg := effectiveConfig()
	if !cfg.Enabled {
		return
	}
	h := getHost()
	if h == nil || h.Audit == nil {
		return
	}

	exists, err := h.Audit.Exists(ctx, ev.ExecutionID, ev.TerminalOutcome)
	if err == nil && exists {
		return
	}
	if !deduper.TryClaim(ev.ExecutionID, ev.TerminalOutcome) {
		return
	}

	var rc *RunContext
	if h.RunContextProvider != nil {
		var gerr error
		rc, gerr = h.RunContextProvider.GetRunContext(ctx, ev)
		if gerr != nil {
			rc = minimalRunContext(ev)
		}
	}
	if rc == nil {
		rc = minimalRunContext(ev)
	}

	engine := getEngine()
	dec, err := engine.Evaluate(ctx, rc, cfg)
	if err != nil {
		return
	}

	if err := h.Audit.SaveDecision(ctx, dec); err != nil {
		return
	}

	switch cfg.ExecutionPhase {
	case PhaseAuditOnly:
		// 仅审计
	case PhaseMemory:
		applyMemory(ctx, h, rc, dec, cfg)
	case PhaseSkill:
		applyMemory(ctx, h, rc, dec, cfg)
		applySkill(ctx, h, rc, dec, cfg)
	default:
		applyMemory(ctx, h, rc, dec, cfg)
		applySkill(ctx, h, rc, dec, cfg)
	}
	_ = h.Audit.MarkApplied(ctx, dec.DecisionID)
}

func minimalRunContext(ev LearningEvent) *RunContext {
	return &RunContext{
		AppName:         ev.AppName,
		RunID:           ev.RunID,
		ExecutionID:     ev.ExecutionID,
		SessionID:       ev.SessionID,
		UserID:          ev.UserID,
		TerminalOutcome: ev.TerminalOutcome,
	}
}

func applyMemory(ctx context.Context, h *Host, rc *RunContext, dec *LearningDecision, cfg Config) {
	if h.Memory == nil {
		return
	}
	for _, ma := range dec.MemoryActions {
		if ma.Action != "add" {
			continue
		}
		exists, err := h.Memory.Exists(ctx, rc.AppName, rc.UserID, ma.Content)
		if err != nil || exists {
			continue
		}
		_ = h.Memory.Add(ctx, rc.AppName, rc.UserID, ma.Category, ma.Content, rc.RunID, ma.Importance)
	}
}

func applySkill(ctx context.Context, h *Host, rc *RunContext, dec *LearningDecision, cfg Config) {
	if h.Skill == nil {
		return
	}
	for _, sa := range dec.SkillActions {
		switch sa.Action {
		case "patch":
			id, _, found, err := h.Skill.FindSimilar(ctx, rc.AppName, sa.SkillName)
			if err == nil && found {
				_ = h.Skill.Patch(ctx, id, sa.PatchNote, "learning-worker")
			}
		case "create":
			if cfg.PatchFirst {
				if id, _, found, err := h.Skill.FindSimilar(ctx, rc.AppName, sa.SkillName); err == nil && found {
					_ = h.Skill.Patch(ctx, id, sa.PatchNote, "learning-worker")
					continue
				}
			}
			_, _ = h.Skill.Create(ctx, rc.AppName, sa.SkillName, sa.FullContent, "learning-worker")
		}
	}
}
