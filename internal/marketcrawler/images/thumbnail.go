package images

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

const DefaultThumbnailSize = 256

// ThumbnailJPEG downloads an image, center-crops it to a square, and returns JPEG bytes.
func ThumbnailJPEG(httpClient *http.Client, sourceURL string, size int) ([]byte, error) {
	if size <= 0 {
		size = DefaultThumbnailSize
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return nil, errEmptySourceURL
	}
	client := httpClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "image/*")
	req.Header.Set("User-Agent", "guitars-market-crawler/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errDownloadFailed
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	thumb := imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)
	var out bytes.Buffer
	if err := imaging.Encode(&out, thumb, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
