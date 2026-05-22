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

	// FindAll returns every guitar in the collection. For a personal
	// collection the cardinality is small, so no pagination is exposed yet.
	FindAll(ctx context.Context) ([]*Guitar, error)

	// Delete removes the guitar with the given id. Deleting an unknown id
	// returns ErrGuitarNotFound so callers can produce the correct 404.
	Delete(ctx context.Context, id string) error
}
