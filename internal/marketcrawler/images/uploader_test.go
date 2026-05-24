package images

import (
	"bytes"
	"context"
	"image/color"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/disintegration/imaging"
)

func newTestImageServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(body)
	}))
}

func TestUploader_Upload(t *testing.T) {
	t.Parallel()

	src := newTestImageServer(t, mustJPEG(t, 300, 150))
	t.Cleanup(src.Close)

	var uploaded []byte
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.NotFound(w, r)
			return
		}
		buf, _ := io.ReadAll(r.Body)
		uploaded = buf
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(uploadServer.Close)

	uploader := &Uploader{
		Presign:    &stubPresign{uploadURL: uploadServer.URL, publicURL: "https://cdn.example/images/market-logs/test.jpg"},
		HTTPClient: src.Client(),
	}
	publicURL, err := uploader.Upload(context.Background(), src.URL)
	if err != nil {
		t.Fatal(err)
	}
	if publicURL != "https://cdn.example/images/market-logs/test.jpg" {
		t.Fatalf("unexpected public url: %q", publicURL)
	}
	if len(uploaded) == 0 {
		t.Fatal("expected uploaded bytes")
	}
}

type stubPresign struct {
	uploadURL string
	publicURL string
}

func (s *stubPresign) PresignMarketLogImage(_ context.Context) (string, string, error) {
	return s.uploadURL, s.publicURL, nil
}

func mustJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := imaging.New(width, height, color.White)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
