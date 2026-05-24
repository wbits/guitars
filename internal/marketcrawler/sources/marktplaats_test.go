package sources

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wbits/guitars/internal/marketcrawler"
)

func TestMarktplaats_listingRegex(t *testing.T) {
	html := `<a href="/v/muziek/a1524055119-gibson-custom-shop"><div class="hz-Listing-title">Gibson Custom Shop CS336</div></a><span>€` + "\u00a0" + `4.099,00</span>`
	m := marktplaatsListingRe.FindStringSubmatch(html)
	if m == nil {
		t.Fatal("expected listing match")
	}
	if m[1] != "/v/muziek/a1524055119-gibson-custom-shop" {
		t.Fatalf("unexpected link %q", m[1])
	}
	if m[2] != "Gibson Custom Shop CS336" {
		t.Fatalf("unexpected title %q", m[2])
	}
	amount, ok := parseEuroMinor(m[3])
	if !ok || amount != 409900 {
		t.Fatalf("parseEuroMinor(%q) = %d, %v", m[3], amount, ok)
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
	m := &Marktplaats{HTTPClient: httpDefaultClient()}
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
	if strings.TrimSpace(findings[0].ListingTitle) == "" {
		t.Fatal("expected listing title")
	}
}

func TestMarktplaats_Search_FallsBackForSpecificGuitar(t *testing.T) {
	if os.Getenv("LIVE_MARKET_CRAWL") == "" {
		t.Skip("set LIVE_MARKET_CRAWL=1 to run")
	}
	m := &Marktplaats{HTTPClient: httpDefaultClient()}
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
}

func httpDefaultClient() *http.Client {
	return http.DefaultClient
}
