package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

func mustGuitar(t *testing.T, id, brand string) *domain.Guitar {
	t.Helper()
	price, err := domain.NewMoney(100000, domain.EUR)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID:        id,
		Brand:     brand,
		TypeName:  "Stratocaster",
		BuildYear: 2000,
		Price:     price,
	})
	if err != nil {
		t.Fatalf("guitar: %v", err)
	}
	return g
}

func TestMemoryRepository_SaveAndFind(t *testing.T) {
	r := NewMemoryRepository()
	g := mustGuitar(t, "g-1", "Fender")
	if err := r.Save(context.Background(), g); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := r.FindByID(context.Background(), "g-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Brand() != "Fender" {
		t.Errorf("got brand %s", got.Brand())
	}
}

func TestMemoryRepository_FindByID_NotFound(t *testing.T) {
	r := NewMemoryRepository()
	_, err := r.FindByID(context.Background(), "nope")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestMemoryRepository_FindAll_SortedByID(t *testing.T) {
	r := NewMemoryRepository()
	_ = r.Save(context.Background(), mustGuitar(t, "g-2", "Gibson"))
	_ = r.Save(context.Background(), mustGuitar(t, "g-1", "Fender"))
	all, err := r.FindAll(context.Background())
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(all) != 2 || all[0].ID() != "g-1" || all[1].ID() != "g-2" {
		t.Errorf("expected stable order g-1, g-2; got %v", []string{all[0].ID(), all[1].ID()})
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	r := NewMemoryRepository()
	_ = r.Save(context.Background(), mustGuitar(t, "g-1", "Fender"))
	if err := r.Delete(context.Background(), "g-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := r.FindByID(context.Background(), "g-1"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("guitar should be gone, got %v", err)
	}
}

func TestMemoryRepository_Delete_NotFound(t *testing.T) {
	r := NewMemoryRepository()
	if err := r.Delete(context.Background(), "nope"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("expected ErrGuitarNotFound, got %v", err)
	}
}
