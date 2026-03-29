package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

func TestMetaWhatsAppProviderSendText(t *testing.T) {
	fixedNow := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		if got, want := r.URL.Path, "/v23.0/PHONE123/messages"; got != want {
			t.Fatalf("expected path %s, got %s", want, got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("expected bearer token header, got %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("json decode failed: %v", err)
		}
		if got := payload["to"]; got != "+62812" {
			t.Fatalf("expected normalized to field, got %#v", got)
		}
		if got := payload["type"]; got != "text" {
			t.Fatalf("expected text message type payload, got %#v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"messages":[{"id":"wamid.123"}]}`)
	}))
	defer server.Close()

	provider := NewMetaWhatsAppProvider(MetaWhatsAppProviderConfig{
		AccessToken:   "token-123",
		PhoneNumberID: "PHONE123",
		APIBaseURL:    server.URL,
		APIVersion:    "v23.0",
		Clock: func() time.Time {
			return fixedNow
		},
	})

	result, err := provider.Send(context.Background(), models.WhatsAppSendRequest{
		To: "+62812",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: "Halo",
		},
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if result.Provider != models.WhatsAppProviderDirectWhatsApp {
		t.Fatalf("expected provider %q, got %q", models.WhatsAppProviderDirectWhatsApp, result.Provider)
	}
	if result.ProviderMessageID != "wamid.123" {
		t.Fatalf("expected provider message id wamid.123, got %q", result.ProviderMessageID)
	}
	if !result.AcceptedAt.Equal(fixedNow) {
		t.Fatalf("expected acceptedAt %s, got %s", fixedNow, result.AcceptedAt)
	}
}

func TestMetaWhatsAppProviderSendTemplateReturnsActionableError(t *testing.T) {
	provider := NewMetaWhatsAppProvider(MetaWhatsAppProviderConfig{AccessToken: "token", PhoneNumberID: "PHONE123"})

	_, err := provider.Send(context.Background(), models.WhatsAppSendRequest{
		To: "+62812",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeTemplate,
		},
	})
	if err == nil {
		t.Fatal("expected template send to fail for direct WhatsApp provider")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "template") {
		t.Fatalf("expected actionable template error, got %v", err)
	}
}

func TestMetaWhatsAppProviderParseInboundWebhook(t *testing.T) {
	provider := NewMetaWhatsAppProvider(MetaWhatsAppProviderConfig{})
	payload := `{"object":"whatsapp_business_account","entry":[{"id":"entry-1","changes":[{"field":"messages","value":{"metadata":{"display_phone_number":"15550001","phone_number_id":"PHONE123"},"contacts":[{"profile":{"name":"Buyer Name"},"wa_id":"62812"}],"messages":[{"from":"62812","id":"wamid.111","timestamp":"1710000000","type":"text","text":{"body":"Hai"}}]}}]}]}`

	req := httptest.NewRequest(http.MethodPost, "https://example.com/meta/webhook", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	envelope, err := provider.ParseInboundWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("ParseInboundWebhook returned error: %v", err)
	}
	if envelope.Provider != models.WhatsAppProviderDirectWhatsApp {
		t.Fatalf("expected provider %q, got %q", models.WhatsAppProviderDirectWhatsApp, envelope.Provider)
	}
	if envelope.ProviderMessageID != "wamid.111" {
		t.Fatalf("expected provider message id wamid.111, got %q", envelope.ProviderMessageID)
	}
	if envelope.Type != models.WhatsAppMessageTypeText {
		t.Fatalf("expected text message type, got %q", envelope.Type)
	}
	if envelope.Text != "Hai" {
		t.Fatalf("expected text payload to be normalized, got %q", envelope.Text)
	}
	if envelope.Metadata["contact_name"] != "Buyer Name" {
		t.Fatalf("expected contact metadata, got %#v", envelope.Metadata)
	}
}

func TestMetaWhatsAppProviderParseDeliveryStatusWebhook(t *testing.T) {
	provider := NewMetaWhatsAppProvider(MetaWhatsAppProviderConfig{})
	payload := `{"object":"whatsapp_business_account","entry":[{"id":"entry-1","changes":[{"field":"messages","value":{"statuses":[{"id":"wamid.111","status":"failed","timestamp":"1710000010","errors":[{"code":131026,"title":"Message undeliverable","message":"Business phone not eligible"}]}]}}]}]}`

	req := httptest.NewRequest(http.MethodPost, "https://example.com/meta/webhook", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	events, err := provider.ParseDeliveryStatusWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("ParseDeliveryStatusWebhook returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one status event, got %d", len(events))
	}
	event := events[0]
	if event.Status != models.WhatsAppDeliveryStatusFailed {
		t.Fatalf("expected failed status mapping, got %q", event.Status)
	}
	if event.ErrorCode != "131026" {
		t.Fatalf("expected error code to be preserved, got %q", event.ErrorCode)
	}
	if !strings.Contains(event.ErrorMessage, "Message undeliverable") {
		t.Fatalf("expected error message to include provider title, got %q", event.ErrorMessage)
	}
}

func TestMetaWhatsAppProviderVerifyWebhookRequest(t *testing.T) {
	provider := NewMetaWhatsAppProvider(MetaWhatsAppProviderConfig{AppSecret: "meta-secret"})
	payload := `{"entry":[{"changes":[]}]}`
	req := httptest.NewRequest(http.MethodPost, "https://example.com/meta/webhook", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", metaSignature("meta-secret", payload))

	if err := provider.VerifyWebhookRequest(req); err != nil {
		t.Fatalf("VerifyWebhookRequest returned error: %v", err)
	}

	reqInvalid := httptest.NewRequest(http.MethodPost, "https://example.com/meta/webhook", strings.NewReader(payload))
	reqInvalid.Header.Set("Content-Type", "application/json")
	reqInvalid.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	if err := provider.VerifyWebhookRequest(reqInvalid); err == nil {
		t.Fatal("expected invalid signature to be rejected")
	}
}

func metaSignature(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
