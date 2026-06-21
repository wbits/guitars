package guitaranalysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// CoverPictureURL returns the URL at pictures[coverPictureIndex], or empty when unavailable.
func CoverPictureURL(g *domain.Guitar) string {
	if g == nil {
		return ""
	}
	pictures := g.Pictures()
	if len(pictures) == 0 {
		return ""
	}
	idx := g.CoverPictureIndex()
	if idx < 0 || idx >= len(pictures) {
		return ""
	}
	return strings.TrimSpace(pictures[idx])
}

// CoverFingerprint fingerprints the cover selection for change detection.
// Analysis re-runs when the cover index or cover URL changes.
func CoverFingerprint(coverIndex int, coverURL string) string {
	coverURL = strings.TrimSpace(coverURL)
	payload := fmt.Sprintf("%d\n%s", coverIndex, coverURL)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// CoverFingerprintForGuitar fingerprints the guitar's current cover picture selection.
func CoverFingerprintForGuitar(g *domain.Guitar) string {
	if g == nil {
		return ""
	}
	return CoverFingerprint(g.CoverPictureIndex(), CoverPictureURL(g))
}
