package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wbits/guitars/internal/userprofile/domain"
)

// DynamoAPI is the subset of DynamoDB used by the profile repository.
type DynamoAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoRepository implements domain.Repository on DynamoDB.
type DynamoRepository struct {
	client    DynamoAPI
	table     string
	usernameIndex string
}

// NewDynamoRepository constructs a DynamoRepository.
func NewDynamoRepository(client DynamoAPI, table, usernameIndex string) *DynamoRepository {
	if usernameIndex == "" {
		usernameIndex = "usernameIndex"
	}
	return &DynamoRepository{client: client, table: table, usernameIndex: usernameIndex}
}

type profileItem struct {
	UserID   string `dynamodbav:"userId"`
	Username string `dynamodbav:"username,omitempty"`
	Email    string `dynamodbav:"email,omitempty"`
}

func toItem(profile *domain.Profile) profileItem {
	return profileItem{
		UserID:   profile.UserID(),
		Username: profile.Username(),
		Email:    profile.Email(),
	}
}

func (item profileItem) toDomain() (*domain.Profile, error) {
	return domain.NewProfile(domain.ProfileProps{
		UserID:   item.UserID,
		Username: item.Username,
		Email:    item.Email,
	})
}

// FindByUserID implements domain.Repository.
func (r *DynamoRepository) FindByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"userId": &ddbtypes.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	var item profileItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}
	return item.toDomain()
}

// FindByUsername implements domain.Repository.
func (r *DynamoRepository) FindByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String(r.usernameIndex),
		KeyConditionExpression: aws.String("#username = :username"),
		ExpressionAttributeNames: map[string]string{
			"#username": "username",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":username": &ddbtypes.AttributeValueMemberS{Value: username},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb Query: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, nil
	}
	var item profileItem
	if err := attributevalue.UnmarshalMap(out.Items[0], &item); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}
	return item.toDomain()
}

// FindByUserIDs implements domain.Repository.
func (r *DynamoRepository) FindByUserIDs(ctx context.Context, userIDs []string) (map[string]*domain.Profile, error) {
	out := make(map[string]*domain.Profile, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"userId": &ddbtypes.AttributeValueMemberS{Value: userID},
		})
	}
	resp, err := r.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			r.table: {Keys: keys},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb BatchGetItem: %w", err)
	}
	items := resp.Responses[r.table]
	for _, raw := range items {
		var item profileItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return nil, fmt.Errorf("unmarshal profile: %w", err)
		}
		profile, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		out[profile.UserID()] = profile
	}
	return out, nil
}

// Save implements domain.Repository.
func (r *DynamoRepository) Save(ctx context.Context, profile *domain.Profile) error {
	av, err := attributevalue.MarshalMap(toItem(profile))
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	})
	if err != nil {
		var cfe *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return domain.ErrUsernameTaken
		}
		return fmt.Errorf("dynamodb PutItem: %w", err)
	}
	return nil
}
