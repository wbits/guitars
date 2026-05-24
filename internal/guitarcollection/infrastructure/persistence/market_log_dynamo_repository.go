package persistence

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wbits/guitars/internal/guitarcollection/domain"
)

// MarketLogDynamoAPI is the DynamoDB client surface used by MarketLogDynamoRepository.
type MarketLogDynamoAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// MarketLogDynamoRepository implements domain.MarketLogRepository on DynamoDB.
//
// Table schema:
//   - PK: id
//   - GSI guitarIdIndex: guitarId (HASH) + observedAt (RANGE)
type MarketLogDynamoRepository struct {
	client MarketLogDynamoAPI
	table  string
}

// NewMarketLogDynamoRepository constructs a repository for the given table.
func NewMarketLogDynamoRepository(client MarketLogDynamoAPI, table string) *MarketLogDynamoRepository {
	return &MarketLogDynamoRepository{client: client, table: table}
}

type marketLogItem struct {
	ID                string `dynamodbav:"id"`
	GuitarID          string `dynamodbav:"guitarId"`
	ObservedAt        string `dynamodbav:"observedAt"`
	Source            string `dynamodbav:"source"`
	Action            string `dynamodbav:"action"`
	PriceAmount       int64  `dynamodbav:"priceAmount"`
	PriceCurrency     string `dynamodbav:"priceCurrency"`
	ListingURL        string `dynamodbav:"listingUrl,omitempty"`
	ListingTitle      string `dynamodbav:"listingTitle,omitempty"`
	ExternalListingID string `dynamodbav:"externalListingId,omitempty"`
	ListingImageURL   string `dynamodbav:"listingImageUrl,omitempty"`
}

func marketLogToItem(log *domain.MarketLog) marketLogItem {
	return marketLogItem{
		ID:                log.ID(),
		GuitarID:          log.GuitarID(),
		ObservedAt:        log.ObservedAt().UTC().Format(timeRFC3339Nano),
		Source:            string(log.Source()),
		Action:            string(log.Action()),
		PriceAmount:       log.Price().Amount(),
		PriceCurrency:     string(log.Price().Currency()),
		ListingURL:        log.ListingURL(),
		ListingTitle:      log.ListingTitle(),
		ExternalListingID: log.ExternalListingID(),
		ListingImageURL:   log.ListingImageURL(),
	}
}

func (i marketLogItem) toDomain() (*domain.MarketLog, error) {
	observedAt, err := parseTimeRFC3339(i.ObservedAt)
	if err != nil {
		return nil, fmt.Errorf("corrupt observedAt for market log %s: %w", i.ID, err)
	}
	price, err := domain.NewMoney(i.PriceAmount, domain.Currency(i.PriceCurrency))
	if err != nil {
		return nil, fmt.Errorf("corrupt price for market log %s: %w", i.ID, err)
	}
	return domain.NewMarketLog(domain.MarketLogProps{
		ID:                i.ID,
		GuitarID:          i.GuitarID,
		ObservedAt:        observedAt,
		Source:            domain.MarketSource(i.Source),
		Action:            domain.MarketAction(i.Action),
		Price:             price,
		ListingURL:        i.ListingURL,
		ListingTitle:      i.ListingTitle,
		ExternalListingID: i.ExternalListingID,
		ListingImageURL:   i.ListingImageURL,
	})
}

// Save persists the market log (upsert).
func (r *MarketLogDynamoRepository) Save(ctx context.Context, log *domain.MarketLog) error {
	av, err := attributevalue.MarshalMap(marketLogToItem(log))
	if err != nil {
		return fmt.Errorf("marshal market log: %w", err)
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

// FindByGuitarID returns logs for a guitar via the guitarIdIndex GSI.
func (r *MarketLogDynamoRepository) FindByGuitarID(ctx context.Context, guitarID string) ([]*domain.MarketLog, error) {
	var items []marketLogItem
	var startKey map[string]ddbtypes.AttributeValue
	for {
		out, err := r.client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(r.table),
			IndexName:              aws.String("guitarIdIndex"),
			KeyConditionExpression: aws.String("guitarId = :gid"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":gid": &ddbtypes.AttributeValueMemberS{Value: guitarID},
			},
			ScanIndexForward:  aws.Bool(false),
			ExclusiveStartKey: startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("dynamodb Query: %w", err)
		}
		var batch []marketLogItem
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshal query page: %w", err)
		}
		items = append(items, batch...)
		if len(out.LastEvaluatedKey) == 0 {
			break
		}
		startKey = out.LastEvaluatedKey
	}
	logs := make([]*domain.MarketLog, 0, len(items))
	for _, it := range items {
		log, err := it.toDomain()
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ObservedAt().After(logs[j].ObservedAt())
	})
	return logs, nil
}

var _ domain.MarketLogRepository = (*MarketLogDynamoRepository)(nil)
