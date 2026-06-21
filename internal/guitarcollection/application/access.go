package application

import (
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func guitarReadableBy(g *domain.Guitar, userID string) bool {
	if !g.HiddenInCollection() {
		return true
	}
	return guitarWritableBy(g, userID)
}

func guitarWritableBy(g *domain.Guitar, userID string) bool {
	owner := g.Owner()
	if owner == "" {
		return true
	}
	return owner == strings.TrimSpace(userID)
}

// MarketLogWritableBy reports whether caller may append market observations to a guitar.
func MarketLogWritableBy(
	g *domain.Guitar,
	callerID, callerEmail string,
	crawlerEmails, crawlerUserIDs map[string]struct{},
	ownerMarketCrawlEnabled bool,
) bool {
	if guitarWritableBy(g, callerID) {
		return true
	}
	if !isMarketCrawler(callerID, callerEmail, crawlerEmails, crawlerUserIDs) {
		return false
	}
	if !ownerMarketCrawlEnabled {
		return false
	}
	return guitarReadableBy(g, callerID)
}

func isMarketCrawler(callerID, callerEmail string, emails, userIDs map[string]struct{}) bool {
	if len(emails) == 0 && len(userIDs) == 0 {
		return false
	}
	if id := strings.TrimSpace(callerID); id != "" {
		if _, ok := userIDs[id]; ok {
			return true
		}
	}
	email := strings.ToLower(strings.TrimSpace(callerEmail))
	if email != "" {
		if _, ok := emails[email]; ok {
			return true
		}
	}
	return false
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

// ParseCrawlerUserIDs splits a comma-separated allowlist of crawler Cognito subs.
func ParseCrawlerUserIDs(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for part := range strings.SplitSeq(raw, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		out[id] = struct{}{}
	}
	return out
}

func resolveOwnerForUpdate(g *domain.Guitar, userID string) string {
	if owner := g.Owner(); owner != "" {
		return owner
	}
	return strings.TrimSpace(userID)
}
