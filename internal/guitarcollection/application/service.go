package application

import (
	"context"

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
func (s *Service) AddGuitar(ctx context.Context, in GuitarInput) (*domain.Guitar, error) {
	price, err := domain.NewMoney(in.PriceAmount, domain.Currency(in.PriceCurrency))
	if err != nil {
		return nil, err
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID:                s.ids.NewID(),
		SerialNumber:      in.SerialNumber,
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
// when the input violates an invariant.
func (s *Service) UpdateGuitar(ctx context.Context, id string, in GuitarInput) (*domain.Guitar, error) {
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	price, err := domain.NewMoney(in.PriceAmount, domain.Currency(in.PriceCurrency))
	if err != nil {
		return nil, err
	}
	if err := g.Update(domain.GuitarProps{
		SerialNumber:      in.SerialNumber,
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

// GetGuitar returns a single guitar by id.
func (s *Service) GetGuitar(ctx context.Context, id string) (*domain.Guitar, error) {
	return s.repo.FindByID(ctx, id)
}

// ListGuitars returns all guitars in the collection.
func (s *Service) ListGuitars(ctx context.Context) ([]*domain.Guitar, error) {
	return s.repo.FindAll(ctx)
}

// DeleteGuitar removes the guitar with the given id.
func (s *Service) DeleteGuitar(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
