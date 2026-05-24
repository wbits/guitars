package marketcrawler

import "testing"

func TestSearchQueries_BroadensSpecificGuitar(t *testing.T) {
	queries := SearchQueries(GuitarSummary{
		Brand:     "Gibson",
		TypeName:  "60th Anniversary '52 Les Paul Gold Top",
		BuildYear: 2012,
	})
	if len(queries) < 3 {
		t.Fatalf("want at least 3 queries, got %v", queries)
	}
	if queries[0] != "Gibson 60th Anniversary '52 Les Paul Gold Top 2012" {
		t.Fatalf("unexpected first query: %q", queries[0])
	}
	foundFamily := false
	for _, q := range queries {
		if q == "Gibson Les Paul" {
			foundFamily = true
			break
		}
	}
	if !foundFamily {
		t.Fatalf("expected Gibson Les Paul fallback in %v", queries)
	}
}

func TestSearchQueries_DeduplicatesRepeatedPhrases(t *testing.T) {
	queries := SearchQueries(GuitarSummary{
		Brand:    "Gibson",
		TypeName: "Les Paul Standard",
	})
	seen := make(map[string]struct{})
	for _, q := range queries {
		if _, ok := seen[q]; ok {
			t.Fatalf("duplicate query %q in %v", q, queries)
		}
		seen[q] = struct{}{}
	}
}
