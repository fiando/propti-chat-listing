package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const dokuCheckoutPath = "/checkout/v1/payment"

type DOKUConfig struct {
	ClientID   string
	SecretKey  string
	BaseURL    string
	HTTPClient *http.Client
}

type DOKUProvider struct {
	clientID   string
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

func NewDOKUProvider(cfg DOKUConfig) *DOKUProvider {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api-sandbox.doku.com"
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &DOKUProvider{
		clientID:   cfg.ClientID,
		secretKey:  cfg.SecretKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (p *DOKUProvider) Name() string {
	return ProviderDOKU
}

func (p *DOKUProvider) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	payload := map[string]any{
		"order": map[string]any{
			"invoice_number": input.OrderID,
			"amount":         input.Amount,
			"currency":       defaultString(input.Currency, "IDR"),
		},
		"payment": map[string]any{
			"type":             "SALE",
			"payment_due_date": 60,
		},
		"customer": map[string]any{
			"id":    input.Customer.ID,
			"name":  input.Customer.Name,
			"email": input.Customer.Email,
			"phone": normalizeDOKUPhone(input.Customer.Phone, p.baseURL),
		},
	}

	if strings.TrimSpace(input.Description) != "" {
		payload["order"].(map[string]any)["line_items"] = []map[string]any{
			{
				"id":       input.OrderID,
				"name":     input.Description,
				"quantity": 1,
				"price":    input.Amount,
			},
		}
	}
	if strings.TrimSpace(input.NotificationURL) != "" {
		payload["additional_info"] = map[string]any{
			"override_notification_url": input.NotificationURL,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal doku checkout request: %w", err)
	}

	requestID := uuid.NewString()
	requestTimestamp := time.Now().UTC().Format(time.RFC3339)
	signature := buildSignature(p.clientID, requestID, requestTimestamp, dokuCheckoutPath, string(body), p.secretKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+dokuCheckoutPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create doku request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Client-Id", p.clientID)
	req.Header.Set("Request-Id", requestID)
	req.Header.Set("Request-Timestamp", requestTimestamp)
	req.Header.Set("Signature", signature)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doku checkout request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read doku response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("doku checkout returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var checkoutResponse struct {
		Response struct {
			Order struct {
				InvoiceNumber string `json:"invoice_number"`
			} `json:"order"`
			Payment struct {
				TokenID string `json:"token_id"`
				URL     string `json:"url"`
			} `json:"payment"`
		} `json:"response"`
	}
	if err := json.Unmarshal(responseBody, &checkoutResponse); err != nil {
		return nil, fmt.Errorf("decode doku response: %w", err)
	}
	if checkoutResponse.Response.Payment.URL == "" {
		return nil, fmt.Errorf("doku checkout response missing payment url")
	}

	return &CreatePaymentResult{
		Provider:          ProviderDOKU,
		ProviderOrderID:   defaultString(checkoutResponse.Response.Order.InvoiceNumber, input.OrderID),
		ProviderPaymentID: checkoutResponse.Response.Payment.TokenID,
		PaymentURL:        checkoutResponse.Response.Payment.URL,
	}, nil
}

func (p *DOKUProvider) GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/checkout/v1/payment/%s/check-status", p.baseURL, paymentID), nil)
	if err != nil {
		return PaymentStatusPending, fmt.Errorf("create doku status request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "bearer authnologin")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return PaymentStatusPending, fmt.Errorf("doku status request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PaymentStatusPending, fmt.Errorf("read doku status response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return PaymentStatusPending, fmt.Errorf("doku status returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var statusResponse struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(responseBody, &statusResponse); err != nil {
		return PaymentStatusPending, fmt.Errorf("decode doku status response: %w", err)
	}

	return mapDOKUStatus(statusResponse.Status), nil
}

func (p *DOKUProvider) ParseCallback(headers map[string]string, path string, body []byte) (*CallbackResult, error) {
	clientID := getHeader(headers, "Client-Id")
	requestID := getHeader(headers, "Request-Id")
	requestTimestamp := getHeader(headers, "Request-Timestamp")
	signature := getHeader(headers, "Signature")

	expected := buildSignature(clientID, requestID, requestTimestamp, path, string(body), p.secretKey)
	if signature == "" || !hmac.Equal([]byte(signature), []byte(expected)) {
		return nil, fmt.Errorf("invalid doku callback signature")
	}

	var notification struct {
		Order struct {
			InvoiceNumber string `json:"invoice_number"`
		} `json:"order"`
		Transaction struct {
			Status            string `json:"status"`
			OriginalRequestID string `json:"original_request_id"`
		} `json:"transaction"`
		AuthorizeID string `json:"authorize_id"`
	}
	if err := json.Unmarshal(body, &notification); err != nil {
		return nil, fmt.Errorf("decode doku callback: %w", err)
	}

	paymentID := strings.TrimSpace(notification.AuthorizeID)
	if paymentID == "" {
		paymentID = strings.TrimSpace(notification.Transaction.OriginalRequestID)
	}

	return &CallbackResult{
		Provider:          ProviderDOKU,
		ProviderOrderID:   notification.Order.InvoiceNumber,
		ProviderPaymentID: paymentID,
		Status:            mapDOKUStatus(notification.Transaction.Status),
	}, nil
}

func mapDOKUStatus(status string) PaymentStatus {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS", "PAID":
		return PaymentStatusSucceeded
	case "FAILED", "EXPIRED", "CANCEL", "CANCELLED":
		return PaymentStatusFailed
	default:
		return PaymentStatusPending
	}
}

func buildSignature(clientID, requestID, requestTimestamp, requestTarget, body, secret string) string {
	component := strings.Join([]string{
		"Client-Id:" + clientID,
		"Request-Id:" + requestID,
		"Request-Timestamp:" + requestTimestamp,
		"Request-Target:" + requestTarget,
		"Digest:" + generateDigest(body),
	}, "\n")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(component))
	return "HMACSHA256=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func generateDigest(body string) string {
	sum := sha256.Sum256([]byte(body))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func getHeader(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func normalizeDOKUPhone(phone, _ string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	switch {
	case strings.HasPrefix(digits, "0"):
		digits = "62" + strings.TrimPrefix(digits, "0")
	case strings.HasPrefix(digits, "8"):
		digits = "62" + digits
	}

	if len(digits) >= 5 && len(digits) <= 16 {
		return digits
	}
	return "6281234567890"
}
