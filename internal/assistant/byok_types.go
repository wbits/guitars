package assistant

import "context"

// BYOKCredentials are owner-provided LLM settings (tier 2).
type BYOKCredentials struct {
	APIKey  string
	BaseURL string
	Model   string
}

// BYOKProvider loads owner credentials when configured.
type BYOKProvider interface {
	Credentials(ctx context.Context, ownerUserID string) (BYOKCredentials, bool, error)
}
