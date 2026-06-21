package guitaranalysis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wbits/guitars/internal/guitaranalysis"
	"github.com/wbits/guitars/internal/guitaranalysis/persistence"
	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

type stubOwners struct {
	enabled bool
	creds   guitaranalysis.VisionCredentials
	ok      bool
}

func (s stubOwners) PhotoAnalysisEnabled(_ context.Context, _ string) (bool, error) {
	return s.enabled, nil
}

func (s stubOwners) VisionCredentials(_ context.Context, _ string) (guitaranalysis.VisionCredentials, bool, error) {
	return s.creds, s.ok, nil
}

func TestAnalyzeIfEligible_SkipsWhenDisabled(t *testing.T) {
	repo := persistence.NewMemoryRepository()
	svc := guitaranalysis.NewService(repo, stubOwners{enabled: false}, nil)
	guitar := testGuitar(t)
	if _, err := svc.AnalyzeIfEligible(context.Background(), guitar); err != nil {
		t.Fatal(err)
	}
	rec, err := repo.FindByGuitarID(context.Background(), guitar.ID())
	if err != nil {
		t.Fatal(err)
	}
	if rec != nil {
		t.Fatal("expected no analysis record when disabled")
	}
}

func TestAnalyzeIfEligible_StoresReadyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"visualSummary":"Cherry sunburst finish with maple neck.","tags":["sunburst","maple-neck"],"confidence":0.9}`}},
			},
		})
	}))
	defer server.Close()

	repo := persistence.NewMemoryRepository()
	vision := &guitaranalysis.VisionAnalyzer{Client: server.Client()}
	svc := guitaranalysis.NewService(
		repo,
		stubOwners{
			enabled: true,
			ok:      true,
			creds: guitaranalysis.VisionCredentials{
				APIKey:  "sk-test",
				BaseURL: server.URL,
				Model:   "gpt-4o-mini",
			},
		},
		vision,
	)
	guitar := testGuitar(t)
	rec, err := svc.AnalyzeIfEligible(context.Background(), guitar)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status() != guitaranalysis.StatusReady {
		t.Fatalf("status: %s", rec.Status())
	}
	if rec.VisualSummary() == "" || len(rec.Tags()) == 0 {
		t.Fatalf("unexpected record: %+v", rec)
	}
}

func testGuitar(t *testing.T) *domain.Guitar {
	t.Helper()
	price, err := domain.NewMoney(100000, domain.EUR)
	if err != nil {
		t.Fatal(err)
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID: "g1", Owner: "owner-1", Brand: "Fender", TypeName: "Strat",
		BuildYear: 1996, Price: price, Pictures: []string{"https://example.com/a.jpg"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}
