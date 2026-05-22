package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wbits/guitars/internal/marketcrawler"
)

const reverbListingsURL = "https://api.reverb.com/api/listings"

// Reverb searches Reverb.com listings via the public Reverb API.
type Reverb struct {
	HTTPClient *http.Client
	PerPage    int
}

func (r *Reverb) Name() string { return "reverb" }

func (r *Reverb) client() *http.Client {
	if r.HTTPClient != nil {
		return r.HTTPClient
	}
	return &http.Client{Timeout: 20 * time.Second}
}

func (r *Reverb) limit() int {
	if r.PerPage <= 0 {
		return 15
	}
	return r.PerPage
}

// Search returns active and sold Reverb listings matching the guitar.
func (r *Reverb) Search(ctx context.Context, guitar marketcrawler.GuitarSummary) ([]marketcrawler.Finding, error) {
	query := marketcrawler.SearchQuery(guitar)
	findings, err := r.fetch(ctx, query, false)
	if err != nil {
		return nil, err
	}
	sold, err := r.fetch(ctx, query, true)
	if err != nil {
		return findings, nil
	}
	return append(findings, sold...), nil
}

func (r *Reverb) fetch(ctx context.Context, query string, soldOnly bool) ([]marketcrawler.Finding, error) {
	u, err := url.Parse(reverbListingsURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("per_page", strconv.Itoa(r.limit()))
	if soldOnly {
		q.Set("filter", "sold")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/hal+json")
	req.Header.Set("Accept-Version", "3.0")

	resp, err := r.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("reverb api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload reverbListingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := make([]marketcrawler.Finding, 0, len(payload.Listings))
	for _, listing := range payload.Listings {
		action := listing.action(soldOnly)
		amount, currency, ok := listing.priceMinorUnits()
		if !ok {
			continue
		}
		out = append(out, marketcrawler.Finding{
			Source:            "reverb",
			Action:            action,
			PriceAmount:       amount,
			PriceCurrency:     currency,
			ListingURL:        listing.webURL(),
			ListingTitle:      strings.TrimSpace(listing.Title),
			ExternalListingID: strconv.FormatInt(listing.ID, 10),
			ObservedAt:        now,
		})
	}
	return out, nil
}

type reverbListingsResponse struct {
	Listings []reverbListing `json:"listings"`
}

type reverbListing struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	State struct {
		Slug string `json:"slug"`
	} `json:"state"`
	Links struct {
		Web struct {
			Href string `json:"href"`
		} `json:"web"`
	} `json:"_links"`
	Price *struct {
		AmountCents int64       `json:"amount_cents"`
		Amount      json.Number `json:"amount"`
		Currency    string      `json:"currency"`
	} `json:"price"`
}

func (l reverbListing) webURL() string {
	return strings.TrimSpace(l.Links.Web.Href)
}

func (l reverbListing) action(soldOnly bool) string {
	if soldOnly || strings.EqualFold(l.State.Slug, "sold") {
		return "sold"
	}
	return "for_sale"
}

func (l reverbListing) priceMinorUnits() (int64, string, bool) {
	if l.Price == nil {
		return 0, "", false
	}
	currency := strings.ToUpper(strings.TrimSpace(l.Price.Currency))
	if currency == "" {
		currency = "USD"
	}
	if l.Price.AmountCents > 0 {
		return l.Price.AmountCents, currency, true
	}
	amount, err := l.Price.Amount.Float64()
	if err != nil || amount <= 0 {
		return 0, "", false
	}
	return int64(amount * 100), currency, true
}
