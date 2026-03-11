package models

import "time"

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
	PK                string            `json:"pk" dynamodbav:"PK"`
	SK                string            `json:"sk" dynamodbav:"SK"`
	TransactionID     string            `json:"transactionId" dynamodbav:"transactionId"`
	UserID            string            `json:"userId" dynamodbav:"userId"`
	Type              TransactionType   `json:"type" dynamodbav:"type"`
	ListingID         string            `json:"listingId,omitempty" dynamodbav:"listingId,omitempty"`
	Amount            float64           `json:"amount" dynamodbav:"amount"`
	Currency          string            `json:"currency" dynamodbav:"currency"`
	Status            TransactionStatus `json:"status" dynamodbav:"status"`
	MidtransPaymentID string            `json:"midtransPaymentId,omitempty" dynamodbav:"midtransPaymentId,omitempty"`
	MidtransOrderID   string            `json:"midtransOrderId,omitempty" dynamodbav:"midtransOrderId,omitempty"`
	PaymentURL        string            `json:"paymentUrl,omitempty" dynamodbav:"paymentUrl,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt" dynamodbav:"updatedAt"`
}

type FeatureListingRequest struct {
	ListingID    string `json:"listingId"`
	DurationDays int    `json:"durationDays"`
	Type         string `json:"type"` // featured | promotion
}

type PremiumUpgradeRequest struct {
	Tier string `json:"tier"`
}

type PaymentResponse struct {
	TransactionID string `json:"transactionId"`
	PaymentURL    string `json:"paymentUrl"`
	OrderID       string `json:"orderId"`
}
