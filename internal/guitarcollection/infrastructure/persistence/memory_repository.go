package persistence

import (
	"context"
	"sort"
	"sync"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// MemoryRepository is a goroutine-safe, in-memory implementation of
// domain.Repository. It exists primarily for tests and local smoke runs that
// don't want to depend on DynamoDB/LocalStack.
type MemoryRepository struct {
	mu      sync.RWMutex
	guitars map[string]*domain.Guitar
}

// NewMemoryRepository returns an empty MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{guitars: map[string]*domain.Guitar{}}
}

// Save implements domain.Repository with upsert semantics.
func (r *MemoryRepository) Save(_ context.Context, g *domain.Guitar) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.guitars[g.ID()] = g
	return nil
}

// FindByID implements domain.Repository.
func (r *MemoryRepository) FindByID(_ context.Context, id string) (*domain.Guitar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.guitars[id]
	if !ok {
		return nil, domain.ErrGuitarNotFound
	}
	return g, nil
}

// FindAll implements domain.Repository. Results are sorted by id for a stable
// API response (this matters for clients and for tests).
func (r *MemoryRepository) FindAll(_ context.Context) ([]*domain.Guitar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Guitar, 0, len(r.guitars))
	for _, g := range r.guitars {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out, nil
}

// Delete implements domain.Repository.
func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.guitars[id]; !ok {
		return domain.ErrGuitarNotFound
	}
	delete(r.guitars, id)
	return nil
}
