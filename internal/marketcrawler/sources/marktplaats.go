package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/wbits/guitars/internal/marketcrawler"
)

var (
	marktplaatsNextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)
)

var marktplaatsGuitarCategories = []string{
	"muziek-en-instrumenten/snaarinstrumenten-gitaren-elektrisch",
	"muziek-en-instrumenten/snaarinstrumenten-gitaren-akoestisch",
}

// Marktplaats searches marktplaats.nl guitar categories via embedded page JSON.
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

// Search returns for-sale guitar listings scraped from marktplaats.nl search results.
func (m *Marktplaats) Search(ctx context.Context, guitar marketcrawler.GuitarSummary) ([]marketcrawler.Finding, error) {
	var best []marketcrawler.Finding
	for _, query := range marketcrawler.SearchQueries(guitar) {
		for _, category := range marktplaatsGuitarCategories {
			findings, err := m.searchCategoryQuery(ctx, category, query)
			if err != nil {
				return nil, err
			}
			if len(findings) > len(best) {
				best = findings
			}
		}
	}
	if len(best) == 0 {
		return nil, nil
	}
	return best, nil
}

func (m *Marktplaats) searchCategoryQuery(ctx context.Context, category, query string) ([]marketcrawler.Finding, error) {
	slug := marktplaatsQuerySlug(query)
	u := fmt.Sprintf("https://www.marktplaats.nl/l/%s/q/%s/", category, url.PathEscape(slug))

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
	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	listings, err := parseMarktplaatsListingsHTML(string(htmlBytes))
	if err != nil {
		return nil, err
	}
	if len(listings) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	limit := m.limit()
	if len(listings) < limit {
		limit = len(listings)
	}
	out := make([]marketcrawler.Finding, 0, limit)
	for i := 0; i < limit; i++ {
		listing := listings[i]
		if listing.PriceCents <= 0 {
			continue
		}
		listingURL := marktplaatsAbsoluteURL(listing.VipURL)
		out = append(out, marketcrawler.Finding{
			Source:            "marktplaats",
			Action:            "for_sale",
			PriceAmount:       listing.PriceCents,
			PriceCurrency:     "EUR",
			ListingURL:        listingURL,
			ListingTitle:      listing.Title,
			ExternalListingID: listing.ItemID,
			SourceImageURL:    listing.ImageURL,
			ObservedAt:        now,
		})
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

type marktplaatsListing struct {
	ItemID     string
	Title      string
	VipURL     string
	PriceCents int64
	ImageURL   string
}

func parseMarktplaatsListingsHTML(html string) ([]marktplaatsListing, error) {
	match := marktplaatsNextDataRe.FindStringSubmatch(html)
	if len(match) < 2 {
		return nil, nil
	}
	var payload any
	if err := json.Unmarshal([]byte(match[1]), &payload); err != nil {
		return nil, fmt.Errorf("parse marktplaats next data: %w", err)
	}
	raw := collectMarktplaatsJSONListings(payload)
	seen := make(map[string]struct{}, len(raw))
	out := make([]marktplaatsListing, 0, len(raw))
	for _, listing := range raw {
		if !isMarktplaatsGuitarListing(listing.VipURL) {
			continue
		}
		if _, ok := seen[listing.ItemID]; ok {
			continue
		}
		seen[listing.ItemID] = struct{}{}
		out = append(out, listing)
	}
	return out, nil
}

func collectMarktplaatsJSONListings(v any) []marktplaatsListing {
	switch node := v.(type) {
	case map[string]any:
		if listing, ok := marktplaatsListingFromMap(node); ok {
			return []marktplaatsListing{listing}
		}
		out := make([]marktplaatsListing, 0)
		for _, child := range node {
			out = append(out, collectMarktplaatsJSONListings(child)...)
		}
		return out
	case []any:
		out := make([]marktplaatsListing, 0)
		for _, child := range node {
			out = append(out, collectMarktplaatsJSONListings(child)...)
		}
		return out
	default:
		return nil
	}
}

func marktplaatsListingFromMap(node map[string]any) (marktplaatsListing, bool) {
	itemID, _ := node["itemId"].(string)
	title, _ := node["title"].(string)
	vipURL, _ := node["vipUrl"].(string)
	if itemID == "" || title == "" || vipURL == "" {
		return marktplaatsListing{}, false
	}
	priceInfo, _ := node["priceInfo"].(map[string]any)
	priceCents := jsonInt64(priceInfo["priceCents"])
	if priceCents <= 0 {
		return marktplaatsListing{}, false
	}
	return marktplaatsListing{
		ItemID:     itemID,
		Title:      strings.TrimSpace(title),
		VipURL:     vipURL,
		PriceCents: priceCents,
		ImageURL:   marktplaatsListingImage(node),
	}, true
}

func marktplaatsListingImage(node map[string]any) string {
	candidates := make([]string, 0, 4)
	if pictures, ok := node["pictures"].([]any); ok {
		for _, pic := range pictures {
			picMap, ok := pic.(map[string]any)
			if !ok {
				continue
			}
			if raw, ok := picMap["url"].(string); ok {
				candidates = append(candidates, normalizeMarktplaatsURL(raw))
			}
		}
	}
	if imageURLs, ok := node["imageUrls"].([]any); ok {
		for _, raw := range imageURLs {
			if s, ok := raw.(string); ok {
				candidates = append(candidates, normalizeMarktplaatsURL(s))
			}
		}
	}
	for _, candidate := range candidates {
		if isMarktplaatsListingPhoto(candidate) {
			return candidate
		}
	}
	return ""
}

func isMarktplaatsGuitarListing(vipURL string) bool {
	vipURL = strings.ToLower(strings.TrimSpace(vipURL))
	return strings.Contains(vipURL, "/snaarinstrumenten-gitaren-")
}

func isMarktplaatsListingPhoto(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || strings.Contains(raw, "admarkt-cdn.marktplaats.com") {
		return false
	}
	return strings.Contains(raw, "images.marktplaats.com") ||
		strings.Contains(raw, "hz-mp-pro-listing") ||
		strings.Contains(raw, "hzcdn.io")
}

func normalizeMarktplaatsURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	return raw
}

func marktplaatsAbsoluteURL(vipURL string) string {
	vipURL = strings.TrimSpace(vipURL)
	if strings.HasPrefix(vipURL, "http://") || strings.HasPrefix(vipURL, "https://") {
		return vipURL
	}
	return "https://www.marktplaats.nl" + vipURL
}

func marktplaatsQuerySlug(query string) string {
	slug := strings.ToLower(strings.TrimSpace(query))
	slug = strings.NewReplacer("'", "", "’", "", "`", "").Replace(slug)
	slug = strings.Join(strings.Fields(slug), "-")
	return slug
}

func jsonInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}
