package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// ErrBYOKNotConfigured is returned when BYOK endpoints are unavailable.
var ErrBYOKNotConfigured = errors.New("assistant BYOK is not configured on the server")

// ErrBYOKDecryptFailed is returned when stored credentials cannot be decrypted
// (for example after the server encryption key changed).
var ErrBYOKDecryptFailed = errors.New("assistant BYOK credentials could not be decrypted")

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

// AssistantBYOKMeStatus reports whether BYOK is stored and whether it can be decrypted for use.
func (s *Service) AssistantBYOKMeStatus(ctx context.Context, userID string) (configured bool, usable bool, err error) {
	if s == nil || s.repo == nil {
		return false, false, nil
	}
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return false, false, err
	}
	if profile == nil || !profile.AssistantBYOKConfigured() {
		return false, false, nil
	}
	_, ok, err := s.AssistantBYOKCredentialsForUser(ctx, userID)
	if IsBYOKDecryptFailed(err) {
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return true, ok, nil
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
		return AssistantBYOKCredentials{}, false, fmt.Errorf("%w: %v", ErrBYOKDecryptFailed, err)
	}
	return AssistantBYOKCredentials{
		APIKey:  apiKey,
		BaseURL: profile.AssistantLLMBaseURL(),
		Model:   profile.AssistantLLMModel(),
	}, true, nil
}
