package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/payments"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakePremiumUserStore struct {
	byID map[string]*models.User
}

func (f *fakePremiumUserStore) GetByID(_ context.Context, userID string) (*models.User, error) {
	user := f.byID[userID]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (f *fakePremiumUserStore) Put(_ context.Context, user *models.User) error {
	copy := *user
	f.byID[user.UserID] = &copy
	return nil
}

type fakePremiumTransactionStore struct {
	items []*models.Transaction
	byID  map[string]*models.Transaction
}

func (f *fakePremiumTransactionStore) Put(_ context.Context, tx *models.Transaction) error {
	copy := *tx
	f.items = append(f.items, &copy)
	if f.byID == nil {
		f.byID = map[string]*models.Transaction{}
	}
	f.byID[copy.OrderID] = &copy
	return nil
}

func (f *fakePremiumTransactionStore) GetByOrderID(_ context.Context, orderID string) (*models.Transaction, error) {
	tx := f.byID[orderID]
	if tx == nil {
		return nil, nil
	}
	copy := *tx
	return &copy, nil
}

func (f *fakePremiumTransactionStore) UpdateStatus(_ context.Context, transactionID, createdAt string, status models.TransactionStatus) error {
	for _, tx := range f.items {
		if tx.TransactionID == transactionID && tx.SK == createdAt {
			tx.Status = status
			tx.UpdatedAt = time.Now().UTC()
		}
	}
	for _, tx := range f.byID {
		if tx.TransactionID == transactionID && tx.SK == createdAt {
			tx.Status = status
			tx.UpdatedAt = time.Now().UTC()
		}
	}
	return nil
}

func (f *fakePremiumTransactionStore) ListByUserID(_ context.Context, userID string, limit int32) ([]models.Transaction, error) {
	if limit <= 0 {
		limit = int32(len(f.items))
	}

	result := make([]models.Transaction, 0, len(f.items))
	for _, tx := range f.items {
		if tx.UserID != userID {
			continue
		}
		result = append(result, *tx)
		if int32(len(result)) == limit {
			break
		}
	}

	return result, nil
}

type fakePremiumListingService struct {
	lastUserID      string
	lastListingID   string
	lastFeatureType string
	lastUntil       time.Time
}

func (f *fakePremiumListingService) FeatureListing(_ context.Context, userID, listingID, featureType string, until time.Time) error {
	f.lastUserID = userID
	f.lastListingID = listingID
	f.lastFeatureType = featureType
	f.lastUntil = until
	return nil
}

type fakePaymentProvider struct {
	createResult   *payments.CreatePaymentResult
	createErr      error
	callbackResult *payments.CallbackResult
	callbackErr    error
	statusResult   payments.PaymentStatus
	statusErr      error
	lastInput      payments.CreatePaymentInput
}

func (f *fakePaymentProvider) Name() string {
	return payments.ProviderDOKU
}

func (f *fakePaymentProvider) CreatePayment(_ context.Context, input payments.CreatePaymentInput) (*payments.CreatePaymentResult, error) {
	f.lastInput = input
	return f.createResult, f.createErr
}

func (f *fakePaymentProvider) GetPaymentStatus(_ context.Context, _ string) (payments.PaymentStatus, error) {
	return f.statusResult, f.statusErr
}

func (f *fakePaymentProvider) ParseCallback(_ map[string]string, _ string, _ []byte) (*payments.CallbackResult, error) {
	return f.callbackResult, f.callbackErr
}

func authHeaderForPremiumTest(t *testing.T, userID string) map[string]string {
	t.Helper()
	token, err := utils.GenerateToken(userID, "test@example.com")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

func TestGenerateOrderIDUsesCompactCollisionResistantSuffix(t *testing.T) {
	t.Parallel()

	id, err := generateOrderID("PROPTI-PREM-")
	if err != nil {
		t.Fatalf("generateOrderID returned error: %v", err)
	}

	const prefix = "PROPTI-PREM-"
	if len(id) != len(prefix)+16 {
		t.Fatalf("expected %d-char order id, got %d (%q)", len(prefix)+16, len(id), id)
	}
	if got := id[:len(prefix)]; got != prefix {
		t.Fatalf("expected prefix %q, got %q", prefix, got)
	}
}

func TestPremiumHandlerUpgradeToPremiumUsesPaymentProvider(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}
	txStore := &fakePremiumTransactionStore{}
	listingService := &fakePremiumListingService{}
	provider := &fakePaymentProvider{
		createResult: &payments.CreatePaymentResult{
			Provider:   payments.ProviderDOKU,
			OrderID:    "PROPTI-PREM-001",
			PaymentID:  "tok-123",
			PaymentURL: "https://sandbox.doku.com/checkout-link-v2/tok-123",
		},
	}

	handler := NewPremiumHandler(userStore, txStore, listingService, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body models.PaymentResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body.PaymentURL != "https://sandbox.doku.com/checkout-link-v2/tok-123" {
		t.Fatalf("unexpected payment url: %q", body.PaymentURL)
	}
	if body.OrderID != "PROPTI-PREM-001" {
		t.Fatalf("unexpected order id: %q", body.OrderID)
	}
	if body.Amount != 129000 {
		t.Fatalf("expected amount 129000, got %f", body.Amount)
	}

	if provider.lastInput.NotificationURL != "https://api.propti.test/premium/callback" {
		t.Fatalf("expected callback url to be passed to provider, got %q", provider.lastInput.NotificationURL)
	}
	if provider.lastInput.CallbackURL != "https://propti.test/profile#premium" {
		t.Fatalf("expected callback url result to point to profile, got %q", provider.lastInput.CallbackURL)
	}
	if provider.lastInput.ResultURL != "https://propti.test/profile#premium" {
		t.Fatalf("expected result url to point to profile, got %q", provider.lastInput.ResultURL)
	}
	if !provider.lastInput.AutoRedirect {
		t.Fatalf("expected auto redirect to be enabled")
	}
	if provider.lastInput.Customer.Email != "bob@example.com" {
		t.Fatalf("expected customer email to be forwarded, got %q", provider.lastInput.Customer.Email)
	}
	if len(txStore.items) != 1 {
		t.Fatalf("expected one stored transaction, got %d", len(txStore.items))
	}
	if txStore.items[0].Provider != payments.ProviderDOKU {
		t.Fatalf("expected stored provider %q, got %q", payments.ProviderDOKU, txStore.items[0].Provider)
	}
}

func TestPremiumHandlerUpgradeToPremiumReusesPendingTransactionWithinPaymentWindow(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}

	createdAt := time.Now().UTC().Add(-20 * time.Minute)
	existingTx := &models.Transaction{
		PK:            "tx-existing",
		SK:            createdAt.Format(time.RFC3339),
		TransactionID: "tx-existing",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Amount:        129000,
		Currency:      "IDR",
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		PaymentID:     "tok-existing",
		OrderID:       "PROPTI-PREM-EXISTING",
		PaymentURL:    "https://sandbox.doku.com/checkout-link-v2/tok-existing",
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
	txStore := &fakePremiumTransactionStore{
		items: []*models.Transaction{existingTx},
		byID: map[string]*models.Transaction{
			"PROPTI-PREM-EXISTING": existingTx,
		},
	}
	provider := &fakePaymentProvider{
		createErr: errors.New("CreatePayment should not be called when pending premium checkout is still active"),
	}

	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body models.PaymentResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body.TransactionID != "tx-existing" {
		t.Fatalf("expected existing transaction id to be reused, got %q", body.TransactionID)
	}
	if body.OrderID != "PROPTI-PREM-EXISTING" {
		t.Fatalf("expected existing order id to be reused, got %q", body.OrderID)
	}
	if body.PaymentURL != "https://sandbox.doku.com/checkout-link-v2/tok-existing" {
		t.Fatalf("expected existing payment url to be reused, got %q", body.PaymentURL)
	}
	if len(txStore.items) != 1 {
		t.Fatalf("expected no new transaction to be stored, got %d items", len(txStore.items))
	}
}


func TestPremiumHandlerUpgradeToPremiumCancelsMismatchedPendingTierAndCreatesNewOrder(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}

	createdAt := time.Now().UTC().Add(-10 * time.Minute)
	existingTx := &models.Transaction{
		PK:            "tx-existing-pro",
		SK:            createdAt.Format(time.RFC3339),
		TransactionID: "tx-existing-pro",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Amount:        199000,
		Currency:      "IDR",
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		PaymentID:     "tok-existing-pro",
		OrderID:       "PROPTI-PREM-EXISTING-PRO",
		PaymentURL:    "https://sandbox.doku.com/checkout-link-v2/tok-existing-pro",
		Metadata: map[string]string{
			"tier": "pro",
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	 txStore := &fakePremiumTransactionStore{
		items: []*models.Transaction{existingTx},
		byID: map[string]*models.Transaction{
			"PROPTI-PREM-EXISTING-PRO": existingTx,
		},
	}
	provider := &fakePaymentProvider{
		createResult: &payments.CreatePaymentResult{
			Provider:   payments.ProviderDOKU,
			OrderID:    "PROPTI-PREM-NEW-PREMIUM",
			PaymentID:  "tok-new-premium",
			PaymentURL: "https://sandbox.doku.com/checkout-link-v2/tok-new-premium",
		},
	}

	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
		Body:       `{"tier":"premium"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body models.PaymentResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body.OrderID != "PROPTI-PREM-NEW-PREMIUM" {
		t.Fatalf("expected a new order id for changed tier, got %q", body.OrderID)
	}
	if len(txStore.items) != 2 {
		t.Fatalf("expected a new transaction to be stored, got %d items", len(txStore.items))
	}
	if got := txStore.byID["PROPTI-PREM-EXISTING-PRO"].Status; got == models.TransactionStatusPending {
		t.Fatalf("expected previous mismatched-tier transaction to be canceled/invalidated, still pending")
	}
}

func TestPremiumHandlerUpgradeToPremiumCreatesNewTransactionAfterPendingPaymentExpires(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}

	expiredAt := time.Now().UTC().Add(-61 * time.Minute)
	expiredTx := &models.Transaction{
		PK:            "tx-expired",
		SK:            expiredAt.Format(time.RFC3339),
		TransactionID: "tx-expired",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Amount:        129000,
		Currency:      "IDR",
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		PaymentID:     "tok-expired",
		OrderID:       "PROPTI-PREM-EXPIRED",
		PaymentURL:    "https://sandbox.doku.com/checkout-link-v2/tok-expired",
		CreatedAt:     expiredAt,
		UpdatedAt:     expiredAt,
	}
	txStore := &fakePremiumTransactionStore{
		items: []*models.Transaction{expiredTx},
		byID: map[string]*models.Transaction{
			"PROPTI-PREM-EXPIRED": expiredTx,
		},
	}
	provider := &fakePaymentProvider{
		createResult: &payments.CreatePaymentResult{
			Provider:   payments.ProviderDOKU,
			OrderID:    "PROPTI-PREM-NEW",
			PaymentID:  "tok-new",
			PaymentURL: "https://sandbox.doku.com/checkout-link-v2/tok-new",
		},
	}

	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body models.PaymentResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body.OrderID != "PROPTI-PREM-NEW" {
		t.Fatalf("expected a new order id after expiry, got %q", body.OrderID)
	}
	if body.PaymentURL != "https://sandbox.doku.com/checkout-link-v2/tok-new" {
		t.Fatalf("expected a new payment url after expiry, got %q", body.PaymentURL)
	}
	if len(txStore.items) != 2 {
		t.Fatalf("expected a new transaction to be stored, got %d items", len(txStore.items))
	}
}

func TestPremiumHandlerCallbackCompletesPremiumUpgrade(t *testing.T) {
	t.Parallel()

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Email:  "bob@example.com",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}
	tx := &models.Transaction{
		PK:            "tx-1",
		SK:            time.Now().UTC().Format(time.RFC3339),
		TransactionID: "tx-1",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		OrderID:       "PROPTI-PREM-001",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	txStore := &fakePremiumTransactionStore{
		items: []*models.Transaction{tx},
		byID: map[string]*models.Transaction{
			"PROPTI-PREM-001": tx,
		},
	}
	provider := &fakePaymentProvider{
		callbackResult: &payments.CallbackResult{
			Provider:  payments.ProviderDOKU,
			OrderID:   "PROPTI-PREM-001",
			PaymentID: "auth-789",
			Status:    payments.PaymentStatusSucceeded,
		},
	}

	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/callback",
		Headers: map[string]string{
			"Client-Id": "client-123",
		},
		Body: `{"ok":true}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	if got := userStore.byID["user-1"].Subscription.Tier; got != models.SubscriptionPremium {
		t.Fatalf("expected premium subscription after callback, got %q", got)
	}
	if got := txStore.byID["PROPTI-PREM-001"].Status; got != models.TransactionStatusCompleted {
		t.Fatalf("expected completed transaction, got %q", got)
	}
}

func TestMain(m *testing.M) {
	_ = os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}

func TestPremiumHandlerRenewalRejectedWhenNotInWindow(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	futureExpiry := time.Now().UTC().Add(30 * 24 * time.Hour)
	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: &futureExpiry,
				},
			},
		},
	}
	handler := NewPremiumHandler(userStore, &fakePremiumTransactionStore{}, &fakePremiumListingService{}, &fakePaymentProvider{})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when renewal not yet open, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if !strings.Contains(resp.Body, "7 days") {
		t.Fatalf("expected error message about 7-day window, got: %s", resp.Body)
	}
}

func TestPremiumHandlerRenewalAllowedWhenExpiringSoon(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	expiringSoon := time.Now().UTC().Add(3 * 24 * time.Hour)
	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Phone:  "6281234567890",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: &expiringSoon,
				},
			},
		},
	}
	provider := &fakePaymentProvider{
		createResult: &payments.CreatePaymentResult{
			Provider:   payments.ProviderDOKU,
			OrderID:    "PROPTI-PREM-RENEW-001",
			PaymentID:  "tok-renew",
			PaymentURL: "https://sandbox.doku.com/checkout-link-v2/tok-renew",
		},
	}
	handler := NewPremiumHandler(userStore, &fakePremiumTransactionStore{}, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for renewal within window, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func TestPremiumHandlerUpgradeToPremiumUsesCorrectPricing(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}
	provider := &fakePaymentProvider{
		createResult: &payments.CreatePaymentResult{
			Provider:   payments.ProviderDOKU,
			OrderID:    "PROPTI-PREM-PREMIUM-001",
			PaymentID:  "tok-premium",
			PaymentURL: "https://sandbox.doku.com/checkout-link-v2/tok-premium",
		},
	}
	txStore := &fakePremiumTransactionStore{}
	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, provider)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
		Body:       `{"tier":"premium"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	var body models.PaymentResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body.Amount != 99000 {
		t.Fatalf("expected amount 99000 for premium, got %f", body.Amount)
	}
	if len(txStore.items) != 1 || txStore.items[0].Metadata["tier"] != "premium" {
		t.Fatalf("expected premium tier metadata in transaction, got %#v", txStore.items)
	}
}

func TestPremiumHandlerUpgradeRejectsBasicTier(t *testing.T) {
	t.Setenv("PUBLIC_API_BASE_URL", "https://api.propti.test")

	userStore := &fakePremiumUserStore{
		byID: map[string]*models.User{
			"user-1": {
				UserID: "user-1",
				Name:   "Bobby",
				Email:  "bob@example.com",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
	}
	txStore := &fakePremiumTransactionStore{}
	handler := NewPremiumHandler(userStore, txStore, &fakePremiumListingService{}, &fakePaymentProvider{})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/premium/upgrade",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
		Body:       `{"tier":"basic"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for basic tier (no longer purchasable), got %d body=%s", resp.StatusCode, resp.Body)
	}
}
