package domain

import (
	"testing"
	"time"
)

func validMarketLogProps() MarketLogProps {
	price, _ := NewMoney(150000, EUR)
	return MarketLogProps{
		ID:         "ml-1",
		GuitarID:   "g-1",
		ObservedAt: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC),
		Source:     MarketSourceReverb,
		Action:     MarketActionForSale,
		Price:      price,
		ListingURL: "https://reverb.com/item/123",
	}
}

func TestNewMarketLog_AcceptsForSaleAndSold(t *testing.T) {
	for _, action := range []MarketAction{MarketActionForSale, MarketActionSold} {
		p := validMarketLogProps()
		p.Action = action
		log, err := NewMarketLog(p)
		if err != nil {
			t.Fatalf("action %s: %v", action, err)
		}
		if log.Action() != action {
			t.Fatalf("want action %s, got %s", action, log.Action())
		}
	}
}

func TestNewMarketLog_AcceptsListingImageURL(t *testing.T) {
	p := validMarketLogProps()
	p.ListingImageURL = "https://cdn.example/images/market-logs/thumb.jpg"
	log, err := NewMarketLog(p)
	if err != nil {
		t.Fatal(err)
	}
	if log.ListingImageURL() != p.ListingImageURL {
		t.Fatalf("want image url, got %q", log.ListingImageURL())
	}
}

func TestNewMarketLog_RejectsUnknownSource(t *testing.T) {
	p := validMarketLogProps()
	p.Source = "unknown"
	if _, err := NewMarketLog(p); err == nil {
		t.Fatal("expected validation error")
	}
}
