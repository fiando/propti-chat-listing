package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/payments"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type premiumUserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Put(ctx context.Context, user *models.User) error
}

type premiumTransactionStore interface {
	Put(ctx context.Context, tx *models.Transaction) error
	GetByProviderOrderID(ctx context.Context, orderID string) (*models.Transaction, error)
	UpdateStatus(ctx context.Context, transactionID, createdAt string, status models.TransactionStatus) error
}

type premiumListingService interface {
	FeatureListing(ctx context.Context, userID, listingID, featureType string, until time.Time) error
}

// PremiumHandler handles subscription upgrades, featured listings, and payment callbacks.
type PremiumHandler struct {
	userRepo        premiumUserStore
	transactionRepo premiumTransactionStore
	listingService  premiumListingService
	paymentProvider payments.Provider
}

// NewPremiumHandler creates a PremiumHandler.
func NewPremiumHandler(
	userRepo premiumUserStore,
	transactionRepo premiumTransactionStore,
	listingService premiumListingService,
	paymentProvider payments.Provider,
) *PremiumHandler {
	return &PremiumHandler{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		listingService:  listingService,
		paymentProvider: paymentProvider,
	}
}

// Handle routes premium API requests.
func (h *PremiumHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == http.MethodOptions:
		return jsonResponse(http.StatusOK, ""), nil
	case req.HTTPMethod == http.MethodPost && req.Path == "/premium/upgrade":
		return h.upgradeToPremium(ctx, req)
	case req.HTTPMethod == http.MethodPost && req.Path == "/premium/feature-listing":
		return h.featureListing(ctx, req)
	case req.HTTPMethod == http.MethodPost && req.Path == "/premium/callback":
		return h.paymentCallback(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

var (
	_ premiumUserStore        = (*repository.UserRepo)(nil)
	_ premiumTransactionStore = (*repository.TransactionRepo)(nil)
	_ premiumListingService   = (*services.ListingService)(nil)
)

// upgradeToPremium initiates a payment to upgrade the user to premium.
func (h *PremiumHandler) upgradeToPremium(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	if user.Subscription.Tier == models.SubscriptionPremium {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "already on premium tier"))), nil
	}

	orderID := fmt.Sprintf("PROPTI-PREM-%s", uuid.NewString())
	amount := 49000.0 // Rp 49,000/month

	tx := &models.Transaction{
		PK:              uuid.NewString(),
		SK:              time.Now().UTC().Format(time.RFC3339),
		TransactionID:   uuid.NewString(),
		UserID:          userID,
		Type:            models.TransactionTypePremiumTier,
		Amount:          amount,
		Currency:        "IDR",
		Status:          models.TransactionStatusPending,
		Provider:        h.paymentProvider.Name(),
		ProviderOrderID: orderID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	tx.PK = tx.TransactionID

	paymentResult, err := h.paymentProvider.CreatePayment(ctx, payments.CreatePaymentInput{
		OrderID:         orderID,
		Amount:          amount,
		Currency:        "IDR",
		Description:     "Propti Premium",
		NotificationURL: buildPremiumCallbackURL(),
		Customer: payments.Customer{
			ID:    userID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
	})
	if err != nil {
		utils.LogError("create payment checkout", err, "userId", userID, "provider", h.paymentProvider.Name())
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	tx.ProviderPaymentID = paymentResult.ProviderPaymentID
	tx.ProviderOrderID = paymentResult.ProviderOrderID
	tx.PaymentURL = paymentResult.PaymentURL

	if err := h.transactionRepo.Put(ctx, tx); err != nil {
		utils.LogError("save transaction", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp := models.PaymentResponse{
		TransactionID: tx.TransactionID,
		PaymentURL:    paymentResult.PaymentURL,
		OrderID:       paymentResult.ProviderOrderID,
		Amount:        amount,
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// featureListing initiates payment to feature or promote a listing.
func (h *PremiumHandler) featureListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var featReq models.FeatureListingRequest
	if err := json.Unmarshal([]byte(req.Body), &featReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if err := utils.ValidateFeatureListingRequest(&featReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, err.Error()))), nil
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	// Pricing: Rp 50,000/day for featured; Rp 30,000/day for promotion.
	pricePerDay := 50000.0
	if featReq.Type == "promotion" {
		pricePerDay = 30000.0
	}
	amount := pricePerDay * float64(featReq.DurationDays)

	orderID := fmt.Sprintf("PROPTI-FEAT-%s", uuid.NewString())

	txType := models.TransactionTypeFeatured
	if featReq.Type == "promotion" {
		txType = models.TransactionTypePromotion
	}

	tx := &models.Transaction{
		TransactionID:   uuid.NewString(),
		UserID:          userID,
		Type:            txType,
		ListingID:       featReq.ListingID,
		Amount:          amount,
		Currency:        "IDR",
		Status:          models.TransactionStatusPending,
		Provider:        h.paymentProvider.Name(),
		ProviderOrderID: orderID,
		Metadata: map[string]string{
			"featureType":  featReq.Type,
			"durationDays": fmt.Sprintf("%d", featReq.DurationDays),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	tx.PK = tx.TransactionID
	tx.SK = tx.CreatedAt.Format(time.RFC3339)

	description := "Propti Featured Listing"
	if featReq.Type == "promotion" {
		description = "Propti Listing Promotion"
	}
	paymentResult, err := h.paymentProvider.CreatePayment(ctx, payments.CreatePaymentInput{
		OrderID:         orderID,
		Amount:          amount,
		Currency:        "IDR",
		Description:     description,
		NotificationURL: buildPremiumCallbackURL(),
		Customer: payments.Customer{
			ID:    userID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
	})
	if err != nil {
		utils.LogError("create payment checkout", err, "userId", userID, "provider", h.paymentProvider.Name())
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	tx.ProviderPaymentID = paymentResult.ProviderPaymentID
	tx.ProviderOrderID = paymentResult.ProviderOrderID
	tx.PaymentURL = paymentResult.PaymentURL

	if err := h.transactionRepo.Put(ctx, tx); err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp := models.PaymentResponse{
		TransactionID: tx.TransactionID,
		PaymentURL:    paymentResult.PaymentURL,
		OrderID:       paymentResult.ProviderOrderID,
		Amount:        amount,
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// paymentCallback handles payment notification webhooks.
func (h *PremiumHandler) paymentCallback(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	callbackResult, err := h.paymentProvider.ParseCallback(req.Headers, req.Path, []byte(req.Body))
	if err != nil {
		utils.LogWarn("payment callback rejected", "provider", h.paymentProvider.Name(), "error", err.Error())
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	tx, err := h.transactionRepo.GetByProviderOrderID(ctx, callbackResult.ProviderOrderID)
	if err != nil || tx == nil {
		utils.LogWarn("transaction not found for callback", "provider", h.paymentProvider.Name(), "orderId", callbackResult.ProviderOrderID)
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	switch callbackResult.Status {
	case payments.PaymentStatusSucceeded:
		if err := h.handleSuccessfulPayment(ctx, tx); err != nil {
			utils.LogError("handle successful payment", err, "transactionId", tx.TransactionID)
			return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
		}
	case payments.PaymentStatusFailed:
		_ = h.transactionRepo.UpdateStatus(ctx, tx.TransactionID, tx.SK, models.TransactionStatusFailed)
	}

	return jsonResponse(http.StatusOK, `{"status":"ok"}`), nil
}

// handleSuccessfulPayment fulfils the transaction: upgrade user tier or feature a listing.
func (h *PremiumHandler) handleSuccessfulPayment(ctx context.Context, tx *models.Transaction) error {
	if tx.Status == models.TransactionStatusCompleted {
		return nil
	}

	if err := h.transactionRepo.UpdateStatus(ctx, tx.TransactionID, tx.SK, models.TransactionStatusCompleted); err != nil {
		return err
	}

	switch tx.Type {
	case models.TransactionTypePremiumTier:
		user, err := h.userRepo.GetByID(ctx, tx.UserID)
		if err != nil || user == nil {
			return fmt.Errorf("user not found for premium upgrade: %s", tx.UserID)
		}
		renewDate := time.Now().UTC().AddDate(0, 1, 0)
		user.Subscription.Tier = models.SubscriptionPremium
		user.Subscription.RenewDate = &renewDate
		return h.userRepo.Put(ctx, user)

	case models.TransactionTypeFeatured, models.TransactionTypePromotion:
		if tx.ListingID == "" {
			return nil
		}
		days := 7 // default duration
		if tx.Metadata != nil {
			if d, ok := tx.Metadata["durationDays"]; ok {
				if n, err := strconv.Atoi(d); err == nil && n > 0 {
					days = n
				} else {
					utils.LogWarn("invalid durationDays in metadata, using default", "value", d, "transactionId", tx.TransactionID)
				}
			}
		}
		featureType := "featured"
		if tx.Type == models.TransactionTypePromotion {
			featureType = "promotion"
		}
		until := time.Now().UTC().AddDate(0, 0, days)
		return h.listingService.FeatureListing(ctx, tx.UserID, tx.ListingID, featureType, until)
	}

	return nil
}

func buildPremiumCallbackURL() string {
	baseURL := strings.TrimRight(os.Getenv("PUBLIC_API_BASE_URL"), "/")
	if baseURL == "" {
		return ""
	}
	return baseURL + "/premium/callback"
}
