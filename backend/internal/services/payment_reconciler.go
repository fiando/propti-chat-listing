package services

import (
	"context"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/payments"
)

type reconcileUserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Put(ctx context.Context, user *models.User) error
}

type reconcileTransactionStore interface {
	ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Transaction, error)
	UpdateStatus(ctx context.Context, transactionID, createdAt string, status models.TransactionStatus) error
}

type reconcilePaymentProvider interface {
	GetPaymentStatus(ctx context.Context, paymentID string) (payments.PaymentStatus, error)
}

type PaymentReconciler struct {
	userRepo        reconcileUserStore
	transactionRepo reconcileTransactionStore
	paymentProvider reconcilePaymentProvider
}

func NewPaymentReconciler(userRepo reconcileUserStore, transactionRepo reconcileTransactionStore, paymentProvider reconcilePaymentProvider) *PaymentReconciler {
	return &PaymentReconciler{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		paymentProvider: paymentProvider,
	}
}

func (r *PaymentReconciler) ReconcileUser(ctx context.Context, userID string) error {
	if r == nil || r.paymentProvider == nil || r.transactionRepo == nil || r.userRepo == nil {
		return nil
	}

	txs, err := r.transactionRepo.ListByUserID(ctx, userID, 10)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		if tx.Type != models.TransactionTypePremiumTier || tx.Status != models.TransactionStatusPending || tx.Provider != payments.ProviderDOKU || tx.ProviderPaymentID == "" {
			continue
		}

		status, err := r.paymentProvider.GetPaymentStatus(ctx, tx.ProviderPaymentID)
		if err != nil || status != payments.PaymentStatusSucceeded {
			return err
		}

		if err := r.transactionRepo.UpdateStatus(ctx, tx.TransactionID, tx.SK, models.TransactionStatusCompleted); err != nil {
			return err
		}

		user, err := r.userRepo.GetByID(ctx, tx.UserID)
		if err != nil || user == nil {
			return err
		}

		renewDate := time.Now().UTC().AddDate(0, 1, 0)
		user.Subscription.Tier = models.SubscriptionPremium
		user.Subscription.RenewDate = &renewDate
		return r.userRepo.Put(ctx, user)
	}

	return nil
}
