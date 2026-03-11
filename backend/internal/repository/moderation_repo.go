package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fiando/propti/backend/internal/models"
)

// ModerationRepo provides write/read access to the propti-moderations DynamoDB table.
type ModerationRepo struct {
	db *DynamoDB
}

// NewModerationRepo creates a new ModerationRepo.
func NewModerationRepo(db *DynamoDB) *ModerationRepo {
	return &ModerationRepo{db: db}
}

// Put writes (creates or replaces) a moderation record.
func (r *ModerationRepo) Put(ctx context.Context, m *models.Moderation) error {
	item, err := attributevalue.MarshalMap(m)
	if err != nil {
		return fmt.Errorf("marshal moderation: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.ModerationsTable),
		Item:      item,
	})
	return err
}

// GetByID retrieves a moderation record by its primary key pair.
func (r *ModerationRepo) GetByID(ctx context.Context, moderationID, createdAt string) (*models.Moderation, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.ModerationsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: moderationID},
			"SK": &types.AttributeValueMemberS{Value: createdAt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get moderation: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var m models.Moderation
	if err := attributevalue.UnmarshalMap(result.Item, &m); err != nil {
		return nil, fmt.Errorf("unmarshal moderation: %w", err)
	}
	return &m, nil
}

// ListByListingID returns all moderation records for a listing via a GSI.
func (r *ModerationRepo) ListByListingID(ctx context.Context, listingID string) ([]models.Moderation, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.ModerationsTable),
		IndexName:              aws.String("listingId-index"),
		KeyConditionExpression: aws.String("listingId = :lid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lid": &types.AttributeValueMemberS{Value: listingID},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("list moderations by listingId: %w", err)
	}

	mods := make([]models.Moderation, 0, len(result.Items))
	for _, item := range result.Items {
		var m models.Moderation
		if err := attributevalue.UnmarshalMap(item, &m); err != nil {
			continue
		}
		mods = append(mods, m)
	}
	return mods, nil
}
