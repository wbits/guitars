package marketcrawler

import (
	"strconv"
	"strings"
	"time"
)

// DedupeFindingsPerRun collapses duplicate observations from a single crawl run.
// Duplicates share source, price, observed calendar date, and either external listing
// id or listing title. When two rows differ only by action, sold is kept over for_sale.
func DedupeFindingsPerRun(findings []Finding) []Finding {
	if len(findings) <= 1 {
		return findings
	}
	byKey := make(map[string]Finding, len(findings))
	order := make([]string, 0, len(findings))
	for _, f := range findings {
		key := findingDedupeKey(f)
		existing, ok := byKey[key]
		if !ok {
			byKey[key] = f
			order = append(order, key)
			continue
		}
		byKey[key] = preferFinding(existing, f)
	}
	out := make([]Finding, 0, len(order))
	for _, key := range order {
		out = append(out, byKey[key])
	}
	return out
}

func findingDedupeKey(f Finding) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(f.Source))
	b.WriteByte(0)
	if id := strings.TrimSpace(f.ExternalListingID); id != "" {
		b.WriteString(id)
	} else {
		b.WriteString(strings.TrimSpace(f.ListingTitle))
	}
	b.WriteByte(0)
	b.WriteString(strconv.FormatInt(f.PriceAmount, 10))
	b.WriteByte(0)
	b.WriteString(strings.ToUpper(strings.TrimSpace(f.PriceCurrency)))
	b.WriteByte(0)
	if f.ObservedAt.IsZero() {
		b.WriteString("unknown")
	} else {
		b.WriteString(f.ObservedAt.UTC().Format(time.DateOnly))
	}
	return b.String()
}

func preferFinding(existing, candidate Finding) Finding {
	switch {
	case candidate.Action == "sold" && existing.Action != "sold":
		return mergeFinding(candidate, existing)
	case existing.Action == "sold" && candidate.Action != "sold":
		return mergeFinding(existing, candidate)
	default:
		return mergeFinding(existing, candidate)
	}
}

func mergeFinding(primary, other Finding) Finding {
	out := primary
	if strings.TrimSpace(out.ListingURL) == "" {
		out.ListingURL = other.ListingURL
	}
	if strings.TrimSpace(out.ListingTitle) == "" {
		out.ListingTitle = other.ListingTitle
	}
	if strings.TrimSpace(out.ExternalListingID) == "" {
		out.ExternalListingID = other.ExternalListingID
	}
	if strings.TrimSpace(out.SourceImageURL) == "" {
		out.SourceImageURL = other.SourceImageURL
	}
	if strings.TrimSpace(out.ListingImageURL) == "" {
		out.ListingImageURL = other.ListingImageURL
	}
	return out
}
