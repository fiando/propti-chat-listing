package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakeWhatsAppIdentityService struct {
	eligibilityResult *services.WhatsAppWriteEligibility
	eligibilityErr    error

	challengeResult *services.WhatsAppLinkChallenge
	challengeErr    error
	challengeInput  struct {
		userID string
		phone  string
	}

	verifyResult *services.WhatsAppWriteEligibility
	verifyErr    error
	verifyInput  struct {
		userID      string
		challengeID string
		otpCode     string
	}

	disconnectResult *services.WhatsAppWriteEligibility
	disconnectErr    error
	disconnectUserID string
}

func (f *fakeWhatsAppIdentityService) GetWriteEligibility(_ context.Context, userID string) (*services.WhatsAppWriteEligibility, error) {
	return f.eligibilityResult, f.eligibilityErr
}

func (f *fakeWhatsAppIdentityService) StartLink(_ context.Context, userID, phone string) (*services.WhatsAppLinkChallenge, error) {
	f.challengeInput.userID = userID
	f.challengeInput.phone = phone
	return f.challengeResult, f.challengeErr
}

func (f *fakeWhatsAppIdentityService) VerifyLink(_ context.Context, userID, challengeID, otpCode string) (*services.WhatsAppWriteEligibility, error) {
	f.verifyInput.userID = userID
	f.verifyInput.challengeID = challengeID
	f.verifyInput.otpCode = otpCode
	return f.verifyResult, f.verifyErr
}

func (f *fakeWhatsAppIdentityService) DisconnectLink(_ context.Context, userID string) (*services.WhatsAppWriteEligibility, error) {
	f.disconnectUserID = userID
	return f.disconnectResult, f.disconnectErr
}

func TestWhatsAppIdentityHandlerLinkStatusUnauthorized(t *testing.T) {
	handler := NewWhatsAppIdentityHandler(&fakeWhatsAppIdentityService{})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/auth/whatsapp/link-status",
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func TestWhatsAppIdentityHandlerLinkStatusSuccess(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	service := &fakeWhatsAppIdentityService{
		eligibilityResult: &services.WhatsAppWriteEligibility{
			Eligible:    true,
			LinkedPhone: "+628123456789",
			VerifiedAt:  &now,
		},
	}
	handler := NewWhatsAppIdentityHandler(service)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/auth/whatsapp/link-status",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid json: %v", err)
	}
	if body["eligible"] != true {
		t.Fatalf("expected eligible=true, got %#v", body["eligible"])
	}
	if body["linkedPhone"] != "+628123456789" {
		t.Fatalf("expected linkedPhone +628123456789, got %#v", body["linkedPhone"])
	}
}

func TestWhatsAppIdentityHandlerLinkChallengePassesRequestToService(t *testing.T) {
	service := &fakeWhatsAppIdentityService{
		challengeResult: &services.WhatsAppLinkChallenge{
			ChallengeID: "challenge-1",
			Phone:       "+628123456789",
			ExpiresAt:   time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC),
			RetryCount:  1,
		},
	}
	handler := NewWhatsAppIdentityHandler(service)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/auth/whatsapp/link-challenge",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
		Body:       `{"phone":"0812-3456-789"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if service.challengeInput.userID != "user-1" {
		t.Fatalf("expected userID user-1, got %q", service.challengeInput.userID)
	}
	if service.challengeInput.phone != "0812-3456-789" {
		t.Fatalf("expected phone to be forwarded, got %q", service.challengeInput.phone)
	}
}

func TestWhatsAppIdentityHandlerLinkVerifyReturnsServiceAppError(t *testing.T) {
	service := &fakeWhatsAppIdentityService{
		verifyErr: utils.NewAppError(401, "invalid otp code"),
	}
	handler := NewWhatsAppIdentityHandler(service)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/auth/whatsapp/link-verify",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
		Body:       `{"challengeId":"challenge-1","otpCode":"000000"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.StatusCode, resp.Body)
	}
}

func TestWhatsAppIdentityHandlerDisconnectSuccess(t *testing.T) {
	service := &fakeWhatsAppIdentityService{
		disconnectResult: &services.WhatsAppWriteEligibility{
			Eligible: false,
			Reason:   "whatsapp phone is not linked",
		},
	}
	handler := NewWhatsAppIdentityHandler(service)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodDelete,
		Path:       "/auth/whatsapp/link",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if service.disconnectUserID != "user-1" {
		t.Fatalf("expected disconnect to be called with user-1, got %q", service.disconnectUserID)
	}
}

func TestAuthHandlerRoutesWhatsAppIdentityEndpoints(t *testing.T) {
	auth := &AuthHandler{
		whatsAppIdentity: NewWhatsAppIdentityHandler(&fakeWhatsAppIdentityService{
			eligibilityResult: &services.WhatsAppWriteEligibility{Eligible: false},
		}),
	}

	resp, err := auth.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/auth/whatsapp/link-status",
		Headers:    authHeaderForPremiumTest(t, "user-1"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}
}
