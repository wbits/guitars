package marketcrawler

import "context"

// Source searches a marketplace for listings similar to a guitar.
type Source interface {
	Name() string
	Search(ctx context.Context, guitar GuitarSummary) ([]Finding, error)
}
