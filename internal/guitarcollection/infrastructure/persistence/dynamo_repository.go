package persistence

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// DynamoAPI is the subset of the DynamoDB client that the repository uses.
// Defining it as an interface keeps DynamoRepository unit-testable without
// requiring an actual DynamoDB or LocalStack instance.
type DynamoAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// DynamoRepository implements domain.Repository on top of DynamoDB.
//
// The table schema is intentionally simple: id is the partition key and there
// is no sort key. All other fields are top-level attributes on the item.
type DynamoRepository struct {
	client DynamoAPI
	table  string
}

// NewDynamoRepository constructs a DynamoRepository using the supplied client
// and the table name to operate on.
func NewDynamoRepository(client DynamoAPI, table string) *DynamoRepository {
	return &DynamoRepository{client: client, table: table}
}

// guitarItem is the on-the-wire DynamoDB representation of a Guitar. Keeping
// this type local to the persistence package ensures the domain layer stays
// free of DynamoDB tags and concerns.
type guitarItem struct {
	ID                string   `dynamodbav:"id"`
	SerialNumber      string   `dynamodbav:"serialNumber,omitempty"`
	Pictures          []string `dynamodbav:"pictures,omitempty"`
	CoverPictureIndex int      `dynamodbav:"coverPictureIndex,omitempty"`
	Description       string   `dynamodbav:"description,omitempty"`
	Brand             string   `dynamodbav:"brand"`
	TypeName          string   `dynamodbav:"typeName"`
	BuildYear         int      `dynamodbav:"buildYear"`
	PriceAmount       int64    `dynamodbav:"priceAmount"`
	PriceCurrency     string   `dynamodbav:"priceCurrency"`
}

func toItem(g *domain.Guitar) guitarItem {
	return guitarItem{
		ID:                g.ID(),
		SerialNumber:      g.SerialNumber(),
		Pictures:          g.Pictures(),
		CoverPictureIndex: g.CoverPictureIndex(),
		Description:       g.Description(),
		Brand:             g.Brand(),
		TypeName:          g.TypeName(),
		BuildYear:         g.BuildYear(),
		PriceAmount:       g.Price().Amount(),
		PriceCurrency:     string(g.Price().Currency()),
	}
}

func (i guitarItem) toDomain() (*domain.Guitar, error) {
	price, err := domain.NewMoney(i.PriceAmount, domain.Currency(i.PriceCurrency))
	if err != nil {
		return nil, fmt.Errorf("corrupt price for guitar %s: %w", i.ID, err)
	}
	return domain.NewGuitar(domain.GuitarProps{
		ID:                i.ID,
		SerialNumber:      i.SerialNumber,
		Pictures:          i.Pictures,
		CoverPictureIndex: i.CoverPictureIndex,
		Description:       i.Description,
		Brand:             i.Brand,
		TypeName:          i.TypeName,
		BuildYear:         i.BuildYear,
		Price:             price,
	})
}

// Save persists the guitar (upsert).
func (r *DynamoRepository) Save(ctx context.Context, g *domain.Guitar) error {
	av, err := attributevalue.MarshalMap(toItem(g))
	if err != nil {
		return fmt.Errorf("marshal guitar: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("dynamodb PutItem: %w", err)
	}
	return nil
}

// FindByID looks up a guitar by primary key.
func (r *DynamoRepository) FindByID(ctx context.Context, id string) (*domain.Guitar, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"id": &ddbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if len(out.Item) == 0 {
		return nil, domain.ErrGuitarNotFound
	}
	var item guitarItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal guitar: %w", err)
	}
	return item.toDomain()
}

// FindAll returns every guitar in the collection. Uses Scan with paging since
// a personal guitar collection is small.
func (r *DynamoRepository) FindAll(ctx context.Context) ([]*domain.Guitar, error) {
	var items []guitarItem
	var startKey map[string]ddbtypes.AttributeValue
	for {
		out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(r.table),
			ExclusiveStartKey: startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("dynamodb Scan: %w", err)
		}
		var batch []guitarItem
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshal scan page: %w", err)
		}
		items = append(items, batch...)
		if len(out.LastEvaluatedKey) == 0 {
			break
		}
		startKey = out.LastEvaluatedKey
	}
	guitars := make([]*domain.Guitar, 0, len(items))
	for _, it := range items {
		g, err := it.toDomain()
		if err != nil {
			return nil, err
		}
		guitars = append(guitars, g)
	}
	sort.Slice(guitars, func(i, j int) bool { return guitars[i].ID() < guitars[j].ID() })
	return guitars, nil
}

// Delete removes a guitar. Uses a conditional expression so we can return
// domain.ErrGuitarNotFound on a missing id.
func (r *DynamoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"id": &ddbtypes.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	})
	if err != nil {
		var cfe *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return domain.ErrGuitarNotFound
		}
		return fmt.Errorf("dynamodb DeleteItem: %w", err)
	}
	return nil
}
