package images

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/wbits/guitars/internal/marketcrawler"
)

// PresignClient requests upload URLs from the GuitarCollection API.
type PresignClient interface {
	PresignMarketLogImage(ctx context.Context) (uploadURL, publicURL string, err error)
}

// Uploader mirrors listing images to the collection CDN as square thumbnails.
type Uploader struct {
	Presign    PresignClient
	HTTPClient *http.Client
}

// Upload downloads sourceURL, crops a thumbnail, and stores it on the CDN.
func (u *Uploader) Upload(ctx context.Context, sourceURL string) (string, error) {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return "", errEmptySourceURL
	}
	client := u.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	jpeg, err := ThumbnailJPEG(client, sourceURL, DefaultThumbnailSize)
	if err != nil {
		return "", err
	}
	uploadURL, publicURL, err := u.Presign.PresignMarketLogImage(ctx)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(jpeg))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "image/jpeg")
	req.ContentLength = int64(len(jpeg))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("%w: status %d: %s", errUploadFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return publicURL, nil
}

// NewUploaderFromAPI constructs an Uploader backed by the GuitarCollection API presign endpoint.
func NewUploaderFromAPI(api *marketcrawler.APIClient) *Uploader {
	return &Uploader{
		Presign:    &apiPresignAdapter{api: api},
		HTTPClient: api.HTTPClient,
	}
}

type apiPresignAdapter struct {
	api *marketcrawler.APIClient
}

func (a *apiPresignAdapter) PresignMarketLogImage(ctx context.Context) (string, string, error) {
	return a.api.PresignUpload(ctx, "image/jpeg")
}
