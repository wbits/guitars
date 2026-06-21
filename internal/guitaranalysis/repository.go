package guitaranalysis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Repository persists guitar analysis records.
type Repository interface {
	FindByGuitarID(ctx context.Context, guitarID string) (*Record, error)
	FindByGuitarIDs(ctx context.Context, guitarIDs []string) (map[string]*Record, error)
	Save(ctx context.Context, record *Record) error
	DeleteByGuitarID(ctx context.Context, guitarID string) error
}

// PicturesFingerprint hashes the picture URL set for change detection.
func PicturesFingerprint(urls []string) string {
	clean := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u != "" {
			clean = append(clean, u)
		}
	}
	sort.Strings(clean)
	sum := sha256.Sum256([]byte(strings.Join(clean, "\n")))
	return hex.EncodeToString(sum[:])
}
