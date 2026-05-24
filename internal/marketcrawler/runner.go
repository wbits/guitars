package marketcrawler

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Runner coordinates source adapters and uploads findings to the API.
type Runner struct {
	API     *APIClient
	Sources []Source
	Images  ImageUploader
	Logger  *log.Logger
}

// ImageUploader stores listing images on the collection CDN.
type ImageUploader interface {
	Upload(ctx context.Context, sourceURL string) (cdnURL string, err error)
}

// RunAll crawls every guitar in the collection.
func (r *Runner) RunAll(ctx context.Context) error {
	logger := r.Logger
	if logger == nil {
		logger = log.Default()
	}
	guitars, err := r.API.ListGuitars(ctx)
	if err != nil {
		return err
	}
	logger.Printf("found %d guitars to crawl", len(guitars))
	if len(guitars) == 0 {
		logger.Print("no guitars found across collections")
		return nil
	}
	for _, g := range guitars {
		if err := r.RunGuitar(ctx, GuitarSummary(g)); err != nil {
			return fmt.Errorf("crawl guitar %s: %w", g.ID, err)
		}
	}
	return nil
}

// RunGuitar crawls marketplaces for a single guitar and uploads findings.
func (r *Runner) RunGuitar(ctx context.Context, guitar GuitarSummary) error {
	logger := r.Logger
	if logger == nil {
		logger = log.Default()
	}
	var all []Finding
	for _, source := range r.Sources {
		findings, err := source.Search(ctx, guitar)
		if err != nil {
			logger.Printf("source %s guitar %s: %v", source.Name(), guitar.ID, err)
			continue
		}
		logger.Printf("source %s guitar %s: %d findings", source.Name(), guitar.ID, len(findings))
		all = append(all, findings...)
	}
	if r.Images != nil {
		for i := range all {
			if strings.TrimSpace(all[i].SourceImageURL) == "" {
				continue
			}
			cdnURL, err := r.Images.Upload(ctx, all[i].SourceImageURL)
			if err != nil {
				logger.Printf("image guitar %s listing %s: %v", guitar.ID, all[i].ExternalListingID, err)
				continue
			}
			all[i].ListingImageURL = cdnURL
		}
	}
	if err := r.API.UploadMarketLogs(ctx, guitar.ID, all); err != nil {
		return err
	}
	if len(all) == 0 {
		logger.Printf("no marketplace findings for guitar %s", guitar.ID)
	} else {
		logger.Printf("uploaded %d market logs for guitar %s", len(all), guitar.ID)
	}
	return nil
}
