package application

import (
	"context"
	"errors"
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
	marketSvc := NewMarketLogService(guitars, logs, ids, nil, nil, nil)

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
	crawlerUserIDs := ParseCrawlerUserIDs("crawler-sub")
	crawlChecker := stubCrawlChecker{enabled: map[string]bool{"real-owner": true}}
	marketSvc := NewMarketLogService(guitars, logs, ids, crawlerEmails, crawlerUserIDs, crawlChecker)

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

func TestMarketLogService_CrawlerBlockedWhenMarketCrawlDisabled(t *testing.T) {
	guitars := persistence.NewMemoryRepository()
	logs := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "ml-1"}}
	crawlerEmails := ParseCrawlerEmails("info@wbits.net")
	crawlerUserIDs := ParseCrawlerUserIDs("crawler-sub")
	crawlChecker := stubCrawlChecker{enabled: map[string]bool{"real-owner": false}}
	marketSvc := NewMarketLogService(guitars, logs, ids, crawlerEmails, crawlerUserIDs, crawlChecker)

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
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("want ErrGuitarNotFound, got %v", err)
	}
}

func TestMarketLogService_ClearCollectionMarketLogs(t *testing.T) {
	guitars := persistence.NewMemoryRepository()
	logs := persistence.NewMemoryMarketLogRepository()
	ids := &sequentialIDs{ids: []string{"g-1", "g-2", "ml-1", "ml-2", "ml-3"}}
	marketSvc := NewMarketLogService(guitars, logs, ids, nil, nil, nil)

	ctx := context.Background()
	price, _ := domain.NewMoney(199900, domain.EUR)
	for _, id := range []string{"g-1", "g-2"} {
		g, err := domain.NewGuitar(domain.GuitarProps{
			ID: id, Owner: marketLogTestOwner, Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996, Price: price,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := guitars.Save(ctx, g); err != nil {
			t.Fatal(err)
		}
	}

	for _, guitarID := range []string{"g-1", "g-2", "g-2"} {
		if _, err := marketSvc.AddMarketLog(ctx, marketLogTestOwner, "owner@example.com", guitarID, MarketLogInput{
			Source:        "reverb",
			Action:        "for_sale",
			PriceAmount:   150000,
			PriceCurrency: "EUR",
		}); err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := marketSvc.ClearCollectionMarketLogs(ctx, marketLogTestOwner)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 3 {
		t.Fatalf("want 3 deleted logs, got %d", deleted)
	}
	for _, guitarID := range []string{"g-1", "g-2"} {
		listed, err := marketSvc.ListMarketLogs(ctx, marketLogTestOwner, guitarID)
		if err != nil {
			t.Fatal(err)
		}
		if len(listed) != 0 {
			t.Fatalf("want 0 logs for %s, got %d", guitarID, len(listed))
		}
	}
}

type stubCrawlChecker struct {
	enabled map[string]bool
}

func (s stubCrawlChecker) MarketCrawlEnabledForUser(_ context.Context, userID string) (bool, error) {
	return s.enabled[userID], nil
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
