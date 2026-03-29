package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type fakeWhatsAppSessionStore struct {
	trackCalls       int
	getCalls         int
	trackedCustomer  string
	trackedBusiness  string
	trackedInboundAt time.Time
	lastInboundAt    *time.Time
	trackErr         error
	getErr           error
}

func (f *fakeWhatsAppSessionStore) TrackCustomerInbound(_ context.Context, customerID, businessID string, inboundAt time.Time) error {
	f.trackCalls++
	f.trackedCustomer = customerID
	f.trackedBusiness = businessID
	f.trackedInboundAt = inboundAt
	if f.trackErr != nil {
		return f.trackErr
	}
	return nil
}

func (f *fakeWhatsAppSessionStore) GetLastCustomerInbound(_ context.Context, customerID, businessID string) (*time.Time, error) {
	f.getCalls++
	f.trackedCustomer = customerID
	f.trackedBusiness = businessID
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.lastInboundAt, nil
}

func TestWhatsAppPolicyRecordInboundMessageTracksSession(t *testing.T) {
	store := &fakeWhatsAppSessionStore{}
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	receivedAt := time.Date(2026, 4, 5, 15, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
	err = policy.RecordInboundMessage(context.Background(), models.WhatsAppMessageEnvelope{
		From:       "customer-1",
		To:         "business-1",
		ReceivedAt: receivedAt,
	})
	if err != nil {
		t.Fatalf("RecordInboundMessage returned error: %v", err)
	}

	if store.trackCalls != 1 {
		t.Fatalf("expected 1 track call, got %d", store.trackCalls)
	}
	if store.trackedCustomer != "customer-1" || store.trackedBusiness != "business-1" {
		t.Fatalf("unexpected tracked participants: customer=%q business=%q", store.trackedCustomer, store.trackedBusiness)
	}
	if !store.trackedInboundAt.Equal(receivedAt.UTC()) {
		t.Fatalf("expected UTC inbound timestamp %s, got %s", receivedAt.UTC(), store.trackedInboundAt)
	}
}

func TestWhatsAppPolicyRecordInboundMessageDefaultsTimestampFromClock(t *testing.T) {
	store := &fakeWhatsAppSessionStore{}
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.RecordInboundMessage(context.Background(), models.WhatsAppMessageEnvelope{From: "customer-1", To: "business-1"})
	if err != nil {
		t.Fatalf("RecordInboundMessage returned error: %v", err)
	}
	if !store.trackedInboundAt.Equal(now) {
		t.Fatalf("expected default inbound timestamp %s, got %s", now, store.trackedInboundAt)
	}
}

func TestWhatsAppPolicyEnsureOutboundAllowedAllowsTemplateOutOfWindow(t *testing.T) {
	store := &fakeWhatsAppSessionStore{}
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.EnsureOutboundAllowed(context.Background(), models.WhatsAppSendRequest{
		To:   "customer-1",
		From: "business-1",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeTemplate,
		},
	})
	if err != nil {
		t.Fatalf("EnsureOutboundAllowed returned error for template: %v", err)
	}
	if store.getCalls != 0 {
		t.Fatalf("expected no session lookup for template message, got %d", store.getCalls)
	}
}

func TestWhatsAppPolicyEnsureOutboundAllowedBlocksWhenNoInboundSession(t *testing.T) {
	store := &fakeWhatsAppSessionStore{}
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.EnsureOutboundAllowed(context.Background(), models.WhatsAppSendRequest{
		To:   "customer-1",
		From: "business-1",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: "halo",
		},
	})
	if !errors.Is(err, ErrWhatsAppFreeFormOutsideSessionWindow) {
		t.Fatalf("expected ErrWhatsAppFreeFormOutsideSessionWindow, got %v", err)
	}
}

func TestWhatsAppPolicyEnsureOutboundAllowedAllowsFreeFormWithinSessionWindow(t *testing.T) {
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	lastInbound := now.Add(-23 * time.Hour)
	store := &fakeWhatsAppSessionStore{lastInboundAt: &lastInbound}
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.EnsureOutboundAllowed(context.Background(), models.WhatsAppSendRequest{
		To:   "customer-1",
		From: "business-1",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeMedia,
		},
	})
	if err != nil {
		t.Fatalf("EnsureOutboundAllowed returned error: %v", err)
	}
}

func TestWhatsAppPolicyEnsureOutboundAllowedBlocksAfterSessionWindowCloses(t *testing.T) {
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	lastInbound := now.Add(-24*time.Hour - time.Second)
	store := &fakeWhatsAppSessionStore{lastInboundAt: &lastInbound}
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.EnsureOutboundAllowed(context.Background(), models.WhatsAppSendRequest{
		To:   "customer-1",
		From: "business-1",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: "halo",
		},
	})
	if !errors.Is(err, ErrWhatsAppFreeFormOutsideSessionWindow) {
		t.Fatalf("expected ErrWhatsAppFreeFormOutsideSessionWindow, got %v", err)
	}
}

func TestWhatsAppPolicyEnsureOutboundAllowedWrapsSessionLookupErrors(t *testing.T) {
	store := &fakeWhatsAppSessionStore{getErr: errors.New("dynamo unavailable")}
	policy, err := NewWhatsAppPolicy(store, WhatsAppPolicyOptions{})
	if err != nil {
		t.Fatalf("NewWhatsAppPolicy returned error: %v", err)
	}

	err = policy.EnsureOutboundAllowed(context.Background(), models.WhatsAppSendRequest{
		To:   "customer-1",
		From: "business-1",
		Message: models.WhatsAppOutboundMessage{
			Type: models.WhatsAppMessageTypeText,
			Text: "halo",
		},
	})
	if err == nil || err.Error() == "dynamo unavailable" {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestIsCustomerServiceWindowOpen(t *testing.T) {
	now := time.Date(2026, 4, 5, 8, 30, 0, 0, time.UTC)
	inside := now.Add(-2 * time.Hour)
	outside := now.Add(-25 * time.Hour)

	if !IsCustomerServiceWindowOpen(&inside, now, 24*time.Hour) {
		t.Fatal("expected inside timestamp to be open")
	}
	if IsCustomerServiceWindowOpen(&outside, now, 24*time.Hour) {
		t.Fatal("expected outside timestamp to be closed")
	}
	if IsCustomerServiceWindowOpen(nil, now, 24*time.Hour) {
		t.Fatal("expected nil timestamp to be closed")
	}
}
