package application

import (
	"context"
	"testing"
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
	"github.com/wbits/guitars/internal/guitarcollection/infrastructure/persistence"
)

const marketLogTestOwner = "user-1"

func TestMarketLogService_AddAndList(t *testing.T) {
	guitars := persistence.NewMemoryRepository()
	logs := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "ml-1"}}
	marketSvc := NewMarketLogService(guitars, logs, ids, nil)

	ctx := context.Background()
	price, _ := domain.NewMoney(199900, domain.EUR)
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g-1", Owner: marketLogTestOwner, Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996, Price: price,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := guitars.Save(ctx, g); err != nil {
		t.Fatal(err)
	}

	observed := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	created, err := marketSvc.AddMarketLog(ctx, marketLogTestOwner, "owner@example.com", "g-1", MarketLogInput{
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

	listed, err := marketSvc.ListMarketLogs(ctx, marketLogTestOwner, "g-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("want 1 log, got %d", len(listed))
	}
}

func TestMarketLogService_CrawlerCanWriteForOtherOwner(t *testing.T) {
	guitars := persistence.NewMemoryRepository()
	logs := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "ml-1"}}
	crawlerEmails := ParseCrawlerEmails("info@wbits.net")
	marketSvc := NewMarketLogService(guitars, logs, ids, crawlerEmails)

	ctx := context.Background()
	price, _ := domain.NewMoney(199900, domain.EUR)
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g-1", Owner: "real-owner", Brand: "Gibson", TypeName: "Les Paul", BuildYear: 2017, Price: price,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := guitars.Save(ctx, g); err != nil {
		t.Fatal(err)
	}

	_, err = marketSvc.AddMarketLog(ctx, "crawler-sub", "info@wbits.net", "g-1", MarketLogInput{
		Source:        "reverb",
		Action:        "for_sale",
		PriceAmount:   150000,
		PriceCurrency: "EUR",
	})
	if err != nil {
		t.Fatal(err)
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
