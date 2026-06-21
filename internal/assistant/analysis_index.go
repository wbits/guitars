package assistant

import (
	"context"

	"github.com/wbits/guitars/internal/guitaranalysis"
)

// GuitarAnalysisIndex loads AI metadata for assistant filtering.
type GuitarAnalysisIndex struct {
	Service *guitaranalysis.Service
}

func (g *GuitarAnalysisIndex) MapForGuitars(ctx context.Context, guitarIDs []string) (map[string]AnalysisSearch, error) {
	out := map[string]AnalysisSearch{}
	if g == nil || g.Service == nil {
		return out, nil
	}
	records, err := g.Service.MapForGuitars(ctx, guitarIDs)
	if err != nil {
		return nil, err
	}
	for id, rec := range records {
		out[id] = rec
	}
	return out, nil
}
