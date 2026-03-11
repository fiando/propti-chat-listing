package handlers

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
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
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

// PremiumHandler handles subscription upgrades, featured listings, and payment callbacks.
type PremiumHandler struct {
	userRepo        *repository.UserRepo
	listingRepo     *repository.ListingRepo
	transactionRepo *repository.TransactionRepo
	listingService  *services.ListingService
}

// NewPremiumHandler creates a PremiumHandler.
func NewPremiumHandler(
	userRepo *repository.UserRepo,
	listingRepo *repository.ListingRepo,
	transactionRepo *repository.TransactionRepo,
	listingService *services.ListingService,
) *PremiumHandler {
	return &PremiumHandler{
		userRepo:        userRepo,
		listingRepo:     listingRepo,
		transactionRepo: transactionRepo,
		listingService:  listingService,
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
		return h.midtransCallback(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

// upgradeToPremium initiates a Midtrans payment to upgrade the user to premium.
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
		PK:            uuid.NewString(),
		SK:            time.Now().UTC().Format(time.RFC3339),
		TransactionID: uuid.NewString(),
		UserID:        userID,
		Type:          models.TransactionTypePremiumTier,
		Amount:        amount,
		Currency:      "IDR",
		Status:        models.TransactionStatusPending,
		MidtransOrderID: orderID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	tx.PK = tx.TransactionID

	paymentURL, midtransID, err := createMidtransSnapToken(ctx, orderID, amount, user)
	if err != nil {
		utils.LogError("create midtrans snap token", err, "userId", userID)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	tx.MidtransPaymentID = midtransID
	tx.PaymentURL = paymentURL

	if err := h.transactionRepo.Put(ctx, tx); err != nil {
		utils.LogError("save transaction", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp := models.PaymentResponse{
		TransactionID: tx.TransactionID,
		PaymentURL:    paymentURL,
		OrderID:       orderID,
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
		MidtransOrderID: orderID,
		Metadata: map[string]string{
			"featureType":  featReq.Type,
			"durationDays": fmt.Sprintf("%d", featReq.DurationDays),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	tx.PK = tx.TransactionID
	tx.SK = tx.CreatedAt.Format(time.RFC3339)

	paymentURL, midtransID, err := createMidtransSnapToken(ctx, orderID, amount, user)
	if err != nil {
		utils.LogError("create midtrans snap token", err, "userId", userID)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	tx.MidtransPaymentID = midtransID
	tx.PaymentURL = paymentURL

	if err := h.transactionRepo.Put(ctx, tx); err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp := models.PaymentResponse{
		TransactionID: tx.TransactionID,
		PaymentURL:    paymentURL,
		OrderID:       orderID,
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// midtransCallback handles Midtrans payment notification webhooks.
func (h *PremiumHandler) midtransCallback(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var notification midtransNotification
	if err := json.Unmarshal([]byte(req.Body), &notification); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	// Verify signature key per Midtrans spec:
	// SHA-512(order_id + status_code + gross_amount + server_key)
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	raw := notification.OrderID + notification.StatusCode + notification.GrossAmount + serverKey
	h512 := sha512.New()
	h512.Write([]byte(raw))
	expectedSig := hex.EncodeToString(h512.Sum(nil))

	if notification.SignatureKey != expectedSig {
		utils.LogWarn("midtrans signature mismatch", "orderId", notification.OrderID)
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	tx, err := h.transactionRepo.GetByMidtransOrderID(ctx, notification.OrderID)
	if err != nil || tx == nil {
		utils.LogWarn("transaction not found for callback", "orderId", notification.OrderID)
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	switch notification.TransactionStatus {
	case "capture", "settlement":
		if err := h.handleSuccessfulPayment(ctx, tx); err != nil {
			utils.LogError("handle successful payment", err, "transactionId", tx.TransactionID)
			return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
		}
	case "deny", "cancel", "expire":
		_ = h.transactionRepo.UpdateStatus(ctx, tx.TransactionID, tx.SK, models.TransactionStatusFailed)
	}

	return jsonResponse(http.StatusOK, `{"status":"ok"}`), nil
}

// handleSuccessfulPayment fulfils the transaction: upgrade user tier or feature a listing.
func (h *PremiumHandler) handleSuccessfulPayment(ctx context.Context, tx *models.Transaction) error {
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

// --- Midtrans Snap API integration ---

type midtransSnapRequest struct {
	TransactionDetails struct {
		OrderID     string  `json:"order_id"`
		GrossAmount float64 `json:"gross_amount"`
	} `json:"transaction_details"`
	CustomerDetails struct {
		FirstName string `json:"first_name"`
		Email     string `json:"email"`
	} `json:"customer_details"`
}

type midtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type midtransNotification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
}

// createMidtransSnapToken calls the Midtrans Snap API and returns the redirect URL and token.
func createMidtransSnapToken(_ context.Context, orderID string, amount float64, user *models.User) (paymentURL, token string, err error) {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	isProduction := os.Getenv("MIDTRANS_ENV") == "production"

	baseURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if isProduction {
		baseURL = "https://app.midtrans.com/snap/v1/transactions"
	}

	snapReq := midtransSnapRequest{}
	snapReq.TransactionDetails.OrderID = orderID
	snapReq.TransactionDetails.GrossAmount = amount
	snapReq.CustomerDetails.FirstName = user.Name
	snapReq.CustomerDetails.Email = user.Email

	payload, err := json.Marshal(snapReq)
	if err != nil {
		return "", "", fmt.Errorf("marshal snap request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Midtrans uses HTTP Basic Auth with server key as username and empty password.
	httpReq.SetBasicAuth(serverKey, "")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("snap api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("snap api returned status %d", resp.StatusCode)
	}

	var snapResp midtransSnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		return "", "", fmt.Errorf("decode snap response: %w", err)
	}

	if snapResp.RedirectURL == "" {
		snapResp.RedirectURL = buildSnapRedirectURL(snapResp.Token, isProduction)
	}

	return snapResp.RedirectURL, snapResp.Token, nil
}

func buildSnapRedirectURL(token string, isProduction bool) string {
	base := "https://app.sandbox.midtrans.com/snap/v2/vtweb/"
	if isProduction {
		base = "https://app.midtrans.com/snap/v2/vtweb/"
	}
	return base + strings.TrimSpace(token)
}
