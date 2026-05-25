package application

import (
	"context"
	"strings"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// IDGenerator abstracts id creation so that production uses UUIDv4 while tests
// can inject a deterministic generator.
type IDGenerator interface {
	NewID() string
}

// Service is the application-level use-case layer for the GuitarCollection
// bounded context. It coordinates the domain aggregate and the repository
// port; it does not contain business rules itself.
type Service struct {
	repo domain.Repository
	ids  IDGenerator
}

// NewService wires the application service with its required collaborators.
func NewService(repo domain.Repository, ids IDGenerator) *Service {
	return &Service{repo: repo, ids: ids}
}

// AddGuitar creates a new guitar from the given input, persists it and returns
// the resulting aggregate. The id is assigned by the supplied IDGenerator.
func (s *Service) AddGuitar(ctx context.Context, ownerID string, in GuitarInput) (*domain.Guitar, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domain.InvalidField("owner", "is required")
	}
	price, err := domain.NewMoney(in.PriceAmount, domain.Currency(in.PriceCurrency))
	if err != nil {
		return nil, err
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID:                s.ids.NewID(),
		Owner:             ownerID,
		SerialNumber:      in.SerialNumber,
		Color:             in.Color,
		Country:           in.Country,
		Factory:           in.Factory,
		Pictures:          in.Pictures,
		CoverPictureIndex: in.CoverPictureIndex,
		Description:       in.Description,
		Brand:             in.Brand,
		TypeName:          in.TypeName,
		BuildYear:         in.BuildYear,
		Price:             price,
	})
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// UpdateGuitar applies the given input to the guitar identified by id. Returns
// domain.ErrGuitarNotFound when no such guitar exists, or a ValidationError
// when the input violates an invariant. Guitars without an owner are assigned
// to the caller on update.
func (s *Service) UpdateGuitar(ctx context.Context, ownerID, id string, in GuitarInput) (*domain.Guitar, error) {
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !guitarWritableBy(g, ownerID) {
		return nil, domain.ErrGuitarNotFound
	}
	price, err := domain.NewMoney(in.PriceAmount, domain.Currency(in.PriceCurrency))
	if err != nil {
		return nil, err
	}
	if err := g.Update(domain.GuitarProps{
		Owner:             resolveOwnerForUpdate(g, ownerID),
		SerialNumber:      in.SerialNumber,
		Color:             in.Color,
		Country:           in.Country,
		Factory:           in.Factory,
		Pictures:          in.Pictures,
		CoverPictureIndex: in.CoverPictureIndex,
		Description:       in.Description,
		Brand:             in.Brand,
		TypeName:          in.TypeName,
		BuildYear:         in.BuildYear,
		Price:             price,
	}); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// GetGuitar returns a single guitar by id when visible to the caller.
func (s *Service) GetGuitar(ctx context.Context, ownerID, id string) (*domain.Guitar, error) {
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !guitarReadableBy(g, ownerID) {
		return nil, domain.ErrGuitarNotFound
	}
	return g, nil
}

// ListGuitars returns guitars owned by the caller.
func (s *Service) ListGuitars(ctx context.Context, ownerID string) ([]*domain.Guitar, error) {
	return s.repo.FindByOwner(ctx, strings.TrimSpace(ownerID))
}

// ListUserGuitars returns guitars owned by the given user id.
func (s *Service) ListUserGuitars(ctx context.Context, userID string) ([]*domain.Guitar, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.InvalidField("userId", "is required")
	}
	return s.repo.FindByOwner(ctx, userID)
}

// ListCollectionOwners returns every user id that owns at least one guitar.
func (s *Service) ListCollectionOwners(ctx context.Context) ([]string, error) {
	return s.repo.FindDistinctOwners(ctx)
}

// DeleteGuitar removes the guitar with the given id when the caller may modify it.
func (s *Service) DeleteGuitar(ctx context.Context, ownerID, id string) error {
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !guitarWritableBy(g, ownerID) {
		return domain.ErrGuitarNotFound
	}
	return s.repo.Delete(ctx, id)
}
