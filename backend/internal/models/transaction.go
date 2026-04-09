package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeFeatured    TransactionType = "featured"
	TransactionTypePromotion   TransactionType = "promotion"
	TransactionTypePremiumTier TransactionType = "premium_tier"

	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	PK            string            `json:"pk" dynamodbav:"PK"`
	SK            string            `json:"sk" dynamodbav:"SK"`
	TransactionID string            `json:"transactionId" dynamodbav:"transactionId"`
	UserID        string            `json:"userId" dynamodbav:"userId"`
	Type          TransactionType   `json:"type" dynamodbav:"type"`
	ListingID     string            `json:"listingId,omitempty" dynamodbav:"listingId,omitempty"`
	Amount        float64           `json:"amount" dynamodbav:"amount"`
	Currency      string            `json:"currency" dynamodbav:"currency"`
	Status        TransactionStatus `json:"status" dynamodbav:"status"`
	Provider      string            `json:"provider,omitempty" dynamodbav:"provider,omitempty"`
	PaymentID     string            `json:"paymentId,omitempty" dynamodbav:"paymentId,omitempty"`
	OrderID       string            `json:"orderId,omitempty" dynamodbav:"orderId,omitempty"`
	PaymentURL    string            `json:"paymentUrl,omitempty" dynamodbav:"paymentUrl,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt" dynamodbav:"updatedAt"`
}

type FeatureListingRequest struct {
	ListingID    string `json:"listingId"`
	DurationDays int    `json:"durationDays"`
	Type         string `json:"type"` // featured | promotion
}

type PremiumUpgradeRequest struct {
	Tier string `json:"tier"`
}

// GrantTrialRequest is the body for POST /admin/grant-trial.
type GrantTrialRequest struct {
	UserID         string `json:"userId"`
	Tier           string `json:"tier"`           // optional, defaults to "basic"
	DurationMonths int    `json:"durationMonths"` // optional, defaults to 3
}

// GrantTrialResponse is returned on a successful trial grant.
type GrantTrialResponse struct {
	UserID    string `json:"userId"`
	Tier      string `json:"tier"`
	RenewDate string `json:"renewDate"`
}

type PaymentResponse struct {
	TransactionID string  `json:"transactionId"`
	PaymentURL    string  `json:"paymentUrl"`
	OrderID       string  `json:"orderId"`
	Amount        float64 `json:"amount"`
}

type transactionAlias Transaction

func (t *Transaction) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	m, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("unexpected transaction attribute value type %T", av)
	}

	normalized := make(map[string]types.AttributeValue, len(m.Value)+2)
	for key, value := range m.Value {
		normalized[key] = value
	}

	var decoded transactionAlias
	if err := attributevalue.UnmarshalMap(normalized, &decoded); err != nil {
		return err
	}

	*t = Transaction(decoded)
	return nil
}
