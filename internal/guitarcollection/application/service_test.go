package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

const testOwner = "user-1"

func (r *fakeRepo) FindByOwner(_ context.Context, owner string) ([]*domain.Guitar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.Guitar, 0)
	for _, g := range r.guitars {
		if g.Owner() == owner {
			out = append(out, g)
		}
	}
	return out, nil
}

func (r *fakeRepo) FindDistinctOwners(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]struct{}{}
	for _, g := range r.guitars {
		if owner := g.Owner(); owner != "" {
			seen[owner] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for owner := range seen {
		out = append(out, owner)
	}
	return out, nil
}

func TestService_AddGuitar_PersistsAndAssignsID(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	g, err := svc.AddGuitar(context.Background(), testOwner, validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ID() != "guitar-1" {
		t.Errorf("expected id guitar-1, got %s", g.ID())
	}
	if g.Owner() != testOwner {
		t.Errorf("expected owner %s, got %s", testOwner, g.Owner())
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
	_, err := svc.AddGuitar(context.Background(), testOwner, in)
	if !domain.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestService_AddGuitar_PropagatesPriceValidationError(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	in := validInput()
	in.PriceCurrency = "GBP"
	_, err := svc.AddGuitar(context.Background(), testOwner, in)
	if !domain.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestService_GetGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	_, err := svc.GetGuitar(context.Background(), testOwner, "missing")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_GetGuitar_AllowsOtherOwners(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})
	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatal(err)
	}
	g, err := svc.GetGuitar(context.Background(), "other-user", "guitar-1")
	if err != nil {
		t.Fatalf("expected readable guitar, got %v", err)
	}
	if g.ID() != "guitar-1" {
		t.Fatalf("unexpected guitar: %+v", g)
	}
}

func TestService_UpdateGuitar_RejectsOtherOwners(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})
	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatal(err)
	}
	_, err := svc.UpdateGuitar(context.Background(), "other-user", "guitar-1", validInput())
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_UpdateGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	_, err := svc.UpdateGuitar(context.Background(), testOwner, "missing", validInput())
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_UpdateGuitar_Persists(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	in := validInput()
	in.Brand = "Gibson"
	in.TypeName = "Les Paul"
	g, err := svc.UpdateGuitar(context.Background(), testOwner, "guitar-1", in)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if g.Brand() != "Gibson" || g.TypeName() != "Les Paul" {
		t.Errorf("brand/type not updated: %s / %s", g.Brand(), g.TypeName())
	}
}

func TestService_UpdateGuitar_BackfillsOwner(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})
	price, _ := domain.NewMoney(199900, domain.EUR)
	legacy, err := domain.NewGuitar(domain.GuitarProps{
		ID: "guitar-1", Brand: "Fender", TypeName: "Stratocaster", BuildYear: 1996, Price: price,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), legacy); err != nil {
		t.Fatal(err)
	}

	g, err := svc.UpdateGuitar(context.Background(), testOwner, "guitar-1", validInput())
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if g.Owner() != testOwner {
		t.Fatalf("want owner backfilled to %s, got %s", testOwner, g.Owner())
	}
	listed, err := svc.ListGuitars(context.Background(), testOwner, true)
	if err != nil || len(listed) != 1 {
		t.Fatalf("want listed guitar after backfill, got %d err=%v", len(listed), err)
	}
}

func TestService_ListGuitars(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1", "guitar-2"}})

	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatalf("seed 1 failed: %v", err)
	}
	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatalf("seed 2 failed: %v", err)
	}
	all, err := svc.ListGuitars(context.Background(), testOwner, true)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 guitars, got %d", len(all))
	}
}

func TestService_ListGuitars_OmitsHiddenByDefault(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1", "guitar-2"}})

	visible, err := svc.AddGuitar(context.Background(), testOwner, validInput())
	if err != nil {
		t.Fatal(err)
	}
	hidden, err := svc.AddGuitar(context.Background(), testOwner, validInput())
	if err != nil {
		t.Fatal(err)
	}
	hidden.SetHiddenInCollection(true)
	if err := repo.Save(context.Background(), hidden); err != nil {
		t.Fatal(err)
	}

	listed, err := svc.ListGuitars(context.Background(), testOwner, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID() != visible.ID() {
		t.Fatalf("expected only visible guitar, got %+v", listed)
	}

	withHidden, err := svc.ListGuitars(context.Background(), testOwner, true)
	if err != nil || len(withHidden) != 2 {
		t.Fatalf("expected 2 guitars with includeHidden, got %d err=%v", len(withHidden), err)
	}
}

func TestService_SetGuitarHidden(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})
	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatal(err)
	}

	hidden, err := svc.SetGuitarHidden(context.Background(), testOwner, "guitar-1", true)
	if err != nil || !hidden.HiddenInCollection() {
		t.Fatalf("expected hidden guitar, got err=%v hidden=%v", err, hidden)
	}
	_, err = svc.GetGuitar(context.Background(), "other-user", "guitar-1")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("other users should not read hidden guitar, got %v", err)
	}
	if _, err := svc.GetGuitar(context.Background(), testOwner, "guitar-1"); err != nil {
		t.Fatalf("owner should read hidden guitar: %v", err)
	}

	shown, err := svc.SetGuitarHidden(context.Background(), testOwner, "guitar-1", false)
	if err != nil || shown.HiddenInCollection() {
		t.Fatalf("expected visible guitar, got err=%v hidden=%v", err, shown)
	}
}

func TestService_UpdateGuitar_PreservesHiddenFlag(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})
	g, err := svc.AddGuitar(context.Background(), testOwner, validInput())
	if err != nil {
		t.Fatal(err)
	}
	g.SetHiddenInCollection(true)
	if err := repo.Save(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	in := validInput()
	in.Brand = "Gibson"
	updated, err := svc.UpdateGuitar(context.Background(), testOwner, "guitar-1", in)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.HiddenInCollection() || updated.Brand() != "Gibson" {
		t.Fatalf("expected hidden Gibson, got hidden=%v brand=%s", updated.HiddenInCollection(), updated.Brand())
	}
}

func TestService_DeleteGuitar_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	err := svc.DeleteGuitar(context.Background(), testOwner, "missing")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Fatalf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestService_DeleteGuitar_Removes(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fixedIDs{ids: []string{"guitar-1"}})

	if _, err := svc.AddGuitar(context.Background(), testOwner, validInput()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if err := svc.DeleteGuitar(context.Background(), testOwner, "guitar-1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := svc.GetGuitar(context.Background(), testOwner, "guitar-1"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("guitar should be gone, got err=%v", err)
	}
}

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
