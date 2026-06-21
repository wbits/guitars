package assistant

import (
	"context"
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

type stubGuitars struct {
	guitars []*domain.Guitar
}

func (s stubGuitars) ListUserGuitars(_ context.Context, _ string) ([]*domain.Guitar, error) {
	return s.guitars, nil
}

type stubBYOK struct {
	creds BYOKCredentials
	ok    bool
}

func (s stubBYOK) Credentials(_ context.Context, _ string) (BYOKCredentials, bool, error) {
	return s.creds, s.ok, nil
}

type countingLLM struct {
	called bool
}

func (c *countingLLM) ParseFilter(_ context.Context, message string, guitars []*domain.Guitar) (Filter, string, error) {
	c.called = true
	f, reply := ParseRules(message, guitars)
	return f, reply, nil
}

func TestService_ResolveTier_UsesBYOKOnOwnCollection(t *testing.T) {
	hosted := &countingLLM{}
	price, err := domain.NewMoney(100000, domain.EUR)
	if err != nil {
		t.Fatal(err)
	}
	guitar, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g1", Brand: "Fender", TypeName: "Strat", BuildYear: 2020, Price: price,
	})
	svc := NewService(
		stubGuitars{guitars: []*domain.Guitar{guitar}},
		hosted,
		NewMemoryRateLimiter(1),
		&stubBYOK{creds: BYOKCredentials{APIKey: "sk-owner"}, ok: true},
		NewMemoryRateLimiter(5),
	)
	_, err = svc.Chat(context.Background(), ChatRequest{
		CollectionUserID: "owner-1",
		CallerUserID:     "owner-1",
		Message:          "Fender",
	})
	if err != nil {
		t.Fatal(err)
	}
	if hosted.called {
		t.Fatal("expected BYOK LLM, hosted was called")
	}
}

func TestService_ResolveTier_UsesHostedForOtherCollections(t *testing.T) {
	hosted := &countingLLM{}
	svc := NewService(
		stubGuitars{},
		hosted,
		NewMemoryRateLimiter(5),
		&stubBYOK{creds: BYOKCredentials{APIKey: "sk-owner"}, ok: true},
		NewMemoryRateLimiter(100),
	)
	_, err := svc.Chat(context.Background(), ChatRequest{
		CollectionUserID: "owner-1",
		CallerUserID:     "visitor-1",
		Message:          "Fender",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hosted.called {
		t.Fatal("expected hosted LLM for visitor")
	}
}
