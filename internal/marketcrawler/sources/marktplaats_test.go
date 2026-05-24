package sources

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/wbits/guitars/internal/marketcrawler"
)

func TestMarktplaats_priceRegex(t *testing.T) {
	for _, html := range []string{
		"€\u00a04.099,00",
		"€ 1.250,00",
		"€12,50",
	} {
		m := marktplaatsPriceRe.FindStringSubmatch(html)
		if m == nil {
			t.Fatalf("no match for %q", html)
		}
		amount, ok := parseEuroMinor(m[1])
		if !ok || amount <= 0 {
			t.Fatalf("parseEuroMinor(%q) = %d, %v", m[1], amount, ok)
		}
	}
}

func TestMarktplaats_Search_LiveGibsonLesPaul(t *testing.T) {
	if os.Getenv("LIVE_MARKET_CRAWL") == "" {
		t.Skip("set LIVE_MARKET_CRAWL=1 to run")
	}
	req, err := http.NewRequest(http.MethodGet, "https://www.marktplaats.nl/q/gibson-les-paul-2017/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "guitars-market-crawler/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	prices := marktplaatsPriceRe.FindAllStringSubmatch(html, -1)
	t.Logf("html size=%d prices=%d", len(html), len(prices))

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
}
