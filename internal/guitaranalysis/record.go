package guitaranalysis

import (
	"strings"
	"time"
)

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
)

// Record holds AI-derived metadata for a guitar (advisory, not authoritative).
type Record struct {
	guitarID           string
	ownerID            string
	status             string
	picturesFingerprint string // cover index + URL fingerprint (legacy field name)
	visualSummary      string
	tags               []string
	confidence         float64
	failureReason      string
	analyzedAt         time.Time
	updatedAt          time.Time
}

// NewRecord validates and constructs a Record.
func NewRecord(guitarID, ownerID, status, fingerprint string) (*Record, error) {
	guitarID = strings.TrimSpace(guitarID)
	ownerID = strings.TrimSpace(ownerID)
	if guitarID == "" {
		return nil, InvalidField("guitarId", "is required")
	}
	if ownerID == "" {
		return nil, InvalidField("ownerId", "is required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = StatusPending
	}
	now := time.Now().UTC()
	return &Record{
		guitarID:            guitarID,
		ownerID:             ownerID,
		status:              status,
		picturesFingerprint: strings.TrimSpace(fingerprint),
		updatedAt:           now,
	}, nil
}

func (r *Record) GuitarID() string            { return r.guitarID }
func (r *Record) OwnerID() string             { return r.ownerID }
func (r *Record) Status() string              { return r.status }
func (r *Record) PicturesFingerprint() string { return r.picturesFingerprint }
func (r *Record) VisualSummary() string       { return r.visualSummary }
func (r *Record) Tags() []string {
	out := make([]string, len(r.tags))
	copy(out, r.tags)
	return out
}
func (r *Record) Confidence() float64  { return r.confidence }
func (r *Record) FailureReason() string { return r.failureReason }
func (r *Record) AnalyzedAt() time.Time { return r.analyzedAt }
func (r *Record) UpdatedAt() time.Time  { return r.updatedAt }

func (r *Record) SetPending(fingerprint string) {
	r.status = StatusPending
	r.picturesFingerprint = strings.TrimSpace(fingerprint)
	r.visualSummary = ""
	r.tags = nil
	r.confidence = 0
	r.failureReason = ""
	r.analyzedAt = time.Time{}
	r.updatedAt = time.Now().UTC()
}

func (r *Record) SetReady(fingerprint, summary string, tags []string, confidence float64) {
	r.status = StatusReady
	r.picturesFingerprint = strings.TrimSpace(fingerprint)
	r.visualSummary = strings.TrimSpace(summary)
	r.tags = normalizeTags(tags)
	r.confidence = confidence
	r.failureReason = ""
	r.analyzedAt = time.Now().UTC()
	r.updatedAt = r.analyzedAt
}

func (r *Record) SetFailed(fingerprint, reason string) {
	r.status = StatusFailed
	r.picturesFingerprint = strings.TrimSpace(fingerprint)
	r.failureReason = strings.TrimSpace(reason)
	r.updatedAt = time.Now().UTC()
}

// SearchBlob returns lowercase text used for assistant filtering.
func (r *Record) SearchBlob() string {
	if r == nil || r.status != StatusReady {
		return ""
	}
	parts := append([]string{r.visualSummary}, r.tags...)
	return strings.ToLower(strings.Join(parts, " "))
}

func normalizeTags(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
