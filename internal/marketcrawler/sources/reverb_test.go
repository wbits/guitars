package sources

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/wbits/guitars/internal/marketcrawler"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

const reverbSampleListing = `{
  "listings": [{
    "id": 96078243,
    "title": "2012 Gibson Les Paul",
    "state": {"slug": "live", "description": "Live"},
    "_links": {"web": {"href": "https://reverb.com/item/96078243-example"}},
    "price": {"amount": "12115.67", "amount_cents": 1211567, "currency": "USD"}
  }]
}`

func TestReverb_Search_ParsesLiveListing(t *testing.T) {
	calls := 0
	r := &Reverb{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			return jsonResponse(reverbSampleListing), nil
		})},
		PerPage: 1,
	}
	findings, err := r.Search(context.Background(), marketcrawler.GuitarSummary{
		Brand: "Gibson", TypeName: "Les Paul", BuildYear: 2012,
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("want 2 API calls (live + sold), got %d", calls)
	}
	if len(findings) != 2 {
		t.Fatalf("want 2 findings, got %d", len(findings))
	}
	if findings[0].Action != "for_sale" {
		t.Fatalf("want for_sale, got %s", findings[0].Action)
	}
	if findings[0].PriceAmount != 1211567 {
		t.Fatalf("want 1211567 cents, got %d", findings[0].PriceAmount)
	}
}
