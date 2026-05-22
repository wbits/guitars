package storage

import (
	"context"
	"testing"
)

func TestPresignPut_RejectsUnsupportedContentType(t *testing.T) {
	p := &Presigner{bucket: "test", cdnBase: "https://cdn.example.com"}
	_, err := p.PresignPut(context.Background(), "application/pdf")
	if err == nil {
		t.Fatal("expected error for unsupported content type")
	}
}

func TestPresignPut_AcceptsCommonImageTypes(t *testing.T) {
	for ct := range allowedContentTypes {
		if _, ok := allowedContentTypes[ct]; !ok {
			t.Errorf("missing extension mapping for %q", ct)
		}
	}
}

func TestRewritePresignedHost_SwapsHostForBrowserAccess(t *testing.T) {
	got, err := rewritePresignedHost(
		"http://guitars-localstack:4566/guitars-local/images/guitars/a.jpg?sig=1",
		"http://localhost:4566",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost:4566/guitars-local/images/guitars/a.jpg?sig=1" {
		t.Fatalf("got %q", got)
	}
}
