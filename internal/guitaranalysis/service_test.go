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

func TestReanalyze_ForcesRerunWhenReady(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"visualSummary":"Updated summary.","tags":["black"],"confidence":0.8}`}},
			},
		})
	}))
	defer server.Close()

	repo := persistence.NewMemoryRepository()
	vision := &guitaranalysis.VisionAnalyzer{Client: server.Client()}
	owners := stubOwners{
		enabled: false,
		ok:      true,
		creds: guitaranalysis.VisionCredentials{
			APIKey:  "sk-test",
			BaseURL: server.URL,
		},
	}
	svc := guitaranalysis.NewService(repo, owners, vision)
	guitar := testGuitar(t)

	if _, err := svc.AnalyzeIfEligible(context.Background(), guitar); err == nil {
		// opt-in disabled; seed a ready record manually
		rec, err := guitaranalysis.NewRecord(guitar.ID(), guitar.Owner(), guitaranalysis.StatusReady, guitaranalysis.PicturesFingerprint(guitar.Pictures()))
		if err != nil {
			t.Fatal(err)
		}
		rec.SetReady(guitaranalysis.PicturesFingerprint(guitar.Pictures()), "Old summary", []string{"old"}, 0.9)
		if err := repo.Save(context.Background(), rec); err != nil {
			t.Fatal(err)
		}
	}

	updated, err := svc.Reanalyze(context.Background(), guitar)
	if err != nil {
		t.Fatal(err)
	}
	if updated.VisualSummary() != "Updated summary." {
		t.Fatalf("summary: %q", updated.VisualSummary())
	}
	if callCount != 1 {
		t.Fatalf("vision calls: %d", callCount)
	}
}

func TestReanalyzeCollection_CountsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"visualSummary":"Done.","tags":["tag"],"confidence":0.8}`}},
			},
		})
	}))
	defer server.Close()

	repo := persistence.NewMemoryRepository()
	vision := &guitaranalysis.VisionAnalyzer{Client: server.Client()}
	svc := guitaranalysis.NewService(
		repo,
		stubOwners{
			ok: true,
			creds: guitaranalysis.VisionCredentials{
				APIKey:  "sk-test",
				BaseURL: server.URL,
			},
		},
		vision,
	)
	withPictures := testGuitar(t)
	withoutPictures := testGuitar(t)
	withoutPictures, _ = domain.NewGuitar(domain.GuitarProps{
		ID: "g2", Owner: "owner-1", Brand: "Gibson", TypeName: "Les Paul",
		BuildYear: 2010, Price: withPictures.Price(),
	})

	result, err := svc.ReanalyzeCollection(context.Background(), "owner-1", []*domain.Guitar{withPictures, withoutPictures})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 || result.Analyzed != 1 || result.Skipped != 1 || result.Failed != 0 {
		t.Fatalf("result: %+v", result)
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
