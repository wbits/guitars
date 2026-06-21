package assistant

import (
	"context"

	profileapp "github.com/wbits/guitars/internal/userprofile/application"
)

// ProfileBYOKProvider loads decrypted owner credentials from user profiles.
type ProfileBYOKProvider struct {
	Profiles *profileapp.Service
}

// Credentials implements BYOKProvider.
func (p *ProfileBYOKProvider) Credentials(ctx context.Context, ownerUserID string) (BYOKCredentials, bool, error) {
	if p == nil || p.Profiles == nil {
		return BYOKCredentials{}, false, nil
	}
	creds, ok, err := p.Profiles.AssistantBYOKCredentialsForUser(ctx, ownerUserID)
	if err != nil || !ok {
		return BYOKCredentials{}, false, err
	}
	return BYOKCredentials{
		APIKey:  creds.APIKey,
		BaseURL: creds.BaseURL,
		Model:   creds.Model,
	}, true, nil
}
