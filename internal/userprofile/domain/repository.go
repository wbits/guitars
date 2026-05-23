package domain

import "context"

// Repository persists user profile aggregates.
type Repository interface {
	FindByUserID(ctx context.Context, userID string) (*Profile, error)
	FindByUsername(ctx context.Context, username string) (*Profile, error)
	FindByUserIDs(ctx context.Context, userIDs []string) (map[string]*Profile, error)
	Save(ctx context.Context, profile *Profile) error
}
