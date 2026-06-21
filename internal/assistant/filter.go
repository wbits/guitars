package assistant

import (
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// Filter describes client-side gallery filters. Prices use major units (e.g. euros).
type Filter struct {
	Brand         string   `json:"brand,omitempty"`
	TypeName      string   `json:"typeName,omitempty"`
	Color         string   `json:"color,omitempty"`
	MinPriceMajor *float64 `json:"minPriceMajor,omitempty"`
	MaxPriceMajor *float64 `json:"maxPriceMajor,omitempty"`
	MinYear       *int     `json:"minYear,omitempty"`
	MaxYear       *int     `json:"maxYear,omitempty"`
}

func (f Filter) isEmpty() bool {
	return f.Brand == "" && f.TypeName == "" && f.Color == "" &&
		f.MinPriceMajor == nil && f.MaxPriceMajor == nil &&
		f.MinYear == nil && f.MaxYear == nil
}

// ApplyFilter returns guitars matching all non-empty filter fields.
func ApplyFilter(guitars []*domain.Guitar, f Filter) []*domain.Guitar {
	if f.isEmpty() {
		out := make([]*domain.Guitar, len(guitars))
		copy(out, guitars)
		return out
	}
	out := make([]*domain.Guitar, 0, len(guitars))
	for _, g := range guitars {
		if matchesFilter(g, f) {
			out = append(out, g)
		}
	}
	return out
}

func matchesFilter(g *domain.Guitar, f Filter) bool {
	if f.Brand != "" && !containsFold(g.Brand(), f.Brand) {
		return false
	}
	if f.TypeName != "" && !containsFold(g.TypeName(), f.TypeName) {
		return false
	}
	if f.Color != "" && !containsFold(g.Color(), f.Color) {
		return false
	}
	if f.MinYear != nil && g.BuildYear() < *f.MinYear {
		return false
	}
	if f.MaxYear != nil && g.BuildYear() > *f.MaxYear {
		return false
	}
	priceMajor := float64(g.Price().Amount()) / 100.0
	if f.MinPriceMajor != nil && priceMajor < *f.MinPriceMajor {
		return false
	}
	if f.MaxPriceMajor != nil && priceMajor > *f.MaxPriceMajor {
		return false
	}
	return true
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(strings.TrimSpace(needle)))
}

func guitarIDs(guitars []*domain.Guitar) []string {
	ids := make([]string, len(guitars))
	for i, g := range guitars {
		ids[i] = g.ID()
	}
	return ids
}
