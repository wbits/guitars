package images

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/disintegration/imaging"
)

func TestThumbnailJPEG_CropsToSquare(t *testing.T) {
	t.Parallel()

	src := imaging.New(400, 200, color.White)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, src, imaging.JPEG); err != nil {
		t.Fatal(err)
	}

	server := newTestImageServer(t, buf.Bytes())
	t.Cleanup(server.Close)

	out, err := ThumbnailJPEG(server.Client(), server.URL, 128)
	if err != nil {
		t.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != 128 || b.Dy() != 128 {
		t.Fatalf("want 128x128 thumbnail, got %dx%d", b.Dx(), b.Dy())
	}
}
