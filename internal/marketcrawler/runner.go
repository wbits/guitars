package marketcrawler

import (
	"context"
	"fmt"
	"log"
)

// Runner coordinates source adapters and uploads findings to the API.
type Runner struct {
	API     *APIClient
	Sources []Source
	Logger  *log.Logger
}

// RunAll crawls every guitar in the collection.
func (r *Runner) RunAll(ctx context.Context) error {
	guitars, err := r.API.ListGuitars(ctx)
	if err != nil {
		return err
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
	if err := r.API.UploadMarketLogs(ctx, guitar.ID, all); err != nil {
		return err
	}
	logger.Printf("uploaded %d market logs for guitar %s", len(all), guitar.ID)
	return nil
}
