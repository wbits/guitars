package guitaranalysis

import "context"

// Repository persists guitar analysis records.
type Repository interface {
	FindByGuitarID(ctx context.Context, guitarID string) (*Record, error)
	FindByGuitarIDs(ctx context.Context, guitarIDs []string) (map[string]*Record, error)
	Save(ctx context.Context, record *Record) error
	DeleteByGuitarID(ctx context.Context, guitarID string) error
}
