package services

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

const (
	whatsAppVoiceDefaultMinConfidence = 0.60
	whatsAppVoiceFallbackMessage      = "Maaf, voice note belum cukup jelas. Coba kirim ulang voice note yang lebih jelas atau kirim perintah dalam teks ya."
	openAIWhisperModel                = "whisper-1"
)

type WhatsAppVoiceStatus string

const (
	WhatsAppVoiceStatusProcessed WhatsAppVoiceStatus = "processed"
	WhatsAppVoiceStatusFallback  WhatsAppVoiceStatus = "fallback"
)

type WhatsAppVoiceRequest struct {
	UserID   string
	Envelope models.WhatsAppMessageEnvelope
}

type WhatsAppVoiceResponse struct {
	Status           WhatsAppVoiceStatus      `json:"status"`
	TranscribedText  string                   `json:"transcribedText,omitempty"`
	Confidence       float64                  `json:"confidence,omitempty"`
	DurationSeconds  int                      `json:"durationSeconds,omitempty"`
	RemainingSeconds int                      `json:"remainingSeconds,omitempty"`
	CommandResponse  *WhatsAppCommandResponse `json:"commandResponse,omitempty"`
	FallbackMessage  string                   `json:"fallbackMessage,omitempty"`
}

type WhatsAppVoiceTranscription struct {
	Text            string
	Confidence      float64
	DurationSeconds int
}

type WhatsAppVoiceUserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Put(ctx context.Context, user *models.User) error
}

type WhatsAppVoiceMediaDownloader interface {
	DownloadMedia(ctx context.Context, envelope models.WhatsAppMessageEnvelope, media models.WhatsAppMediaItem) ([]byte, error)
}

type WhatsAppAudioTranscriber interface {
	TranscribeAudio(ctx context.Context, audio []byte, mimeType string) (*WhatsAppVoiceTranscription, error)
}

type WhatsAppVoiceCommandOrchestrator interface {
	HandleText(ctx context.Context, req WhatsAppCommandRequest) (*WhatsAppCommandResponse, error)
}

type WhatsAppVoiceServiceOptions struct {
	Now           func() time.Time
	MinConfidence float64
	MetricsSink   WhatsAppVoiceMetricsSink
}

type WhatsAppVoiceMetricsSink interface {
	RecordVoiceUsage(ctx context.Context, metric WhatsAppVoiceUsageMetric) error
	RecordUpgradeConversionHint(ctx context.Context, metric WhatsAppUpgradeConversionHintMetric) error
}

type WhatsAppVoiceService struct {
	userStore     WhatsAppVoiceUserStore
	downloader    WhatsAppVoiceMediaDownloader
	transcriber   WhatsAppAudioTranscriber
	orchestrator  WhatsAppVoiceCommandOrchestrator
	now           func() time.Time
	minConfidence float64
	metricsSink   WhatsAppVoiceMetricsSink
}

func NewWhatsAppVoiceService(
	userStore WhatsAppVoiceUserStore,
	downloader WhatsAppVoiceMediaDownloader,
	transcriber WhatsAppAudioTranscriber,
	orchestrator WhatsAppVoiceCommandOrchestrator,
	opts WhatsAppVoiceServiceOptions,
) (*WhatsAppVoiceService, error) {
	if userStore == nil {
		return nil, fmt.Errorf("user store is required")
	}
	if downloader == nil {
		return nil, fmt.Errorf("media downloader is required")
	}
	if transcriber == nil {
		return nil, fmt.Errorf("audio transcriber is required")
	}
	if orchestrator == nil {
		return nil, fmt.Errorf("command orchestrator is required")
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}
	minConfidence := opts.MinConfidence
	if minConfidence <= 0 || minConfidence > 1 {
		minConfidence = whatsAppVoiceDefaultMinConfidence
	}

	return &WhatsAppVoiceService{
		userStore:     userStore,
		downloader:    downloader,
		transcriber:   transcriber,
		orchestrator:  orchestrator,
		now:           nowFn,
		minConfidence: minConfidence,
		metricsSink:   opts.MetricsSink,
	}, nil
}

func (s *WhatsAppVoiceService) HandleInboundVoice(ctx context.Context, req WhatsAppVoiceRequest) (*WhatsAppVoiceResponse, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, utils.NewAppError(400, "user id is required")
	}

	media, err := pickVoiceMedia(req.Envelope.Media)
	if err != nil {
		return nil, err
	}

	user, err := s.userStore.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	tier := effectiveTierForUser(user, s.now())
	quotaSeconds := TierEntitlementFor(tier).VoiceMinutesPerMonth * 60
	if quotaSeconds <= 0 {
		return nil, utils.NewAppError(403, "free tier does not allow WhatsApp voice commands")
	}

	audioBytes, err := s.downloader.DownloadMedia(ctx, req.Envelope, media)
	if err != nil {
		return &WhatsAppVoiceResponse{Status: WhatsAppVoiceStatusFallback, FallbackMessage: whatsAppVoiceFallbackMessage}, nil
	}

	transcription, err := s.transcriber.TranscribeAudio(ctx, audioBytes, media.MimeType)
	if err != nil {
		return &WhatsAppVoiceResponse{Status: WhatsAppVoiceStatusFallback, FallbackMessage: whatsAppVoiceFallbackMessage}, nil
	}

	text := strings.TrimSpace(transcription.Text)
	if text == "" || transcription.Confidence < s.minConfidence {
		return &WhatsAppVoiceResponse{Status: WhatsAppVoiceStatusFallback, FallbackMessage: whatsAppVoiceFallbackMessage}, nil
	}

	currentMonth := s.now().UTC().Format("2006-01")
	usedSeconds := user.Subscription.VoiceSecondsUsed
	if user.Subscription.VoiceUsageMonth != currentMonth {
		usedSeconds = 0
	}

	durationSeconds := transcription.DurationSeconds
	if durationSeconds < 0 {
		durationSeconds = 0
	}
	if usedSeconds+durationSeconds > quotaSeconds {
		s.recordUpgradeConversionHint(ctx, user.UserID, tier, "voice_quota_exceeded")
		return nil, utils.NewAppError(403, fmt.Sprintf("%s tier allows at most %d voice minute(s) per month", tier, TierEntitlementFor(tier).VoiceMinutesPerMonth))
	}

	commandResp, err := s.orchestrator.HandleText(ctx, WhatsAppCommandRequest{UserID: req.UserID, Text: text})
	if err != nil {
		return nil, err
	}

	user.Subscription.VoiceUsageMonth = currentMonth
	user.Subscription.VoiceSecondsUsed = usedSeconds + durationSeconds
	if err := s.userStore.Put(ctx, user); err != nil {
		return nil, utils.ErrInternal
	}
	s.recordVoiceUsageMetric(ctx, user.UserID, tier, durationSeconds, user.Subscription.VoiceSecondsUsed, quotaSeconds)

	return &WhatsAppVoiceResponse{
		Status:           WhatsAppVoiceStatusProcessed,
		TranscribedText:  text,
		Confidence:       transcription.Confidence,
		DurationSeconds:  durationSeconds,
		RemainingSeconds: quotaSeconds - user.Subscription.VoiceSecondsUsed,
		CommandResponse:  commandResp,
	}, nil
}

func (s *WhatsAppVoiceService) recordVoiceUsageMetric(ctx context.Context, userID string, tier models.SubscriptionTier, durationSeconds, usedSeconds, quotaSeconds int) {
	if s.metricsSink == nil {
		return
	}
	_ = s.metricsSink.RecordVoiceUsage(ctx, WhatsAppVoiceUsageMetric{
		UserID:              userID,
		Tier:                tier,
		VoiceDurationSecond: durationSeconds,
		VoiceUsedSeconds:    usedSeconds,
		VoiceQuotaSeconds:   quotaSeconds,
		Source:              "whatsapp_voice",
	})
}

func (s *WhatsAppVoiceService) recordUpgradeConversionHint(ctx context.Context, userID string, tier models.SubscriptionTier, hint string) {
	if s.metricsSink == nil {
		return
	}
	_ = s.metricsSink.RecordUpgradeConversionHint(ctx, WhatsAppUpgradeConversionHintMetric{
		UserID:         userID,
		Tier:           tier,
		ConversionHint: hint,
		Source:         "voice_quota",
	})
}

func pickVoiceMedia(items []models.WhatsAppMediaItem) (models.WhatsAppMediaItem, error) {
	for _, item := range items {
		if strings.TrimSpace(item.URL) == "" {
			continue
		}
		mime := strings.ToLower(strings.TrimSpace(item.MimeType))
		if mime == "" || strings.HasPrefix(mime, "audio/") {
			return item, nil
		}
	}
	return models.WhatsAppMediaItem{}, utils.NewAppError(400, "audio media is required for voice command")
}

func (s *AIService) TranscribeAudio(ctx context.Context, audio []byte, mimeType string) (*WhatsAppVoiceTranscription, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("audio payload is required")
	}
	response, err := s.client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openAIWhisperModel,
		Reader:   bytes.NewReader(audio),
		FilePath: "voice" + extensionFromMimeType(mimeType),
		Format:   openai.AudioResponseFormatVerboseJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("openai transcription request: %w", err)
	}

	text := strings.TrimSpace(response.Text)
	if text == "" {
		return nil, fmt.Errorf("transcription returned empty text")
	}

	duration := int(math.Round(response.Duration))
	if duration < 0 {
		duration = 0
	}

	return &WhatsAppVoiceTranscription{
		Text:            text,
		DurationSeconds: duration,
		Confidence:      transcriptionConfidence(response),
	}, nil
}

func transcriptionConfidence(resp openai.AudioResponse) float64 {
	if len(resp.Segments) == 0 {
		return 0.85
	}

	total := 0.0
	for _, segment := range resp.Segments {
		logProbComponent := clamp01(segment.AvgLogprob + 1)
		noSpeechComponent := clamp01(1 - segment.NoSpeechProb)
		total += (logProbComponent + noSpeechComponent) / 2
	}
	return clamp01(total / float64(len(resp.Segments)))
}

func extensionFromMimeType(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "audio/ogg", "audio/opus", "audio/ogg; codecs=opus":
		return ".ogg"
	case "audio/mpeg":
		return ".mp3"
	case "audio/mp4", "audio/aac":
		return ".m4a"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	default:
		return ".ogg"
	}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
