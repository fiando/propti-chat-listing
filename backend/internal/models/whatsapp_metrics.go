package models

import "time"

type WhatsAppMetricEvent string

const (
	WhatsAppMetricEventChatFirstCompletion         WhatsAppMetricEvent = "whatsapp_chat_first_completion"
	WhatsAppMetricEventVoiceUsage                  WhatsAppMetricEvent = "whatsapp_voice_usage"
	WhatsAppMetricEventVoiceQuotaPressure          WhatsAppMetricEvent = "whatsapp_voice_quota_pressure"
	WhatsAppMetricEventUpgradeIntent               WhatsAppMetricEvent = "whatsapp_upgrade_intent"
	WhatsAppMetricEventUpgradeConversionHint       WhatsAppMetricEvent = "whatsapp_upgrade_conversion_hint"
	WhatsAppMetricEventZeroContextSwitchCompletion WhatsAppMetricEvent = "whatsapp_zero_context_switch_completion"
)

type WhatsAppMetric struct {
	Event               WhatsAppMetricEvent `json:"event"`
	UserID              string              `json:"userId,omitempty"`
	Tier                SubscriptionTier    `json:"tier"`
	Intent              string              `json:"intent,omitempty"`
	Source              string              `json:"source,omitempty"`
	ZeroContextSwitch   bool                `json:"zeroContextSwitch"`
	VoiceDurationSecond int                 `json:"voiceDurationSecond,omitempty"`
	VoiceUsedSeconds    int                 `json:"voiceUsedSeconds,omitempty"`
	VoiceQuotaSeconds   int                 `json:"voiceQuotaSeconds,omitempty"`
	VoiceRemaining      int                 `json:"voiceRemaining,omitempty"`
	UpgradeHint         string              `json:"upgradeHint,omitempty"`
	ConversionHint      string              `json:"conversionHint,omitempty"`
	OccurredAt          time.Time           `json:"occurredAt"`
	Metadata            map[string]string   `json:"metadata,omitempty"`
}
