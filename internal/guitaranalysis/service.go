package guitaranalysis

import (
	"context"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// OwnerSettingsLoader returns whether photo analysis is enabled and BYOK credentials.
type OwnerSettingsLoader interface {
	PhotoAnalysisEnabled(ctx context.Context, ownerID string) (bool, error)
	VisionCredentials(ctx context.Context, ownerID string) (VisionCredentials, bool, error)
}

// Service coordinates analysis persistence and vision calls.
type Service struct {
	repo    Repository
	owners  OwnerSettingsLoader
	vision  *VisionAnalyzer
}

func NewService(repo Repository, owners OwnerSettingsLoader, vision *VisionAnalyzer) *Service {
	return &Service{repo: repo, owners: owners, vision: vision}
}

func (s *Service) Get(ctx context.Context, guitarID string) (*Record, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.FindByGuitarID(ctx, guitarID)
}

func (s *Service) MapForGuitars(ctx context.Context, guitarIDs []string) (map[string]*Record, error) {
	if s == nil || s.repo == nil || len(guitarIDs) == 0 {
		return map[string]*Record{}, nil
	}
	return s.repo.FindByGuitarIDs(ctx, guitarIDs)
}

// AnalyzeIfEligible runs vision analysis when the owner opted in with BYOK and pictures changed.
func (s *Service) AnalyzeIfEligible(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	if s == nil || s.repo == nil || guitar == nil {
		return nil, nil
	}
	ownerID := strings.TrimSpace(guitar.Owner())
	if ownerID == "" {
		return nil, nil
	}
	enabled, err := s.owners.PhotoAnalysisEnabled(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}
	creds, ok, err := s.owners.VisionCredentials(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !ok || strings.TrimSpace(creds.APIKey) == "" {
		return nil, nil
	}
	pictures := guitar.Pictures()
	if len(pictures) == 0 {
		return nil, nil
	}
	fingerprint := PicturesFingerprint(pictures)
	existing, err := s.repo.FindByGuitarID(ctx, guitar.ID())
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.PicturesFingerprint() == fingerprint && existing.Status() == StatusReady {
		return existing, nil
	}
	record, err := s.pendingRecord(guitar.ID(), ownerID, fingerprint, existing)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, record); err != nil {
		return nil, err
	}
	if s.vision == nil {
		record.SetFailed(fingerprint, "vision analyzer not configured")
		_ = s.repo.Save(ctx, record)
		return record, nil
	}
	result, err := s.vision.AnalyzePictures(ctx, creds, pictures, guitar.Brand(), guitar.TypeName())
	if err != nil {
		record.SetFailed(fingerprint, err.Error())
		_ = s.repo.Save(ctx, record)
		return record, fmt.Errorf("analyze guitar %s: %w", guitar.ID(), err)
	}
	record.SetReady(fingerprint, result.VisualSummary, result.Tags, result.Confidence)
	if err := s.repo.Save(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) pendingRecord(guitarID, ownerID, fingerprint string, existing *Record) (*Record, error) {
	if existing != nil {
		existing.SetPending(fingerprint)
		return existing, nil
	}
	return NewRecord(guitarID, ownerID, StatusPending, fingerprint)
}

func (s *Service) DeleteForGuitar(ctx context.Context, guitarID string) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.DeleteByGuitarID(ctx, guitarID)
}
