package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type whatsAppSessionDynamoClient interface {
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

type WhatsAppSessionRepo struct {
	client    whatsAppSessionDynamoClient
	tableName string
	now       func() time.Time
}

func NewWhatsAppSessionRepo(db *DynamoDB) *WhatsAppSessionRepo {
	return &WhatsAppSessionRepo{
		client:    db.Client,
		tableName: db.WhatsAppSessionsTable,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func newWhatsAppSessionRepoWithClient(client whatsAppSessionDynamoClient, tableName string) *WhatsAppSessionRepo {
	return &WhatsAppSessionRepo{
		client:    client,
		tableName: tableName,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (r *WhatsAppSessionRepo) TrackCustomerInbound(ctx context.Context, customerID, businessID string, inboundAt time.Time) error {
	return r.trackCustomerInbound(ctx, customerID, businessID, inboundAt, r.now())
}

func (r *WhatsAppSessionRepo) trackCustomerInbound(ctx context.Context, customerID, businessID string, inboundAt, updatedAt time.Time) error {
	if customerID == "" {
		return fmt.Errorf("customer id is required")
	}
	if businessID == "" {
		return fmt.Errorf("business id is required")
	}
	if inboundAt.IsZero() {
		return fmt.Errorf("inbound timestamp is required")
	}

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"customerId": &types.AttributeValueMemberS{Value: customerID},
			"businessId": &types.AttributeValueMemberS{Value: businessID},
		},
		UpdateExpression:    aws.String("SET lastCustomerInboundAt = :inboundAt, updatedAt = :updatedAt"),
		ConditionExpression: aws.String("attribute_not_exists(lastCustomerInboundAt) OR lastCustomerInboundAt < :inboundAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inboundAt": &types.AttributeValueMemberS{Value: inboundAt.UTC().Format(time.RFC3339Nano)},
			":updatedAt": &types.AttributeValueMemberS{Value: updatedAt.UTC().Format(time.RFC3339Nano)},
		},
	})
	if err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
			return nil
		}
		return fmt.Errorf("track whatsapp inbound session: %w", err)
	}

	return nil
}

func (r *WhatsAppSessionRepo) GetLastCustomerInbound(ctx context.Context, customerID, businessID string) (*time.Time, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customer id is required")
	}
	if businessID == "" {
		return nil, fmt.Errorf("business id is required")
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"customerId": &types.AttributeValueMemberS{Value: customerID},
			"businessId": &types.AttributeValueMemberS{Value: businessID},
		},
		ProjectionExpression: aws.String("lastCustomerInboundAt"),
	})
	if err != nil {
		return nil, fmt.Errorf("get whatsapp inbound session: %w", err)
	}
	if len(result.Item) == 0 {
		return nil, nil
	}

	raw, ok := result.Item["lastCustomerInboundAt"].(*types.AttributeValueMemberS)
	if !ok || raw.Value == "" {
		return nil, nil
	}

	parsed, err := parseRFC3339Timestamp(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("parse last inbound timestamp: %w", err)
	}

	parsed = parsed.UTC()
	return &parsed, nil
}

func parseRFC3339Timestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, nil
	}

	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, err
}
