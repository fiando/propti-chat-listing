package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

const whatsAppLinkConfirmationMessage = "✅ Nomor WhatsApp kamu berhasil terhubung ke Propti! Sekarang kamu bisa kirim iklan properti langsung via WhatsApp."

type whatsAppWebhookProvider interface {
	ParseInboundWebhook(ctx context.Context, request *http.Request) (*models.WhatsAppMessageEnvelope, error)
	ParseDeliveryStatusWebhook(ctx context.Context, request *http.Request) ([]models.WhatsAppDeliveryStatusEvent, error)
	VerifyWebhookRequest(request *http.Request) error
}

type whatsAppWebhookUserStore interface {
	GetByWhatsAppPhone(ctx context.Context, phone string) (*models.User, error)
}

type whatsAppWebhookCommandOrchestrator interface {
	HandleText(ctx context.Context, req services.WhatsAppCommandRequest) (*services.WhatsAppCommandResponse, error)
}

type whatsAppWebhookVoiceService interface {
	HandleInboundVoice(ctx context.Context, req services.WhatsAppVoiceRequest) (*services.WhatsAppVoiceResponse, error)
}

type whatsAppWebhookPolicy interface {
	RecordInboundMessage(ctx context.Context, envelope models.WhatsAppMessageEnvelope) error
}

type whatsAppInboundIdentityVerifier interface {
	VerifyLinkFromInbound(ctx context.Context, fromPhone, text string) (bool, error)
}

type whatsAppLinkConfirmSender interface {
	Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error)
}

type whatsAppCommandReplySender interface {
	Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error)
}

type whatsAppWebhookStatusSink interface {
	HandleDeliveryStatus(ctx context.Context, event models.WhatsAppDeliveryStatusEvent) error
}

type WhatsAppHandlerDependencies struct {
	Provider            whatsAppWebhookProvider
	UserStore           whatsAppWebhookUserStore
	CommandOrchestrator whatsAppWebhookCommandOrchestrator
	VoiceService        whatsAppWebhookVoiceService
	Policy              whatsAppWebhookPolicy
	IdentityVerifier    whatsAppInboundIdentityVerifier
	ConfirmSender       whatsAppLinkConfirmSender
	CommandReplySender  whatsAppCommandReplySender
	StatusSink          whatsAppWebhookStatusSink
	MetaVerifyToken     string
}

type WhatsAppHandler struct {
	provider            whatsAppWebhookProvider
	userStore           whatsAppWebhookUserStore
	commandOrchestrator whatsAppWebhookCommandOrchestrator
	voiceService        whatsAppWebhookVoiceService
	policy              whatsAppWebhookPolicy
	identityVerifier    whatsAppInboundIdentityVerifier
	confirmSender       whatsAppLinkConfirmSender
	commandReplySender  whatsAppCommandReplySender
	statusSink          whatsAppWebhookStatusSink
	metaVerifyToken     string
}

func NewWhatsAppHandler(deps WhatsAppHandlerDependencies) *WhatsAppHandler {
	return &WhatsAppHandler{
		provider:            deps.Provider,
		userStore:           deps.UserStore,
		commandOrchestrator: deps.CommandOrchestrator,
		voiceService:        deps.VoiceService,
		policy:              deps.Policy,
		identityVerifier:    deps.IdentityVerifier,
		confirmSender:       deps.ConfirmSender,
		commandReplySender:  deps.CommandReplySender,
		statusSink:          deps.StatusSink,
		metaVerifyToken:     strings.TrimSpace(deps.MetaVerifyToken),
	}
}

func (h *WhatsAppHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if h.provider == nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	switch {
	case req.HTTPMethod == http.MethodGet && req.Path == "/whatsapp/webhook":
		return h.handleMetaWebhookVerification(req), nil
	case req.HTTPMethod == http.MethodPost && req.Path == "/whatsapp/webhook/inbound":
		return h.handleInbound(ctx, req), nil
	case req.HTTPMethod == http.MethodPost && req.Path == "/whatsapp/webhook/status":
		return h.handleStatus(ctx, req), nil
	case req.HTTPMethod == http.MethodPost && req.Path == "/whatsapp/webhook":
		return h.handleUnifiedWebhook(ctx, req), nil
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

func (h *WhatsAppHandler) handleMetaWebhookVerification(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	mode := strings.TrimSpace(req.QueryStringParameters["hub.mode"])
	token := strings.TrimSpace(req.QueryStringParameters["hub.verify_token"])
	challenge := req.QueryStringParameters["hub.challenge"]

	if mode != "subscribe" || challenge == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest))
	}
	if h.metaVerifyToken == "" || token != h.metaVerifyToken {
		return jsonResponse(http.StatusForbidden, utils.MarshalErrorResponse(utils.ErrForbidden))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                "text/plain",
			"Access-Control-Allow-Origin": "*",
		},
		Body: challenge,
	}
}

func (h *WhatsAppHandler) handleUnifiedWebhook(ctx context.Context, req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	if resp, ok := h.processInbound(ctx, req); ok {
		return resp
	}
	return h.processStatus(ctx, req)
}

func (h *WhatsAppHandler) handleInbound(ctx context.Context, req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	resp, _ := h.processInbound(ctx, req)
	return resp
}

func (h *WhatsAppHandler) processInbound(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, bool) {
	httpReq, err := toHTTPRequest(req)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), true
	}
	if err := h.provider.VerifyWebhookRequest(httpReq); err != nil {
		utils.LogWarn("verify whatsapp inbound webhook signature", "error", err.Error())
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), true
	}

	envelope, err := h.provider.ParseInboundWebhook(ctx, httpReq)
	if err != nil {
		utils.LogWarn("parse whatsapp inbound webhook", "error", err.Error())
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), false
	}
	if envelope == nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), true
	}

	if h.policy != nil {
		if err := h.policy.RecordInboundMessage(ctx, *envelope); err != nil {
			utils.LogWarn("record whatsapp inbound session", "error", err.Error())
		}
	}
	if h.identityVerifier != nil {
		verified, err := h.identityVerifier.VerifyLinkFromInbound(ctx, envelope.From, envelope.Text)
		if err != nil {
			return appErrorResponse(err), true
		}
		if verified {
			if h.confirmSender != nil {
				_, sendErr := h.confirmSender.Send(ctx, models.WhatsAppSendRequest{
					To: envelope.From,
					Message: models.WhatsAppOutboundMessage{
						Type: models.WhatsAppMessageTypeText,
						Text: whatsAppLinkConfirmationMessage,
					},
				})
				if sendErr != nil {
					utils.LogWarn("send whatsapp link confirmation", "error", sendErr.Error())
				}
			}
			return jsonResponse(http.StatusOK, `{"status":"ok","reason":"whatsapp_link_verified"}`), true
		}
	}

	userID, err := h.lookupUserIDByWhatsAppPhone(ctx, envelope.From)
	if err != nil {
		utils.LogWarn("resolve whatsapp user", "error", err.Error())
		return jsonResponse(http.StatusOK, `{"status":"ignored","reason":"user_not_linked"}`), true
	}

	if envelope.Type == models.WhatsAppMessageTypeMedia && hasAudioMedia(envelope.Media) {
		if h.voiceService == nil {
			return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "voice service unavailable"))), true
		}
		voiceResp, err := h.voiceService.HandleInboundVoice(ctx, services.WhatsAppVoiceRequest{
			UserID:   userID,
			Envelope: *envelope,
		})
		if err != nil {
			return appErrorResponse(err), true
		}
		body, _ := json.Marshal(voiceResp)
		return jsonResponse(http.StatusOK, string(body)), true
	}

	if h.commandOrchestrator == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "whatsapp command service unavailable"))), true
	}
	commandResp, err := h.commandOrchestrator.HandleText(ctx, services.WhatsAppCommandRequest{
		UserID: userID,
		Text:   envelope.Text,
	})
	if err != nil {
		return appErrorResponse(err), true
	}
	if commandResp.Message != "" && h.commandReplySender != nil {
		_, sendErr := h.commandReplySender.Send(ctx, models.WhatsAppSendRequest{
			To: envelope.From,
			Message: models.WhatsAppOutboundMessage{
				Type: models.WhatsAppMessageTypeText,
				Text: commandResp.Message,
			},
		})
		if sendErr != nil {
			utils.LogWarn("send whatsapp command reply", "error", sendErr.Error())
		}
	}
	body, _ := json.Marshal(commandResp)
	return jsonResponse(http.StatusOK, string(body)), true
}

func (h *WhatsAppHandler) handleStatus(ctx context.Context, req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	return h.processStatus(ctx, req)
}

func (h *WhatsAppHandler) processStatus(ctx context.Context, req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	httpReq, err := toHTTPRequest(req)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest))
	}
	if err := h.provider.VerifyWebhookRequest(httpReq); err != nil {
		utils.LogWarn("verify whatsapp status webhook signature", "error", err.Error())
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized))
	}

	statusEvents, err := h.provider.ParseDeliveryStatusWebhook(ctx, httpReq)
	if err != nil {
		utils.LogWarn("parse whatsapp status webhook", "error", err.Error())
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest))
	}
	if h.statusSink != nil {
		for _, event := range statusEvents {
			if err := h.statusSink.HandleDeliveryStatus(ctx, event); err != nil {
				utils.LogWarn("handle whatsapp delivery status event", "error", err.Error(), "providerMessageId", event.ProviderMessageID)
			}
		}
	}

	body, _ := json.Marshal(map[string]any{
		"status":   "ok",
		"received": len(statusEvents),
	})
	return jsonResponse(http.StatusOK, string(body))
}

func (h *WhatsAppHandler) lookupUserIDByWhatsAppPhone(ctx context.Context, phone string) (string, error) {
	if h.userStore == nil {
		return "", errors.New("whatsapp user store is not configured")
	}
	normalized, err := normalizeIncomingWhatsAppPhone(phone)
	if err != nil {
		return "", err
	}
	user, err := h.userStore.GetByWhatsAppPhone(ctx, normalized)
	if err != nil || user == nil {
		return "", errors.New("user not found for whatsapp phone")
	}
	return user.UserID, nil
}

func normalizeIncomingWhatsAppPhone(raw string) (string, error) {
	phone := strings.TrimSpace(raw)
	phone = strings.TrimPrefix(strings.ToLower(phone), "whatsapp:")
	phone = strings.TrimSpace(phone)
	return utils.NormalizeWhatsAppPhone(phone)
}

func hasAudioMedia(items []models.WhatsAppMediaItem) bool {
	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.MimeType)), "audio/") {
			return true
		}
	}
	return false
}

func toHTTPRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	rawBody := req.Body
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		rawBody = string(decoded)
	}

	proto := headerValue(req.Headers, "X-Forwarded-Proto")
	if proto == "" {
		proto = "https"
	}

	host := headerValue(req.Headers, "Host")
	if host == "" {
		host = headerValue(req.Headers, "X-Forwarded-Host")
	}
	if host == "" {
		host = "webhook.local"
	}

	targetURL := proto + "://" + host + req.Path
	if len(req.QueryStringParameters) > 0 {
		query := url.Values{}
		for key, val := range req.QueryStringParameters {
			query.Set(key, val)
		}
		targetURL += "?" + query.Encode()
	}

	httpReq, err := http.NewRequest(req.HTTPMethod, targetURL, io.NopCloser(strings.NewReader(rawBody)))
	if err != nil {
		return nil, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	return httpReq, nil
}

func headerValue(headers map[string]string, key string) string {
	if headers == nil {
		return ""
	}
	for k, v := range headers {
		if strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
