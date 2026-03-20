package services

import (
	"context"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/payments"
)

type fakeReconcileUserStore struct {
	user *models.User
}

func (f *fakeReconcileUserStore) GetByID(_ context.Context, userID string) (*models.User, error) {
	if f.user == nil || f.user.UserID != userID {
		return nil, nil
	}
	copy := *f.user
	return &copy, nil
}

func (f *fakeReconcileUserStore) Put(_ context.Context, user *models.User) error {
	copy := *user
	f.user = &copy
	return nil
}

type fakeReconcileTransactionStore struct {
	items   []models.Transaction
	updated []models.TransactionStatus
}

func (f *fakeReconcileTransactionStore) ListByUserID(_ context.Context, userID string, limit int32) ([]models.Transaction, error) {
	return f.items, nil
}

func (f *fakeReconcileTransactionStore) UpdateStatus(_ context.Context, transactionID, createdAt string, status models.TransactionStatus) error {
	f.updated = append(f.updated, status)
	for i := range f.items {
		if f.items[i].TransactionID == transactionID && f.items[i].SK == createdAt {
			f.items[i].Status = status
		}
	}
	return nil
}

type fakeReconcileProvider struct {
	status payments.PaymentStatus
}

func (f *fakeReconcileProvider) GetPaymentStatus(_ context.Context, paymentID string) (payments.PaymentStatus, error) {
	return f.status, nil
}

func TestPaymentReconcilerUpgradesUserWhenPendingPremiumPaymentIsPaid(t *testing.T) {
	t.Parallel()

	userStore := &fakeReconcileUserStore{user: &models.User{
		UserID:       "user-1",
		Subscription: models.Subscription{Tier: models.SubscriptionFree},
	}}
	txStore := &fakeReconcileTransactionStore{items: []models.Transaction{{
		TransactionID: "tx-1",
		SK:            "2026-03-16T06:00:00Z",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		PaymentID:     "tok-123",
	}}}
	reconciler := NewPaymentReconciler(userStore, txStore, &fakeReconcileProvider{status: payments.PaymentStatusSucceeded})

	if err := reconciler.ReconcileUser(context.Background(), "user-1"); err != nil {
		t.Fatalf("ReconcileUser returned error: %v", err)
	}

	if got := userStore.user.Subscription.Tier; got != models.SubscriptionPremium {
		t.Fatalf("expected premium tier, got %q", got)
	}
	if userStore.user.Subscription.RenewDate == nil || !userStore.user.Subscription.RenewDate.After(time.Now().UTC()) {
		t.Fatal("expected renew date to be set in the future")
	}
	if len(txStore.updated) != 1 || txStore.updated[0] != models.TransactionStatusCompleted {
		t.Fatalf("expected transaction to be marked completed, got %#v", txStore.updated)
	}
}

func TestPaymentReconcilerLeavesUserFreeWhenPaymentStillPending(t *testing.T) {
	t.Parallel()

	userStore := &fakeReconcileUserStore{user: &models.User{
		UserID:       "user-1",
		Subscription: models.Subscription{Tier: models.SubscriptionFree},
	}}
	txStore := &fakeReconcileTransactionStore{items: []models.Transaction{{
		TransactionID: "tx-1",
		SK:            "2026-03-16T06:00:00Z",
		UserID:        "user-1",
		Type:          models.TransactionTypePremiumTier,
		Status:        models.TransactionStatusPending,
		Provider:      payments.ProviderDOKU,
		PaymentID:     "tok-123",
	}}}
	reconciler := NewPaymentReconciler(userStore, txStore, &fakeReconcileProvider{status: payments.PaymentStatusPending})

	if err := reconciler.ReconcileUser(context.Background(), "user-1"); err != nil {
		t.Fatalf("ReconcileUser returned error: %v", err)
	}

	if got := userStore.user.Subscription.Tier; got != models.SubscriptionFree {
		t.Fatalf("expected free tier, got %q", got)
	}
	if len(txStore.updated) != 0 {
		t.Fatalf("expected no status updates, got %#v", txStore.updated)
	}
}

func TestPaymentReconcilerExtendsRenewDateWhenRenewingBeforeExpiry(t *testing.T) {
t.Parallel()

existingExpiry := time.Now().UTC().Add(5 * 24 * time.Hour) // 5 days remaining
userStore := &fakeReconcileUserStore{user: &models.User{
UserID: "user-1",
Subscription: models.Subscription{
Tier:      models.SubscriptionPremium,
RenewDate: &existingExpiry,
},
}}
txStore := &fakeReconcileTransactionStore{items: []models.Transaction{{
TransactionID: "tx-renew",
SK:            "2026-03-16T06:00:00Z",
UserID:        "user-1",
Type:          models.TransactionTypePremiumTier,
Status:        models.TransactionStatusPending,
Provider:      payments.ProviderDOKU,
PaymentID:     "tok-456",
}}}
reconciler := NewPaymentReconciler(userStore, txStore, &fakeReconcileProvider{status: payments.PaymentStatusSucceeded})

if err := reconciler.ReconcileUser(context.Background(), "user-1"); err != nil {
t.Fatalf("ReconcileUser returned error: %v", err)
}

expected := existingExpiry.AddDate(0, 1, 0)
got := userStore.user.Subscription.RenewDate
if got == nil {
t.Fatal("expected renewDate to be set")
}
// Allow 1 second tolerance for test timing
diff := got.Sub(expected)
if diff < -time.Second || diff > time.Second {
t.Fatalf("expected renewDate ~%v (extended from existing), got %v", expected, *got)
}
}

func TestPaymentReconcilerStartsFreshRenewDateWhenRenewingAfterExpiry(t *testing.T) {
t.Parallel()

expiredDate := time.Now().UTC().Add(-24 * time.Hour) // already expired
userStore := &fakeReconcileUserStore{user: &models.User{
UserID: "user-1",
Subscription: models.Subscription{
Tier:      models.SubscriptionPremium,
RenewDate: &expiredDate,
},
}}
txStore := &fakeReconcileTransactionStore{items: []models.Transaction{{
TransactionID: "tx-restart",
SK:            "2026-03-16T06:00:00Z",
UserID:        "user-1",
Type:          models.TransactionTypePremiumTier,
Status:        models.TransactionStatusPending,
Provider:      payments.ProviderDOKU,
PaymentID:     "tok-789",
}}}
reconciler := NewPaymentReconciler(userStore, txStore, &fakeReconcileProvider{status: payments.PaymentStatusSucceeded})

before := time.Now().UTC()
if err := reconciler.ReconcileUser(context.Background(), "user-1"); err != nil {
t.Fatalf("ReconcileUser returned error: %v", err)
}
after := time.Now().UTC()

got := userStore.user.Subscription.RenewDate
if got == nil {
t.Fatal("expected renewDate to be set")
}
expectedMin := before.AddDate(0, 1, 0)
expectedMax := after.AddDate(0, 1, 0)
if got.Before(expectedMin) || got.After(expectedMax) {
t.Fatalf("expected new renewDate ~1 month from now, got %v", *got)
}
}
