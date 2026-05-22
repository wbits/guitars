package application

import (
	"context"
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// MarketLogService coordinates market price log use cases.
type MarketLogService struct {
	guitars    domain.Repository
	marketLogs domain.MarketLogRepository
	ids        IDGenerator
	clock      func() time.Time
}

// NewMarketLogService wires the market log application service.
func NewMarketLogService(guitars domain.Repository, marketLogs domain.MarketLogRepository, ids IDGenerator) *MarketLogService {
	return &MarketLogService{
		guitars:    guitars,
		marketLogs: marketLogs,
		ids:        ids,
		clock:      time.Now,
	}
}

// AddMarketLog records a single marketplace observation for a guitar.
func (s *MarketLogService) AddMarketLog(ctx context.Context, guitarID string, in MarketLogInput) (*domain.MarketLog, error) {
	if _, err := s.guitars.FindByID(ctx, guitarID); err != nil {
		return nil, err
	}
	log, err := s.buildMarketLog(guitarID, in)
	if err != nil {
		return nil, err
	}
	if err := s.marketLogs.Save(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

// AddMarketLogs records multiple observations for a guitar in one call.
func (s *MarketLogService) AddMarketLogs(ctx context.Context, guitarID string, inputs []MarketLogInput) ([]*domain.MarketLog, error) {
	if _, err := s.guitars.FindByID(ctx, guitarID); err != nil {
		return nil, err
	}
	out := make([]*domain.MarketLog, 0, len(inputs))
	for _, in := range inputs {
		log, err := s.buildMarketLog(guitarID, in)
		if err != nil {
			return nil, err
		}
		if err := s.marketLogs.Save(ctx, log); err != nil {
			return nil, err
		}
		out = append(out, log)
	}
	return out, nil
}

// ListMarketLogs returns observations for a guitar, newest first.
func (s *MarketLogService) ListMarketLogs(ctx context.Context, guitarID string) ([]*domain.MarketLog, error) {
	if _, err := s.guitars.FindByID(ctx, guitarID); err != nil {
		return nil, err
	}
	return s.marketLogs.FindByGuitarID(ctx, guitarID)
}

func (s *MarketLogService) buildMarketLog(guitarID string, in MarketLogInput) (*domain.MarketLog, error) {
	price, err := domain.NewMoney(in.PriceAmount, domain.Currency(in.PriceCurrency))
	if err != nil {
		return nil, err
	}
	observedAt := in.ObservedAt
	if observedAt.IsZero() {
		observedAt = s.clock().UTC()
	}
	return domain.NewMarketLog(domain.MarketLogProps{
		ID:                s.ids.NewID(),
		GuitarID:          guitarID,
		ObservedAt:        observedAt,
		Source:            domain.MarketSource(in.Source),
		Action:            domain.MarketAction(in.Action),
		Price:             price,
		ListingURL:        in.ListingURL,
		ListingTitle:      in.ListingTitle,
		ExternalListingID: in.ExternalListingID,
	})
}
