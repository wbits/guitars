package guitaranalysis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wbits/guitars/internal/guitaranalysis"
	"github.com/wbits/guitars/internal/guitaranalysis/persistence"
)

func TestAnalyzePictureForCatalog_ReturnsSuggestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{
					"visualSummary":"Sunburst Strat with maple neck.",
					"tags":["sunburst","maple-neck"],
					"confidence":0.88,
					"brand":"Fender",
					"typeName":"Stratocaster",
					"color":"Sunburst",
					"buildYear":1996,
					"description":"Sunburst Strat with maple neck."
				}`}},
			},
		})
	}))
	defer server.Close()

	vision := &guitaranalysis.VisionAnalyzer{Client: server.Client()}
	svc := guitaranalysis.NewService(
		persistence.NewMemoryRepository(),
		stubOwners{
			enabled: true,
			ok:      true,
			creds: guitaranalysis.VisionCredentials{
				APIKey:  "sk-test",
				BaseURL: server.URL,
			},
		},
		vision,
		nil,
		nil,
	)

	result, err := svc.AnalyzePictureForCatalog(context.Background(), "owner-1", "https://example.com/guitar.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if result.Suggestions.Brand != "Fender" || result.Suggestions.TypeName != "Stratocaster" {
		t.Fatalf("unexpected suggestions: %+v", result.Suggestions)
	}
	if result.VisualSummary == "" || len(result.Tags) == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestAnalyzePictureForCatalog_RequiresOptIn(t *testing.T) {
	svc := guitaranalysis.NewService(
		persistence.NewMemoryRepository(),
		stubOwners{enabled: false},
		nil,
		nil,
		nil,
	)
	_, err := svc.AnalyzePictureForCatalog(context.Background(), "owner-1", "https://example.com/guitar.jpg")
	if err == nil {
		t.Fatal("expected error when photo analysis disabled")
	}
}
