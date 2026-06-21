package assistant

import (
	"context"
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func eur(amount int64) domain.Money {
	m, err := domain.NewMoney(amount, domain.EUR)
	if err != nil {
		panic(err)
	}
	return m
}

func sampleGuitars() []*domain.Guitar {
	g1, _ := domain.NewGuitar(domain.GuitarProps{
		ID: "1", Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996,
		Price: eur(150000), Color: "Sunburst",
	})
	g2, _ := domain.NewGuitar(domain.GuitarProps{
		ID: "2", Brand: "Gibson", TypeName: "Les Paul", BuildYear: 2017,
		Price: eur(80000), Color: "Cherry Red",
	})
	g3, _ := domain.NewGuitar(domain.GuitarProps{
		ID: "3", Brand: "Fender", TypeName: "Telecaster", BuildYear: 2020,
		Price: eur(95000), Color: "Black",
	})
	return []*domain.Guitar{g1, g2, g3}
}

func TestApplyFilter_BrandAndMaxPrice(t *testing.T) {
	matched := ApplyFilter(sampleGuitars(), Filter{Brand: "Fender", MaxPriceMajor: fp(2000)})
	if len(matched) != 2 {
		t.Fatalf("want 2 Fenders under 2000, got %d", len(matched))
	}
}

func TestApplyFilter_Color(t *testing.T) {
	matched := ApplyFilter(sampleGuitars(), Filter{Color: "red"})
	if len(matched) != 1 || matched[0].ID() != "2" {
		t.Fatalf("want cherry red Gibson, got %+v", guitarIDs(matched))
	}
}

func TestParseRules_UnderPrice(t *testing.T) {
	f, _ := ParseRules("show guitars under 1000 euro", sampleGuitars())
	if f.MaxPriceMajor == nil || *f.MaxPriceMajor != 1000 {
		t.Fatalf("max price: %+v", f.MaxPriceMajor)
	}
}

func TestService_Chat_RateLimit(t *testing.T) {
	limiter := NewMemoryRateLimiter(1)
	svc := NewService(stubLister{guitars: sampleGuitars()}, RuleLLM{}, limiter)
	ctx := t.Context()
	req := ChatRequest{CollectionUserID: "owner-1", Message: "Fender", CallerUserID: "viewer-1"}
	if _, err := svc.Chat(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Chat(ctx, req); !IsRateLimited(err) {
		t.Fatalf("want rate limit, got %v", err)
	}
}

func fp(v float64) *float64 { return &v }

type stubLister struct{ guitars []*domain.Guitar }

func (s stubLister) ListUserGuitars(_ context.Context, _ string) ([]*domain.Guitar, error) {
	return s.guitars, nil
}
