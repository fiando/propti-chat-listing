package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakeWhatsAppVoiceUserStore struct {
	user      *models.User
	getErr    error
	putErr    error
	putCalls  int
	storedUser *models.User
}

func (f *fakeWhatsAppVoiceUserStore) GetByID(_ context.Context, _ string) (*models.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.user == nil {
		return nil, nil
	}
	copy := *f.user
	return &copy, nil
}

func (f *fakeWhatsAppVoiceUserStore) Put(_ context.Context, user *models.User) error {
	f.putCalls++
	if f.putErr != nil {
		return f.putErr
	}
	if user != nil {
		copy := *user
		f.storedUser = &copy
	}
	return nil
}

type fakeWhatsAppVoiceMediaDownloader struct {
	data  []byte
	err   error
	calls int
}

func (f *fakeWhatsAppVoiceMediaDownloader) DownloadMedia(_ context.Context, _ models.WhatsAppMessageEnvelope, _ models.WhatsAppMediaItem) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return append([]byte(nil), f.data...), nil
}

type fakeWhatsAppAudioTranscriber struct {
	result WhatsAppVoiceTranscription
	err    error
	calls  int
}

func (f *fakeWhatsAppAudioTranscriber) TranscribeAudio(_ context.Context, _ []byte, _ string) (*WhatsAppVoiceTranscription, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	copy := f.result
	return &copy, nil
}

type fakeWhatsAppVoiceOrchestrator struct {
	response *WhatsAppCommandResponse
	err      error
	calls    int
	lastReq  WhatsAppCommandRequest
}

func (f *fakeWhatsAppVoiceOrchestrator) HandleText(_ context.Context, req WhatsAppCommandRequest) (*WhatsAppCommandResponse, error) {
	f.calls++
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.response != nil {
		copy := *f.response
		return &copy, nil
	}
	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingCreate, Message: "ok"}, nil
}

func TestWhatsAppVoiceServiceBlocksFreeTier(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	service := mustNewWhatsAppVoiceService(t,
		&fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionFree}}},
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{result: WhatsAppVoiceTranscription{Text: "cari rumah", DurationSeconds: 30, Confidence: 0.9}},
		&fakeWhatsAppVoiceOrchestrator{},
		now,
	)

	_, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err == nil {
		t.Fatal("expected free tier to be blocked")
	}
	if appCodeFromVoiceErr(t, err) != 403 {
		t.Fatalf("expected 403, got %d (%v)", appCodeFromVoiceErr(t, err), err)
	}
}

func TestWhatsAppVoiceServiceRoutesTranscriptionAndConsumesQuota(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)
	renew := now.Add(24 * time.Hour)
	users := &fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renew, VoiceUsageMonth: "2026-02", VoiceSecondsUsed: 180}}}
	orchestrator := &fakeWhatsAppVoiceOrchestrator{response: &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentSearch, Message: "search ok"}}
	service := mustNewWhatsAppVoiceService(t,
		users,
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{result: WhatsAppVoiceTranscription{Text: "cari rumah di jogja", DurationSeconds: 120, Confidence: 0.93}},
		orchestrator,
		now,
	)

	resp, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err != nil {
		t.Fatalf("HandleInboundVoice returned error: %v", err)
	}
	if resp.Status != WhatsAppVoiceStatusProcessed {
		t.Fatalf("expected status processed, got %q", resp.Status)
	}
	if resp.CommandResponse == nil || resp.CommandResponse.Intent != WhatsAppCommandIntentSearch {
		t.Fatalf("expected orchestrator response, got %#v", resp.CommandResponse)
	}
	if orchestrator.calls != 1 || orchestrator.lastReq.Text != "cari rumah di jogja" {
		t.Fatalf("expected orchestrator called with transcription, calls=%d text=%q", orchestrator.calls, orchestrator.lastReq.Text)
	}
	if users.putCalls != 1 || users.storedUser == nil {
		t.Fatalf("expected user usage to be persisted once, calls=%d", users.putCalls)
	}
	if users.storedUser.Subscription.VoiceSecondsUsed != 300 {
		t.Fatalf("expected voice seconds usage to become 300, got %d", users.storedUser.Subscription.VoiceSecondsUsed)
	}
}

func TestWhatsAppVoiceServiceBlocksWhenQuotaExceeded(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	renew := now.Add(24 * time.Hour)
	users := &fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renew, VoiceUsageMonth: "2026-02", VoiceSecondsUsed: 1190}}}
	service := mustNewWhatsAppVoiceService(t,
		users,
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{result: WhatsAppVoiceTranscription{Text: "cari rumah", DurationSeconds: 30, Confidence: 0.9}},
		&fakeWhatsAppVoiceOrchestrator{},
		now,
	)

	_, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err == nil {
		t.Fatal("expected quota exceeded error")
	}
	if appCodeFromVoiceErr(t, err) != 403 {
		t.Fatalf("expected 403 quota error, got %d (%v)", appCodeFromVoiceErr(t, err), err)
	}
	if users.putCalls != 0 {
		t.Fatalf("expected no usage persistence on blocked quota, got %d", users.putCalls)
	}
}

func TestWhatsAppVoiceServiceTranscriptionFailureReturnsFallback(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	renew := now.Add(24 * time.Hour)
	service := mustNewWhatsAppVoiceService(t,
		&fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: &renew}}},
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{err: errors.New("openai timeout")},
		&fakeWhatsAppVoiceOrchestrator{},
		now,
	)

	resp, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err != nil {
		t.Fatalf("expected fallback response without hard error, got %v", err)
	}
	if resp.Status != WhatsAppVoiceStatusFallback {
		t.Fatalf("expected fallback status, got %q", resp.Status)
	}
	if resp.FallbackMessage == "" {
		t.Fatal("expected explicit fallback message")
	}
	if resp.CommandResponse != nil {
		t.Fatalf("expected no orchestrator response on transcription failure, got %#v", resp.CommandResponse)
	}
}

func TestWhatsAppVoiceServiceLowConfidenceReturnsFallback(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	renew := now.Add(24 * time.Hour)
	orchestrator := &fakeWhatsAppVoiceOrchestrator{}
	service := mustNewWhatsAppVoiceService(t,
		&fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPro, RenewDate: &renew}}},
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{result: WhatsAppVoiceTranscription{Text: "kurang jelas", DurationSeconds: 50, Confidence: 0.42}},
		orchestrator,
		now,
	)

	resp, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err != nil {
		t.Fatalf("expected fallback response without hard error, got %v", err)
	}
	if resp.Status != WhatsAppVoiceStatusFallback {
		t.Fatalf("expected fallback status, got %q", resp.Status)
	}
	if resp.FallbackMessage == "" {
		t.Fatal("expected explicit fallback message")
	}
	if orchestrator.calls != 0 {
		t.Fatalf("expected low-confidence transcript not routed to orchestrator, got %d calls", orchestrator.calls)
	}
}

func TestWhatsAppVoiceServiceResetsMonthlyQuotaWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	renew := now.Add(24 * time.Hour)
	users := &fakeWhatsAppVoiceUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: &renew, VoiceUsageMonth: "2026-02", VoiceSecondsUsed: 3599}}}
	service := mustNewWhatsAppVoiceService(t,
		users,
		&fakeWhatsAppVoiceMediaDownloader{data: []byte("audio")},
		&fakeWhatsAppAudioTranscriber{result: WhatsAppVoiceTranscription{Text: "listing saya", DurationSeconds: 60, Confidence: 0.95}},
		&fakeWhatsAppVoiceOrchestrator{response: &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingRead}},
		now,
	)

	resp, err := service.HandleInboundVoice(context.Background(), WhatsAppVoiceRequest{UserID: "user-1", Envelope: sampleVoiceEnvelope("audio/ogg")})
	if err != nil {
		t.Fatalf("HandleInboundVoice returned error: %v", err)
	}
	if resp.RemainingSeconds != 3540 {
		t.Fatalf("expected remaining seconds 3540 for premium after monthly reset, got %d", resp.RemainingSeconds)
	}
	if users.storedUser == nil || users.storedUser.Subscription.VoiceUsageMonth != "2026-03" {
		t.Fatalf("expected usage month reset to current month, got %#v", users.storedUser)
	}
}

func mustNewWhatsAppVoiceService(
	t *testing.T,
	users WhatsAppVoiceUserStore,
	downloader WhatsAppVoiceMediaDownloader,
	transcriber WhatsAppAudioTranscriber,
	orchestrator WhatsAppVoiceCommandOrchestrator,
	now time.Time,
) *WhatsAppVoiceService {
	t.Helper()
	service, err := NewWhatsAppVoiceService(users, downloader, transcriber, orchestrator, WhatsAppVoiceServiceOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppVoiceService returned error: %v", err)
	}
	return service
}

func sampleVoiceEnvelope(mime string) models.WhatsAppMessageEnvelope {
	return models.WhatsAppMessageEnvelope{
		From: "whatsapp:+6281234567890",
		Type: models.WhatsAppMessageTypeMedia,
		Media: []models.WhatsAppMediaItem{{URL: "https://example.com/voice.ogg", MimeType: mime}},
	}
}

func appCodeFromVoiceErr(t *testing.T, err error) int {
	t.Helper()
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T (%v)", err, err)
	}
	return appErr.Code
}
