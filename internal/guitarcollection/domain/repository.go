package domain

import "context"

// Repository is the port through which the GuitarCollection domain persists
// and retrieves Guitar aggregates. Implementations live in the infrastructure
// layer (DynamoDB in production, an in-memory map for tests/local dev).
//
// All methods take a context.Context so that cancellation and request-scoped
// values (request id, tracing) flow naturally from the Lambda entry point.
type Repository interface {
	// Save inserts a new guitar or updates an existing one (upsert semantics).
	Save(ctx context.Context, g *Guitar) error

	// FindByID returns the guitar with the given id, or ErrGuitarNotFound.
	FindByID(ctx context.Context, id string) (*Guitar, error)

	// FindByOwner returns guitars owned by the given user id.
	FindByOwner(ctx context.Context, owner string) ([]*Guitar, error)

	// FindDistinctOwners returns sorted user ids that own at least one guitar.
	FindDistinctOwners(ctx context.Context) ([]string, error)

	// Delete removes the guitar with the given id. Deleting an unknown id
	// returns ErrGuitarNotFound so callers can produce the correct 404.
	Delete(ctx context.Context, id string) error
}
