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
	// Marktplaats separates € from the amount with a non-breaking space; Go's \s does not match \u00a0.
	marktplaatsListingRe = regexp.MustCompile(`(?s)href="(/v/[^"]+)"[^>]*>.*?class="[^"]*hz-Listing-title[^"]*"[^>]*>([^<]+)</.*?€(?:\s|` + "\u00a0" + `)*([0-9][0-9.,]*)`)
	marktplaatsImageRe     = regexp.MustCompile(`https://images\.marktplaats\.com/api/v1/[^"'\\]+`)
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
	for _, query := range marketcrawler.SearchQueries(guitar) {
		findings, err := m.searchQuery(ctx, query)
		if err != nil {
			return nil, err
		}
		if len(findings) > 0 {
			return findings, nil
		}
	}
	return nil, nil
}

func (m *Marktplaats) searchQuery(ctx context.Context, query string) ([]marketcrawler.Finding, error) {
	slug := marktplaatsQuerySlug(query)
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
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("marktplaats status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	html := string(htmlBytes)

	listings := marktplaatsListingRe.FindAllStringSubmatch(html, -1)
	if len(listings) == 0 {
		return nil, nil
	}
	images := marktplaatsImageRe.FindAllString(html, -1)

	now := time.Now().UTC()
	limit := m.limit()
	if len(listings) < limit {
		limit = len(listings)
	}
	out := make([]marketcrawler.Finding, 0, limit)
	for i := 0; i < limit; i++ {
		link := strings.TrimSpace(listings[i][1])
		title := strings.TrimSpace(listings[i][2])
		amount, ok := parseEuroMinor(listings[i][3])
		if !ok {
			continue
		}
		listingURL := "https://www.marktplaats.nl" + link
		sourceImageURL := ""
		if i < len(images) {
			sourceImageURL = images[i]
		}
		out = append(out, marketcrawler.Finding{
			Source:            "marktplaats",
			Action:            "for_sale",
			PriceAmount:       amount,
			PriceCurrency:     "EUR",
			ListingURL:        listingURL,
			ListingTitle:      title,
			ExternalListingID: listingURL,
			SourceImageURL:    sourceImageURL,
			ObservedAt:        now,
		})
	}
	return out, nil
}

func marktplaatsQuerySlug(query string) string {
	slug := strings.ToLower(strings.TrimSpace(query))
	slug = strings.NewReplacer("'", "", "’", "", "`", "").Replace(slug)
	slug = strings.Join(strings.Fields(slug), "-")
	return slug
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
