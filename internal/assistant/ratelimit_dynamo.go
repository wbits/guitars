package assistant

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// DynamoRateLimiter enforces a per-user daily quota using a DynamoDB counter with TTL.
type DynamoRateLimiter struct {
	Client dynamoAPI
	Table  string
	Limit  int
}

type usageItem struct {
	PK        string `dynamodbav:"pk"`
	Count     int    `dynamodbav:"count"`
	ExpiresAt int64  `dynamodbav:"expiresAt"`
}

// Allow increments today's counter for userID or returns ErrRateLimited.
func (d *DynamoRateLimiter) Allow(ctx context.Context, userID string) error {
	if d == nil || d.Client == nil || d.Table == "" {
		return nil
	}
	limit := d.Limit
	if limit <= 0 {
		limit = 10
	}
	now := time.Now().UTC()
	pk := fmt.Sprintf("user#%s#%s", userID, utcDay(now))
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	out, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.Table),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return fmt.Errorf("rate limit read: %w", err)
	}

	count := 0
	if len(out.Item) > 0 {
		var item usageItem
		if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
			return fmt.Errorf("rate limit unmarshal: %w", err)
		}
		count = item.Count
	}
	if count >= limit {
		return fmt.Errorf("%w (%d per day)", ErrRateLimited, limit)
	}

	item, err := attributevalue.MarshalMap(usageItem{
		PK:        pk,
		Count:     count + 1,
		ExpiresAt: endOfDay.Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return err
	}
	_, err = d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.Table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("rate limit write: %w", err)
	}
	return nil
}

// ParseDailyLimit reads ASSISTANT_DAILY_LIMIT-style env values.
func ParseDailyLimit(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

// IsRateLimited reports whether err is a quota exhaustion.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}
