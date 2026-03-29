package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type MetaWhatsAppProviderConfig struct {
	AccessToken   string
	PhoneNumberID string
	VerifyToken   string
	AppSecret     string
	APIBaseURL    string
	APIVersion    string
	HTTPClient    *http.Client
	Clock         func() time.Time
}

type MetaWhatsAppProvider struct {
	accessToken   string
	phoneNumberID string
	verifyToken   string
	appSecret     string
	apiBaseURL    string
	apiVersion    string
	httpClient    *http.Client
	now           func() time.Time
}

type metaSendRequest struct {
	MessagingProduct string              `json:"messaging_product"`
	RecipientType    string              `json:"recipient_type"`
	To               string              `json:"to"`
	Type             string              `json:"type"`
	Text             *metaSendText       `json:"text,omitempty"`
	Image            *metaMediaReference `json:"image,omitempty"`
	Document         *metaMediaReference `json:"document,omitempty"`
	Audio            *metaMediaReference `json:"audio,omitempty"`
	Video            *metaMediaReference `json:"video,omitempty"`
	Context          *metaMessageContext `json:"context,omitempty"`
}

type metaSendText struct {
	PreviewURL bool   `json:"preview_url,omitempty"`
	Body       string `json:"body"`
}

type metaMediaReference struct {
	Link    string `json:"link"`
	Caption string `json:"caption,omitempty"`
}

type metaMessageContext struct {
	MessageID string `json:"message_id"`
}

type metaSendResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

type metaWebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				Metadata struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					WAID    string `json:"wa_id"`
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
					Image    *metaWebhookMedia `json:"image"`
					Video    *metaWebhookMedia `json:"video"`
					Document *metaWebhookMedia `json:"document"`
					Audio    *metaWebhookMedia `json:"audio"`
				} `json:"messages"`
				Statuses []struct {
					ID        string `json:"id"`
					Status    string `json:"status"`
					Timestamp string `json:"timestamp"`
					Errors    []struct {
						Code    int    `json:"code"`
						Title   string `json:"title"`
						Message string `json:"message"`
					} `json:"errors"`
				} `json:"statuses"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

type metaWebhookMedia struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption"`
}

func NewMetaWhatsAppProvider(config MetaWhatsAppProviderConfig) *MetaWhatsAppProvider {
	apiBaseURL := strings.TrimRight(strings.TrimSpace(config.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://graph.facebook.com"
	}
	apiVersion := strings.TrimSpace(config.APIVersion)
	if apiVersion == "" {
		apiVersion = "v23.0"
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	now := config.Clock
	if now == nil {
		now = time.Now
	}

	return &MetaWhatsAppProvider{
		accessToken:   strings.TrimSpace(config.AccessToken),
		phoneNumberID: strings.TrimSpace(config.PhoneNumberID),
		verifyToken:   strings.TrimSpace(config.VerifyToken),
		appSecret:     strings.TrimSpace(config.AppSecret),
		apiBaseURL:    apiBaseURL,
		apiVersion:    apiVersion,
		httpClient:    httpClient,
		now:           now,
	}
}

func (p *MetaWhatsAppProvider) Name() string {
	return "direct-whatsapp"
}

func (p *MetaWhatsAppProvider) Kind() models.WhatsAppProviderKind {
	return models.WhatsAppProviderDirectWhatsApp
}

func (p *MetaWhatsAppProvider) Capabilities() models.WhatsAppProviderCapabilities {
	return DirectWhatsAppCapabilities()
}

func (p *MetaWhatsAppProvider) Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error) {
	if p.accessToken == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("meta whatsapp access token is required")
	}
	if p.phoneNumberID == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("meta whatsapp phone number id is required")
	}
	if strings.TrimSpace(request.To) == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("whatsapp recipient is required")
	}

	payload := metaSendRequest{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               strings.TrimSpace(strings.TrimPrefix(normalizeWhatsAppAddress(request.To), "whatsapp:")),
	}

	switch request.Message.Type {
	case models.WhatsAppMessageTypeText:
		if strings.TrimSpace(request.Message.Text) == "" {
			return models.WhatsAppSendResult{}, fmt.Errorf("text message body is required")
		}
		payload.Type = "text"
		payload.Text = &metaSendText{Body: request.Message.Text}
	case models.WhatsAppMessageTypeMedia:
		if len(request.Message.Media) == 0 {
			return models.WhatsAppSendResult{}, fmt.Errorf("at least one media item is required for media messages")
		}
		media := request.Message.Media[0]
		if strings.TrimSpace(media.URL) == "" {
			return models.WhatsAppSendResult{}, fmt.Errorf("media item url is required")
		}
		payloadType, mediaPayload := toMetaMediaPayload(media)
		payload.Type = payloadType
		switch payloadType {
		case "image":
			payload.Image = mediaPayload
		case "video":
			payload.Video = mediaPayload
		case "audio":
			payload.Audio = mediaPayload
		default:
			payload.Document = mediaPayload
		}
	case models.WhatsAppMessageTypeTemplate:
		return models.WhatsAppSendResult{}, fmt.Errorf("template messages are not supported by direct WhatsApp provider; use Twilio template flow")
	default:
		return models.WhatsAppSendResult{}, fmt.Errorf("unsupported whatsapp message type %q for direct whatsapp provider", request.Message.Type)
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("encode meta send payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s/messages", p.apiBaseURL, p.apiVersion, p.phoneNumberID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buf)
	if err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("build meta send request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("meta send request failed: %w", err)
	}
	defer resp.Body.Close()

	var response metaSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.WhatsAppSendResult{}, fmt.Errorf("decode meta send response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if response.Error != nil {
			return models.WhatsAppSendResult{}, fmt.Errorf("meta send failed: %s (code=%d)", response.Error.Message, response.Error.Code)
		}
		return models.WhatsAppSendResult{}, fmt.Errorf("meta send failed: %s", resp.Status)
	}
	if len(response.Messages) == 0 || strings.TrimSpace(response.Messages[0].ID) == "" {
		return models.WhatsAppSendResult{}, fmt.Errorf("meta send response missing message id")
	}

	return models.WhatsAppSendResult{
		Provider:          models.WhatsAppProviderDirectWhatsApp,
		ProviderMessageID: response.Messages[0].ID,
		AcceptedAt:        p.now().UTC(),
	}, nil
}

func (p *MetaWhatsAppProvider) ParseInboundWebhook(_ context.Context, request *http.Request) (*models.WhatsAppMessageEnvelope, error) {
	if request == nil {
		return nil, fmt.Errorf("webhook request is required")
	}
	payload, err := decodeMetaWebhookPayload(request)
	if err != nil {
		return nil, err
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			for _, msg := range change.Value.Messages {
				if strings.TrimSpace(msg.ID) == "" || strings.TrimSpace(msg.From) == "" {
					continue
				}
				envelope := &models.WhatsAppMessageEnvelope{
					Provider:          models.WhatsAppProviderDirectWhatsApp,
					ProviderMessageID: msg.ID,
					From:              normalizeWhatsAppAddress(msg.From),
					To:                normalizeWhatsAppAddress(change.Value.Metadata.DisplayPhoneNumber),
					Type:              mapMetaMessageType(msg.Type),
					Text:              strings.TrimSpace(msg.Text.Body),
					ReceivedAt:        parseUnixTimestampOrNow(msg.Timestamp, p.now),
					Metadata:          map[string]string{},
				}

				if len(change.Value.Contacts) > 0 {
					contact := change.Value.Contacts[0]
					if strings.TrimSpace(contact.Profile.Name) != "" {
						envelope.Metadata["contact_name"] = strings.TrimSpace(contact.Profile.Name)
					}
					if strings.TrimSpace(contact.WAID) != "" {
						envelope.Metadata["wa_id"] = strings.TrimSpace(contact.WAID)
					}
				}

				envelope.Media = append(envelope.Media, metaWebhookMediaToDomain(msg.Image)...)
				envelope.Media = append(envelope.Media, metaWebhookMediaToDomain(msg.Video)...)
				envelope.Media = append(envelope.Media, metaWebhookMediaToDomain(msg.Document)...)
				envelope.Media = append(envelope.Media, metaWebhookMediaToDomain(msg.Audio)...)
				if len(envelope.Media) > 0 {
					envelope.Type = models.WhatsAppMessageTypeMedia
				}
				if len(envelope.Metadata) == 0 {
					envelope.Metadata = nil
				}
				return envelope, nil
			}
		}
	}

	return nil, fmt.Errorf("meta inbound webhook did not contain messages")
}

func (p *MetaWhatsAppProvider) ParseDeliveryStatusWebhook(_ context.Context, request *http.Request) ([]models.WhatsAppDeliveryStatusEvent, error) {
	if request == nil {
		return nil, fmt.Errorf("webhook request is required")
	}
	payload, err := decodeMetaWebhookPayload(request)
	if err != nil {
		return nil, err
	}

	events := make([]models.WhatsAppDeliveryStatusEvent, 0)
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			for _, status := range change.Value.Statuses {
				if strings.TrimSpace(status.ID) == "" {
					continue
				}
				mappedStatus, mapErr := mapMetaDeliveryStatus(status.Status)
				if mapErr != nil {
					return nil, mapErr
				}
				event := models.WhatsAppDeliveryStatusEvent{
					Provider:          models.WhatsAppProviderDirectWhatsApp,
					ProviderMessageID: status.ID,
					Status:            mappedStatus,
					OccurredAt:        parseUnixTimestampOrNow(status.Timestamp, p.now),
				}
				if len(status.Errors) > 0 {
					errItem := status.Errors[0]
					event.ErrorCode = strconv.Itoa(errItem.Code)
					event.ErrorMessage = strings.TrimSpace(strings.Join([]string{strings.TrimSpace(errItem.Title), strings.TrimSpace(errItem.Message)}, ": "))
				}
				events = append(events, event)
			}
		}
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("meta status webhook did not contain status events")
	}
	return events, nil
}

func (p *MetaWhatsAppProvider) VerifyWebhookRequest(request *http.Request) error {
	if request == nil {
		return fmt.Errorf("webhook request is required")
	}
	if p.appSecret == "" {
		return fmt.Errorf("meta app secret is required for webhook verification")
	}

	headerSig := strings.TrimSpace(request.Header.Get("X-Hub-Signature-256"))
	if headerSig == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}
	if !strings.HasPrefix(headerSig, "sha256=") {
		return fmt.Errorf("invalid X-Hub-Signature-256 format")
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("read meta webhook body: %w", err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	mac := hmac.New(sha256.New, []byte(p.appSecret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	provided := strings.TrimPrefix(headerSig, "sha256=")
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(expected)), []byte(strings.ToLower(strings.TrimSpace(provided)))) != 1 {
		return fmt.Errorf("invalid meta webhook signature")
	}

	return nil
}

func decodeMetaWebhookPayload(request *http.Request) (*metaWebhookPayload, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("read meta webhook body: %w", err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	var payload metaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse meta webhook payload: %w", err)
	}
	return &payload, nil
}

func mapMetaMessageType(raw string) models.WhatsAppMessageType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "text":
		return models.WhatsAppMessageTypeText
	case "image", "video", "audio", "document", "sticker":
		return models.WhatsAppMessageTypeMedia
	default:
		return models.WhatsAppMessageTypeText
	}
}

func mapMetaDeliveryStatus(raw string) (models.WhatsAppDeliveryStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "sent":
		return models.WhatsAppDeliveryStatusSent, nil
	case "delivered":
		return models.WhatsAppDeliveryStatusDelivered, nil
	case "read":
		return models.WhatsAppDeliveryStatusRead, nil
	case "failed":
		return models.WhatsAppDeliveryStatusFailed, nil
	default:
		return "", fmt.Errorf("unsupported meta delivery status %q", raw)
	}
}

func parseUnixTimestampOrNow(raw string, fallback func() time.Time) time.Time {
	seconds, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback().UTC()
	}
	return time.Unix(seconds, 0).UTC()
}

func toMetaMediaPayload(media models.WhatsAppMediaItem) (string, *metaMediaReference) {
	mime := strings.ToLower(strings.TrimSpace(media.MimeType))
	typeValue := "document"
	switch {
	case strings.HasPrefix(mime, "image/"):
		typeValue = "image"
	case strings.HasPrefix(mime, "video/"):
		typeValue = "video"
	case strings.HasPrefix(mime, "audio/"):
		typeValue = "audio"
	}
	return typeValue, &metaMediaReference{
		Link:    media.URL,
		Caption: media.Caption,
	}
}

func metaWebhookMediaToDomain(media *metaWebhookMedia) []models.WhatsAppMediaItem {
	if media == nil {
		return nil
	}
	return []models.WhatsAppMediaItem{{
		URL:      strings.TrimSpace(media.ID),
		MimeType: strings.TrimSpace(media.MimeType),
		Caption:  strings.TrimSpace(media.Caption),
	}}
}
