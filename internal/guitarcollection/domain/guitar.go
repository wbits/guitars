package domain

import (
	"net/url"
	"strings"
	"time"
)

// minBuildYear is the earliest year for which a guitar in this collection is
// considered plausible. The very first acoustic guitars predate this, but the
// constraint exists to catch obvious data-entry errors (e.g. year 19 instead
// of 1990).
const minBuildYear = 1800

// Guitar is the sole aggregate root of the GuitarCollection bounded context.
//
// Construction is funneled through NewGuitar / Rehydrate so that invariants
// cannot be bypassed: every Guitar instance that exists in memory is
// guaranteed to be valid.
type Guitar struct {
	id                string
	serialNumber      string
	pictures          []string
	coverPictureIndex int
	description       string
	brand             string
	typeName          string
	buildYear         int
	price             Money
}

// GuitarProps is the data-transfer shape used to create or update a Guitar.
// Using a struct keeps the constructor stable as fields are added.
type GuitarProps struct {
	ID                string
	SerialNumber      string
	Pictures          []string
	CoverPictureIndex int
	Description       string
	Brand             string
	TypeName          string
	BuildYear         int
	Price             Money
}

// NewGuitar validates the supplied props and returns a freshly built Guitar.
// The returned error is always a *ValidationError when validation fails.
func NewGuitar(p GuitarProps) (*Guitar, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, newValidationError("id", "is required")
	}
	if strings.TrimSpace(p.Brand) == "" {
		return nil, newValidationError("brand", "is required")
	}
	if strings.TrimSpace(p.TypeName) == "" {
		return nil, newValidationError("typeName", "is required")
	}

	currentYear := time.Now().UTC().Year()
	if p.BuildYear < minBuildYear || p.BuildYear > currentYear+1 {
		return nil, newValidationError("buildYear", "must be between 1800 and next year")
	}

	if (p.Price == Money{}) {
		return nil, newValidationError("price", "is required")
	}

	pictures, err := validatePictureURLs(p.Pictures)
	if err != nil {
		return nil, err
	}
	coverIndex, err := validateCoverPictureIndex(p.CoverPictureIndex, len(pictures))
	if err != nil {
		return nil, err
	}

	return &Guitar{
		id:                strings.TrimSpace(p.ID),
		serialNumber:      strings.TrimSpace(p.SerialNumber),
		pictures:          pictures,
		coverPictureIndex: coverIndex,
		description:       strings.TrimSpace(p.Description),
		brand:             strings.TrimSpace(p.Brand),
		typeName:          strings.TrimSpace(p.TypeName),
		buildYear:         p.BuildYear,
		price:             p.Price,
	}, nil
}

func validateCoverPictureIndex(index, pictureCount int) (int, error) {
	if pictureCount == 0 {
		return 0, nil
	}
	if index < 0 || index >= pictureCount {
		return 0, newValidationError("coverPictureIndex", "must refer to an existing picture")
	}
	return index, nil
}

func validatePictureURLs(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(in))
	for i, raw := range in {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, newValidationError("pictures", "entry "+itoa(i)+" is not a valid absolute URL")
		}
		out = append(out, raw)
	}
	return out, nil
}

// itoa avoids importing strconv for a single use case in validation.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// --- Getters (the aggregate exposes no setters; mutation is through methods) ---

func (g *Guitar) ID() string           { return g.id }
func (g *Guitar) SerialNumber() string { return g.serialNumber }
func (g *Guitar) Pictures() []string {
	out := make([]string, len(g.pictures))
	copy(out, g.pictures)
	return out
}
func (g *Guitar) CoverPictureIndex() int { return g.coverPictureIndex }
func (g *Guitar) Description() string    { return g.description }
func (g *Guitar) Brand() string       { return g.brand }
func (g *Guitar) TypeName() string    { return g.typeName }
func (g *Guitar) BuildYear() int      { return g.buildYear }
func (g *Guitar) Price() Money        { return g.price }

// Update replaces all mutable details of the guitar in a single atomic step.
// Identity (ID) cannot be changed. The same invariants as for construction
// apply, so callers receive a *ValidationError if any invariant is violated.
func (g *Guitar) Update(p GuitarProps) error {
	p.ID = g.id
	updated, err := NewGuitar(p)
	if err != nil {
		return err
	}
	*g = *updated
	return nil
}
