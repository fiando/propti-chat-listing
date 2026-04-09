package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fiando/propti/backend/internal/models"
)

type LeadRepo struct {
	db *DynamoDB
}

func NewLeadRepo(db *DynamoDB) *LeadRepo {
	return &LeadRepo{db: db}
}

func (r *LeadRepo) Put(ctx context.Context, lead *models.Lead) error {
	item, err := attributevalue.MarshalMap(lead)
	if err != nil {
		return fmt.Errorf("marshal lead: %w", err)
	}
	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.LeadsTable),
		Item:      item,
	})
	return err
}

func (r *LeadRepo) GetByID(ctx context.Context, ownerUserID, leadID string) (*models.Lead, error) {
	pk := fmt.Sprintf("%s#%s", ownerUserID, leadID)
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.LeadsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: leadID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get lead: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var lead models.Lead
	if err := attributevalue.UnmarshalMap(result.Item, &lead); err != nil {
		return nil, fmt.Errorf("unmarshal lead: %w", err)
	}
	return &lead, nil
}

// LeadPage holds a page of leads and an optional cursor for the next page.
type LeadPage struct {
	Leads      []models.Lead
	NextCursor string
}

func (r *LeadRepo) ListByOwner(ctx context.Context, ownerUserID string, limit int32) ([]models.Lead, error) {
	page, err := r.ListByOwnerPaged(ctx, ownerUserID, limit, "")
	if err != nil {
		return nil, err
	}
	return page.Leads, nil
}

func (r *LeadRepo) ListByOwnerPaged(ctx context.Context, ownerUserID string, limit int32, cursor string) (*LeadPage, error) {
	if limit <= 0 {
		limit = 100
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.db.LeadsTable),
		IndexName:              aws.String("ownerUserId-createdAt-index"),
		KeyConditionExpression: aws.String("ownerUserId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: ownerUserID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	}
	if cursor != "" {
		startKey, err := decodeCursor(cursor)
		if err == nil && len(startKey) > 0 {
			input.ExclusiveStartKey = startKey
		}
	}
	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list leads by owner: %w", err)
	}

	leads := make([]models.Lead, 0, len(result.Items))
	for _, item := range result.Items {
		var lead models.Lead
		if err := attributevalue.UnmarshalMap(item, &lead); err != nil {
			continue
		}
		leads = append(leads, lead)
	}

	page := &LeadPage{Leads: leads}
	if result.LastEvaluatedKey != nil {
		page.NextCursor = encodeCursor(result.LastEvaluatedKey)
	}
	return page, nil
}

func encodeCursor(key map[string]types.AttributeValue) string {
	raw := make(map[string]string, len(key))
	for k, v := range key {
		if sv, ok := v.(*types.AttributeValueMemberS); ok {
			raw[k] = sv.Value
		}
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]types.AttributeValue, len(raw))
	for k, v := range raw {
		result[k] = &types.AttributeValueMemberS{Value: v}
	}
	return result, nil
}
