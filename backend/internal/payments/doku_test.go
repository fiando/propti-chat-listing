package payments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDOKUProviderCreatePaymentBuildsSignedCheckoutRequest(t *testing.T) {
	t.Parallel()

	var receivedPath string
	var receivedHeaders http.Header
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedHeaders = r.Header.Clone()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		receivedBody = body

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"response": {
				"order": {
					"invoice_number": "INV-PREM-001"
				},
				"payment": {
					"token_id": "tok-123",
					"url": "https://sandbox.doku.com/checkout-link-v2/tok-123"
				}
			}
		}`))
	}))
	defer server.Close()

	provider := NewDOKUProvider(DOKUConfig{
		ClientID:  "client-123",
		SecretKey: "secret-123",
		BaseURL:   server.URL,
	})

	result, err := provider.CreatePayment(context.Background(), CreatePaymentInput{
		OrderID:         "INV-PREM-001",
		Amount:          49000,
		Currency:        "IDR",
		Description:     "Propti Premium",
		NotificationURL: "https://api.propti.test/premium/callback",
		CallbackURL:     "https://propti.test/profile#premium",
		ResultURL:       "https://propti.test/profile#premium",
		AutoRedirect:    true,
		Customer: Customer{
			ID:    "user-1",
			Name:  "Bobby",
			Email: "bob@example.com",
			Phone: "6281234567890",
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	if receivedPath != "/checkout/v1/payment" {
		t.Fatalf("expected request path /checkout/v1/payment, got %q", receivedPath)
	}
	if got := receivedHeaders.Get("Client-Id"); got != "client-123" {
		t.Fatalf("expected Client-Id header, got %q", got)
	}
	if got := receivedHeaders.Get("Signature"); !strings.HasPrefix(got, "HMACSHA256=") {
		t.Fatalf("expected Signature header to use HMACSHA256, got %q", got)
	}

	var payload map[string]any
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("request body is not valid json: %v", err)
	}

	order := payload["order"].(map[string]any)
	if got := order["invoice_number"]; got != "INV-PREM-001" {
		t.Fatalf("expected invoice_number INV-PREM-001, got %#v", got)
	}
	if got := order["amount"]; got != float64(49000) {
		t.Fatalf("expected amount 49000, got %#v", got)
	}
	if got := order["callback_url"]; got != "https://propti.test/profile#premium" {
		t.Fatalf("expected callback_url to be forwarded, got %#v", got)
	}
	if got := order["callback_url_result"]; got != "https://propti.test/profile#premium" {
		t.Fatalf("expected callback_url_result to be forwarded, got %#v", got)
	}
	if got := order["language"]; got != "ID" {
		t.Fatalf("expected language to default to Indonesian (ID), got %#v", got)
	}
	if got := order["auto_redirect"]; got != true {
		t.Fatalf("expected auto_redirect to be true, got %#v", got)
	}

	additionalInfo := payload["additional_info"].(map[string]any)
	if got := additionalInfo["override_notification_url"]; got != "https://api.propti.test/premium/callback" {
		t.Fatalf("expected override_notification_url to be forwarded, got %#v", got)
	}

	customer := payload["customer"].(map[string]any)
	if got := customer["email"]; got != "bob@example.com" {
		t.Fatalf("expected customer email to be forwarded, got %#v", got)
	}

	if result.Provider != ProviderDOKU {
		t.Fatalf("expected provider %q, got %q", ProviderDOKU, result.Provider)
	}
	if result.OrderID != "INV-PREM-001" {
		t.Fatalf("expected provider order id INV-PREM-001, got %q", result.OrderID)
	}
	if result.PaymentID != "tok-123" {
		t.Fatalf("expected provider payment id tok-123, got %q", result.PaymentID)
	}
	if result.PaymentURL != "https://sandbox.doku.com/checkout-link-v2/tok-123" {
		t.Fatalf("unexpected payment url: %q", result.PaymentURL)
	}
}

func TestDOKUProviderCreatePaymentOmitsCustomerPhoneWhenProvided(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		receivedBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"order":{"invoice_number":"INV-PREM-002"},"payment":{"token_id":"tok-456","url":"https://sandbox.doku.com/checkout-link-v2/tok-456"}}}`))
	}))
	defer server.Close()

	provider := NewDOKUProvider(DOKUConfig{
		ClientID:  "client-123",
		SecretKey: "secret-123",
		BaseURL:   server.URL,
	})

	_, err := provider.CreatePayment(context.Background(), CreatePaymentInput{
		OrderID:         "INV-PREM-002",
		Amount:          49000,
		Currency:        "IDR",
		Description:     "Propti Premium",
		NotificationURL: "https://api.propti.test/premium/callback",
		Customer: Customer{
			ID:    "user-1",
			Name:  "Bobby",
			Email: "bob@example.com",
			Phone: "+62 812-3456-7890",
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("request body is not valid json: %v", err)
	}
	customer := payload["customer"].(map[string]any)
	if _, exists := customer["phone"]; exists {
		t.Fatalf("expected customer phone to be omitted, got %#v", customer["phone"])
	}
}

func TestDOKUProviderCreatePaymentOmitsCustomerPhoneWhenMissing(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		receivedBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"order":{"invoice_number":"INV-PREM-003"},"payment":{"token_id":"tok-789","url":"https://sandbox.doku.com/checkout-link-v2/tok-789"}}}`))
	}))
	defer server.Close()

	provider := NewDOKUProvider(DOKUConfig{
		ClientID:  "client-123",
		SecretKey: "secret-123",
		BaseURL:   server.URL,
	})

	_, err := provider.CreatePayment(context.Background(), CreatePaymentInput{
		OrderID:     "INV-PREM-003",
		Amount:      49000,
		Currency:    "IDR",
		Description: "Propti Premium",
		Customer: Customer{
			ID:    "user-1",
			Name:  "Bobby",
			Email: "bob@example.com",
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("request body is not valid json: %v", err)
	}
	customer := payload["customer"].(map[string]any)
	if _, exists := customer["phone"]; exists {
		t.Fatalf("expected customer phone to be omitted, got %#v", customer["phone"])
	}
}

func TestDOKUProviderGetPaymentStatusMapsPaidToSucceeded(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/checkout/v1/payment/tok-123/check-status" {
			t.Fatalf("expected request path /checkout/v1/payment/tok-123/check-status, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PAID"}`))
	}))
	defer server.Close()

	provider := NewDOKUProvider(DOKUConfig{BaseURL: server.URL})
	status, err := provider.GetPaymentStatus(context.Background(), "tok-123")
	if err != nil {
		t.Fatalf("GetPaymentStatus returned error: %v", err)
	}
	if status != PaymentStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", status)
	}
}

func TestDOKUProviderParseCallbackValidatesSignatureAndMapsSuccess(t *testing.T) {
	t.Parallel()

	provider := NewDOKUProvider(DOKUConfig{
		ClientID:  "client-123",
		SecretKey: "secret-123",
	})

	body := []byte(`{
		"order": {
			"invoice_number": "INV-PREM-001",
			"amount": 49000
		},
		"transaction": {
			"status": "SUCCESS",
			"original_request_id": "req-123"
		},
		"service": {
			"id": "EMONEY"
		},
		"channel": {
			"id": "QRIS"
		},
		"authorize_id": "auth-789"
	}`)

	headers := map[string]string{
		"Client-Id":         "client-123",
		"Request-Id":        "notif-123",
		"Request-Timestamp": "2026-03-16T05:45:00Z",
	}
	headers["Signature"] = buildSignature(headers["Client-Id"], headers["Request-Id"], headers["Request-Timestamp"], "/premium/callback", string(body), "secret-123")

	result, err := provider.ParseCallback(headers, "/premium/callback", body)
	if err != nil {
		t.Fatalf("ParseCallback returned error: %v", err)
	}

	if result.Provider != ProviderDOKU {
		t.Fatalf("expected provider %q, got %q", ProviderDOKU, result.Provider)
	}
	if result.OrderID != "INV-PREM-001" {
		t.Fatalf("expected invoice number INV-PREM-001, got %q", result.OrderID)
	}
	if result.PaymentID != "auth-789" {
		t.Fatalf("expected payment id auth-789, got %q", result.PaymentID)
	}
	if result.Status != PaymentStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", result.Status)
	}
}

func TestDOKUProviderParseCallbackRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	provider := NewDOKUProvider(DOKUConfig{
		ClientID:  "client-123",
		SecretKey: "secret-123",
	})

	_, err := provider.ParseCallback(map[string]string{
		"Client-Id":         "client-123",
		"Request-Id":        "notif-123",
		"Request-Timestamp": "2026-03-16T05:45:00Z",
		"Signature":         "HMACSHA256=invalid",
	}, "/premium/callback", []byte(`{"order":{"invoice_number":"INV-PREM-001"},"transaction":{"status":"SUCCESS"}}`))
	if err == nil {
		t.Fatal("expected ParseCallback to reject an invalid signature")
	}
}
