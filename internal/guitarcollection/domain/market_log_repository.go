package domain

import "context"

// MarketLogRepository persists price observations scraped from marketplaces.
type MarketLogRepository interface {
	Save(ctx context.Context, log *MarketLog) error
	FindByGuitarID(ctx context.Context, guitarID string) ([]*MarketLog, error)
}
