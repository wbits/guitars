package application

import (
	"context"
	"errors"
	"strings"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// ErrBYOKNotConfigured is returned when BYOK endpoints are unavailable.
var ErrBYOKNotConfigured = errors.New("assistant BYOK is not configured on the server")

// BYOKEncryptor encrypts owner API keys at rest.
type BYOKEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AssistantBYOKSettings is the non-secret assistant BYOK configuration exposed to clients.
type AssistantBYOKSettings struct {
	Configured bool
	BaseURL    string
	Model      string
}

// AssistantBYOKCredentials are decrypted credentials for server-side LLM calls.
type AssistantBYOKCredentials struct {
	APIKey  string
	BaseURL string
	Model   string
}

// SetAssistantBYOK stores an encrypted owner API key and optional LLM endpoint settings.
func (s *Service) SetAssistantBYOK(ctx context.Context, userID, email, apiKey, baseURL, model string) (*domain.Profile, error) {
	if s.byokEncryptor == nil {
		return nil, ErrBYOKNotConfigured
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, domain.InvalidField("apiKey", "is required")
	}
	profile, err := s.GetProfile(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	encrypted, err := s.byokEncryptor.Encrypt(apiKey)
	if err != nil {
		return nil, err
	}
	if err := profile.SetAssistantBYOK(encrypted, baseURL, model); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// ClearAssistantBYOK removes stored owner assistant credentials.
func (s *Service) ClearAssistantBYOK(ctx context.Context, userID, email string) (*domain.Profile, error) {
	profile, err := s.GetProfile(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	profile.ClearAssistantBYOK()
	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// AssistantBYOKSettingsForUser returns client-safe BYOK status for the owner.
func (s *Service) AssistantBYOKSettingsForUser(ctx context.Context, userID string) (AssistantBYOKSettings, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return AssistantBYOKSettings{}, err
	}
	if profile == nil || !profile.AssistantBYOKConfigured() {
		return AssistantBYOKSettings{}, nil
	}
	return AssistantBYOKSettings{
		Configured: true,
		BaseURL:    profile.AssistantLLMBaseURL(),
		Model:      profile.AssistantLLMModel(),
	}, nil
}

// AssistantBYOKCredentialsForUser decrypts stored owner credentials when configured.
func (s *Service) AssistantBYOKCredentialsForUser(ctx context.Context, userID string) (AssistantBYOKCredentials, bool, error) {
	if s.byokEncryptor == nil {
		return AssistantBYOKCredentials{}, false, nil
	}
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return AssistantBYOKCredentials{}, false, err
	}
	if profile == nil || !profile.AssistantBYOKConfigured() {
		return AssistantBYOKCredentials{}, false, nil
	}
	apiKey, err := s.byokEncryptor.Decrypt(profile.AssistantEncryptedAPIKey())
	if err != nil {
		return AssistantBYOKCredentials{}, false, err
	}
	return AssistantBYOKCredentials{
		APIKey:  apiKey,
		BaseURL: profile.AssistantLLMBaseURL(),
		Model:   profile.AssistantLLMModel(),
	}, true, nil
}
