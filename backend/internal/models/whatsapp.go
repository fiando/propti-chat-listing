package models

import "time"

type WhatsAppProviderKind string

type WhatsAppMessageType string

type WhatsAppDeliveryStatus string

const (
	WhatsAppProviderTwilio         WhatsAppProviderKind = "twilio"
	WhatsAppProviderDirectWhatsApp WhatsAppProviderKind = "direct_whatsapp"
)

const (
	WhatsAppMessageTypeText     WhatsAppMessageType = "text"
	WhatsAppMessageTypeTemplate WhatsAppMessageType = "template"
	WhatsAppMessageTypeMedia    WhatsAppMessageType = "media"
)

const (
	WhatsAppDeliveryStatusQueued    WhatsAppDeliveryStatus = "queued"
	WhatsAppDeliveryStatusSent      WhatsAppDeliveryStatus = "sent"
	WhatsAppDeliveryStatusDelivered WhatsAppDeliveryStatus = "delivered"
	WhatsAppDeliveryStatusRead      WhatsAppDeliveryStatus = "read"
	WhatsAppDeliveryStatusFailed    WhatsAppDeliveryStatus = "failed"
)

type WhatsAppProviderCapabilities struct {
	SupportsText                bool `json:"supportsText"`
	SupportsTemplate            bool `json:"supportsTemplate"`
	SupportsMedia               bool `json:"supportsMedia"`
	SupportsInbound             bool `json:"supportsInbound"`
	SupportsDeliveryStatus      bool `json:"supportsDeliveryStatus"`
	SupportsWebhookVerification bool `json:"supportsWebhookVerification"`
}

type WhatsAppMediaItem struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type WhatsAppTemplateParameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

type WhatsAppTemplatePayload struct {
	Name       string                      `json:"name"`
	Language   string                      `json:"language,omitempty"`
	Parameters []WhatsAppTemplateParameter `json:"parameters,omitempty"`
}

type WhatsAppOutboundMessage struct {
	Type     WhatsAppMessageType      `json:"type"`
	Text     string                   `json:"text,omitempty"`
	Template *WhatsAppTemplatePayload `json:"template,omitempty"`
	Media    []WhatsAppMediaItem      `json:"media,omitempty"`
}

type WhatsAppSendRequest struct {
	To       string                  `json:"to"`
	From     string                  `json:"from,omitempty"`
	Message  WhatsAppOutboundMessage `json:"message"`
	Metadata map[string]string       `json:"metadata,omitempty"`
}

type WhatsAppSendResult struct {
	Provider          WhatsAppProviderKind `json:"provider"`
	ProviderMessageID string               `json:"providerMessageId"`
	AcceptedAt        time.Time            `json:"acceptedAt"`
}

type WhatsAppMessageEnvelope struct {
	Provider          WhatsAppProviderKind `json:"provider"`
	ProviderMessageID string               `json:"providerMessageId,omitempty"`
	From              string               `json:"from"`
	To                string               `json:"to,omitempty"`
	Type              WhatsAppMessageType  `json:"type"`
	Text              string               `json:"text,omitempty"`
	Media             []WhatsAppMediaItem  `json:"media,omitempty"`
	ReceivedAt        time.Time            `json:"receivedAt"`
	Metadata          map[string]string    `json:"metadata,omitempty"`
}

type WhatsAppDeliveryStatusEvent struct {
	Provider          WhatsAppProviderKind   `json:"provider"`
	ProviderMessageID string                 `json:"providerMessageId"`
	Status            WhatsAppDeliveryStatus `json:"status"`
	OccurredAt        time.Time              `json:"occurredAt"`
	ErrorCode         string                 `json:"errorCode,omitempty"`
	ErrorMessage      string                 `json:"errorMessage,omitempty"`
	Metadata          map[string]string      `json:"metadata,omitempty"`
}
