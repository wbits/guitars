package marketcrawler

import (
	"testing"
	"time"
)

func observed(day string) time.Time {
	t, err := time.Parse(time.DateOnly, day)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

func TestDedupeFindingsPerRun_PrefersSoldOverForSale(t *testing.T) {
	when := observed("2026-05-23")
	in := []Finding{
		{
			Source: "reverb", Action: "for_sale", ListingTitle: "2012 Gibson Les Paul",
			ExternalListingID: "96078243", PriceAmount: 1211567, PriceCurrency: "USD",
			ListingURL: "https://reverb.com/item/96078243", ObservedAt: when,
		},
		{
			Source: "reverb", Action: "sold", ListingTitle: "2012 Gibson Les Paul",
			ExternalListingID: "96078243", PriceAmount: 1211567, PriceCurrency: "USD",
			ListingURL: "https://reverb.com/item/96078243", ObservedAt: when,
		},
	}
	out := DedupeFindingsPerRun(in)
	if len(out) != 1 {
		t.Fatalf("want 1 finding, got %d", len(out))
	}
	if out[0].Action != "sold" {
		t.Fatalf("want sold, got %s", out[0].Action)
	}
}

func TestDedupeFindingsPerRun_CollapsesSameActionDuplicates(t *testing.T) {
	when := observed("2026-05-23")
	in := []Finding{
		{Source: "marktplaats", Action: "for_sale", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: when},
		{Source: "marktplaats", Action: "for_sale", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: when},
	}
	out := DedupeFindingsPerRun(in)
	if len(out) != 1 {
		t.Fatalf("want 1 finding, got %d", len(out))
	}
}

func TestDedupeFindingsPerRun_KeepsDifferentSources(t *testing.T) {
	when := observed("2026-05-23")
	in := []Finding{
		{Source: "reverb", Action: "for_sale", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: when},
		{Source: "ebay", Action: "for_sale", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: when},
	}
	out := DedupeFindingsPerRun(in)
	if len(out) != 2 {
		t.Fatalf("want 2 findings, got %d", len(out))
	}
}

func TestDedupeFindingsPerRun_KeepsDifferentObservedDates(t *testing.T) {
	in := []Finding{
		{Source: "reverb", Action: "sold", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: observed("2026-05-22")},
		{Source: "reverb", Action: "sold", ListingTitle: "Gibson Les Paul", PriceAmount: 250000, PriceCurrency: "EUR", ObservedAt: observed("2026-05-23")},
	}
	out := DedupeFindingsPerRun(in)
	if len(out) != 2 {
		t.Fatalf("want 2 findings across runs/dates, got %d", len(out))
	}
}

func TestDedupeFindingsPerRun_MergesImageFromDiscardedRow(t *testing.T) {
	when := observed("2026-05-23")
	in := []Finding{
		{
			Source: "reverb", Action: "for_sale", ListingTitle: "Gibson Les Paul",
			ExternalListingID: "1", PriceAmount: 100, PriceCurrency: "USD", ObservedAt: when,
			SourceImageURL: "https://example.com/photo.jpg",
		},
		{
			Source: "reverb", Action: "sold", ListingTitle: "Gibson Les Paul",
			ExternalListingID: "1", PriceAmount: 100, PriceCurrency: "USD", ObservedAt: when,
		},
	}
	out := DedupeFindingsPerRun(in)
	if len(out) != 1 {
		t.Fatalf("want 1 finding, got %d", len(out))
	}
	if out[0].Action != "sold" {
		t.Fatalf("want sold, got %s", out[0].Action)
	}
	if out[0].SourceImageURL != "https://example.com/photo.jpg" {
		t.Fatalf("want merged image url, got %q", out[0].SourceImageURL)
	}
}
