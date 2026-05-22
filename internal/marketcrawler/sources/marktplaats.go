package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wbits/guitars/internal/marketcrawler"
)

var (
	marktplaatsPriceRe = regexp.MustCompile(`€\s*([0-9][0-9.,]*)`)
	marktplaatsLinkRe   = regexp.MustCompile(`href="(/v/[^"]+)"`)
)

// Marktplaats searches marktplaats.nl by scraping the public search results page.
// This adapter is best-effort and may break if the site layout changes.
type Marktplaats struct {
	HTTPClient *http.Client
	MaxResults int
}

func (m *Marktplaats) Name() string { return "marktplaats" }

func (m *Marktplaats) client() *http.Client {
	if m.HTTPClient != nil {
		return m.HTTPClient
	}
	return &http.Client{Timeout: 20 * time.Second}
}

func (m *Marktplaats) limit() int {
	if m.MaxResults <= 0 {
		return 10
	}
	return m.MaxResults
}

// Search returns for-sale listings scraped from marktplaats.nl search results.
func (m *Marktplaats) Search(ctx context.Context, guitar marketcrawler.GuitarSummary) ([]marketcrawler.Finding, error) {
	query := marketcrawler.SearchQuery(guitar)
	slug := strings.ReplaceAll(strings.ToLower(query), " ", "-")
	u := fmt.Sprintf("https://www.marktplaats.nl/q/%s/", url.PathEscape(slug))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "guitars-market-crawler/1.0")
	req.Header.Set("Accept-Language", "nl-NL,nl;q=0.9,en;q=0.8")

	resp, err := m.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("marktplaats status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	html := string(htmlBytes)

	prices := marktplaatsPriceRe.FindAllStringSubmatch(html, -1)
	links := marktplaatsLinkRe.FindAllStringSubmatch(html, -1)
	if len(prices) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	limit := m.limit()
	if len(prices) < limit {
		limit = len(prices)
	}
	out := make([]marketcrawler.Finding, 0, limit)
	for i := 0; i < limit; i++ {
		amount, ok := parseEuroMinor(prices[i][1])
		if !ok {
			continue
		}
		listingURL := ""
		if i < len(links) {
			listingURL = "https://www.marktplaats.nl" + links[i][1]
		}
		out = append(out, marketcrawler.Finding{
			Source:            "marktplaats",
			Action:            "for_sale",
			PriceAmount:       amount,
			PriceCurrency:     "EUR",
			ListingURL:        listingURL,
			ListingTitle:      query,
			ExternalListingID: listingURL,
			ObservedAt:        now,
		})
	}
	return out, nil
}

func parseEuroMinor(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ".", "")
	raw = strings.ReplaceAll(raw, ",", ".")
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return int64(value * 100), true
}
