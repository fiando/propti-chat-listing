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
	"github.com/fiando/propti/backend/internal/services"
)

type fakeWebhookProvider struct {
	verifyErr error

	inboundEnvelope *models.WhatsAppMessageEnvelope
	inboundErr      error

	statusEvents []models.WhatsAppDeliveryStatusEvent
	statusErr    error

	verifyCalls int
}

func (f *fakeWebhookProvider) ParseInboundWebhook(_ context.Context, _ *http.Request) (*models.WhatsAppMessageEnvelope, error) {
	return f.inboundEnvelope, f.inboundErr
}

func (f *fakeWebhookProvider) ParseDeliveryStatusWebhook(_ context.Context, _ *http.Request) ([]models.WhatsAppDeliveryStatusEvent, error) {
	return f.statusEvents, f.statusErr
}

func (f *fakeWebhookProvider) VerifyWebhookRequest(_ *http.Request) error {
	f.verifyCalls++
	return f.verifyErr
}

type fakeWhatsAppUserStore struct {
	user *models.User
	err  error
}

func (f *fakeWhatsAppUserStore) GetByWhatsAppPhone(_ context.Context, _ string) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

type fakeWhatsAppCommandOrchestrator struct {
	resp      *services.WhatsAppCommandResponse
	err       error
	lastReq   services.WhatsAppCommandRequest
	callCount int
}

func (f *fakeWhatsAppCommandOrchestrator) HandleText(_ context.Context, req services.WhatsAppCommandRequest) (*services.WhatsAppCommandResponse, error) {
	f.callCount++
	f.lastReq = req
	return f.resp, f.err
}

type fakeWhatsAppVoiceService struct {
	resp      *services.WhatsAppVoiceResponse
	err       error
	lastReq   services.WhatsAppVoiceRequest
	callCount int
}

func (f *fakeWhatsAppVoiceService) HandleInboundVoice(_ context.Context, req services.WhatsAppVoiceRequest) (*services.WhatsAppVoiceResponse, error) {
	f.callCount++
	f.lastReq = req
	return f.resp, f.err
}

type fakeWhatsAppPolicy struct {
	recordErr error
	recorded  *models.WhatsAppMessageEnvelope
}

func (f *fakeWhatsAppPolicy) RecordInboundMessage(_ context.Context, envelope models.WhatsAppMessageEnvelope) error {
	f.recorded = &envelope
	return f.recordErr
}

type fakeStatusSink struct {
	events []models.WhatsAppDeliveryStatusEvent
	err    error
}

func (f *fakeStatusSink) HandleDeliveryStatus(_ context.Context, event models.WhatsAppDeliveryStatusEvent) error {
	f.events = append(f.events, event)
	return f.err
}

func TestWhatsAppHandlerMetaWebhookVerificationHandshake(t *testing.T) {
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:        &fakeWebhookProvider{},
		MetaVerifyToken: "meta-verify-token",
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/whatsapp/webhook",
		QueryStringParameters: map[string]string{
			"hub.mode":         "subscribe",
			"hub.verify_token": "meta-verify-token",
			"hub.challenge":    "challenge-value",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for valid meta challenge, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if resp.Body != "challenge-value" {
		t.Fatalf("expected challenge body to be echoed, got %q", resp.Body)
	}
}

func TestWhatsAppHandlerMetaWebhookVerificationRejectsInvalidToken(t *testing.T) {
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:        &fakeWebhookProvider{},
		MetaVerifyToken: "meta-verify-token",
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/whatsapp/webhook",
		QueryStringParameters: map[string]string{
			"hub.mode":         "subscribe",
			"hub.verify_token": "wrong-token",
			"hub.challenge":    "challenge-value",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for invalid token, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func TestWhatsAppHandlerInboundTextRoutesToCommandOrchestrator(t *testing.T) {
	provider := &fakeWebhookProvider{
		inboundEnvelope: &models.WhatsAppMessageEnvelope{
			Provider:          models.WhatsAppProviderTwilio,
			ProviderMessageID: "SM123",
			From:              "whatsapp:+628123456789",
			To:                "whatsapp:+14155550000",
			Type:              models.WhatsAppMessageTypeText,
			Text:              "cari rumah di sleman",
			ReceivedAt:        time.Now().UTC(),
		},
	}
	command := &fakeWhatsAppCommandOrchestrator{
		resp: &services.WhatsAppCommandResponse{
			Intent:  services.WhatsAppCommandIntentSearch,
			Message: "ok",
		},
	}
	policy := &fakeWhatsAppPolicy{}
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:            provider,
		UserStore:           &fakeWhatsAppUserStore{user: &models.User{UserID: "user-1"}},
		CommandOrchestrator: command,
		Policy:              policy,
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook/inbound",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"entry":[]}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if command.callCount != 1 {
		t.Fatalf("expected command orchestrator to be called once, got %d", command.callCount)
	}
	if command.lastReq.UserID != "user-1" {
		t.Fatalf("expected command request userID user-1, got %q", command.lastReq.UserID)
	}
	if command.lastReq.Text != "cari rumah di sleman" {
		t.Fatalf("expected command text to be forwarded, got %q", command.lastReq.Text)
	}
	if policy.recorded == nil || policy.recorded.ProviderMessageID != "SM123" {
		t.Fatalf("expected inbound event to be recorded, got %#v", policy.recorded)
	}
}

func TestWhatsAppHandlerInboundVoiceRoutesToVoiceService(t *testing.T) {
	provider := &fakeWebhookProvider{
		inboundEnvelope: &models.WhatsAppMessageEnvelope{
			Provider:          models.WhatsAppProviderTwilio,
			ProviderMessageID: "SMVOICE",
			From:              "whatsapp:+628123456789",
			To:                "whatsapp:+14155550000",
			Type:              models.WhatsAppMessageTypeMedia,
			Media: []models.WhatsAppMediaItem{
				{URL: "https://media.example.com/voice.ogg", MimeType: "audio/ogg"},
			},
			ReceivedAt: time.Now().UTC(),
		},
	}
	voice := &fakeWhatsAppVoiceService{
		resp: &services.WhatsAppVoiceResponse{Status: services.WhatsAppVoiceStatusProcessed},
	}
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:     provider,
		UserStore:    &fakeWhatsAppUserStore{user: &models.User{UserID: "user-1"}},
		VoiceService: voice,
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook/inbound",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"entry":[]}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if voice.callCount != 1 {
		t.Fatalf("expected voice service to be called once, got %d", voice.callCount)
	}
	if voice.lastReq.UserID != "user-1" {
		t.Fatalf("expected voice request userID user-1, got %q", voice.lastReq.UserID)
	}
}

func TestWhatsAppHandlerStatusCallbackParsesAndDispatchesEvents(t *testing.T) {
	provider := &fakeWebhookProvider{
		statusEvents: []models.WhatsAppDeliveryStatusEvent{
			{ProviderMessageID: "SM1", Status: models.WhatsAppDeliveryStatusDelivered},
			{ProviderMessageID: "SM2", Status: models.WhatsAppDeliveryStatusFailed},
		},
	}
	sink := &fakeStatusSink{}
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:   provider,
		StatusSink: sink,
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook/status",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"entry":[]}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if len(sink.events) != 2 {
		t.Fatalf("expected 2 status events forwarded, got %d", len(sink.events))
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("expected json response body, got err=%v body=%s", err, resp.Body)
	}
	if body["received"] != float64(2) {
		t.Fatalf("expected received=2 in response body, got %#v", body["received"])
	}
}

func TestWhatsAppHandlerInboundRejectsInvalidSignature(t *testing.T) {
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider: &fakeWebhookProvider{verifyErr: errors.New("invalid signature")},
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook/inbound",
		Body:       `{"entry":[]}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid webhook signature, got %d", resp.StatusCode)
	}
}

type fakeIdentityVerifier struct {
	verified bool
	err      error
	lastFrom string
	lastText string
}

func (f *fakeIdentityVerifier) VerifyLinkFromInbound(_ context.Context, fromPhone, text string) (bool, error) {
	f.lastFrom = fromPhone
	f.lastText = text
	return f.verified, f.err
}

func TestWhatsAppHandlerInboundLinkVerificationReturnsVerifiedBeforeUserLookup(t *testing.T) {
	provider := &fakeWebhookProvider{
		inboundEnvelope: &models.WhatsAppMessageEnvelope{
			Provider:          models.WhatsAppProviderTwilio,
			ProviderMessageID: "SM999",
			From:              "whatsapp:+628123456789",
			To:                "whatsapp:+14155550000",
			Type:              models.WhatsAppMessageTypeText,
			Text:              "PROPTI LINK 182542",
			ReceivedAt:        time.Now().UTC(),
		},
	}
	verifier := &fakeIdentityVerifier{verified: true}
	handler := NewWhatsAppHandler(WhatsAppHandlerDependencies{
		Provider:         provider,
		UserStore:        &fakeWhatsAppUserStore{}, // no linked user
		IdentityVerifier: verifier,
	})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook/inbound",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"entry":[]}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if !strings.Contains(resp.Body, "whatsapp_link_verified") {
		t.Fatalf("expected whatsapp_link_verified response, got %s", resp.Body)
	}
	if verifier.lastFrom != "whatsapp:+628123456789" {
		t.Fatalf("expected verifier to receive raw from phone, got %q", verifier.lastFrom)
	}
}

func TestWhatsAppHandlerTemplateWiresWebhookLambdaRoutes(t *testing.T) {
	template, err := os.ReadFile("../../template.yaml")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	content := string(template)

	if !strings.Contains(content, "CodeUri: ./cmd/whatsapp/") {
		t.Fatal("expected SAM template to wire whatsapp lambda bootstrap from ./cmd/whatsapp/")
	}
	if !strings.Contains(content, "Path: /whatsapp/webhook") {
		t.Fatal("expected SAM template to expose /whatsapp/webhook route")
	}
	if !strings.Contains(content, "Path: /whatsapp/webhook/inbound") {
		t.Fatal("expected SAM template to expose /whatsapp/webhook/inbound route")
	}
	if !strings.Contains(content, "Path: /whatsapp/webhook/status") {
		t.Fatal("expected SAM template to expose /whatsapp/webhook/status route")
	}
}

func TestToHTTPRequestPrefersForwardedURLForWebhookSignature(t *testing.T) {
	httpReq, err := toHTTPRequest(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook",
		Headers: map[string]string{
			"X-Forwarded-Proto": "https",
			"X-Forwarded-Host":  "api.propti.id",
			"Content-Type":      "application/x-www-form-urlencoded",
		},
		Body: "MessageSid=SM123&Body=hi",
	})
	if err != nil {
		t.Fatalf("toHTTPRequest returned error: %v", err)
	}

	if got, want := httpReq.URL.String(), "https://api.propti.id/whatsapp/webhook"; got != want {
		t.Fatalf("expected forwarded webhook URL %q, got %q", want, got)
	}
}

func TestToHTTPRequestPrefersHostHeaderOverForwardedHost(t *testing.T) {
	httpReq, err := toHTTPRequest(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/whatsapp/webhook",
		Headers: map[string]string{
			"Host":              "api.propti.id",
			"X-Forwarded-Proto": "https",
			"X-Forwarded-Host":  "4s6kgocbpg.execute-api.ap-southeast-1.amazonaws.com",
			"Content-Type":      "application/x-www-form-urlencoded",
		},
		Body: "MessageSid=SM123&Body=halo",
	})
	if err != nil {
		t.Fatalf("toHTTPRequest returned error: %v", err)
	}

	if got, want := httpReq.URL.String(), "https://api.propti.id/whatsapp/webhook"; got != want {
		t.Fatalf("expected host header URL %q, got %q", want, got)
	}
}
