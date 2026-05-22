package domain

import (
	"strings"
	"time"
)

// MarketAction describes whether a listing is currently for sale or has sold.
type MarketAction string

const (
	MarketActionForSale MarketAction = "for_sale"
	MarketActionSold    MarketAction = "sold"
)

// MarketSource identifies the marketplace where a listing was observed.
type MarketSource string

const (
	MarketSourceReverb      MarketSource = "reverb"
	MarketSourceEbay        MarketSource = "ebay"
	MarketSourceMarktplaats MarketSource = "marktplaats"
)

var knownMarketSources = map[MarketSource]struct{}{
	MarketSourceReverb:      {},
	MarketSourceEbay:        {},
	MarketSourceMarktplaats: {},
}

// MarketLog records a price observation for a guitar in the collection on an
// external marketplace.
type MarketLog struct {
	id                string
	guitarID          string
	observedAt        time.Time
	source            MarketSource
	action            MarketAction
	price             Money
	listingURL        string
	listingTitle      string
	externalListingID string
}

// MarketLogProps is the data-transfer shape used to create a MarketLog.
type MarketLogProps struct {
	ID                string
	GuitarID          string
	ObservedAt        time.Time
	Source            MarketSource
	Action            MarketAction
	Price             Money
	ListingURL        string
	ListingTitle      string
	ExternalListingID string
}

// NewMarketLog validates props and returns a MarketLog.
func NewMarketLog(p MarketLogProps) (*MarketLog, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, newValidationError("id", "is required")
	}
	if strings.TrimSpace(p.GuitarID) == "" {
		return nil, newValidationError("guitarId", "is required")
	}
	if p.ObservedAt.IsZero() {
		return nil, newValidationError("observedAt", "is required")
	}
	if _, ok := knownMarketSources[p.Source]; !ok {
		return nil, newValidationError("source", "must be reverb, ebay, or marktplaats")
	}
	switch p.Action {
	case MarketActionForSale, MarketActionSold:
	default:
		return nil, newValidationError("action", "must be for_sale or sold")
	}
	if (p.Price == Money{}) {
		return nil, newValidationError("price", "is required")
	}
	if strings.TrimSpace(p.ListingURL) != "" {
		if _, err := validatePictureURLs([]string{p.ListingURL}); err != nil {
			return nil, newValidationError("listingUrl", "must be a valid absolute URL")
		}
	}
	return &MarketLog{
		id:                strings.TrimSpace(p.ID),
		guitarID:          strings.TrimSpace(p.GuitarID),
		observedAt:        p.ObservedAt.UTC(),
		source:            p.Source,
		action:            p.Action,
		price:             p.Price,
		listingURL:        strings.TrimSpace(p.ListingURL),
		listingTitle:      strings.TrimSpace(p.ListingTitle),
		externalListingID: strings.TrimSpace(p.ExternalListingID),
	}, nil
}

func (m *MarketLog) ID() string                { return m.id }
func (m *MarketLog) GuitarID() string          { return m.guitarID }
func (m *MarketLog) ObservedAt() time.Time     { return m.observedAt }
func (m *MarketLog) Source() MarketSource      { return m.source }
func (m *MarketLog) Action() MarketAction      { return m.action }
func (m *MarketLog) Price() Money              { return m.price }
func (m *MarketLog) ListingURL() string        { return m.listingURL }
func (m *MarketLog) ListingTitle() string      { return m.listingTitle }
func (m *MarketLog) ExternalListingID() string { return m.externalListingID }
