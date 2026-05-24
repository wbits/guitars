package auth

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// GroupsFromClaims extracts Cognito group membership from JWT claims.
func GroupsFromClaims(claims jwt.MapClaims) []string {
	raw, ok := claims["cognito:groups"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return normalizeGroups(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return normalizeGroups(out)
	default:
		return nil
	}
}

func normalizeGroups(groups []string) []string {
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		if trimmed := strings.TrimSpace(group); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// IsAdmin reports whether principal belongs to the configured admin group.
func IsAdmin(p Principal, adminGroup string) bool {
	adminGroup = strings.TrimSpace(adminGroup)
	if adminGroup == "" {
		return false
	}
	for _, group := range p.Groups {
		if group == adminGroup {
			return true
		}
	}
	return false
}
