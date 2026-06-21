package guitaranalysis

import (
	"context"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// CatalogSuggestions are AI guesses for catalog fields (advisory).
type CatalogSuggestions struct {
	Brand       string `json:"brand,omitempty"`
	TypeName    string `json:"typeName,omitempty"`
	Color       string `json:"color,omitempty"`
	BuildYear   int    `json:"buildYear,omitempty"`
	Description string `json:"description,omitempty"`
}

// CatalogAnalysisResult is ephemeral vision output for the add-from-photo flow.
type CatalogAnalysisResult struct {
	PictureURL    string             `json:"pictureUrl"`
	VisualSummary string             `json:"visualSummary"`
	Tags          []string           `json:"tags,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Suggestions   CatalogSuggestions `json:"suggestions"`
}

type catalogVisionResult struct {
	VisualSummary string   `json:"visualSummary"`
	Tags          []string `json:"tags"`
	Confidence    float64  `json:"confidence"`
	Brand         string   `json:"brand"`
	TypeName      string   `json:"typeName"`
	Color         string   `json:"color"`
	BuildYear     int      `json:"buildYear"`
	Description   string   `json:"description"`
}

// AnalyzePictureForCatalog runs vision against a picture URL without a saved guitar.
func (s *Service) AnalyzePictureForCatalog(ctx context.Context, ownerID, pictureURL string) (CatalogAnalysisResult, error) {
	if s == nil {
		return CatalogAnalysisResult{}, fmt.Errorf("analysis service not configured")
	}
	ownerID = strings.TrimSpace(ownerID)
	pictureURL = strings.TrimSpace(pictureURL)
	if ownerID == "" {
		return CatalogAnalysisResult{}, InvalidField("owner", "is required")
	}
	if pictureURL == "" {
		return CatalogAnalysisResult{}, InvalidField("pictureUrl", "is required")
	}
	if s.owners == nil {
		return CatalogAnalysisResult{}, fmt.Errorf("owner settings not configured")
	}
	enabled, err := s.owners.PhotoAnalysisEnabled(ctx, ownerID)
	if err != nil {
		return CatalogAnalysisResult{}, err
	}
	if !enabled {
		return CatalogAnalysisResult{}, ErrPhotoAnalysisDisabled
	}
	creds, ok, err := s.owners.VisionCredentials(ctx, ownerID)
	if err != nil {
		return CatalogAnalysisResult{}, err
	}
	if !ok || strings.TrimSpace(creds.APIKey) == "" {
		return CatalogAnalysisResult{}, ErrBYOKNotConfigured
	}
	if s.vision == nil {
		return CatalogAnalysisResult{}, fmt.Errorf("vision analyzer not configured")
	}
	result, err := s.vision.AnalyzePictureForCatalog(ctx, creds, pictureURL)
	if err != nil {
		return CatalogAnalysisResult{}, err
	}
	return CatalogAnalysisResult{
		PictureURL:    pictureURL,
		VisualSummary: result.VisualSummary,
		Tags:          result.Tags,
		Confidence:    result.Confidence,
		Suggestions: CatalogSuggestions{
			Brand:       result.Brand,
			TypeName:    result.TypeName,
			Color:       result.Color,
			BuildYear:   result.BuildYear,
			Description: result.Description,
		},
	}, nil
}

// SeedFromCatalogAnalysis stores a ready analysis record without re-running vision.
func (s *Service) SeedFromCatalogAnalysis(ctx context.Context, guitar *domain.Guitar, result CatalogAnalysisResult) error {
	if s == nil || s.repo == nil || guitar == nil {
		return nil
	}
	ownerID := strings.TrimSpace(guitar.Owner())
	if ownerID == "" {
		return nil
	}
	fingerprint := CoverFingerprintForGuitar(guitar)
	record, err := NewRecord(guitar.ID(), ownerID, StatusReady, fingerprint)
	if err != nil {
		return err
	}
	summary := strings.TrimSpace(result.VisualSummary)
	if summary == "" {
		summary = strings.TrimSpace(result.Suggestions.Description)
	}
	confidence := result.Confidence
	if confidence <= 0 {
		confidence = 0.7
	}
	record.SetReady(fingerprint, summary, result.Tags, confidence)
	return s.repo.Save(ctx, record)
}
