package httpapi

import (
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// guitarRequest is the JSON payload accepted by POST /guitar and PUT /guitar/{id}.
type guitarRequest struct {
	SerialNumber      string   `json:"serialNumber,omitempty"`
	Pictures          []string `json:"pictures,omitempty"`
	CoverPictureIndex int      `json:"coverPictureIndex,omitempty"`
	Description       string   `json:"description,omitempty"`
	Brand             string   `json:"brand"`
	TypeName          string   `json:"typeName"`
	BuildYear         int      `json:"buildYear"`
	PriceAmount       int64    `json:"priceAmount"`
	PriceCurrency     string   `json:"priceCurrency"`
}

// guitarResponse is the JSON projection of a Guitar aggregate sent to clients.
type guitarResponse struct {
	ID                string   `json:"id"`
	SerialNumber      string   `json:"serialNumber,omitempty"`
	Pictures          []string `json:"pictures"`
	CoverPictureIndex int      `json:"coverPictureIndex"`
	Description       string   `json:"description,omitempty"`
	Brand             string   `json:"brand"`
	TypeName          string   `json:"typeName"`
	BuildYear         int      `json:"buildYear"`
	PriceAmount       int64    `json:"priceAmount"`
	PriceCurrency     string   `json:"priceCurrency"`
}

func toResponse(g *domain.Guitar) guitarResponse {
	pictures := g.Pictures()
	if pictures == nil {
		pictures = []string{}
	}
	return guitarResponse{
		ID:                g.ID(),
		SerialNumber:      g.SerialNumber(),
		Pictures:          pictures,
		CoverPictureIndex: g.CoverPictureIndex(),
		Description:       g.Description(),
		Brand:             g.Brand(),
		TypeName:          g.TypeName(),
		BuildYear:         g.BuildYear(),
		PriceAmount:       g.Price().Amount(),
		PriceCurrency:     string(g.Price().Currency()),
	}
}

// presignUploadRequest is the JSON payload for POST /upload/presign.
type presignUploadRequest struct {
	ContentType string `json:"contentType"`
}

// presignUploadResponse is returned so the client can PUT directly to S3.
type presignUploadResponse struct {
	UploadURL string `json:"uploadUrl"`
	PublicURL string `json:"publicUrl"`
	Key       string `json:"key"`
}

// errorResponse is the JSON envelope used for non-2xx responses.
type errorResponse struct {
	Error string `json:"error"`
}

// marketLogRequest is the JSON payload for POST /guitar/{id}/market-log.
type marketLogRequest struct {
	ObservedAt        string `json:"observedAt,omitempty"`
	Source            string `json:"source"`
	Action            string `json:"action"`
	PriceAmount       int64  `json:"priceAmount"`
	PriceCurrency     string `json:"priceCurrency"`
	ListingURL        string `json:"listingUrl,omitempty"`
	ListingTitle      string `json:"listingTitle,omitempty"`
	ExternalListingID string `json:"externalListingId,omitempty"`
}

// marketLogResponse is the JSON projection of a MarketLog aggregate.
type marketLogResponse struct {
	ID                string `json:"id"`
	GuitarID          string `json:"guitarId"`
	ObservedAt        string `json:"observedAt"`
	Source            string `json:"source"`
	Action            string `json:"action"`
	PriceAmount       int64  `json:"priceAmount"`
	PriceCurrency     string `json:"priceCurrency"`
	ListingURL        string `json:"listingUrl,omitempty"`
	ListingTitle      string `json:"listingTitle,omitempty"`
	ExternalListingID string `json:"externalListingId,omitempty"`
}

func toMarketLogResponse(log *domain.MarketLog) marketLogResponse {
	return marketLogResponse{
		ID:                log.ID(),
		GuitarID:          log.GuitarID(),
		ObservedAt:        log.ObservedAt().UTC().Format(time.RFC3339),
		Source:            string(log.Source()),
		Action:            string(log.Action()),
		PriceAmount:       log.Price().Amount(),
		PriceCurrency:     string(log.Price().Currency()),
		ListingURL:        log.ListingURL(),
		ListingTitle:      log.ListingTitle(),
		ExternalListingID: log.ExternalListingID(),
	}
}
