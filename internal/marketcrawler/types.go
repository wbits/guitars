package marketcrawler

import "time"

// Finding is a normalized marketplace observation produced by a source adapter.
type Finding struct {
	Source            string
	Action            string
	PriceAmount       int64
	PriceCurrency     string
	ListingURL        string
	ListingTitle      string
	ExternalListingID string
	SourceImageURL    string
	ListingImageURL   string
	ObservedAt        time.Time
}

// GuitarSummary is the collection guitar data used to build search queries.
type GuitarSummary struct {
	ID        string
	Brand     string
	TypeName  string
	BuildYear int
}
