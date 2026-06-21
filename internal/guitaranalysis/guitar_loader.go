package guitaranalysis

import (
	"context"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// GuitarLoader loads a guitar aggregate for the analysis worker.
type GuitarLoader interface {
	LoadGuitar(ctx context.Context, guitarID string) (*domain.Guitar, error)
}

// GuitarLoaderFunc adapts a function to GuitarLoader.
type GuitarLoaderFunc func(ctx context.Context, guitarID string) (*domain.Guitar, error)

func (f GuitarLoaderFunc) LoadGuitar(ctx context.Context, guitarID string) (*domain.Guitar, error) {
	return f(ctx, guitarID)
}
