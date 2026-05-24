package marketcrawler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient uploads findings to the GuitarCollection API.
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// GuitarFromAPI is the subset of guitar fields returned by GET /guitar/{id}.
type GuitarFromAPI struct {
	ID        string `json:"id"`
	Brand     string `json:"brand"`
	TypeName  string `json:"typeName"`
	BuildYear int    `json:"buildYear"`
}

// NewAPIClient constructs an API client for the crawler.
func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   strings.TrimSpace(token),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListGuitars returns every guitar across all collections.
func (c *APIClient) ListGuitars(ctx context.Context) ([]GuitarFromAPI, error) {
	owners, err := c.listCollections(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	out := make([]GuitarFromAPI, 0)
	for _, owner := range owners {
		guitars, err := c.listUserGuitars(ctx, owner.UserID)
		if err != nil {
			return nil, err
		}
		for _, guitar := range guitars {
			if _, ok := seen[guitar.ID]; ok {
				continue
			}
			seen[guitar.ID] = struct{}{}
			out = append(out, guitar)
		}
	}
	return out, nil
}

type collectionOwnerFromAPI struct {
	UserID string `json:"userId"`
}

func (c *APIClient) listCollections(ctx context.Context) ([]collectionOwnerFromAPI, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/collections", nil)
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}
	var out []collectionOwnerFromAPI
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *APIClient) listUserGuitars(ctx context.Context, userID string) ([]GuitarFromAPI, error) {
	url := fmt.Sprintf("%s/collections/%s/guitar", c.BaseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}
	var out []GuitarFromAPI
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// UploadMarketLogs posts findings for a guitar to POST /guitar/{id}/market-log.
func (c *APIClient) UploadMarketLogs(ctx context.Context, guitarID string, findings []Finding) error {
	if len(findings) == 0 {
		return nil
	}
	payload := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		entry := map[string]any{
			"source":        f.Source,
			"action":        f.Action,
			"priceAmount":   f.PriceAmount,
			"priceCurrency": f.PriceCurrency,
		}
		if !f.ObservedAt.IsZero() {
			entry["observedAt"] = f.ObservedAt.UTC().Format(time.RFC3339)
		}
		if f.ListingURL != "" {
			entry["listingUrl"] = f.ListingURL
		}
		if f.ListingTitle != "" {
			entry["listingTitle"] = f.ListingTitle
		}
		if f.ExternalListingID != "" {
			entry["externalListingId"] = f.ExternalListingID
		}
		payload = append(payload, entry)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/guitar/%s/market-log", c.BaseURL, guitarID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		return apiError(resp)
	}
	return nil
}

func (c *APIClient) applyAuth(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/json")
}

func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Errorf("api request failed with status %d", resp.StatusCode)
	}
	return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, msg)
}
