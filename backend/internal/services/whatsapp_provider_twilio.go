package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type TwilioWhatsAppProviderConfig struct {
	AccountSID string
	AuthToken  string
	From       string
	APIBaseURL string
	HTTPClient *http.Client
	Clock      func() time.Time
}

type TwilioWhatsAppProvider struct {
	accountSID string
	authToken  string
	from       string
	apiBaseURL string
	httpClient *http.Client
	now        func() time.Time
}

type twilioMessageResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    *int   `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func NewTwilioWhatsAppProvider(config TwilioWhatsAppProviderConfig) *TwilioWhatsAppProvider {
	apiBaseURL := strings.TrimRight(strings.TrimSpace(config.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.twilio.com"
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	now := config.Clock
	if now == nil {
		now = time.Now
	}

	return &TwilioWhatsAppProvider{
		accountSID: strings.TrimSpace(config.AccountSID),
		authToken:  strings.TrimSpace(config.AuthToken),
		from:       normalizeWhatsAppAddress(config.From),
		apiBaseURL: apiBaseURL,
		httpClient: httpClient,
		now:        now,
	}
}

func (p *TwilioWhatsAppProvider) Name() string {
	return "twilio"
}

func (p *TwilioWhatsAppProvider) Kind() models.WhatsAppProviderKind {
	return models.WhatsAppProviderTwilio
}

func (p *TwilioWhatsAppProvider) Capabilities() models.WhatsAppProviderCapabilities {
	return TwilioWhatsAppCapabilities()
}

func (p *TwilioWhatsAppProvider) Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error) {
	if p.accountSID == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio account sid is required")
	}
	if p.authToken == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio auth token is required")
	}

	to := normalizeWhatsAppAddress(request.To)
	if to == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("whatsapp recipient is required")
	}

	form := url.Values{}
	form.Set("To", to)

	switch request.Message.Type {
	case models.WhatsAppMessageTypeText:
		if strings.TrimSpace(request.Message.Text) == "" {
			return models.WhatsAppSendResult{}, fmt.Errorf("text message body is required")
		}
		form.Set("Body", request.Message.Text)
	case models.WhatsAppMessageTypeTemplate:
		contentSID := strings.TrimSpace(request.Metadata["twilio_content_sid"])
		if contentSID == "" {
			return models.WhatsAppSendResult{}, fmt.Errorf("twilio_content_sid metadata is required for template messages")
		}
		form.Set("ContentSid", contentSID)
		if variables := strings.TrimSpace(request.Metadata["twilio_content_variables"]); variables != "" {
			form.Set("ContentVariables", variables)
		}
	case models.WhatsAppMessageTypeMedia:
		if strings.TrimSpace(request.Message.Text) != "" {
			form.Set("Body", request.Message.Text)
		}
		if len(request.Message.Media) == 0 {
			return models.WhatsAppSendResult{}, fmt.Errorf("at least one media item is required for media messages")
		}
		for _, media := range request.Message.Media {
			if strings.TrimSpace(media.URL) == "" {
				return models.WhatsAppSendResult{}, fmt.Errorf("media item url is required")
			}
			form.Add("MediaUrl", media.URL)
		}
	default:
		return models.WhatsAppSendResult{}, fmt.Errorf("unsupported whatsapp message type %q for twilio provider", request.Message.Type)
	}

	from := normalizeWhatsAppAddress(request.From)
	if from == "" {
		from = p.from
	}
	if from == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio whatsapp sender is required")
	}
	form.Set("From", from)

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", p.apiBaseURL, p.accountSID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("build twilio send request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(p.accountSID, p.authToken)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio send request failed: %w", err)
	}
	defer resp.Body.Close()

	var payload twilioMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("decode twilio send response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail := payload.ErrorMessage
		if payload.ErrorCode != nil {
			detail = fmt.Sprintf("%s (code=%d)", detail, *payload.ErrorCode)
		}
		detail = strings.TrimSpace(detail)
		if detail == "" {
			detail = resp.Status
		}
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio send failed: %s", detail)
	}

	if strings.TrimSpace(payload.SID) == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("twilio send response missing sid")
	}

	return models.WhatsAppSendResult{
		Provider:          models.WhatsAppProviderTwilio,
		ProviderMessageID: payload.SID,
		AcceptedAt:        p.now().UTC(),
	}, nil
}

func (p *TwilioWhatsAppProvider) ParseInboundWebhook(_ context.Context, request *http.Request) (*models.WhatsAppMessageEnvelope, error) {
	if request == nil {
		return nil, fmt.Errorf("webhook request is required")
	}
	if err := request.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse twilio inbound webhook: %w", err)
	}

	from := normalizeWhatsAppAddress(request.FormValue("From"))
	if from == "" {
		return nil, fmt.Errorf("twilio inbound webhook missing from")
	}
	messageID := strings.TrimSpace(request.FormValue("MessageSid"))
	if messageID == "" {
		messageID = strings.TrimSpace(request.FormValue("SmsSid"))
	}
	if messageID == "" {
		return nil, fmt.Errorf("twilio inbound webhook missing message sid")
	}

	envelope := &models.WhatsAppMessageEnvelope{
		Provider:          models.WhatsAppProviderTwilio,
		ProviderMessageID: messageID,
		From:              from,
		To:                normalizeWhatsAppAddress(request.FormValue("To")),
		Type:              models.WhatsAppMessageTypeText,
		Text:              strings.TrimSpace(request.FormValue("Body")),
		ReceivedAt:        p.now().UTC(),
		Metadata:          map[string]string{},
	}

	if profileName := strings.TrimSpace(request.FormValue("ProfileName")); profileName != "" {
		envelope.Metadata["profile_name"] = profileName
	}
	if waID := strings.TrimSpace(request.FormValue("WaId")); waID != "" {
		envelope.Metadata["wa_id"] = waID
	}

	count, _ := strconv.Atoi(strings.TrimSpace(request.FormValue("NumMedia")))
	if count > 0 {
		envelope.Type = models.WhatsAppMessageTypeMedia
		envelope.Media = make([]models.WhatsAppMediaItem, 0, count)
		for i := 0; i < count; i++ {
			mediaURL := strings.TrimSpace(request.FormValue(fmt.Sprintf("MediaUrl%d", i)))
			if mediaURL == "" {
				continue
			}
			envelope.Media = append(envelope.Media, models.WhatsAppMediaItem{
				URL:      mediaURL,
				MimeType: strings.TrimSpace(request.FormValue(fmt.Sprintf("MediaContentType%d", i))),
			})
		}
	}

	if len(envelope.Metadata) == 0 {
		envelope.Metadata = nil
	}
	return envelope, nil
}

func (p *TwilioWhatsAppProvider) ParseDeliveryStatusWebhook(_ context.Context, request *http.Request) ([]models.WhatsAppDeliveryStatusEvent, error) {
	if request == nil {
		return nil, fmt.Errorf("webhook request is required")
	}
	if err := request.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse twilio status webhook: %w", err)
	}

	messageID := strings.TrimSpace(request.FormValue("MessageSid"))
	if messageID == "" {
		messageID = strings.TrimSpace(request.FormValue("SmsSid"))
	}
	if messageID == "" {
		return nil, fmt.Errorf("twilio status webhook missing message sid")
	}

	statusRaw := strings.TrimSpace(request.FormValue("MessageStatus"))
	if statusRaw == "" {
		statusRaw = strings.TrimSpace(request.FormValue("SmsStatus"))
	}
	status, err := mapTwilioDeliveryStatus(statusRaw)
	if err != nil {
		return nil, err
	}

	event := models.WhatsAppDeliveryStatusEvent{
		Provider:          models.WhatsAppProviderTwilio,
		ProviderMessageID: messageID,
		Status:            status,
		OccurredAt:        p.now().UTC(),
		ErrorCode:         strings.TrimSpace(request.FormValue("ErrorCode")),
		ErrorMessage:      strings.TrimSpace(request.FormValue("ErrorMessage")),
	}
	if channel := strings.TrimSpace(request.FormValue("ChannelToAddress")); channel != "" {
		event.Metadata = map[string]string{"channel_to": channel}
	}

	return []models.WhatsAppDeliveryStatusEvent{event}, nil
}

func (p *TwilioWhatsAppProvider) VerifyWebhookRequest(request *http.Request) error {
	if request == nil {
		return fmt.Errorf("webhook request is required")
	}
	if p.authToken == "" {
		return fmt.Errorf("twilio auth token is required for webhook verification")
	}

	headerSignature := strings.TrimSpace(request.Header.Get("X-Twilio-Signature"))
	if headerSignature == "" {
		return fmt.Errorf("missing X-Twilio-Signature header")
	}

	expected, err := p.computeWebhookSignature(request)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(headerSignature)) != 1 {
		return fmt.Errorf("invalid twilio webhook signature")
	}
	return nil
}

func (p *TwilioWhatsAppProvider) computeWebhookSignature(request *http.Request) (string, error) {
	if request == nil {
		return "", fmt.Errorf("webhook request is required")
	}
	if p.authToken == "" {
		return "", fmt.Errorf("twilio auth token is required for webhook verification")
	}

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return "", fmt.Errorf("read webhook body: %w", err)
	}
	request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	values, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("parse webhook body for signature: %w", err)
	}

	data := request.URL.String()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		vals := append([]string(nil), values[key]...)
		sort.Strings(vals)
		for _, val := range vals {
			data += key + val
		}
	}

	mac := hmac.New(sha1.New, []byte(p.authToken))
	_, _ = mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func mapTwilioDeliveryStatus(raw string) (models.WhatsAppDeliveryStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "queued", "accepted", "scheduled":
		return models.WhatsAppDeliveryStatusQueued, nil
	case "sent":
		return models.WhatsAppDeliveryStatusSent, nil
	case "delivered":
		return models.WhatsAppDeliveryStatusDelivered, nil
	case "read":
		return models.WhatsAppDeliveryStatusRead, nil
	case "failed", "undelivered", "canceled":
		return models.WhatsAppDeliveryStatusFailed, nil
	default:
		return "", fmt.Errorf("unsupported twilio delivery status %q", raw)
	}
}

func normalizeWhatsAppAddress(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "whatsapp:") {
		return "whatsapp:" + strings.TrimSpace(trimmed[len("whatsapp:"):])
	}
	return "whatsapp:" + trimmed
}
