package services

import (
	"context"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type fakeWhatsAppMetricsRecorder struct {
	records []models.WhatsAppMetric
	err     error
}

func (f *fakeWhatsAppMetricsRecorder) Record(_ context.Context, metric models.WhatsAppMetric) error {
	if f.err != nil {
		return f.err
	}
	f.records = append(f.records, metric)
	return nil
}

func TestWhatsAppMetricsServiceRecordsChatFirstCompletionWithTierContext(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	recorder := &fakeWhatsAppMetricsRecorder{}
	service := mustNewWhatsAppMetricsService(t, recorder, now)

	err := service.RecordChatFirstCompletion(context.Background(), WhatsAppChatFirstCompletionMetric{
		UserID:            "user-1",
		Tier:              models.SubscriptionBasic,
		Intent:            string(WhatsAppCommandIntentListingCreate),
		ZeroContextSwitch: true,
	})
	if err != nil {
		t.Fatalf("RecordChatFirstCompletion returned error: %v", err)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("expected 1 metric record, got %d", len(recorder.records))
	}
	got := recorder.records[0]
	if got.Event != models.WhatsAppMetricEventChatFirstCompletion {
		t.Fatalf("expected chat-first event, got %q", got.Event)
	}
	if got.Tier != models.SubscriptionBasic {
		t.Fatalf("expected tier basic, got %q", got.Tier)
	}
	if got.OccurredAt != now {
		t.Fatalf("expected occurredAt %s, got %s", now, got.OccurredAt)
	}
}

func TestWhatsAppMetricsServiceRecordVoiceUsageEmitsQuotaPressureWhenThresholdCrossed(t *testing.T) {
	t.Parallel()

	recorder := &fakeWhatsAppMetricsRecorder{}
	service := mustNewWhatsAppMetricsService(t, recorder, time.Date(2026, 4, 10, 11, 0, 0, 0, time.UTC))

	err := service.RecordVoiceUsage(context.Background(), WhatsAppVoiceUsageMetric{
		UserID:              "user-1",
		Tier:                models.SubscriptionPremium,
		VoiceDurationSecond: 120,
		VoiceUsedSeconds:    2900,
		VoiceQuotaSeconds:   3600,
	})
	if err != nil {
		t.Fatalf("RecordVoiceUsage returned error: %v", err)
	}
	if len(recorder.records) != 2 {
		t.Fatalf("expected usage + pressure metrics, got %d records", len(recorder.records))
	}
	if recorder.records[0].Event != models.WhatsAppMetricEventVoiceUsage {
		t.Fatalf("expected first event voice usage, got %q", recorder.records[0].Event)
	}
	if recorder.records[1].Event != models.WhatsAppMetricEventVoiceQuotaPressure {
		t.Fatalf("expected second event voice quota pressure, got %q", recorder.records[1].Event)
	}
}

func TestWhatsAppMetricsServiceRecordVoiceUsageSkipsQuotaPressureBelowThreshold(t *testing.T) {
	t.Parallel()

	recorder := &fakeWhatsAppMetricsRecorder{}
	service := mustNewWhatsAppMetricsService(t, recorder, time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC))

	err := service.RecordVoiceUsage(context.Background(), WhatsAppVoiceUsageMetric{
		UserID:              "user-2",
		Tier:                models.SubscriptionBasic,
		VoiceDurationSecond: 60,
		VoiceUsedSeconds:    300,
		VoiceQuotaSeconds:   1200,
	})
	if err != nil {
		t.Fatalf("RecordVoiceUsage returned error: %v", err)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("expected only usage metric, got %d records", len(recorder.records))
	}
	if recorder.records[0].Event != models.WhatsAppMetricEventVoiceUsage {
		t.Fatalf("expected voice usage event, got %q", recorder.records[0].Event)
	}
}

func TestWhatsAppMetricsServiceRecordsUpgradeAndZeroContextSignalsWithTier(t *testing.T) {
	t.Parallel()

	recorder := &fakeWhatsAppMetricsRecorder{}
	service := mustNewWhatsAppMetricsService(t, recorder, time.Date(2026, 4, 10, 13, 0, 0, 0, time.UTC))
	ctx := context.Background()

	if err := service.RecordUpgradeIntent(ctx, WhatsAppUpgradeIntentMetric{
		UserID:      "user-1",
		Tier:        models.SubscriptionFree,
		UpgradeHint: "upgrade-to-basic",
		Source:      "subscription_status",
	}); err != nil {
		t.Fatalf("RecordUpgradeIntent returned error: %v", err)
	}

	if err := service.RecordUpgradeConversionHint(ctx, WhatsAppUpgradeConversionHintMetric{
		UserID:         "user-1",
		Tier:           models.SubscriptionBasic,
		ConversionHint: "voice_quota_exceeded",
		Source:         "voice_command",
	}); err != nil {
		t.Fatalf("RecordUpgradeConversionHint returned error: %v", err)
	}

	if err := service.RecordZeroContextSwitchCompletion(ctx, WhatsAppZeroContextSwitchCompletionMetric{
		UserID: "user-1",
		Tier:   models.SubscriptionPro,
		Intent: string(WhatsAppCommandIntentListingCreate),
	}); err != nil {
		t.Fatalf("RecordZeroContextSwitchCompletion returned error: %v", err)
	}

	if len(recorder.records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(recorder.records))
	}
	if recorder.records[0].Tier != models.SubscriptionFree || recorder.records[1].Tier != models.SubscriptionBasic || recorder.records[2].Tier != models.SubscriptionPro {
		t.Fatalf("expected tier context to be kept for all events, got %#v", recorder.records)
	}
}

func mustNewWhatsAppMetricsService(t *testing.T, recorder WhatsAppMetricsRecorder, now time.Time) *WhatsAppMetricsService {
	t.Helper()
	service, err := NewWhatsAppMetricsService(recorder, WhatsAppMetricsServiceOptions{
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewWhatsAppMetricsService returned error: %v", err)
	}
	return service
}
