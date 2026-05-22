package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// --- test doubles ----------------------------------------------------------

type fakeRepo struct {
	mu      sync.Mutex
	guitars map[string]*domain.Guitar
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{guitars: map[string]*domain.Guitar{}}
}

func (r *fakeRepo) Save(_ context.Context, g *domain.Guitar) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.guitars[g.ID()] = g
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*domain.Guitar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.guitars[id]
	if !ok {
		return nil, domain.ErrGuitarNotFound
	}
	return g, nil
}

func (r *fakeRepo) FindAll(_ context.Context) ([]*domain.Guitar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.Guitar, 0, len(r.guitars))
	for _, g := range r.guitars {
		out = append(out, g)
	}
	return out, nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.guitars[id]; !ok {
		return domain.ErrGuitarNotFound
	}
	delete(r.guitars, id)
	return nil
}

type fixedIDs struct {
	ids []string
	i   int
}

func (f *fixedIDs) NewID() string {
	id := f.ids[f.i]
	f.i++
	return id
}

func validInput() GuitarInput {
	return GuitarInput{
		SerialNumber:  "SN-1",
		Pictures:      []string{"https://example.com/a.jpg"},
		Description:   "1996 sunburst",
		Brand:         "Fender",
		TypeName:      "Stratocaster",
		BuildYear:     1996,
		PriceAmount:   199900,
		PriceCurrency: "EUR",
	}
}

// --- tests -----------------------------------------------------------------

func TestService_AddGuitar_PersistsAndAssignsID(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	g, err := svc.AddGuitar(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ID() != "guitar-1" {
		t.Errorf("expected id guitar-1, got %s", g.ID())
	}
	stored, err := repo.FindByID(context.Background(), "guitar-1")
	if err != nil || stored == nil {
		t.Fatalf("guitar not persisted: %v", err)
	}
}

func TestService_AddGuitar_PropagatesValidationError(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	in := validInput()
	in.Brand = ""
	_, err := svc.AddGuitar(context.Background(), in)
	if !domain.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestService_AddGuitar_PropagatesPriceValidationError(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	in := validInput()
	in.PriceCurrency = "GBP"
	_, err := svc.AddGuitar(context.Background(), in)
	if !domain.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestService_GetGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	_, err := svc.GetGuitar(context.Background(), "missing")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_UpdateGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	_, err := svc.UpdateGuitar(context.Background(), "missing", validInput())
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_UpdateGuitar_Persists(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	if _, err := svc.AddGuitar(context.Background(), validInput()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	in := validInput()
	in.Brand = "Gibson"
	in.TypeName = "Les Paul"
	g, err := svc.UpdateGuitar(context.Background(), "guitar-1", in)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if g.Brand() != "Gibson" || g.TypeName() != "Les Paul" {
		t.Errorf("brand/type not updated: %s / %s", g.Brand(), g.TypeName())
	}
}

func TestService_ListGuitars(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1", "guitar-2"}})

	if _, err := svc.AddGuitar(context.Background(), validInput()); err != nil {
		t.Fatalf("seed 1 failed: %v", err)
	}
	if _, err := svc.AddGuitar(context.Background(), validInput()); err != nil {
		t.Fatalf("seed 2 failed: %v", err)
	}
	all, err := svc.ListGuitars(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 guitars, got %d", len(all))
	}
}

func TestService_DeleteGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	err := svc.DeleteGuitar(context.Background(), "missing")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_DeleteGuitar_Removes(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	if _, err := svc.AddGuitar(context.Background(), validInput()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if err := svc.DeleteGuitar(context.Background(), "guitar-1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := svc.GetGuitar(context.Background(), "guitar-1"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("guitar should be gone, got err=%v", err)
	}
}
