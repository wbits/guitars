package httpapi

import "github.com/wbits/guitars/internal/guitarcollection/domain"

// guitarRequest is the JSON payload accepted by POST /guitar and PUT /guitar/{id}.
type guitarRequest struct {
	SerialNumber  string   `json:"serialNumber,omitempty"`
	Pictures      []string `json:"pictures,omitempty"`
	Description   string   `json:"description,omitempty"`
	Brand         string   `json:"brand"`
	TypeName      string   `json:"typeName"`
	BuildYear     int      `json:"buildYear"`
	PriceAmount   int64    `json:"priceAmount"`
	PriceCurrency string   `json:"priceCurrency"`
}

// guitarResponse is the JSON projection of a Guitar aggregate sent to clients.
type guitarResponse struct {
	ID            string   `json:"id"`
	SerialNumber  string   `json:"serialNumber,omitempty"`
	Pictures      []string `json:"pictures"`
	Description   string   `json:"description,omitempty"`
	Brand         string   `json:"brand"`
	TypeName      string   `json:"typeName"`
	BuildYear     int      `json:"buildYear"`
	PriceAmount   int64    `json:"priceAmount"`
	PriceCurrency string   `json:"priceCurrency"`
}

func toResponse(g *domain.Guitar) guitarResponse {
	pictures := g.Pictures()
	if pictures == nil {
		pictures = []string{}
	}
	return guitarResponse{
		ID:            g.ID(),
		SerialNumber:  g.SerialNumber(),
		Pictures:      pictures,
		Description:   g.Description(),
		Brand:         g.Brand(),
		TypeName:      g.TypeName(),
		BuildYear:     g.BuildYear(),
		PriceAmount:   g.Price().Amount(),
		PriceCurrency: string(g.Price().Currency()),
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
