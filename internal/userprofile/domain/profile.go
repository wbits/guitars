package domain

import (
	"regexp"
	"strings"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)

// Profile holds display identity for an authenticated user.
type Profile struct {
	userID                  string
	username                string
	email                   string
	marketCrawlEnabled      bool
	assistantEncryptedAPIKey string
	assistantLLMBaseURL      string
	assistantLLMModel        string
}

// ProfileProps are the mutable fields of a Profile.
type ProfileProps struct {
	UserID                  string
	Username                string
	Email                   string
	MarketCrawlEnabled      bool
	AssistantEncryptedAPIKey string
	AssistantLLMBaseURL      string
	AssistantLLMModel        string
}

// NewProfile validates and constructs a Profile.
func NewProfile(props ProfileProps) (*Profile, error) {
	userID := strings.TrimSpace(props.UserID)
	if userID == "" {
		return nil, InvalidField("userId", "is required")
	}
	username := strings.TrimSpace(props.Username)
	if username != "" {
		if err := ValidateUsername(username); err != nil {
			return nil, err
		}
	}
	return &Profile{
		userID:                  userID,
		username:                username,
		email:                   strings.TrimSpace(props.Email),
		marketCrawlEnabled:      props.MarketCrawlEnabled,
		assistantEncryptedAPIKey: strings.TrimSpace(props.AssistantEncryptedAPIKey),
		assistantLLMBaseURL:      strings.TrimSpace(props.AssistantLLMBaseURL),
		assistantLLMModel:        strings.TrimSpace(props.AssistantLLMModel),
	}, nil
}

// ValidateUsername checks a candidate custom username.
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return InvalidField("username", "is required")
	}
	if !usernamePattern.MatchString(username) {
		return InvalidField("username", "must be 3-30 characters and contain only letters, numbers, underscores, or hyphens")
	}
	return nil
}

func (p *Profile) UserID() string   { return p.userID }
func (p *Profile) Username() string { return p.username }
func (p *Profile) Email() string    { return p.email }

// MarketCrawlEnabled reports whether automated market crawling is enabled for this collection.
func (p *Profile) MarketCrawlEnabled() bool { return p.marketCrawlEnabled }

// SetMarketCrawlEnabled toggles automated market crawling for this collection.
func (p *Profile) SetMarketCrawlEnabled(enabled bool) {
	p.marketCrawlEnabled = enabled
}

// DisplayName returns the username when set, otherwise the email, otherwise the user id.
func (p *Profile) DisplayName() string {
	if name := strings.TrimSpace(p.username); name != "" {
		return name
	}
	if email := strings.TrimSpace(p.email); email != "" {
		return email
	}
	return p.userID
}

// SetUsername updates the custom username. An empty value clears it.
func (p *Profile) SetUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		p.username = ""
		return nil
	}
	if err := ValidateUsername(username); err != nil {
		return err
	}
	p.username = username
	return nil
}

// SetEmail stores the email address when known from authentication.
func (p *Profile) SetEmail(email string) {
	p.email = strings.TrimSpace(email)
}

// AssistantBYOKConfigured reports whether an encrypted assistant API key is stored.
func (p *Profile) AssistantBYOKConfigured() bool {
	return strings.TrimSpace(p.assistantEncryptedAPIKey) != ""
}

func (p *Profile) AssistantEncryptedAPIKey() string { return p.assistantEncryptedAPIKey }
func (p *Profile) AssistantLLMBaseURL() string    { return p.assistantLLMBaseURL }
func (p *Profile) AssistantLLMModel() string      { return p.assistantLLMModel }

// SetAssistantBYOK stores encrypted assistant credentials (never plaintext).
func (p *Profile) SetAssistantBYOK(encryptedAPIKey, baseURL, model string) error {
	encryptedAPIKey = strings.TrimSpace(encryptedAPIKey)
	if encryptedAPIKey == "" {
		return InvalidField("apiKey", "is required")
	}
	p.assistantEncryptedAPIKey = encryptedAPIKey
	p.assistantLLMBaseURL = strings.TrimSpace(baseURL)
	p.assistantLLMModel = strings.TrimSpace(model)
	return nil
}

// ClearAssistantBYOK removes stored assistant credentials.
func (p *Profile) ClearAssistantBYOK() {
	p.assistantEncryptedAPIKey = ""
	p.assistantLLMBaseURL = ""
	p.assistantLLMModel = ""
}
