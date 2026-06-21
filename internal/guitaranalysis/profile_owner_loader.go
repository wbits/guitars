package guitaranalysis

import (
	"context"

	profileapp "github.com/wbits/guitars/internal/userprofile/application"
)

// ProfileOwnerLoader adapts user profiles for analysis eligibility and BYOK vision calls.
type ProfileOwnerLoader struct {
	Profiles *profileapp.Service
}

func (p *ProfileOwnerLoader) PhotoAnalysisEnabled(ctx context.Context, ownerID string) (bool, error) {
	if p == nil || p.Profiles == nil {
		return false, nil
	}
	return p.Profiles.PhotoAnalysisEnabled(ctx, ownerID)
}

func (p *ProfileOwnerLoader) VisionCredentials(ctx context.Context, ownerID string) (VisionCredentials, bool, error) {
	if p == nil || p.Profiles == nil {
		return VisionCredentials{}, false, nil
	}
	creds, ok, err := p.Profiles.AssistantBYOKCredentialsForUser(ctx, ownerID)
	if err != nil || !ok {
		return VisionCredentials{}, false, err
	}
	return VisionCredentials{
		APIKey:  creds.APIKey,
		BaseURL: creds.BaseURL,
		Model:   creds.Model,
	}, true, nil
}
