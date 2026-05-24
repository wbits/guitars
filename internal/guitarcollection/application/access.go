package application

import (
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func guitarReadableBy(_ *domain.Guitar, _ string) bool {
	return true
}

func guitarWritableBy(g *domain.Guitar, userID string) bool {
	owner := g.Owner()
	if owner == "" {
		return true
	}
	return owner == strings.TrimSpace(userID)
}

// MarketLogWritableBy reports whether caller may append market observations to a guitar.
func MarketLogWritableBy(g *domain.Guitar, callerID, callerEmail string, crawlerEmails map[string]struct{}, ownerMarketCrawlEnabled bool) bool {
	if guitarWritableBy(g, callerID) {
		return true
	}
	if len(crawlerEmails) == 0 {
		return false
	}
	email := strings.ToLower(strings.TrimSpace(callerEmail))
	if _, ok := crawlerEmails[email]; !ok {
		return false
	}
	if !ownerMarketCrawlEnabled {
		return false
	}
	return guitarReadableBy(g, callerID)
}

// ParseCrawlerEmails splits a comma-separated allowlist of crawler account emails.
func ParseCrawlerEmails(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for part := range strings.SplitSeq(raw, ",") {
		email := strings.ToLower(strings.TrimSpace(part))
		if email == "" {
			continue
		}
		out[email] = struct{}{}
	}
	return out
}

func resolveOwnerForUpdate(g *domain.Guitar, userID string) string {
	if owner := g.Owner(); owner != "" {
		return owner
	}
	return strings.TrimSpace(userID)
}
