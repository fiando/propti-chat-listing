package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

func TestParseWhatsAppProviderKindDefaultsToTwilio(t *testing.T) {
	provider, err := parseWhatsAppProviderKind("")
	if err != nil {
		t.Fatalf("parseWhatsAppProviderKind returned error: %v", err)
	}
	if provider != models.WhatsAppProviderTwilio {
		t.Fatalf("expected default provider %q, got %q", models.WhatsAppProviderTwilio, provider)
	}
}

func TestParseWhatsAppProviderKindRejectsUnknownValues(t *testing.T) {
	_, err := parseWhatsAppProviderKind("unexpected-provider")
	if err == nil {
		t.Fatal("expected unknown provider to return error")
	}
}

func TestTwilioWhatsAppCapabilities(t *testing.T) {
	capabilities := TwilioWhatsAppCapabilities()

	if !capabilities.SupportsText || !capabilities.SupportsTemplate || !capabilities.SupportsMedia {
		t.Fatalf("expected Twilio capabilities to support text/template/media, got %#v", capabilities)
	}
	if !capabilities.SupportsWebhookVerification || !capabilities.SupportsDeliveryStatus || !capabilities.SupportsInbound {
		t.Fatalf("expected Twilio capabilities to support webhook verification, delivery status, and inbound events, got %#v", capabilities)
	}
}

func TestDirectWhatsAppCapabilities(t *testing.T) {
	capabilities := DirectWhatsAppCapabilities()

	if !capabilities.SupportsText || !capabilities.SupportsMedia {
		t.Fatalf("expected direct WhatsApp capabilities to support text and media, got %#v", capabilities)
	}
	if capabilities.SupportsTemplate {
		t.Fatalf("expected direct WhatsApp capabilities to disable templates by default, got %#v", capabilities)
	}
}

func TestTwilioWhatsAppProviderSendText(t *testing.T) {
	fixedNow := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		if got, want := r.URL.Path, "/2010-04-01/Accounts/AC123/Messages.json"; got != want {
			t.Fatalf("expected path %s, got %s", want, got)
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != "AC123" || password != "secret" {
			t.Fatalf("expected basic auth AC123/secret, got ok=%v username=%q password=%q", ok, username, password)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.FormValue("To"); got != "whatsapp:+62812" {
			t.Fatalf("expected To to be normalized with whatsapp prefix, got %q", got)
		}
		if got := r.FormValue("From"); got != "whatsapp:+1415" {
			t.Fatalf("expected From to be normalized with whatsapp prefix, got %q", got)
		}
		if got := r.FormValue("Body"); got != "halo" {
			t.Fatalf("expected Body to be forwarded, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"sid":"SM123","status":"queued"}`)
	}))
	defer server.Close()

	provider := NewTwilioWhatsAppProvider(TwilioWhatsAppProviderConfig{
		AccountSID: "AC123",
		AuthToken:  "secret",
		From:       "+1415",
		APIBaseURL: server.URL,
		Clock: func() time.Time {
			return fixedNow
		},
	})

	result, err := provider.Send(context.Background(), models.WhatsAppSendRequest{
		To: "+62812",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: "halo",
		},
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if result.Provider != models.WhatsAppProviderTwilio {
		t.Fatalf("expected provider %q, got %q", models.WhatsAppProviderTwilio, result.Provider)
	}
	if result.ProviderMessageID != "SM123" {
		t.Fatalf("expected provider message id SM123, got %q", result.ProviderMessageID)
	}
	if !result.AcceptedAt.Equal(fixedNow) {
		t.Fatalf("expected acceptedAt %s, got %s", fixedNow, result.AcceptedAt)
	}
}

func TestTwilioWhatsAppProviderSendTemplateRequiresContentSID(t *testing.T) {
	provider := NewTwilioWhatsAppProvider(TwilioWhatsAppProviderConfig{AccountSID: "AC123", AuthToken: "secret"})

	_, err := provider.Send(context.Background(), models.WhatsAppSendRequest{
		To: "+62812",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeTemplate,
			Template: &models.WhatsAppTemplatePayload{
				Name: "welcome",
			},
		},
	})
	if err == nil {
		t.Fatal("expected template send to fail without twilio_content_sid metadata")
	}
	if !strings.Contains(err.Error(), "twilio_content_sid") {
		t.Fatalf("expected error to mention twilio_content_sid, got %v", err)
	}
}

func TestTwilioWhatsAppProviderParseInboundWebhook(t *testing.T) {
	provider := NewTwilioWhatsAppProvider(TwilioWhatsAppProviderConfig{})

	body := url.Values{}
	body.Set("MessageSid", "SM999")
	body.Set("From", "whatsapp:+62812")
	body.Set("To", "whatsapp:+1415")
	body.Set("Body", "hello from buyer")
	body.Set("NumMedia", "1")
	body.Set("MediaUrl0", "https://cdn.example.com/image.jpg")
	body.Set("MediaContentType0", "image/jpeg")
	body.Set("ProfileName", "Buyer Name")

	req := httptest.NewRequest(http.MethodPost, "https://example.com/twilio/inbound", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	envelope, err := provider.ParseInboundWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("ParseInboundWebhook returned error: %v", err)
	}
	if envelope.Provider != models.WhatsAppProviderTwilio {
		t.Fatalf("expected provider %q, got %q", models.WhatsAppProviderTwilio, envelope.Provider)
	}
	if envelope.ProviderMessageID != "SM999" {
		t.Fatalf("expected provider message id SM999, got %q", envelope.ProviderMessageID)
	}
	if envelope.Type != models.WhatsAppMessageTypeMedia {
		t.Fatalf("expected message type media, got %q", envelope.Type)
	}
	if len(envelope.Media) != 1 || envelope.Media[0].URL != "https://cdn.example.com/image.jpg" {
		t.Fatalf("expected media payload to be normalized, got %#v", envelope.Media)
	}
	if envelope.Metadata["profile_name"] != "Buyer Name" {
		t.Fatalf("expected metadata profile_name to be populated, got %#v", envelope.Metadata)
	}
}

func TestTwilioWhatsAppProviderParseDeliveryStatusWebhook(t *testing.T) {
	fixedNow := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	provider := NewTwilioWhatsAppProvider(TwilioWhatsAppProviderConfig{Clock: func() time.Time { return fixedNow }})

	body := url.Values{}
	body.Set("MessageSid", "SM555")
	body.Set("MessageStatus", "undelivered")
	body.Set("ErrorCode", "30007")
	body.Set("ErrorMessage", "Carrier violation")

	req := httptest.NewRequest(http.MethodPost, "https://example.com/twilio/status", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	events, err := provider.ParseDeliveryStatusWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("ParseDeliveryStatusWebhook returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 status event, got %d", len(events))
	}
	event := events[0]
	if event.Status != models.WhatsAppDeliveryStatusFailed {
		t.Fatalf("expected mapped status failed, got %q", event.Status)
	}
	if event.ErrorCode != "30007" || event.ErrorMessage != "Carrier violation" {
		t.Fatalf("expected error details to be preserved, got %#v", event)
	}
	if !event.OccurredAt.Equal(fixedNow) {
		t.Fatalf("expected occurredAt %s, got %s", fixedNow, event.OccurredAt)
	}
}

func TestTwilioWhatsAppProviderVerifyWebhookRequest(t *testing.T) {
	provider := NewTwilioWhatsAppProvider(TwilioWhatsAppProviderConfig{AuthToken: "secret"})

	body := url.Values{}
	body.Set("MessageSid", "SM123")
	body.Set("Body", "hello")
	req := httptest.NewRequest(http.MethodPost, "https://example.com/twilio/inbound", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	signature, err := provider.computeWebhookSignature(req)
	if err != nil {
		t.Fatalf("computeWebhookSignature returned error: %v", err)
	}
	req.Header.Set("X-Twilio-Signature", signature)

	if err := provider.VerifyWebhookRequest(req); err != nil {
		t.Fatalf("VerifyWebhookRequest returned error: %v", err)
	}

	reqBad := httptest.NewRequest(http.MethodPost, "https://example.com/twilio/inbound", strings.NewReader(body.Encode()))
	reqBad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqBad.Header.Set("X-Twilio-Signature", "invalid")
	if err := provider.VerifyWebhookRequest(reqBad); err == nil {
		t.Fatal("expected invalid signature to be rejected")
	}
}
