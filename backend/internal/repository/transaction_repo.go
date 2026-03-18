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

// TransactionRepo provides CRUD operations on the propti-transactions DynamoDB table.
type TransactionRepo struct {
	db *DynamoDB
}

// NewTransactionRepo creates a new TransactionRepo.
func NewTransactionRepo(db *DynamoDB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// Put writes (creates or replaces) a transaction.
func (r *TransactionRepo) Put(ctx context.Context, tx *models.Transaction) error {
	item, err := attributevalue.MarshalMap(tx)
	if err != nil {
		return fmt.Errorf("marshal transaction: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TransactionsTable),
		Item:      item,
	})
	return err
}

// GetByID retrieves a transaction by PK (transactionId) and SK (createdAt).
func (r *TransactionRepo) GetByID(ctx context.Context, transactionID, createdAt string) (*models.Transaction, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TransactionsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: transactionID},
			"SK": &types.AttributeValueMemberS{Value: createdAt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var tx models.Transaction
	if err := attributevalue.UnmarshalMap(result.Item, &tx); err != nil {
		return nil, fmt.Errorf("unmarshal transaction: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepo) GetByOrderID(ctx context.Context, orderID string) (*models.Transaction, error) {
	tx, err := r.queryByOrderIDIndex(ctx, "orderId-index", "orderId", orderID)
	if err == nil && tx != nil {
		return tx, nil
	}
	return nil, nil
}

func (r *TransactionRepo) queryByOrderIDIndex(ctx context.Context, indexName, attributeName, orderID string) (*models.Transaction, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.TransactionsTable),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String(attributeName + " = :oid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":oid": &types.AttributeValueMemberS{Value: orderID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query by %s: %w", attributeName, err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var tx models.Transaction
	if err := attributevalue.UnmarshalMap(result.Items[0], &tx); err != nil {
		return nil, fmt.Errorf("unmarshal transaction: %w", err)
	}
	return &tx, nil
}

// ListByUserID returns all transactions for a user (via GSI).
func (r *TransactionRepo) ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Transaction, error) {
	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.db.TransactionsTable),
		IndexName:              aws.String("userId-createdAt-index"),
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list transactions by user: %w", err)
	}

	txs := make([]models.Transaction, 0, len(result.Items))
	for _, item := range result.Items {
		var tx models.Transaction
		if err := attributevalue.UnmarshalMap(item, &tx); err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// UpdateStatus updates just the status field of a transaction.
func (r *TransactionRepo) UpdateStatus(ctx context.Context, transactionID, createdAt string, status models.TransactionStatus) error {
	statusVal, err := attributevalue.Marshal(string(status))
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TransactionsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: transactionID},
			"SK": &types.AttributeValueMemberS{Value: createdAt},
		},
		UpdateExpression: aws.String("SET #s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": statusVal,
		},
	})
	return err
}
