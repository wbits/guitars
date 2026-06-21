package guitaranalysis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// OwnerSettingsLoader returns whether photo analysis is enabled and BYOK credentials.
type OwnerSettingsLoader interface {
	PhotoAnalysisEnabled(ctx context.Context, ownerID string) (bool, error)
	VisionCredentials(ctx context.Context, ownerID string) (VisionCredentials, bool, error)
}

// Service coordinates analysis persistence, queueing, and vision calls.
type Service struct {
	repo    Repository
	owners  OwnerSettingsLoader
	vision  *VisionAnalyzer
	queue   JobQueue
	guitars GuitarLoader
}

func NewService(repo Repository, owners OwnerSettingsLoader, vision *VisionAnalyzer, queue JobQueue, guitars GuitarLoader) *Service {
	return &Service{repo: repo, owners: owners, vision: vision, queue: queue, guitars: guitars}
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

// ReanalyzeCollectionResult summarizes a bulk enqueue run.
type ReanalyzeCollectionResult struct {
	Total   int `json:"total"`
	Queued  int `json:"queued"`
	Skipped int `json:"skipped"`
}

type analyzeOpts struct {
	requireOptIn bool
	force        bool
	runVision    bool
}

// ScheduleIfEligible marks analysis pending when opted in and the cover changed.
func (s *Service) ScheduleIfEligible(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: true, force: false, runVision: false})
}

// ScheduleAndQueueIfEligible marks pending and enqueues a worker job when eligible.
func (s *Service) ScheduleAndQueueIfEligible(ctx context.Context, guitar *domain.Guitar) error {
	rec, err := s.ScheduleIfEligible(ctx, guitar)
	if err != nil || rec == nil || rec.Status() != StatusPending {
		return err
	}
	return s.enqueueOrRun(ctx, guitar, true)
}

// AnalyzeIfEligible runs vision analysis synchronously (tests and queue-less local dev).
func (s *Service) AnalyzeIfEligible(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: true, force: false, runVision: true})
}

// QueueReanalyze marks pending and enqueues a forced re-analysis job.
func (s *Service) QueueReanalyze(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	rec, err := s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: false})
	if err != nil || rec == nil {
		return rec, err
	}
	if err := s.enqueueOrRun(ctx, guitar, true); err != nil {
		return rec, err
	}
	return rec, nil
}

// Reanalyze runs vision synchronously (used in unit tests).
func (s *Service) Reanalyze(ctx context.Context, guitar *domain.Guitar) (*Record, error) {
	return s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: true})
}

// QueueReanalyzeCollection enqueues forced re-analysis for every owned guitar with a cover photo.
func (s *Service) QueueReanalyzeCollection(ctx context.Context, ownerID string, guitars []*domain.Guitar) (ReanalyzeCollectionResult, error) {
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
		rec, err := s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: false})
		if err != nil {
			return result, err
		}
		if rec == nil {
			result.Skipped++
			continue
		}
		if err := s.enqueueOrRun(ctx, guitar, true); err != nil {
			return result, err
		}
		result.Queued++
	}
	return result, nil
}

// ReanalyzeCollection runs vision synchronously for each guitar (unit tests).
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
		rec, err := s.analyze(ctx, guitar, analyzeOpts{requireOptIn: false, force: true, runVision: true})
		if err != nil {
			return result, err
		}
		if rec == nil || rec.Status() != StatusReady {
			continue
		}
		result.Queued++
	}
	return result, nil
}

// ProcessJob runs vision analysis for a queued job (worker Lambda).
func (s *Service) ProcessJob(ctx context.Context, job Job) error {
	if s == nil {
		return nil
	}
	guitarID := strings.TrimSpace(job.GuitarID)
	if guitarID == "" {
		return nil
	}
	if s.guitars == nil {
		return fmt.Errorf("guitar loader not configured")
	}
	guitar, err := s.guitars.LoadGuitar(ctx, guitarID)
	if err != nil {
		if errors.Is(err, domain.ErrGuitarNotFound) {
			return nil
		}
		return err
	}
	if guitar == nil {
		return nil
	}
	ownerID := strings.TrimSpace(job.OwnerID)
	if ownerID != "" && guitar.Owner() != ownerID {
		return nil
	}
	_, err = s.analyze(ctx, guitar, analyzeOpts{
		requireOptIn: !job.Force,
		force:        job.Force,
		runVision:    true,
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrBYOKNotConfigured) {
		return nil
	}
	return err
}

func (s *Service) enqueueOrRun(ctx context.Context, guitar *domain.Guitar, force bool) error {
	if guitar == nil {
		return nil
	}
	job := Job{
		GuitarID: guitar.ID(),
		OwnerID:  guitar.Owner(),
		Force:    force,
	}
	if s.queue != nil {
		return s.queue.Enqueue(ctx, job)
	}
	_, err := s.analyze(ctx, guitar, analyzeOpts{
		requireOptIn: !force,
		force:        force,
		runVision:    true,
	})
	return err
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
