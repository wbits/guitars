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
	Tag           string   `json:"tag,omitempty"`
	SearchText    string   `json:"searchText,omitempty"`
}

func (f Filter) isEmpty() bool {
	return f.Brand == "" && f.TypeName == "" && f.Color == "" &&
		f.MinPriceMajor == nil && f.MaxPriceMajor == nil &&
		f.MinYear == nil && f.MaxYear == nil &&
		f.Tag == "" && f.SearchText == ""
}

// AnalysisSearch provides AI metadata for gallery filtering.
type AnalysisSearch interface {
	SearchBlob() string
	Tags() []string
}

// ApplyFilter returns guitars matching all non-empty filter fields.
func ApplyFilter(guitars []*domain.Guitar, f Filter, analysis map[string]AnalysisSearch) []*domain.Guitar {
	if analysis == nil {
		analysis = map[string]AnalysisSearch{}
	}
	if f.isEmpty() {
		out := make([]*domain.Guitar, len(guitars))
		copy(out, guitars)
		return out
	}
	out := make([]*domain.Guitar, 0, len(guitars))
	for _, g := range guitars {
		if matchesFilter(g, f, analysis[g.ID()]) {
			out = append(out, g)
		}
	}
	return out
}

func matchesFilter(g *domain.Guitar, f Filter, analysis AnalysisSearch) bool {
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
	if f.Tag != "" && !tagMatches(analysis, f.Tag) {
		return false
	}
	if f.SearchText != "" && !searchMatches(g, analysis, f.SearchText) {
		return false
	}
	return true
}

func tagMatches(analysis AnalysisSearch, tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" {
		return true
	}
	if analysis == nil {
		return false
	}
	for _, t := range analysis.Tags() {
		if containsFold(t, tag) {
			return true
		}
	}
	return containsFold(analysis.SearchBlob(), tag)
}

func searchMatches(g *domain.Guitar, analysis AnalysisSearch, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	blob := strings.ToLower(g.Brand() + " " + g.TypeName() + " " + g.Color() + " " + g.Description())
	if analysis != nil {
		blob += " " + analysis.SearchBlob()
	}
	return containsFold(blob, query)
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
