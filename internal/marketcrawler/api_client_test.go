package marketcrawler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIClient_ListGuitars_AcrossCollections(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/collections":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"userId": "owner-a"},
				{"userId": "owner-b"},
			})
		case "/collections/owner-a/guitar":
			_ = json.NewEncoder(w).Encode([]GuitarFromAPI{
				{ID: "g-1", Brand: "Gibson", TypeName: "Les Paul", BuildYear: 2017},
			})
		case "/collections/owner-b/guitar":
			_ = json.NewEncoder(w).Encode([]GuitarFromAPI{
				{ID: "g-2", Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := NewAPIClient(server.URL, "token")
	guitars, err := client.ListGuitars(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(guitars) != 2 {
		t.Fatalf("want 2 guitars, got %d", len(guitars))
	}
}
