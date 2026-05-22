package application

import (
	"context"
	"testing"
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
)

func TestMarketLogService_AddAndList(t *testing.T) {
	guitars := persistence.NewMemoryRepository()
	logs := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "ml-1"}}
	marketSvc := NewMarketLogService(guitars, logs, ids)

	ctx := context.Background()
	price, _ := domain.NewMoney(199900, domain.EUR)
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g-1", Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996, Price: price,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := guitars.Save(ctx, g); err != nil {
		t.Fatal(err)
	}

	observed := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	created, err := marketSvc.AddMarketLog(ctx, "g-1", MarketLogInput{
		ObservedAt:    observed,
		Source:        "reverb",
		Action:        "sold",
		PriceAmount:   150000,
		PriceCurrency: "EUR",
		ListingURL:    "https://reverb.com/item/1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Action() != domain.MarketActionSold {
		t.Fatalf("want sold, got %s", created.Action())
	}

	listed, err := marketSvc.ListMarketLogs(ctx, "g-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("want 1 log, got %d", len(listed))
	}
}

type sequentialIDs struct {
	ids []string
	i   int
}

func (s *sequentialIDs) NewID() string {
	id := s.ids[s.i%len(s.ids)]
	s.i++
	return id
}
