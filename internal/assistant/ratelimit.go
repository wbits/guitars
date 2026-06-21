package assistant

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrRateLimited is returned when the caller exceeded the daily assistant quota.
var ErrRateLimited = errors.New("assistant daily limit reached")

// RateLimiter tracks per-user daily usage.
type RateLimiter interface {
	Allow(ctx context.Context, userID string) error
}

// MemoryRateLimiter is an in-process limiter for tests and local dev.
type MemoryRateLimiter struct {
	Limit int
	usage map[string]int
	day   string
}

func NewMemoryRateLimiter(limit int) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		Limit: limit,
		usage: map[string]int{},
		day:   utcDay(time.Now()),
	}
}

func (m *MemoryRateLimiter) Allow(_ context.Context, userID string) error {
	today := utcDay(time.Now())
	if today != m.day {
		m.day = today
		m.usage = map[string]int{}
	}
	m.usage[userID]++
	if m.usage[userID] > m.Limit {
		return fmt.Errorf("%w (%d per day)", ErrRateLimited, m.Limit)
	}
	return nil
}

func utcDay(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}
