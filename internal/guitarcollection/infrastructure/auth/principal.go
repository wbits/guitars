package auth

// Principal identifies the authenticated caller. UserID is the Cognito sub
// claim for JWT tokens, or a fixed local-dev id for bearer-token auth.
type Principal struct {
	UserID string
}
