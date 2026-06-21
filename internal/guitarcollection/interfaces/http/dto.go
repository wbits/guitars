package httpapi

import (
	"time"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
	profiledomain "github.com/wbits/guitars/internal/userprofile/domain"
)

// guitarRequest is the JSON payload accepted by POST /guitar and PUT /guitar/{id}.
type guitarRequest struct {
	SerialNumber      string   `json:"serialNumber,omitempty"`
	Color             string   `json:"color,omitempty"`
	Country           string   `json:"country,omitempty"`
	Factory           string   `json:"factory,omitempty"`
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
	Owner             string   `json:"owner,omitempty"`
	SerialNumber      string   `json:"serialNumber,omitempty"`
	Color             string   `json:"color,omitempty"`
	Country           string   `json:"country,omitempty"`
	Factory           string   `json:"factory,omitempty"`
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
		Owner:             g.Owner(),
		SerialNumber:      g.SerialNumber(),
		Color:             g.Color(),
		Country:           g.Country(),
		Factory:           g.Factory(),
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
	Purpose     string `json:"purpose,omitempty"`
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

// meResponse is returned by GET /me.
type meResponse struct {
	UserID                   string `json:"userId"`
	Username                 string `json:"username,omitempty"`
	Email                    string `json:"email,omitempty"`
	DisplayName              string `json:"displayName"`
	IsAdmin                  bool   `json:"isAdmin"`
	AssistantByokConfigured  bool   `json:"assistantByokConfigured"`
	AssistantLlmBaseURL      string `json:"assistantLlmBaseUrl,omitempty"`
	AssistantLlmModel        string `json:"assistantLlmModel,omitempty"`
}

// profilePatchRequest is the JSON payload for PATCH /me.
type profilePatchRequest struct {
	Username string `json:"username"`
}

type assistantBYOKPutRequest struct {
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl,omitempty"`
	Model   string `json:"model,omitempty"`
}

// collectionOwnerResponse describes a user that owns at least one guitar.
type collectionOwnerResponse struct {
	UserID             string `json:"userId"`
	Username           string `json:"username,omitempty"`
	Email              string `json:"email,omitempty"`
	DisplayName        string `json:"displayName"`
	GuitarCount        int    `json:"guitarCount"`
	MarketCrawlEnabled bool   `json:"marketCrawlEnabled"`
}

// collectionMarketCrawlPatchRequest is the JSON payload for PATCH /collections/{userId}/market-crawl.
type collectionMarketCrawlPatchRequest struct {
	MarketCrawlEnabled bool `json:"marketCrawlEnabled"`
}

// clearCollectionMarketLogsResponse is returned by DELETE /collections/{userId}/market-log.
type clearCollectionMarketLogsResponse struct {
	DeletedCount int `json:"deletedCount"`
}

func toMeResponse(profile *profiledomain.Profile, isAdmin bool) meResponse {
	return meResponse{
		UserID:                  profile.UserID(),
		Username:                profile.Username(),
		Email:                   profile.Email(),
		DisplayName:             profile.DisplayName(),
		IsAdmin:                 isAdmin,
		AssistantByokConfigured: profile.AssistantBYOKConfigured(),
		AssistantLlmBaseURL:     profile.AssistantLLMBaseURL(),
		AssistantLlmModel:       profile.AssistantLLMModel(),
	}
}

func toCollectionOwnerResponse(userID string, profile *profiledomain.Profile, guitarCount int) collectionOwnerResponse {
	resp := collectionOwnerResponse{
		UserID:      userID,
		GuitarCount: guitarCount,
	}
	if profile != nil {
		resp.Username = profile.Username()
		resp.Email = profile.Email()
		resp.DisplayName = profile.DisplayName()
		resp.MarketCrawlEnabled = profile.MarketCrawlEnabled()
	}
	return resp
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
	ListingImageURL   string `json:"listingImageUrl,omitempty"`
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
	ListingImageURL   string `json:"listingImageUrl,omitempty"`
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
		ListingImageURL:   log.ListingImageURL(),
	}
}
