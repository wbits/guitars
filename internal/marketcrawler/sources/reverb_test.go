package sources

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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
	findings = marketcrawler.DedupeFindingsPerRun(findings)
	if calls != 2 {
		t.Fatalf("want 2 API calls (live + sold), got %d", calls)
	}
	if len(findings) != 1 {
		t.Fatalf("want 1 deduped finding, got %d", len(findings))
	}
	if findings[0].Action != "sold" {
		t.Fatalf("want sold when live and sold searches overlap, got %s", findings[0].Action)
	}
	if findings[0].PriceAmount != 1211567 {
		t.Fatalf("want 1211567 cents, got %d", findings[0].PriceAmount)
	}
}

func TestReverb_fetch_SendsTokenAndUserAgent(t *testing.T) {
	var gotUA, gotAuth string
	r := &Reverb{
		Token: "test-token",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotUA = req.Header.Get("User-Agent")
			gotAuth = req.Header.Get("Authorization")
			return jsonResponse(`{"listings":[]}`), nil
		})},
	}
	if _, err := r.fetch(context.Background(), "Gibson Les Paul", false); err != nil {
		t.Fatal(err)
	}
	if gotUA != reverbUserAgent {
		t.Fatalf("want User-Agent %q, got %q", reverbUserAgent, gotUA)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("want Authorization Bearer test-token, got %q", gotAuth)
	}
}

func TestReverbAPIError_HTMLWithoutToken(t *testing.T) {
	err := reverbAPIError(403, []byte("<!DOCTYPE html><html></html>"), false)
	if err == nil || !strings.Contains(err.Error(), "REVERB_API_TOKEN") {
		t.Fatalf("want REVERB_API_TOKEN hint, got %v", err)
	}
}

func TestReverbAPIError_HTMLWithToken(t *testing.T) {
	err := reverbAPIError(403, []byte("<!DOCTYPE html><html></html>"), true)
	if err == nil || !strings.Contains(err.Error(), "Cloudflare") {
		t.Fatalf("want Cloudflare hint, got %v", err)
	}
}
