package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type whatsAppIdentityService interface {
	GetWriteEligibility(ctx context.Context, userID string) (*services.WhatsAppWriteEligibility, error)
	StartLink(ctx context.Context, userID, phone string) (*services.WhatsAppLinkChallenge, error)
	VerifyLink(ctx context.Context, userID, challengeID, otpCode string) (*services.WhatsAppWriteEligibility, error)
	DisconnectLink(ctx context.Context, userID string) (*services.WhatsAppWriteEligibility, error)
}

type WhatsAppIdentityHandler struct {
	service whatsAppIdentityService
}

type whatsAppLinkChallengeRequest struct {
	Phone string `json:"phone"`
}

type whatsAppLinkVerifyRequest struct {
	ChallengeID string `json:"challengeId"`
	OTPCode     string `json:"otpCode"`
}

type whatsAppEligibilityResponse struct {
	Eligible    bool       `json:"eligible"`
	IsLinked    bool       `json:"isLinked"`
	LinkedPhone string     `json:"linkedPhone,omitempty"`
	VerifiedAt  *time.Time `json:"verifiedAt,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

func NewWhatsAppIdentityHandler(service whatsAppIdentityService) *WhatsAppIdentityHandler {
	return &WhatsAppIdentityHandler{service: service}
}

func (h *WhatsAppIdentityHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == http.MethodGet && req.Path == "/auth/whatsapp/link-status":
		return h.linkStatus(ctx, req)
	case req.HTTPMethod == http.MethodPost && req.Path == "/auth/whatsapp/link-challenge":
		return h.linkChallenge(ctx, req)
	case req.HTTPMethod == http.MethodPost && req.Path == "/auth/whatsapp/link-verify":
		return h.linkVerify(ctx, req)
	case req.HTTPMethod == http.MethodDelete && req.Path == "/auth/whatsapp/link":
		return h.disconnect(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

func (h *WhatsAppIdentityHandler) linkStatus(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	eligibility, err := h.service.GetWriteEligibility(ctx, userID)
	if err != nil {
		return appErrorResponse(err), nil
	}

	return eligibilityResponse(eligibility), nil
}

func (h *WhatsAppIdentityHandler) linkChallenge(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var challengeReq whatsAppLinkChallengeRequest
	if err := json.Unmarshal([]byte(req.Body), &challengeReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if strings.TrimSpace(challengeReq.Phone) == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "phone is required"))), nil
	}

	challenge, err := h.service.StartLink(ctx, userID, challengeReq.Phone)
	if err != nil {
		return appErrorResponse(err), nil
	}

	body, _ := json.Marshal(challenge)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *WhatsAppIdentityHandler) linkVerify(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var verifyReq whatsAppLinkVerifyRequest
	if err := json.Unmarshal([]byte(req.Body), &verifyReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if strings.TrimSpace(verifyReq.ChallengeID) == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "challenge id is required"))), nil
	}
	if strings.TrimSpace(verifyReq.OTPCode) == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "otp code is required"))), nil
	}

	eligibility, err := h.service.VerifyLink(ctx, userID, verifyReq.ChallengeID, verifyReq.OTPCode)
	if err != nil {
		return appErrorResponse(err), nil
	}

	return eligibilityResponse(eligibility), nil
}

func (h *WhatsAppIdentityHandler) disconnect(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	eligibility, err := h.service.DisconnectLink(ctx, userID)
	if err != nil {
		return appErrorResponse(err), nil
	}

	return eligibilityResponse(eligibility), nil
}

func eligibilityResponse(eligibility *services.WhatsAppWriteEligibility) events.APIGatewayProxyResponse {
	resp := whatsAppEligibilityResponse{}
	if eligibility != nil {
		resp.Eligible = eligibility.Eligible
		resp.LinkedPhone = eligibility.LinkedPhone
		resp.VerifiedAt = eligibility.VerifiedAt
		resp.Reason = eligibility.Reason
		resp.IsLinked = strings.TrimSpace(eligibility.LinkedPhone) != ""
	}

	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body))
}
