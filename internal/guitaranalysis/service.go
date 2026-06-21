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

// ReanalyzeCollectionResult summarizes a bulk re-analysis run.
type ReanalyzeCollectionResult struct {
	Total    int `json:"total"`
	Analyzed int `json:"analyzed"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

type analyzeOpts struct {
	requireOptIn bool
	force        bool
	runVision    bool
}

// ScheduleIfEligible marks analysis pending when opted in and pictures changed, without calling vision.
func (s *Service) ScheduleIfEligible(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: true, force: false, runVision: false})
}

// AnalyzeIfEligible runs vision analysis when the owner opted in with BYOK and pictures changed.
func (s *Service) AnalyzeIfEligible(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: true, force: false, runVision: true})
}

// Reanalyze runs vision analysis for an owner guitar using BYOK, even when pictures are unchanged.
func (s *Service) Reanalyze(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: true})
}

// ReanalyzeCollection re-runs photo analysis for every guitar owned by the caller with pictures.
func (s *Service) ReanalyzeCollection(ctx context.Context, ownerID string, guitars []*domain.Guitar) (ReanalyzeCollectionResult, error) {
	result := ReanalyzeCollectionResult{Total: len(guitars)}
	if s == nil || s.repo == nil {
		return result, nil
	}
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return result, fmt.Errorf("owner id is required")
	}
	creds, ok, err := s.owners.VisionCredentials(ctx, ownerID)
	if err != nil {
		return result, err
	}
	if !ok || strings.TrimSpace(creds.APIKey) == "" {
		return result, ErrBYOKNotConfigured
	}
	for _, guitar := range guitars {
		if guitar == nil || guitar.Owner() != ownerID {
			result.Skipped++
			continue
		}
		if len(guitar.Pictures()) == 0 || CoverPictureURL(guitar) == "" {
			result.Skipped++
			continue
		}
		if rec, err := s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: true}); err != nil {
			result.Failed++
			continue
		} else if rec == nil || rec.Status() != StatusReady {
			result.Failed++
			continue
		}
		result.Analyzed++
	}
	return result, nil
}

func (s *Service) analyze(ctx context.Context, guitar *domain.Guitar, opts analyzeOpts) (*Record, error) {
	if s == nil || s.repo == nil || guitar == nil {
		return nil, nil
	}
	ownerID := strings.TrimSpace(guitar.Owner())
	if ownerID == "" {
		return nil, nil
	}
	if opts.requireOptIn {
		enabled, err := s.owners.PhotoAnalysisEnabled(ctx, ownerID)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return nil, nil
		}
	}
	creds, ok, err := s.owners.VisionCredentials(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !ok || strings.TrimSpace(creds.APIKey) == "" {
		if opts.force {
			return nil, ErrBYOKNotConfigured
		}
		return nil, nil
	}
	pictures := guitar.Pictures()
	if len(pictures) == 0 {
		return nil, nil
	}
	coverURL := CoverPictureURL(guitar)
	if coverURL == "" {
		return nil, nil
	}
	fingerprint := CoverFingerprintForGuitar(guitar)
	existing, err := s.repo.FindByGuitarID(ctx, guitar.ID())
	if err != nil {
		return nil, err
	}
	if !opts.force && existing != nil && existing.PicturesFingerprint() == fingerprint && existing.Status() == StatusReady {
		return existing, nil
	}
	record, err := s.pendingRecord(guitar.ID(), ownerID, fingerprint, existing)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, record); err != nil {
		return nil, err
	}
	if !opts.runVision {
		return record, nil
	}
	if s.vision == nil {
		record.SetFailed(fingerprint, "vision analyzer not configured")
		_ = s.repo.Save(ctx, record)
		return record, nil
	}
	result, err := s.vision.AnalyzeCoverPicture(ctx, creds, coverURL, guitar.Brand(), guitar.TypeName())
	if err != nil {
		record.SetFailed(fingerprint, err.Error())
		_ = s.repo.Save(ctx, record)
		return record, nil
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
