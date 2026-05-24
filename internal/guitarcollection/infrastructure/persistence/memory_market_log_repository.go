package persistence

import (
	"context"
	"sort"
	"sync"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// MemoryMarketLogRepository is an in-memory MarketLogRepository for tests.
type MemoryMarketLogRepository struct {
	mu   sync.RWMutex
	logs map[string]*domain.MarketLog
}

// NewMemoryMarketLogRepository constructs an empty in-memory market log store.
func NewMemoryMarketLogRepository() *MemoryMarketLogRepository {
	return &MemoryMarketLogRepository{logs: make(map[string]*domain.MarketLog)}
}

// Save persists a market log (upsert).
func (r *MemoryMarketLogRepository) Save(ctx context.Context, log *domain.MarketLog) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID()] = log
	return nil
}

// FindByGuitarID returns logs for a guitar, newest first.
func (r *MemoryMarketLogRepository) FindByGuitarID(ctx context.Context, guitarID string) ([]*domain.MarketLog, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.MarketLog, 0)
	for _, log := range r.logs {
		if log.GuitarID() == guitarID {
			out = append(out, log)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ObservedAt().After(out[j].ObservedAt())
	})
	return out, nil
}

// DeleteByGuitarID removes every market log for a guitar.
func (r *MemoryMarketLogRepository) DeleteByGuitarID(ctx context.Context, guitarID string) (int, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	deleted := 0
	for id, log := range r.logs {
		if log.GuitarID() == guitarID {
			delete(r.logs, id)
			deleted++
		}
	}
	return deleted, nil
}

// All returns every stored log (test helper).
func (r *MemoryMarketLogRepository) All() []*domain.MarketLog {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.MarketLog, 0, len(r.logs))
	for _, log := range r.logs {
		out = append(out, log)
	}
	return out
}

// Count returns the number of stored logs (test helper).
func (r *MemoryMarketLogRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.logs)
}

var _ domain.MarketLogRepository = (*MemoryMarketLogRepository)(nil)
