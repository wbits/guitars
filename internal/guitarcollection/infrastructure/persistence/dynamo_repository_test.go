package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// fakeDynamo is a hand-rolled in-memory stand-in for the DynamoDB client.
// It implements just enough of DynamoAPI to exercise the repository.
type fakeDynamo struct {
	items map[string]map[string]ddbtypes.AttributeValue
}

func newFakeDynamo() *fakeDynamo {
	return &fakeDynamo{items: map[string]map[string]ddbtypes.AttributeValue{}}
}

func (f *fakeDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	id := in.Item["id"].(*ddbtypes.AttributeValueMemberS).Value
	f.items[id] = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamo) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	id := in.Key["id"].(*ddbtypes.AttributeValueMemberS).Value
	item, ok := f.items[id]
	if !ok {
		return &dynamodb.GetItemOutput{}, nil
	}
	return &dynamodb.GetItemOutput{Item: item}, nil
}

func (f *fakeDynamo) Scan(_ context.Context, _ *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	out := make([]map[string]ddbtypes.AttributeValue, 0, len(f.items))
	for _, v := range f.items {
		out = append(out, v)
	}
	return &dynamodb.ScanOutput{Items: out}, nil
}

func (f *fakeDynamo) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	id := in.Key["id"].(*ddbtypes.AttributeValueMemberS).Value
	if _, ok := f.items[id]; !ok {
		return nil, &ddbtypes.ConditionalCheckFailedException{}
	}
	delete(f.items, id)
	return &dynamodb.DeleteItemOutput{}, nil
}

func buildGuitar(t *testing.T, id string) *domain.Guitar {
	t.Helper()
	price, err := domain.NewMoney(199900, domain.EUR)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	g, err := domain.NewGuitar(domain.GuitarProps{
		ID:           id,
		SerialNumber: "SN-" + id,
		Pictures:     []string{"https://example.com/" + id + ".jpg"},
		Description:  "test",
		Brand:        "Fender",
		TypeName:     "Stratocaster",
		BuildYear:    1996,
		Price:        price,
	})
	if err != nil {
		t.Fatalf("guitar: %v", err)
	}
	return g
}

func TestDynamoRepository_SaveAndFind(t *testing.T) {
	fd := newFakeDynamo()
	repo := NewDynamoRepository(fd, "Guitars")

	g := buildGuitar(t, "g-1")
	if err := repo.Save(context.Background(), g); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := repo.FindByID(context.Background(), "g-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.ID() != "g-1" || got.Brand() != "Fender" || got.SerialNumber() != "SN-g-1" {
		t.Errorf("guitar not faithfully roundtripped: %+v", got)
	}
	if got.Price().Amount() != 199900 || got.Price().Currency() != domain.EUR {
		t.Errorf("price not roundtripped: %+v", got.Price())
	}
}

func TestDynamoRepository_FindByID_NotFound(t *testing.T) {
	fd := newFakeDynamo()
	repo := NewDynamoRepository(fd, "Guitars")
	_, err := repo.FindByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("expected ErrGuitarNotFound, got %v", err)
	}
}

func TestDynamoRepository_FindAll(t *testing.T) {
	fd := newFakeDynamo()
	repo := NewDynamoRepository(fd, "Guitars")
	_ = repo.Save(context.Background(), buildGuitar(t, "g-2"))
	_ = repo.Save(context.Background(), buildGuitar(t, "g-1"))
	all, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("findall: %v", err)
	}
	if len(all) != 2 || all[0].ID() != "g-1" || all[1].ID() != "g-2" {
		t.Errorf("unexpected order/count: %v", []string{all[0].ID(), all[1].ID()})
	}
}

func TestDynamoRepository_Delete(t *testing.T) {
	fd := newFakeDynamo()
	repo := NewDynamoRepository(fd, "Guitars")
	_ = repo.Save(context.Background(), buildGuitar(t, "g-1"))
	if err := repo.Delete(context.Background(), "g-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), "g-1"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("guitar should be gone, got %v", err)
	}
}

func TestDynamoRepository_Delete_NotFound(t *testing.T) {
	fd := newFakeDynamo()
	repo := NewDynamoRepository(fd, "Guitars")
	if err := repo.Delete(context.Background(), "missing"); !errors.Is(err, domain.ErrGuitarNotFound) {
		t.Errorf("expected ErrGuitarNotFound, got %v", err)
	}
}

// Sanity-check that attributevalue tags are wired up correctly.
func TestGuitarItem_RoundTripsThroughAttributeValue(t *testing.T) {
	g := buildGuitar(t, "g-1")
	av, err := attributevalue.MarshalMap(toItem(g))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var item guitarItem
	if err := attributevalue.UnmarshalMap(av, &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if item.ID != g.ID() || item.Brand != g.Brand() || item.PriceAmount != g.Price().Amount() {
		t.Errorf("roundtrip mismatch: %+v", item)
	}
}
