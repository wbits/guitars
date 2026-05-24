package application

import "time"

// MarketLogInput is the application-layer shape for creating a market log entry.
type MarketLogInput struct {
	ObservedAt        time.Time
	Source            string
	Action            string
	PriceAmount       int64
	PriceCurrency     string
	ListingURL        string
	ListingTitle      string
	ExternalListingID string
	ListingImageURL   string
}
