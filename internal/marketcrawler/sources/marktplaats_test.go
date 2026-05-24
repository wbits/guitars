package sources

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wbits/guitars/internal/marketcrawler"
)

func TestParseMarktplaatsListingsHTML_FiltersNonGuitarResults(t *testing.T) {
	html := `<html><script id="__NEXT_DATA__" type="application/json">{
		"page": {
			"listings": [
				{
					"itemId": "a1",
					"title": "PRS Custom 24",
					"vipUrl": "/v/muziek-en-instrumenten/snaarinstrumenten-gitaren-elektrisch/a1-prs-custom-24",
					"priceInfo": {"priceCents": 250000},
					"pictures": [{"url": "https://images.marktplaats.com/api/v1/hz-mp-pro-listing/images/abc.jpg"}],
					"imageUrls": ["//images.marktplaats.com/api/v1/hz-mp-pro-listing/images/abc.jpg"]
				},
				{
					"itemId": "a2",
					"title": "Rainbird sprinkler",
					"vipUrl": "/v/tuin-en-terras/tuinsproeiers/a2-rainbird",
					"priceInfo": {"priceCents": 2500},
					"imageUrls": ["//admarkt-cdn.marktplaats.com/api/v1/icas-mp-pro-admarkt/images/x.jpg"]
				}
			]
		}
	}</script></html>`

	listings, err := parseMarktplaatsListingsHTML(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(listings) != 1 {
		t.Fatalf("want 1 guitar listing, got %d", len(listings))
	}
	if listings[0].Title != "PRS Custom 24" {
		t.Fatalf("unexpected title %q", listings[0].Title)
	}
	if listings[0].ItemID != "a1" {
		t.Fatalf("unexpected item id %q", listings[0].ItemID)
	}
	if !strings.Contains(listings[0].ImageURL, "hz-mp-pro-listing") {
		t.Fatalf("want listing photo, got %q", listings[0].ImageURL)
	}
}

func TestMarktplaatsListingImage_SkipsSponsorCreatives(t *testing.T) {
	node := map[string]any{
		"pictures": []any{
			map[string]any{"url": "https://admarkt-cdn.marktplaats.com/api/v1/icas-mp-pro-admarkt/images/milk.jpg"},
		},
		"imageUrls": []any{"//admarkt-cdn.marktplaats.com/api/v1/icas-mp-pro-admarkt/images/milk.jpg"},
	}
	if got := marktplaatsListingImage(node); got != "" {
		t.Fatalf("want empty image for sponsor-only listing, got %q", got)
	}
}

func TestMarktplaatsListingImage_PrefersListingPhoto(t *testing.T) {
	node := map[string]any{
		"pictures": []any{
			map[string]any{"url": "https://admarkt-cdn.marktplaats.com/api/v1/icas-mp-pro-admarkt/images/ad.jpg"},
			map[string]any{"url": "https://images.marktplaats.com/api/v1/hz-mp-pro-listing/images/guitar.jpg"},
		},
	}
	got := marktplaatsListingImage(node)
	if !strings.Contains(got, "hz-mp-pro-listing") {
		t.Fatalf("want listing photo, got %q", got)
	}
}

func TestMarktplaatsQuerySlug(t *testing.T) {
	got := marktplaatsQuerySlug("Gibson 60th Anniversary '52 Les Paul")
	want := "gibson-60th-anniversary-52-les-paul"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMarktplaats_Search_LiveGibsonLesPaul(t *testing.T) {
	if os.Getenv("LIVE_MARKET_CRAWL") == "" {
		t.Skip("set LIVE_MARKET_CRAWL=1 to run")
	}
	m := &Marktplaats{HTTPClient: http.DefaultClient}
	findings, err := m.Search(context.Background(), marketcrawler.GuitarSummary{
		Brand: "Gibson", TypeName: "Les Paul", BuildYear: 2017,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("findings=%d", len(findings))
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	for _, f := range findings {
		if !strings.Contains(f.ListingURL, "snaarinstrumenten-gitaren-") {
			t.Fatalf("non-guitar listing url %q", f.ListingURL)
		}
		if strings.TrimSpace(f.ListingTitle) == "" {
			t.Fatal("expected listing title")
		}
	}
}

func TestMarktplaats_Search_FallsBackForSpecificGuitar(t *testing.T) {
	if os.Getenv("LIVE_MARKET_CRAWL") == "" {
		t.Skip("set LIVE_MARKET_CRAWL=1 to run")
	}
	m := &Marktplaats{HTTPClient: http.DefaultClient}
	findings, err := m.Search(context.Background(), marketcrawler.GuitarSummary{
		Brand: "Gibson", TypeName: "60th Anniversary '52 Les Paul Gold Top", BuildYear: 2012,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("findings=%d", len(findings))
	if len(findings) == 0 {
		t.Fatal("expected fallback findings")
	}
	for _, f := range findings {
		if !strings.Contains(f.ListingURL, "snaarinstrumenten-gitaren-") {
			t.Fatalf("non-guitar listing url %q title=%q", f.ListingURL, f.ListingTitle)
		}
	}
}

func TestMarktplaats_Search_AvoidsAmbiguousBrandMatches(t *testing.T) {
	if os.Getenv("LIVE_MARKET_CRAWL") == "" {
		t.Skip("set LIVE_MARKET_CRAWL=1 to run")
	}
	m := &Marktplaats{HTTPClient: http.DefaultClient}
	findings, err := m.Search(context.Background(), marketcrawler.GuitarSummary{
		Brand: "PRS", TypeName: "Custom 24", BuildYear: 2020,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("findings=%d", len(findings))
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	for _, f := range findings {
		title := strings.ToLower(f.ListingTitle)
		if strings.Contains(title, "rainbird") || strings.Contains(title, "sproeier") {
			t.Fatalf("unexpected non-guitar listing %q", f.ListingTitle)
		}
	}
}
