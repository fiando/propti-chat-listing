package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

const defaultVoiceQuotaPressureThreshold = 0.80

type WhatsAppMetricsRecorder interface {
	Record(ctx context.Context, metric models.WhatsAppMetric) error
}

type WhatsAppMetricsServiceOptions struct {
	Now                         func() time.Time
	VoiceQuotaPressureThreshold float64
}

type WhatsAppMetricsService struct {
	recorder                    WhatsAppMetricsRecorder
	now                         func() time.Time
	voiceQuotaPressureThreshold float64
}

type WhatsAppChatFirstCompletionMetric struct {
	UserID            string
	Tier              models.SubscriptionTier
	Intent            string
	Source            string
	ZeroContextSwitch bool
	Metadata          map[string]string
}

type WhatsAppVoiceUsageMetric struct {
	UserID              string
	Tier                models.SubscriptionTier
	VoiceDurationSecond int
	VoiceUsedSeconds    int
	VoiceQuotaSeconds   int
	Source              string
	Metadata            map[string]string
}

type WhatsAppUpgradeIntentMetric struct {
	UserID      string
	Tier        models.SubscriptionTier
	UpgradeHint string
	Source      string
	Metadata    map[string]string
}

type WhatsAppUpgradeConversionHintMetric struct {
	UserID         string
	Tier           models.SubscriptionTier
	ConversionHint string
	Source         string
	Metadata       map[string]string
}

type WhatsAppZeroContextSwitchCompletionMetric struct {
	UserID   string
	Tier     models.SubscriptionTier
	Intent   string
	Metadata map[string]string
}

func NewWhatsAppMetricsService(recorder WhatsAppMetricsRecorder, opts WhatsAppMetricsServiceOptions) (*WhatsAppMetricsService, error) {
	if recorder == nil {
		return nil, fmt.Errorf("whatsapp metrics recorder is required")
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}

	threshold := opts.VoiceQuotaPressureThreshold
	if threshold <= 0 || threshold > 1 {
		threshold = defaultVoiceQuotaPressureThreshold
	}

	return &WhatsAppMetricsService{
		recorder:                    recorder,
		now:                         nowFn,
		voiceQuotaPressureThreshold: threshold,
	}, nil
}

func (s *WhatsAppMetricsService) RecordChatFirstCompletion(ctx context.Context, m WhatsAppChatFirstCompletionMetric) error {
	return s.record(ctx, models.WhatsAppMetric{
		Event:             models.WhatsAppMetricEventChatFirstCompletion,
		UserID:            strings.TrimSpace(m.UserID),
		Tier:              m.Tier,
		Intent:            strings.TrimSpace(m.Intent),
		Source:            strings.TrimSpace(m.Source),
		ZeroContextSwitch: m.ZeroContextSwitch,
		Metadata:          m.Metadata,
	})
}

func (s *WhatsAppMetricsService) RecordVoiceUsage(ctx context.Context, m WhatsAppVoiceUsageMetric) error {
	used := m.VoiceUsedSeconds
	quota := m.VoiceQuotaSeconds
	remaining := quota - used
	if remaining < 0 {
		remaining = 0
	}

	if err := s.record(ctx, models.WhatsAppMetric{
		Event:               models.WhatsAppMetricEventVoiceUsage,
		UserID:              strings.TrimSpace(m.UserID),
		Tier:                m.Tier,
		Source:              strings.TrimSpace(m.Source),
		VoiceDurationSecond: m.VoiceDurationSecond,
		VoiceUsedSeconds:    used,
		VoiceQuotaSeconds:   quota,
		VoiceRemaining:      remaining,
		Metadata:            m.Metadata,
	}); err != nil {
		return err
	}

	if quota > 0 {
		ratio := float64(used) / float64(quota)
		if ratio >= s.voiceQuotaPressureThreshold {
			return s.record(ctx, models.WhatsAppMetric{
				Event:               models.WhatsAppMetricEventVoiceQuotaPressure,
				UserID:              strings.TrimSpace(m.UserID),
				Tier:                m.Tier,
				Source:              strings.TrimSpace(m.Source),
				VoiceDurationSecond: m.VoiceDurationSecond,
				VoiceUsedSeconds:    used,
				VoiceQuotaSeconds:   quota,
				VoiceRemaining:      remaining,
				Metadata:            m.Metadata,
			})
		}
	}

	return nil
}

func (s *WhatsAppMetricsService) RecordUpgradeIntent(ctx context.Context, m WhatsAppUpgradeIntentMetric) error {
	return s.record(ctx, models.WhatsAppMetric{
		Event:       models.WhatsAppMetricEventUpgradeIntent,
		UserID:      strings.TrimSpace(m.UserID),
		Tier:        m.Tier,
		Source:      strings.TrimSpace(m.Source),
		UpgradeHint: strings.TrimSpace(m.UpgradeHint),
		Metadata:    m.Metadata,
	})
}

func (s *WhatsAppMetricsService) RecordUpgradeConversionHint(ctx context.Context, m WhatsAppUpgradeConversionHintMetric) error {
	return s.record(ctx, models.WhatsAppMetric{
		Event:          models.WhatsAppMetricEventUpgradeConversionHint,
		UserID:         strings.TrimSpace(m.UserID),
		Tier:           m.Tier,
		Source:         strings.TrimSpace(m.Source),
		ConversionHint: strings.TrimSpace(m.ConversionHint),
		Metadata:       m.Metadata,
	})
}

func (s *WhatsAppMetricsService) RecordZeroContextSwitchCompletion(ctx context.Context, m WhatsAppZeroContextSwitchCompletionMetric) error {
	return s.record(ctx, models.WhatsAppMetric{
		Event:             models.WhatsAppMetricEventZeroContextSwitchCompletion,
		UserID:            strings.TrimSpace(m.UserID),
		Tier:              m.Tier,
		Intent:            strings.TrimSpace(m.Intent),
		Source:            "whatsapp",
		ZeroContextSwitch: true,
		Metadata:          m.Metadata,
	})
}

func (s *WhatsAppMetricsService) record(ctx context.Context, metric models.WhatsAppMetric) error {
	if metric.OccurredAt.IsZero() {
		metric.OccurredAt = s.now().UTC()
	}
	return s.recorder.Record(ctx, metric)
}
