package application

import (
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func TestMarketLogWritableBy(t *testing.T) {
	price, err := domain.NewMoney(100000, domain.EUR)
	if err != nil {
		t.Fatal(err)
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID:        "g-1",
		Owner:     "owner-1",
		Brand:     "Gibson",
		TypeName:  "Les Paul",
		BuildYear: 2017,
		Price:     price,
	})
	if err != nil {
		t.Fatal(err)
	}
	crawlerEmails := ParseCrawlerEmails("info@wbits.net")

	if !MarketLogWritableBy(g, "owner-1", "owner@example.com", crawlerEmails) {
		t.Fatal("owner should write market logs")
	}
	if MarketLogWritableBy(g, "other", "other@example.com", crawlerEmails) {
		t.Fatal("non-crawler should not write market logs for owned guitar")
	}
	if !MarketLogWritableBy(g, "crawler-sub", "info@wbits.net", crawlerEmails) {
		t.Fatal("configured crawler should write market logs")
	}
}

func TestParseCrawlerEmails(t *testing.T) {
	got := ParseCrawlerEmails(" Info@wbits.net , ")
	if len(got) != 1 {
		t.Fatalf("want 1 email, got %d", len(got))
	}
	if _, ok := got["info@wbits.net"]; !ok {
		t.Fatalf("unexpected keys: %v", got)
	}
}
