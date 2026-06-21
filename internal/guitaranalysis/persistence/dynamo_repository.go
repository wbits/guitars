package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wbits/guitars/internal/guitaranalysis"
)

type DynamoAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type DynamoRepository struct {
	client DynamoAPI
	table  string
}

func NewDynamoRepository(client DynamoAPI, table string) *DynamoRepository {
	return &DynamoRepository{client: client, table: table}
}

type analysisItem struct {
	GuitarID            string   `dynamodbav:"guitarId"`
	OwnerID             string   `dynamodbav:"ownerId"`
	Status              string   `dynamodbav:"status"`
	PicturesFingerprint string   `dynamodbav:"picturesFingerprint,omitempty"`
	VisualSummary       string   `dynamodbav:"visualSummary,omitempty"`
	Tags                []string `dynamodbav:"tags,omitempty"`
	Confidence          float64  `dynamodbav:"confidence,omitempty"`
	FailureReason       string   `dynamodbav:"failureReason,omitempty"`
	AnalyzedAt          string   `dynamodbav:"analyzedAt,omitempty"`
	UpdatedAt           string   `dynamodbav:"updatedAt"`
}

func toItem(rec *guitaranalysis.Record) analysisItem {
	item := analysisItem{
		GuitarID:            rec.GuitarID(),
		OwnerID:             rec.OwnerID(),
		Status:              rec.Status(),
		PicturesFingerprint: rec.PicturesFingerprint(),
		VisualSummary:       rec.VisualSummary(),
		Tags:                rec.Tags(),
		Confidence:          rec.Confidence(),
		FailureReason:       rec.FailureReason(),
		UpdatedAt:           rec.UpdatedAt().UTC().Format(time.RFC3339),
	}
	if !rec.AnalyzedAt().IsZero() {
		item.AnalyzedAt = rec.AnalyzedAt().UTC().Format(time.RFC3339)
	}
	return item
}

func (item analysisItem) toDomain() (*guitaranalysis.Record, error) {
	rec, err := guitaranalysis.NewRecord(item.GuitarID, item.OwnerID, item.Status, item.PicturesFingerprint)
	if err != nil {
		return nil, err
	}
	switch item.Status {
	case guitaranalysis.StatusReady:
		rec.SetReady(item.PicturesFingerprint, item.VisualSummary, item.Tags, item.Confidence)
	case guitaranalysis.StatusFailed:
		rec.SetFailed(item.PicturesFingerprint, item.FailureReason)
	default:
		rec.SetPending(item.PicturesFingerprint)
	}
	return rec, nil
}

func (r *DynamoRepository) FindByGuitarID(ctx context.Context, guitarID string) (*guitaranalysis.Record, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"guitarId": &ddbtypes.AttributeValueMemberS{Value: guitarID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var item analysisItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, err
	}
	return item.toDomain()
}

func (r *DynamoRepository) FindByGuitarIDs(ctx context.Context, guitarIDs []string) (map[string]*guitaranalysis.Record, error) {
	result := make(map[string]*guitaranalysis.Record, len(guitarIDs))
	if len(guitarIDs) == 0 {
		return result, nil
	}
	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(guitarIDs))
	for _, id := range guitarIDs {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"guitarId": &ddbtypes.AttributeValueMemberS{Value: id},
		})
	}
	out, err := r.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			r.table: {Keys: keys},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb BatchGetItem: %w", err)
	}
	for _, raw := range out.Responses[r.table] {
		var item analysisItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, err
		}
		rec, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		result[rec.GuitarID()] = rec
	}
	return result, nil
}

func (r *DynamoRepository) Save(ctx context.Context, record *guitaranalysis.Record) error {
	av, err := attributevalue.MarshalMap(toItem(record))
	if err != nil {
		return err
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

func (r *DynamoRepository) DeleteByGuitarID(ctx context.Context, guitarID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"guitarId": &ddbtypes.AttributeValueMemberS{Value: guitarID},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb DeleteItem: %w", err)
	}
	return nil
}
