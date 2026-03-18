package payments

import "context"

const (
	ProviderDOKU = "doku"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type Customer struct {
	ID    string
	Name  string
	Email string
	Phone string
}

type CreatePaymentInput struct {
	OrderID         string
	Amount          float64
	Currency        string
	Description     string
	NotificationURL string
	Customer        Customer
}

type CreatePaymentResult struct {
	Provider   string
	OrderID    string
	PaymentID  string
	PaymentURL string
}

type CallbackResult struct {
	Provider  string
	OrderID   string
	PaymentID string
	Status    PaymentStatus
}

type Provider interface {
	Name() string
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
	ParseCallback(headers map[string]string, path string, body []byte) (*CallbackResult, error)
}
