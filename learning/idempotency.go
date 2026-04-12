package learning

import (
	"sync"
	"time"
)

type entry struct {
	at time.Time
}

// memoryDeduper 进程内去重（TTL 防止内存无限增长）；跨进程幂等仍依赖 LearningAuditStore.Exists。
type memoryDeduper struct {
	mu      sync.Mutex
	entries map[string]entry
	ttl     time.Duration
}

func newMemoryDeduper(ttl time.Duration) *memoryDeduper {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &memoryDeduper{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
}

func (d *memoryDeduper) key(executionID string, outcome TerminalOutcome) string {
	return executionID + "|" + string(outcome)
}

// TryClaim 若尚未见过该键则登记并返回 true；已存在则返回 false。
func (d *memoryDeduper) TryClaim(executionID string, outcome TerminalOutcome) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	k := d.key(executionID, outcome)
	if e, ok := d.entries[k]; ok && now.Sub(e.at) < d.ttl {
		return false
	}
	d.entries[k] = entry{at: now}
	d.pruneLocked(now)
	return true
}

func (d *memoryDeduper) pruneLocked(now time.Time) {
	if len(d.entries) < 10000 {
		return
	}
	for k, e := range d.entries {
		if now.Sub(e.at) > d.ttl {
			delete(d.entries, k)
		}
	}
}
