package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wbits/guitars/internal/marketcrawler"
)

const (
	ebayTokenURL  = "https://api.ebay.com/identity/v1/oauth2/token"
	ebaySearchURL = "https://api.ebay.com/buy/browse/v1/item_summary/search"
)

// Ebay searches eBay listings via the Browse API (requires client credentials).
type Ebay struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	Limit        int

	mu    sync.Mutex
	token string
	exp   time.Time
}

func (e *Ebay) Name() string { return "ebay" }

// Configured reports whether eBay API credentials are present.
func (e *Ebay) Configured() bool {
	return strings.TrimSpace(e.ClientID) != "" && strings.TrimSpace(e.ClientSecret) != ""
}

func (e *Ebay) configured() bool {
	return e.Configured()
}

func (e *Ebay) marketplaceID() string {
	if id := strings.TrimSpace(os.Getenv("EBAY_MARKETPLACE_ID")); id != "" {
		return id
	}
	return "EBAY_NL"
}

func (e *Ebay) client() *http.Client {
	if e.HTTPClient != nil {
		return e.HTTPClient
	}
	return &http.Client{Timeout: 20 * time.Second}
}

// Search returns active eBay listings. Sold/completed listings require separate
// eBay market-data APIs and are not available through Browse search.
func (e *Ebay) Search(ctx context.Context, guitar marketcrawler.GuitarSummary) ([]marketcrawler.Finding, error) {
	if !e.configured() {
		return nil, nil
	}
	token, err := e.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	for _, query := range marketcrawler.SearchQueries(guitar) {
		findings, err := e.searchQuery(ctx, token, query)
		if err != nil {
			return nil, err
		}
		if len(findings) > 0 {
			return findings, nil
		}
	}
	return nil, nil
}

func (e *Ebay) searchQuery(ctx context.Context, token, query string) ([]marketcrawler.Finding, error) {
	u, err := url.Parse(ebaySearchURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", e.resultLimit()))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", e.marketplaceID())

	resp, err := e.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("ebay api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload ebaySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := make([]marketcrawler.Finding, 0, len(payload.ItemSummaries))
	for _, item := range payload.ItemSummaries {
		amount, currency, ok := item.priceMinorUnits()
		if !ok {
			continue
		}
		out = append(out, marketcrawler.Finding{
			Source:            "ebay",
			Action:            "for_sale",
			PriceAmount:       amount,
			PriceCurrency:     currency,
			ListingURL:        strings.TrimSpace(item.ItemWebURL),
			ListingTitle:      strings.TrimSpace(item.Title),
			ExternalListingID: strings.TrimSpace(item.ItemID),
			SourceImageURL:    item.imageURL(),
			ObservedAt:        now,
		})
	}
	return out, nil
}

func (e *Ebay) resultLimit() int {
	if e.Limit <= 0 {
		return 15
	}
	return e.Limit
}

func (e *Ebay) accessToken(ctx context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.token != "" && time.Now().Before(e.exp.Add(-1*time.Minute)) {
		return e.token, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "https://api.ebay.com/oauth/api_scope/buy.item.browse")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ebayTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(e.ClientID, e.ClientSecret)
	resp, err := e.client().Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("ebay token status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}
	e.token = tok.AccessToken
	e.exp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	return e.token, nil
}

// NewEbayFromEnv builds an Ebay source from EBAY_CLIENT_ID and EBAY_CLIENT_SECRET.
func NewEbayFromEnv() *Ebay {
	return &Ebay{
		ClientID:     os.Getenv("EBAY_CLIENT_ID"),
		ClientSecret: os.Getenv("EBAY_CLIENT_SECRET"),
	}
}

type ebaySearchResponse struct {
	ItemSummaries []ebayItemSummary `json:"itemSummaries"`
}

type ebayItemSummary struct {
	ItemID     string `json:"itemId"`
	Title      string `json:"title"`
	ItemWebURL string `json:"itemWebUrl"`
	Image      *struct {
		ImageURL string `json:"imageUrl"`
	} `json:"image"`
	Price      *struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"price"`
}

func (i ebayItemSummary) imageURL() string {
	if i.Image == nil {
		return ""
	}
	return strings.TrimSpace(i.Image.ImageURL)
}

func (i ebayItemSummary) priceMinorUnits() (int64, string, bool) {
	if i.Price == nil {
		return 0, "", false
	}
	value, err := parseDecimal(i.Price.Value)
	if err != nil || value <= 0 {
		return 0, "", false
	}
	currency := strings.ToUpper(strings.TrimSpace(i.Price.Currency))
	if currency == "" {
		currency = "USD"
	}
	return int64(math.Round(value * 100)), currency, true
}

func parseDecimal(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", "")
	return json.Number(raw).Float64()
}
