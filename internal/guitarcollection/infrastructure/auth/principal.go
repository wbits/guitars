package auth

// Principal identifies the authenticated caller. UserID is the Cognito sub
// claim for JWT tokens, or a fixed local-dev id for bearer-token auth.
// Email is populated from the token when available. Groups comes from the
// Cognito cognito:groups claim when present.
type Principal struct {
	UserID string
	Email  string
	Groups []string
}
