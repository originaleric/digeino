package learning

import "github.com/originaleric/digeino/config"

// Config 学习子系统运行时配置（亦可由全局 config 注入）。
type Config struct {
	Enabled bool

	Async bool

	// MemoryNudgeInterval / SkillNudgeInterval 预留：按 run 序号取模控制学习频率（0 表示不启用）。
	MemoryNudgeInterval int
	SkillNudgeInterval  int

	MinToolCallsForSkill int
	MinConfidence        float64
	PatchFirst           bool

	RetryMaxAttempts int
	RetryBackoffMs   int

	// ExecutionPhase 当前执行阶段（审计 / +memory / +skill）。
	ExecutionPhase Phase
}

// DefaultConfig 默认关闭学习。
func DefaultConfig() Config {
	return Config{
		Enabled:              false,
		Async:                true,
		MemoryNudgeInterval:  0,
		SkillNudgeInterval:   0,
		MinToolCallsForSkill: 3,
		MinConfidence:        0.65,
		PatchFirst:           true,
		RetryMaxAttempts:     2,
		RetryBackoffMs:       500,
		ExecutionPhase:       PhaseAuditOnly,
	}
}

// ConfigFromGlobal 从 digeino 全局配置映射（若 Learning 未配置则返回 DefaultConfig）。
func ConfigFromGlobal() Config {
	cfg := config.Get()
	if cfg == nil || !cfg.Learning.Enabled {
		c := DefaultConfig()
		return c
	}
	l := cfg.Learning
	c := Config{
		Enabled:              l.Enabled,
		Async:                l.Async,
		MemoryNudgeInterval:  l.MemoryNudgeInterval,
		SkillNudgeInterval:   l.SkillNudgeInterval,
		MinToolCallsForSkill: l.MinToolCallsForSkill,
		MinConfidence:        l.MinConfidence,
		PatchFirst:           l.PatchFirst,
		RetryMaxAttempts:     l.Retry.MaxAttempts,
		RetryBackoffMs:       l.Retry.BackoffMs,
		ExecutionPhase:       Phase(l.ExecutionPhase),
	}
	if c.MinToolCallsForSkill <= 0 {
		c.MinToolCallsForSkill = 3
	}
	if c.MinConfidence <= 0 {
		c.MinConfidence = 0.65
	}
	if c.RetryMaxAttempts < 0 {
		c.RetryMaxAttempts = 0
	}
	if c.RetryBackoffMs < 0 {
		c.RetryBackoffMs = 0
	}
	return c
}
