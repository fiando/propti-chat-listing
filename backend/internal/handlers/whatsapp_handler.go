package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type whatsAppWebhookProvider interface {
	Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error)
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
	EnsureOutboundAllowed(ctx context.Context, req models.WhatsAppSendRequest) error
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
	StatusSink          whatsAppWebhookStatusSink
	MetaVerifyToken     string
	WebBaseURL          string
}

type WhatsAppHandler struct {
	provider            whatsAppWebhookProvider
	userStore           whatsAppWebhookUserStore
	commandOrchestrator whatsAppWebhookCommandOrchestrator
	voiceService        whatsAppWebhookVoiceService
	policy              whatsAppWebhookPolicy
	statusSink          whatsAppWebhookStatusSink
	metaVerifyToken     string
	webBaseURL          string
}

func NewWhatsAppHandler(deps WhatsAppHandlerDependencies) *WhatsAppHandler {
	webBase := strings.TrimSpace(deps.WebBaseURL)
	if webBase == "" {
		webBase = "https://propti.id"
	}
	return &WhatsAppHandler{
		provider:            deps.Provider,
		userStore:           deps.UserStore,
		commandOrchestrator: deps.CommandOrchestrator,
		voiceService:        deps.VoiceService,
		policy:              deps.Policy,
		statusSink:          deps.StatusSink,
		metaVerifyToken:     strings.TrimSpace(deps.MetaVerifyToken),
		webBaseURL:          webBase,
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
			h.sendErrorReply(ctx, envelope.From, envelope.To, err)
			return appErrorResponse(err), true
		}
		if voiceResp.CommandResponse != nil {
			h.sendCommandReply(ctx, envelope.From, envelope.To, voiceResp.CommandResponse)
		} else if voiceResp.FallbackMessage != "" {
			h.sendTextReply(ctx, envelope.From, envelope.To, voiceResp.FallbackMessage)
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
		h.sendErrorReply(ctx, envelope.From, envelope.To, err)
		return appErrorResponse(err), true
	}
	h.sendCommandReply(ctx, envelope.From, envelope.To, commandResp)
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

// sendCommandReply sends a human-readable WhatsApp reply for a successfully processed command.
// Failures to send are logged and silently swallowed so the webhook always returns 200.
func (h *WhatsAppHandler) sendCommandReply(ctx context.Context, to, from string, resp *services.WhatsAppCommandResponse) {
	if resp == nil {
		return
	}
	msg := formatCommandReply(resp, h.webBaseURL)
	if msg == "" {
		return
	}
	h.sendTextReply(ctx, to, from, msg)
}

// sendErrorReply sends a user-friendly WhatsApp error message back to the sender.
func (h *WhatsAppHandler) sendErrorReply(ctx context.Context, to, from string, err error) {
	msg := formatErrorReply(err)
	if msg == "" {
		return
	}
	h.sendTextReply(ctx, to, from, msg)
}

// sendTextReply sends a plain text WhatsApp message. It checks the outbound policy (24-hour
// customer-service window) before sending. Errors are logged but never propagated so the
// webhook handler can always return 200 OK.
func (h *WhatsAppHandler) sendTextReply(ctx context.Context, to, from, text string) {
	if h.provider == nil || strings.TrimSpace(to) == "" || strings.TrimSpace(text) == "" {
		return
	}
	sendReq := models.WhatsAppSendRequest{
		To:   to,
		From: from,
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: text,
		},
	}
	if h.policy != nil {
		if err := h.policy.EnsureOutboundAllowed(ctx, sendReq); err != nil {
			utils.LogWarn("whatsapp outbound policy blocked reply", "to", to, "error", err.Error())
			return
		}
	}
	if _, err := h.provider.Send(ctx, sendReq); err != nil {
		utils.LogWarn("whatsapp send reply failed", "to", to, "error", err.Error())
	}
}

// formatCommandReply converts a WhatsAppCommandResponse into a human-readable Indonesian
// WhatsApp message. It returns an empty string if there is nothing meaningful to send.
func formatCommandReply(resp *services.WhatsAppCommandResponse, webBaseURL string) string {
	if resp == nil {
		return ""
	}
	switch resp.Intent {
	case services.WhatsAppCommandIntentListingCreate:
		return formatListingCreateReply(resp.Listing, resp.WebDeepLink)
	case services.WhatsAppCommandIntentListingEdit:
		return formatListingEditReply(resp.Listing, resp.WebDeepLink)
	case services.WhatsAppCommandIntentListingDelete:
		return "✅ Iklan berhasil dihapus."
	case services.WhatsAppCommandIntentListingRead:
		return formatListingReadReply(resp.Listings, webBaseURL)
	case services.WhatsAppCommandIntentSearch:
		return formatSearchReply(resp.Listings, resp.WebDeepLink)
	case services.WhatsAppCommandIntentSubscriptionStatus:
		return formatSubscriptionReply(resp.Subscription)
	default:
		return strings.TrimSpace(resp.Message)
	}
}

func formatListingCreateReply(listing *models.Listing, link string) string {
	if listing == nil {
		return "✅ Iklan berhasil dibuat!"
	}
	var b strings.Builder
	b.WriteString("✅ Iklan berhasil dibuat!\n\n")
	if listing.Title != "" {
		b.WriteString("🏠 " + listing.Title + "\n")
	}
	if listing.Price > 0 {
		b.WriteString("💰 " + formatPrice(listing.Price, listing.PriceUnit) + "\n")
	}
	if listing.Location.City != "" || listing.Location.Province != "" {
		b.WriteString("📍 " + formatLocation(listing.Location) + "\n")
	}
	b.WriteString("\n📊 Status: Aktif — sedang menunggu moderasi\n")
	b.WriteString("Iklan kamu sudah langsung tayang sambil menunggu verifikasi.\n")
	if link != "" {
		b.WriteString("\n🔗 Lihat iklan: " + link)
	}
	return b.String()
}

func formatListingEditReply(listing *models.Listing, link string) string {
	if listing == nil {
		return "✅ Iklan berhasil diperbarui!"
	}
	var b strings.Builder
	b.WriteString("✅ Iklan berhasil diperbarui!\n\n")
	if listing.Title != "" {
		b.WriteString("🏠 " + listing.Title + "\n")
	}
	if listing.ModerationStatus == models.ModerationStatusPending {
		b.WriteString("📊 Sedang menunggu moderasi ulang.\n")
	}
	if link != "" {
		b.WriteString("\n🔗 Lihat iklan: " + link)
	}
	return b.String()
}

func formatListingReadReply(listings []models.Listing, webBaseURL string) string {
	if len(listings) == 0 {
		return "📋 Kamu belum memiliki iklan aktif."
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📋 Iklan kamu (%d iklan):\n", len(listings)))
	base := strings.TrimRight(webBaseURL, "/")
	limit := len(listings)
	if limit > 5 {
		limit = 5
	}
	for i, l := range listings[:limit] {
		b.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, l.Title))
		if l.Price > 0 {
			b.WriteString("   💰 " + formatPrice(l.Price, l.PriceUnit) + "\n")
		}
		b.WriteString("   📊 " + formatListingStatus(l.Status, l.ModerationStatus) + "\n")
		if l.ListingID != "" && base != "" {
			b.WriteString("   🔗 " + base + "/listings/" + l.ListingID + "\n")
		}
	}
	if len(listings) > 5 {
		b.WriteString(fmt.Sprintf("\n...dan %d iklan lainnya.", len(listings)-5))
	}
	return b.String()
}

func formatSearchReply(listings []models.Listing, webDeepLink string) string {
	var b strings.Builder
	if len(listings) == 0 {
		b.WriteString("🔍 Tidak ada iklan yang ditemukan untuk pencarian ini.")
		if webDeepLink != "" {
			b.WriteString("\n\n🌐 Coba cari di web: " + webDeepLink)
		}
		return b.String()
	}
	b.WriteString(fmt.Sprintf("🔍 Ditemukan %d iklan:\n", len(listings)))
	limit := len(listings)
	if limit > 5 {
		limit = 5
	}
	for i, l := range listings[:limit] {
		b.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, l.Title))
		if l.Price > 0 {
			b.WriteString("   💰 " + formatPrice(l.Price, l.PriceUnit) + "\n")
		}
		if l.Location.City != "" || l.Location.Province != "" {
			b.WriteString("   📍 " + formatLocation(l.Location) + "\n")
		}
	}
	if webDeepLink != "" {
		b.WriteString("\n🌐 Lihat semua: " + webDeepLink)
	}
	return b.String()
}

func formatSubscriptionReply(sub *services.WhatsAppSubscriptionSummary) string {
	if sub == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("📊 Status Paket Propti\n\n")
	b.WriteString("Paket: " + formatTierName(string(sub.Tier)) + "\n")
	b.WriteString("Status: " + formatSubscriptionStatus(sub.Status) + "\n")
	b.WriteString(fmt.Sprintf("Iklan: %d/%d (%d tersisa)\n", sub.UsedListings, sub.LimitListings, sub.RemainingListings))
	if sub.UpgradeGuidance != "" {
		b.WriteString("\n" + sub.UpgradeGuidance)
	}
	return b.String()
}

// formatErrorReply converts an error into a user-friendly Indonesian message.
func formatErrorReply(err error) string {
	if err == nil {
		return ""
	}
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case 400:
			return "❌ Perintah tidak valid. Coba periksa kembali format pesan kamu."
		case 403:
			return "❌ " + appErr.Message
		case 429:
			return "❌ Terlalu banyak permintaan. Silakan coba lagi beberapa saat."
		case 404:
			return "❌ Iklan tidak ditemukan."
		}
	}
	return "❌ Terjadi kesalahan. Silakan coba lagi."
}

func formatPrice(price float64, unit string) string {
	formatted := formatRupiah(price)
	if unit != "" && unit != "total" {
		return formatted + " " + unit
	}
	return formatted
}

func formatRupiah(amount float64) string {
	if amount >= 1_000_000_000 {
		return fmt.Sprintf("Rp %.2fM", amount/1_000_000_000)
	}
	if amount >= 1_000_000 {
		return fmt.Sprintf("Rp %.0f jt", amount/1_000_000)
	}
	if amount >= 1_000 {
		return fmt.Sprintf("Rp %.0f rb", amount/1_000)
	}
	return fmt.Sprintf("Rp %.0f", amount)
}

func formatLocation(loc models.Location) string {
	parts := []string{}
	if loc.City != "" {
		parts = append(parts, loc.City)
	}
	if loc.Province != "" && loc.Province != loc.City {
		parts = append(parts, loc.Province)
	}
	return strings.Join(parts, ", ")
}

func formatListingStatus(status models.ListingStatus, modStatus models.ModerationStatus) string {
	statusStr := "Aktif"
	switch status {
	case models.ListingStatusSold:
		statusStr = "Terjual"
	case models.ListingStatusArchived:
		statusStr = "Diarsipkan"
	}
	modStr := ""
	switch modStatus {
	case models.ModerationStatusPending:
		modStr = "sedang diproses moderasi"
	case models.ModerationStatusRejected:
		modStr = "ditolak moderasi"
	}
	if modStr != "" {
		return statusStr + " (" + modStr + ")"
	}
	return statusStr
}

func formatSubscriptionStatus(status models.SubscriptionStatus) string {
	switch status {
	case models.SubscriptionActive:
		return "Aktif"
	case models.SubscriptionExpiringSoon:
		return "Akan habis masa berlakunya"
	case models.SubscriptionExpired:
		return "Kadaluarsa"
	default:
		return string(status)
	}
}

func formatTierName(tier string) string {
	if tier == "" {
		return "Free"
	}
	return strings.ToUpper(tier[:1]) + tier[1:]
}
