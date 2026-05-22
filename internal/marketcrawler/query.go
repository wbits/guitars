package marketcrawler

import (
	"fmt"
	"strings"
)

// SearchQuery builds a marketplace search phrase from guitar attributes.
func SearchQuery(g GuitarSummary) string {
	parts := make([]string, 0, 3)
	if brand := strings.TrimSpace(g.Brand); brand != "" {
		parts = append(parts, brand)
	}
	if model := strings.TrimSpace(g.TypeName); model != "" {
		parts = append(parts, model)
	}
	if g.BuildYear > 0 {
		parts = append(parts, fmt.Sprintf("%d", g.BuildYear))
	}
	return strings.Join(parts, " ")
}
