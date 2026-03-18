package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fiando/propti/backend/internal/models"
)

type UploadSessionRepo struct {
	db *DynamoDB
}

func NewUploadSessionRepo(db *DynamoDB) *UploadSessionRepo {
	return &UploadSessionRepo{db: db}
}

func (r *UploadSessionRepo) Put(ctx context.Context, session *models.UploadSession) error {
	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		return fmt.Errorf("marshal upload session: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.UploadSessionsTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put upload session: %w", err)
	}
	return nil
}

func (r *UploadSessionRepo) GetBySessionID(ctx context.Context, sessionID string) (*models.UploadSession, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.UploadSessionsTable),
		Key: map[string]types.AttributeValue{
			"sessionId": &types.AttributeValueMemberS{Value: sessionID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get upload session: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var session models.UploadSession
	if err := attributevalue.UnmarshalMap(result.Item, &session); err != nil {
		return nil, fmt.Errorf("unmarshal upload session: %w", err)
	}
	return &session, nil
}

func (r *UploadSessionRepo) Consume(ctx context.Context, sessionID, listingID string, consumedAt time.Time) error {
	_, err := r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.UploadSessionsTable),
		Key: map[string]types.AttributeValue{
			"sessionId": &types.AttributeValueMemberS{Value: sessionID},
		},
		UpdateExpression:    aws.String("SET consumedAt = :consumedAt, listingId = :listingId"),
		ConditionExpression: aws.String("attribute_exists(sessionId) AND attribute_not_exists(consumedAt)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":consumedAt": &types.AttributeValueMemberS{Value: consumedAt.UTC().Format(time.RFC3339Nano)},
			":listingId":  &types.AttributeValueMemberS{Value: listingID},
		},
	})
	if err != nil {
		return fmt.Errorf("consume upload session: %w", err)
	}
	return nil
}
