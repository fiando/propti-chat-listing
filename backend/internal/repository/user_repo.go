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

// UserRepo provides CRUD operations on the propti-users DynamoDB table.
type UserRepo struct {
	db *DynamoDB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *DynamoDB) *UserRepo {
	return &UserRepo{db: db}
}

// Put writes (creates or replaces) a user.
func (r *UserRepo) Put(ctx context.Context, user *models.User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.UsersTable),
		Item:      item,
	})
	return err
}

// GetByID retrieves a user by their userId (PK = userId, SK = "metadata").
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.UsersTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: userID},
			"SK": &types.AttributeValueMemberS{Value: "metadata"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

// GetByGoogleID looks up a user by their Google subject ID via a GSI.
func (r *UserRepo) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.UsersTable),
		IndexName:              aws.String("googleId-index"),
		KeyConditionExpression: aws.String("googleId = :gid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gid": &types.AttributeValueMemberS{Value: googleID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query user by googleId: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

// Update performs a partial update on a user item.
func (r *UserRepo) Update(ctx context.Context, userID string, updates map[string]types.AttributeValue) error {
	updateExpr := "SET "
	exprAttrNames := map[string]string{}
	i := 0
	for k := range updates {
		if i > 0 {
			updateExpr += ", "
		}
		placeholder := fmt.Sprintf("#f%d", i)
		valKey := fmt.Sprintf(":v%d", i)
		updateExpr += fmt.Sprintf("%s = %s", placeholder, valKey)
		exprAttrNames[placeholder] = k
		i++
	}

	exprAttrValues := map[string]types.AttributeValue{}
	i = 0
	for _, v := range updates {
		valKey := fmt.Sprintf(":v%d", i)
		exprAttrValues[valKey] = v
		i++
	}

	_, err := r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.UsersTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: userID},
			"SK": &types.AttributeValueMemberS{Value: "metadata"},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})
	return err
}

// GetByWhatsAppPhone looks up a user by linked WhatsApp phone.
func (r *UserRepo) GetByWhatsAppPhone(ctx context.Context, phone string) (*models.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone is required")
	}

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.db.UsersTable),
		FilterExpression: aws.String("whatsAppLinkedPhone = :phone"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":phone": &types.AttributeValueMemberS{Value: phone},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("scan user by whatsapp phone: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		return nil, fmt.Errorf("unmarshal user by whatsapp phone: %w", err)
	}
	return &user, nil
}
