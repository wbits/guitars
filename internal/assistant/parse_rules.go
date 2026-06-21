package assistant

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

var (
	underPriceRe = regexp.MustCompile(`(?i)(?:under|below|less than|max|maximum|<=?)\s*(?:€|eur(?:os?)?|\$|usd)?\s*([0-9]+(?:[.,][0-9]{1,2})?)`)
	overPriceRe  = regexp.MustCompile(`(?i)(?:over|above|more than|min|minimum|>=?)\s*(?:€|eur(?:os?)?|\$|usd)?\s*([0-9]+(?:[.,][0-9]{1,2})?)`)
	betweenRe    = regexp.MustCompile(`(?i)(?:between|from)\s*(?:€|eur(?:os?)?|\$|usd)?\s*([0-9]+(?:[.,][0-9]{1,2})?)\s*(?:and|-)\s*(?:€|eur(?:os?)?|\$|usd)?\s*([0-9]+(?:[.,][0-9]{1,2})?)`)
	yearRe       = regexp.MustCompile(`\b(19|20[0-9]{2})\b`)
)

// ParseRules extracts a filter from natural language without an LLM.
func ParseRules(message string, guitars []*domain.Guitar) (Filter, string) {
	msg := strings.TrimSpace(message)
	lower := strings.ToLower(msg)
	f := Filter{}
	priceNumbers := map[string]struct{}{}

	if m := betweenRe.FindStringSubmatch(msg); len(m) == 3 {
		if minP, ok := parseMajorPrice(m[1]); ok {
			f.MinPriceMajor = &minP
			priceNumbers[m[1]] = struct{}{}
		}
		if maxP, ok := parseMajorPrice(m[2]); ok {
			f.MaxPriceMajor = &maxP
			priceNumbers[m[2]] = struct{}{}
		}
	} else {
		if m := underPriceRe.FindStringSubmatch(msg); len(m) == 2 {
			if maxP, ok := parseMajorPrice(m[1]); ok {
				f.MaxPriceMajor = &maxP
				priceNumbers[m[1]] = struct{}{}
			}
		}
		if m := overPriceRe.FindStringSubmatch(msg); len(m) == 2 {
			if minP, ok := parseMajorPrice(m[1]); ok {
				f.MinPriceMajor = &minP
				priceNumbers[m[1]] = struct{}{}
			}
		}
	}

	for _, g := range guitars {
		brand := strings.TrimSpace(g.Brand())
		if brand != "" && strings.Contains(lower, strings.ToLower(brand)) {
			f.Brand = brand
			break
		}
	}

	colorHints := []string{"red", "black", "white", "sunburst", "blue", "green", "natural", "cherry", "gold", "brown"}
	for _, hint := range colorHints {
		if strings.Contains(lower, hint) {
			f.Color = hint
			break
		}
	}

	for _, m := range yearRe.FindAllStringSubmatch(msg, -1) {
		if len(m) != 2 {
			continue
		}
		if _, usedForPrice := priceNumbers[m[1]]; usedForPrice {
			continue
		}
		if y, err := strconv.Atoi(m[1]); err == nil {
			f.MinYear = &y
			f.MaxYear = &y
			break
		}
	}

	reply := buildRulesReply(f)
	return f, reply
}

func parseMajorPrice(raw string) (float64, bool) {
	raw = strings.ReplaceAll(raw, ",", ".")
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}

func buildRulesReply(f Filter) string {
	parts := make([]string, 0, 4)
	if f.Brand != "" {
		parts = append(parts, f.Brand)
	}
	if f.Color != "" {
		parts = append(parts, f.Color)
	}
	if f.MinPriceMajor != nil && f.MaxPriceMajor != nil {
		parts = append(parts, formatPriceRange(*f.MinPriceMajor, *f.MaxPriceMajor))
	} else if f.MaxPriceMajor != nil {
		parts = append(parts, "under "+formatMajor(*f.MaxPriceMajor))
	} else if f.MinPriceMajor != nil {
		parts = append(parts, "over "+formatMajor(*f.MinPriceMajor))
	}
	if f.MinYear != nil && f.MaxYear != nil && *f.MinYear == *f.MaxYear {
		parts = append(parts, strconv.Itoa(*f.MinYear))
	}
	if len(parts) == 0 {
		return "Showing the full collection. Try asking about a brand, color, or price range."
	}
	return "Filtering for " + strings.Join(parts, ", ") + "."
}

func formatPriceRange(min, max float64) string {
	return formatMajor(min) + "–" + formatMajor(max)
}

func formatMajor(v float64) string {
	if v == float64(int(v)) {
		return "€" + strconv.Itoa(int(v))
	}
	return "€" + strconv.FormatFloat(v, 'f', 2, 64)
}
