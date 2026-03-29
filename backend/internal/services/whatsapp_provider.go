package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fiando/propti/backend/internal/models"
)

type WhatsAppProvider interface {
	Name() string
	Kind() models.WhatsAppProviderKind
	Capabilities() models.WhatsAppProviderCapabilities
	Send(ctx context.Context, request models.WhatsAppSendRequest) (models.WhatsAppSendResult, error)
	ParseInboundWebhook(ctx context.Context, request *http.Request) (*models.WhatsAppMessageEnvelope, error)
	ParseDeliveryStatusWebhook(ctx context.Context, request *http.Request) ([]models.WhatsAppDeliveryStatusEvent, error)
	VerifyWebhookRequest(request *http.Request) error
}

func parseWhatsAppProviderKind(raw string) (models.WhatsAppProviderKind, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(models.WhatsAppProviderTwilio):
		return models.WhatsAppProviderTwilio, nil
	case string(models.WhatsAppProviderDirectWhatsApp), "direct-whatsapp":
		return models.WhatsAppProviderDirectWhatsApp, nil
	default:
		return "", fmt.Errorf("unsupported whatsapp provider %q", raw)
	}
}

func TwilioWhatsAppCapabilities() models.WhatsAppProviderCapabilities {
	return models.WhatsAppProviderCapabilities{
		SupportsText:                true,
		SupportsTemplate:            true,
		SupportsMedia:               true,
		SupportsInbound:             true,
		SupportsDeliveryStatus:      true,
		SupportsWebhookVerification: true,
	}
}

func DirectWhatsAppCapabilities() models.WhatsAppProviderCapabilities {
	return models.WhatsAppProviderCapabilities{
		SupportsText:                true,
		SupportsTemplate:            false,
		SupportsMedia:               true,
		SupportsInbound:             true,
		SupportsDeliveryStatus:      true,
		SupportsWebhookVerification: true,
	}
}
