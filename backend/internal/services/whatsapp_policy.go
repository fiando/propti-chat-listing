package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

const DefaultWhatsAppCustomerServiceWindow = 24 * time.Hour

var ErrWhatsAppFreeFormOutsideSessionWindow = errors.New("whatsapp free-form messaging is outside customer-service window")

type WhatsAppSessionStore interface {
	TrackCustomerInbound(ctx context.Context, customerID, businessID string, inboundAt time.Time) error
	GetLastCustomerInbound(ctx context.Context, customerID, businessID string) (*time.Time, error)
}

type WhatsAppPolicyOptions struct {
	Now           func() time.Time
	SessionWindow time.Duration
}

type WhatsAppPolicy struct {
	sessionStore  WhatsAppSessionStore
	now           func() time.Time
	sessionWindow time.Duration
}

func NewWhatsAppPolicy(sessionStore WhatsAppSessionStore, opts WhatsAppPolicyOptions) (*WhatsAppPolicy, error) {
	if sessionStore == nil {
		return nil, fmt.Errorf("whatsapp session store is required")
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = func() time.Time {
			return time.Now().UTC()
		}
	}

	sessionWindow := opts.SessionWindow
	if sessionWindow == 0 {
		sessionWindow = DefaultWhatsAppCustomerServiceWindow
	}
	if sessionWindow < 0 {
		return nil, fmt.Errorf("session window must be greater than or equal to zero")
	}

	return &WhatsAppPolicy{
		sessionStore:  sessionStore,
		now:           nowFn,
		sessionWindow: sessionWindow,
	}, nil
}

func (p *WhatsAppPolicy) RecordInboundMessage(ctx context.Context, envelope models.WhatsAppMessageEnvelope) error {
	if envelope.From == "" {
		return fmt.Errorf("record inbound whatsapp session: customer id is required")
	}
	if envelope.To == "" {
		return fmt.Errorf("record inbound whatsapp session: business id is required")
	}

	inboundAt := envelope.ReceivedAt
	if inboundAt.IsZero() {
		inboundAt = p.now()
	}

	if err := p.sessionStore.TrackCustomerInbound(ctx, envelope.From, envelope.To, inboundAt.UTC()); err != nil {
		return fmt.Errorf("record inbound whatsapp session: %w", err)
	}
	return nil
}

func (p *WhatsAppPolicy) EnsureOutboundAllowed(ctx context.Context, req models.WhatsAppSendRequest) error {
	if req.Message.Type == models.WhatsAppMessageTypeTemplate {
		return nil
	}

	if req.To == "" {
		return fmt.Errorf("evaluate whatsapp outbound policy: recipient is required")
	}
	if req.From == "" {
		return fmt.Errorf("evaluate whatsapp outbound policy: sender is required")
	}

	lastInboundAt, err := p.sessionStore.GetLastCustomerInbound(ctx, req.To, req.From)
	if err != nil {
		return fmt.Errorf("evaluate whatsapp outbound policy: %w", err)
	}

	if !IsCustomerServiceWindowOpen(lastInboundAt, p.now(), p.sessionWindow) {
		return ErrWhatsAppFreeFormOutsideSessionWindow
	}

	return nil
}

func IsCustomerServiceWindowOpen(lastCustomerInboundAt *time.Time, now time.Time, sessionWindow time.Duration) bool {
	if lastCustomerInboundAt == nil || sessionWindow <= 0 {
		return false
	}

	elapsed := now.UTC().Sub(lastCustomerInboundAt.UTC())
	if elapsed < 0 {
		return false
	}

	return elapsed <= sessionWindow
}
