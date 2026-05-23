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

func guitarOwnedBy(g *domain.Guitar, userID string) bool {
	return g.Owner() == strings.TrimSpace(userID)
}

func resolveOwnerForUpdate(g *domain.Guitar, userID string) string {
	if owner := g.Owner(); owner != "" {
		return owner
	}
	return strings.TrimSpace(userID)
}
