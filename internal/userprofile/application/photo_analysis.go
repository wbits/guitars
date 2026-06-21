package application

import (
	"context"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// SetPhotoAnalysisEnabled toggles automatic photo analysis on upload.
func (s *Service) SetPhotoAnalysisEnabled(ctx context.Context, userID, email string, enabled bool) (*domain.Profile, error) {
	profile, err := s.GetProfile(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	if err := profile.SetPhotoAnalysisEnabled(enabled); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// PhotoAnalysisEnabled reports whether analysis should run for this owner's guitars.
func (s *Service) PhotoAnalysisEnabled(ctx context.Context, userID string) (bool, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if profile == nil {
		return false, nil
	}
	return profile.PhotoAnalysisEnabled(), nil
}
