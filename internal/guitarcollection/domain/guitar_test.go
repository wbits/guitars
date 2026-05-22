package domain

import (
	"testing"
	"time"
)

func validProps(t *testing.T) GuitarProps {
	t.Helper()
	price, err := NewMoney(199900, EUR)
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}
	return GuitarProps{
		ID:           "11111111-1111-1111-1111-111111111111",
		SerialNumber: "SN-12345",
		Pictures:     []string{"https://example.com/front.jpg", "https://example.com/back.jpg"},
		Description:  "A lovely sunburst",
		Brand:        "Fender",
		TypeName:     "Stratocaster",
		BuildYear:    1996,
		Price:        price,
	}
}

func TestNewGuitar_HappyPath(t *testing.T) {
	g, err := NewGuitar(validProps(t))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if g.ID() == "" {
		t.Errorf("id should be set")
	}
	if g.Brand() != "Fender" {
		t.Errorf("brand: want Fender, got %s", g.Brand())
	}
	if g.TypeName() != "Stratocaster" {
		t.Errorf("typeName: want Stratocaster, got %s", g.TypeName())
	}
	if len(g.Pictures()) != 2 {
		t.Errorf("pictures: want 2, got %d", len(g.Pictures()))
	}
}

func TestNewGuitar_AcceptsMissingSerialNumber(t *testing.T) {
	p := validProps(t)
	p.SerialNumber = ""
	g, err := NewGuitar(p)
	if err != nil {
		t.Fatalf("serial number must be optional, got %v", err)
	}
	if g.SerialNumber() != "" {
		t.Errorf("serial number should be empty")
	}
}

func TestNewGuitar_AcceptsEmptyPictures(t *testing.T) {
	p := validProps(t)
	p.Pictures = nil
	if _, err := NewGuitar(p); err != nil {
		t.Fatalf("pictures must be optional, got %v", err)
	}
}

func TestNewGuitar_RejectsMissingID(t *testing.T) {
	p := validProps(t)
	p.ID = "  "
	_, err := NewGuitar(p)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError for missing id, got %v", err)
	}
}

func TestNewGuitar_RejectsMissingBrand(t *testing.T) {
	p := validProps(t)
	p.Brand = ""
	_, err := NewGuitar(p)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError for missing brand, got %v", err)
	}
}

func TestNewGuitar_RejectsMissingTypeName(t *testing.T) {
	p := validProps(t)
	p.TypeName = ""
	_, err := NewGuitar(p)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError for missing type name, got %v", err)
	}
}

func TestNewGuitar_RejectsImplausibleBuildYear(t *testing.T) {
	p := validProps(t)
	p.BuildYear = 1700
	if _, err := NewGuitar(p); !IsValidationError(err) {
		t.Errorf("expected validation error for too-old year, got %v", err)
	}
	p.BuildYear = time.Now().UTC().Year() + 5
	if _, err := NewGuitar(p); !IsValidationError(err) {
		t.Errorf("expected validation error for far-future year, got %v", err)
	}
}

func TestNewGuitar_RejectsZeroPrice(t *testing.T) {
	p := validProps(t)
	p.Price = Money{}
	_, err := NewGuitar(p)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError for missing price, got %v", err)
	}
}

func TestNewGuitar_RejectsInvalidPictureURL(t *testing.T) {
	p := validProps(t)
	p.Pictures = []string{"not-a-url"}
	_, err := NewGuitar(p)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError for bad URL, got %v", err)
	}
}

func TestGuitar_PicturesIsDefensiveCopy(t *testing.T) {
	g, err := NewGuitar(validProps(t))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	pics := g.Pictures()
	pics[0] = "https://evil.example.com/oops.jpg"
	if g.Pictures()[0] == "https://evil.example.com/oops.jpg" {
		t.Errorf("Pictures() must return a defensive copy")
	}
}

func TestGuitar_Update_PreservesID(t *testing.T) {
	g, err := NewGuitar(validProps(t))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	originalID := g.ID()
	newProps := validProps(t)
	newProps.ID = "different-id"
	newProps.Brand = "Gibson"
	if err := g.Update(newProps); err != nil {
		t.Fatalf("update should succeed: %v", err)
	}
	if g.ID() != originalID {
		t.Errorf("id must not change on update; want %q, got %q", originalID, g.ID())
	}
	if g.Brand() != "Gibson" {
		t.Errorf("brand should have changed to Gibson")
	}
}

func TestGuitar_Update_PropagatesValidationErrors(t *testing.T) {
	g, _ := NewGuitar(validProps(t))
	bad := validProps(t)
	bad.Brand = ""
	if err := g.Update(bad); !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if g.Brand() != "Fender" {
		t.Errorf("aggregate must not be mutated on a failed update; brand = %s", g.Brand())
	}
}
