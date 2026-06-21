package persistence

import (
	"context"
	"sync"

	"github.com/wbits/guitars/internal/guitaranalysis"
)

// MemoryRepository is an in-memory analysis store for tests.
type MemoryRepository struct {
	mu      sync.RWMutex
	records map[string]*guitaranalysis.Record
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{records: map[string]*guitaranalysis.Record{}}
}

func (r *MemoryRepository) FindByGuitarID(_ context.Context, guitarID string) (*guitaranalysis.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.records[guitarID]
	if !ok {
		return nil, nil
	}
	return cloneRecord(rec)
}

func (r *MemoryRepository) FindByGuitarIDs(_ context.Context, guitarIDs []string) (map[string]*guitaranalysis.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]*guitaranalysis.Record, len(guitarIDs))
	for _, id := range guitarIDs {
		if rec, ok := r.records[id]; ok {
			cloned, err := cloneRecord(rec)
			if err != nil {
				return nil, err
			}
			out[id] = cloned
		}
	}
	return out, nil
}

func (r *MemoryRepository) Save(_ context.Context, record *guitaranalysis.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cloned, err := cloneRecord(record)
	if err != nil {
		return err
	}
	r.records[record.GuitarID()] = cloned
	return nil
}

func (r *MemoryRepository) DeleteByGuitarID(_ context.Context, guitarID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, guitarID)
	return nil
}

func cloneRecord(rec *guitaranalysis.Record) (*guitaranalysis.Record, error) {
	out, err := guitaranalysis.NewRecord(rec.GuitarID(), rec.OwnerID(), rec.Status(), rec.PicturesFingerprint())
	if err != nil {
		return nil, err
	}
	switch rec.Status() {
	case guitaranalysis.StatusReady:
		out.SetReady(rec.PicturesFingerprint(), rec.VisualSummary(), rec.Tags(), rec.Confidence())
	case guitaranalysis.StatusFailed:
		out.SetFailed(rec.PicturesFingerprint(), rec.FailureReason())
	default:
		out.SetPending(rec.PicturesFingerprint())
	}
	return out, nil
}
