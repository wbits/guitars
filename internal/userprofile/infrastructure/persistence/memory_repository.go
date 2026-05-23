package persistence

import (
	"context"
	"sort"
	"sync"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// MemoryRepository is an in-memory profile store for tests and local runs.
type MemoryRepository struct {
	mu       sync.RWMutex
	profiles map[string]*domain.Profile
}

// NewMemoryRepository returns an empty MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{profiles: map[string]*domain.Profile{}}
}

// FindByUserID implements domain.Repository.
func (r *MemoryRepository) FindByUserID(_ context.Context, userID string) (*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.profiles[userID]
	if !ok {
		return nil, nil
	}
	return cloneProfile(profile), nil
}

// FindByUsername implements domain.Repository.
func (r *MemoryRepository) FindByUsername(_ context.Context, username string) (*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, profile := range r.profiles {
		if profile.Username() == username {
			return cloneProfile(profile), nil
		}
	}
	return nil, nil
}

// FindByUserIDs implements domain.Repository.
func (r *MemoryRepository) FindByUserIDs(_ context.Context, userIDs []string) (map[string]*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]*domain.Profile, len(userIDs))
	for _, userID := range userIDs {
		if profile, ok := r.profiles[userID]; ok {
			out[userID] = cloneProfile(profile)
		}
	}
	return out, nil
}

// Save implements domain.Repository.
func (r *MemoryRepository) Save(_ context.Context, profile *domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[profile.UserID()] = cloneProfile(profile)
	return nil
}

func cloneProfile(profile *domain.Profile) *domain.Profile {
	cloned, err := domain.NewProfile(domain.ProfileProps{
		UserID:   profile.UserID(),
		Username: profile.Username(),
		Email:    profile.Email(),
	})
	if err != nil {
		panic(err)
	}
	return cloned
}

// SortedUserIDs returns profile user ids in stable order for tests.
func (r *MemoryRepository) SortedUserIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.profiles))
	for userID := range r.profiles {
		out = append(out, userID)
	}
	sort.Strings(out)
	return out
}
