package application

import (
	"context"
	"errors"
	"strings"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// Service coordinates user profile use cases.
type Service struct {
	repo          domain.Repository
	byokEncryptor BYOKEncryptor
}

// NewService wires the profile application service. byokEncryptor may be nil to disable BYOK storage.
func NewService(repo domain.Repository, byokEncryptor BYOKEncryptor) *Service {
	return &Service{repo: repo, byokEncryptor: byokEncryptor}
}

// GetProfile returns the caller's profile, creating a stub record when needed.
func (s *Service) GetProfile(ctx context.Context, userID, email string) (*domain.Profile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.InvalidField("userId", "is required")
	}
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile, err = domain.NewProfile(domain.ProfileProps{
			UserID: userID,
			Email:  email,
		})
		if err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, profile); err != nil {
			return nil, err
		}
		return profile, nil
	}
	if email != "" && profile.Email() == "" {
		profile.SetEmail(email)
		if err := s.repo.Save(ctx, profile); err != nil {
			return nil, err
		}
	}
	return profile, nil
}

// UpdateUsername sets the caller's custom username.
func (s *Service) UpdateUsername(ctx context.Context, userID, email, username string) (*domain.Profile, error) {
	profile, err := s.GetProfile(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	if err := profile.SetUsername(username); err != nil {
		return nil, err
	}
	if profile.Username() != "" {
		existing, err := s.repo.FindByUsername(ctx, profile.Username())
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.UserID() != profile.UserID() {
			return nil, domain.ErrUsernameTaken
		}
	}
	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfilesByUserIDs returns stored profiles keyed by user id.
func (s *Service) GetProfilesByUserIDs(ctx context.Context, userIDs []string) (map[string]*domain.Profile, error) {
	return s.repo.FindByUserIDs(ctx, userIDs)
}

// SetMarketCrawlEnabled toggles market crawl for a user collection.
func (s *Service) SetMarketCrawlEnabled(ctx context.Context, userID string, enabled bool) (*domain.Profile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.InvalidField("userId", "is required")
	}
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile, err = domain.NewProfile(domain.ProfileProps{UserID: userID})
		if err != nil {
			return nil, err
		}
	}
	profile.SetMarketCrawlEnabled(enabled)
	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// MarketCrawlEnabledForUser reports whether market crawl is enabled for the given user.
func (s *Service) MarketCrawlEnabledForUser(ctx context.Context, userID string) (bool, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if profile == nil {
		return false, nil
	}
	return profile.MarketCrawlEnabled(), nil
}

// DisplayNameForUser resolves a display label for a user id using stored profile data.
func DisplayNameForUser(userID string, profile *domain.Profile) string {
	if profile != nil {
		return profile.DisplayName()
	}
	if userID == "local-dev-user" {
		return "local-dev@example.com"
	}
	return userID
}

// IsValidationError reports whether err is a profile validation failure.
func IsValidationError(err error) bool {
	return domain.IsValidationError(err)
}

// IsUsernameTaken reports whether err is ErrUsernameTaken.
func IsUsernameTaken(err error) bool {
	return errors.Is(err, domain.ErrUsernameTaken)
}

// IsBYOKNotConfigured reports whether BYOK storage is unavailable on the server.
func IsBYOKNotConfigured(err error) bool {
	return errors.Is(err, ErrBYOKNotConfigured)
}
