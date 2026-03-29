package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/fiando/propti/backend/internal/data"
	"github.com/fiando/propti/backend/internal/handlers"
	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type whatsAppDeliveryStatusLogger struct{}

func (l *whatsAppDeliveryStatusLogger) HandleDeliveryStatus(_ context.Context, event models.WhatsAppDeliveryStatusEvent) error {
	utils.LogInfo("whatsapp delivery status",
		"provider", event.Provider,
		"providerMessageId", event.ProviderMessageID,
		"status", event.Status,
		"errorCode", event.ErrorCode,
		"errorMessage", event.ErrorMessage,
	)
	return nil
}

type whatsAppMetricsLogger struct{}

func (l *whatsAppMetricsLogger) Record(_ context.Context, metric models.WhatsAppMetric) error {
	utils.LogInfo("whatsapp metric",
		"event", metric.Event,
		"userId", metric.UserID,
		"tier", metric.Tier,
		"intent", metric.Intent,
		"source", metric.Source,
		"zeroContextSwitch", metric.ZeroContextSwitch,
		"voiceDurationSecond", metric.VoiceDurationSecond,
		"voiceUsedSeconds", metric.VoiceUsedSeconds,
		"voiceQuotaSeconds", metric.VoiceQuotaSeconds,
		"voiceRemaining", metric.VoiceRemaining,
		"upgradeHint", metric.UpgradeHint,
		"conversionHint", metric.ConversionHint,
		"occurredAt", metric.OccurredAt,
		"metadata", metric.Metadata,
	)
	return nil
}

type whatsAppHTTPMediaDownloader struct {
	client           *http.Client
	twilioAccountSID string
	twilioAuthToken  string
}

func (d *whatsAppHTTPMediaDownloader) DownloadMedia(ctx context.Context, _ models.WhatsAppMessageEnvelope, media models.WhatsAppMediaItem) ([]byte, error) {
	if strings.TrimSpace(media.URL) == "" {
		return nil, fmt.Errorf("voice media url is required")
	}
	client := d.client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, media.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build voice media request: %w", err)
	}
	if d.twilioAccountSID != "" && d.twilioAuthToken != "" && strings.Contains(strings.ToLower(media.URL), "twilio.com") {
		req.SetBasicAuth(d.twilioAccountSID, d.twilioAuthToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download voice media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("download voice media failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read voice media response: %w", err)
	}
	return body, nil
}

func main() {
	ctx := context.Background()

	db, err := repository.NewDynamoDB(ctx)
	if err != nil {
		utils.LogError("init dynamodb", err)
		panic(err)
	}

	userRepo := repository.NewUserRepo(db)
	listingRepo := repository.NewListingRepo(db)
	sessionRepo := repository.NewWhatsAppSessionRepo(db)
	otpRepo := repository.NewOTPRepo(db)

	identitySvc, err := services.NewWhatsAppIdentityService(userRepo, otpRepo, services.WhatsAppIdentityOptions{})
	if err != nil {
		utils.LogError("init whatsapp identity service", err)
		panic(err)
	}

	policy, err := services.NewWhatsAppPolicy(sessionRepo, services.WhatsAppPolicyOptions{})
	if err != nil {
		utils.LogError("init whatsapp policy", err)
		panic(err)
	}

	var aiSvc *services.AIService
	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		aiSvc = services.NewAIService(key)
	}

	var s3Svc services.MediaStorage
	if bucket := strings.TrimSpace(os.Getenv("S3_MEDIA_BUCKET")); bucket != "" {
		s3Svc, err = services.NewS3Service(ctx, bucket)
		if err != nil {
			utils.LogError("init s3 service", err)
			panic(err)
		}
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()
	locationCatalog, err := services.NewLocationCatalogFromReader(bytes.NewReader(data.IndonesiaLocationsJSON))
	if err != nil {
		utils.LogError("init location catalog", err)
		panic(err)
	}

	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc, locationCatalog)
	listingSvc.SetWriteEligibilityGuard(identitySvc)
	searchIntentSvc := services.NewSearchIntentService(aiSvc, locationCatalog)
	metricsSvc, err := services.NewWhatsAppMetricsService(&whatsAppMetricsLogger{}, services.WhatsAppMetricsServiceOptions{})
	if err != nil {
		utils.LogError("init whatsapp metrics service", err)
		panic(err)
	}

	commandOrchestrator, err := services.NewWhatsAppCommandOrchestrator(
		listingSvc,
		searchIntentSvc,
		userRepo,
		identitySvc,
		services.WhatsAppCommandOrchestratorOptions{MetricsSink: metricsSvc},
	)
	if err != nil {
		utils.LogError("init whatsapp command orchestrator", err)
		panic(err)
	}

	var voiceSvc *services.WhatsAppVoiceService
	if aiSvc != nil {
		voiceSvc, err = services.NewWhatsAppVoiceService(
			userRepo,
			&whatsAppHTTPMediaDownloader{
				client:           http.DefaultClient,
				twilioAccountSID: strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID")),
				twilioAuthToken:  strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN")),
			},
			aiSvc,
			commandOrchestrator,
			services.WhatsAppVoiceServiceOptions{MetricsSink: metricsSvc},
		)
		if err != nil {
			utils.LogError("init whatsapp voice service", err)
			panic(err)
		}
	}

	provider, metaVerifyToken, err := buildWhatsAppProvider()
	if err != nil {
		utils.LogError("init whatsapp provider", err)
		panic(err)
	}

	handler := handlers.NewWhatsAppHandler(handlers.WhatsAppHandlerDependencies{
		Provider:            provider,
		UserStore:           userRepo,
		CommandOrchestrator: commandOrchestrator,
		VoiceService:        voiceSvc,
		Policy:              policy,
		StatusSink:          &whatsAppDeliveryStatusLogger{},
		MetaVerifyToken:     metaVerifyToken,
	})

	lambda.Start(handler.Handle)
}

func buildWhatsAppProvider() (services.WhatsAppProvider, string, error) {
	rawProvider := strings.ToLower(strings.TrimSpace(os.Getenv("WHATSAPP_PROVIDER")))
	switch rawProvider {
	case "", "twilio":
		return services.NewTwilioWhatsAppProvider(services.TwilioWhatsAppProviderConfig{
			AccountSID: strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID")),
			AuthToken:  strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN")),
			From:       strings.TrimSpace(os.Getenv("TWILIO_WHATSAPP_FROM")),
			APIBaseURL: strings.TrimSpace(os.Getenv("TWILIO_API_BASE_URL")),
		}), "", nil
	case "meta", "direct_whatsapp", "direct-whatsapp":
		verifyToken := strings.TrimSpace(os.Getenv("META_WHATSAPP_VERIFY_TOKEN"))
		return services.NewMetaWhatsAppProvider(services.MetaWhatsAppProviderConfig{
			AccessToken:   strings.TrimSpace(os.Getenv("META_WHATSAPP_ACCESS_TOKEN")),
			PhoneNumberID: strings.TrimSpace(os.Getenv("META_WHATSAPP_PHONE_NUMBER_ID")),
			VerifyToken:   verifyToken,
			AppSecret:     strings.TrimSpace(os.Getenv("META_WHATSAPP_APP_SECRET")),
			APIBaseURL:    strings.TrimSpace(os.Getenv("META_WHATSAPP_API_BASE_URL")),
			APIVersion:    strings.TrimSpace(os.Getenv("META_WHATSAPP_API_VERSION")),
		}), verifyToken, nil
	default:
		return nil, "", fmt.Errorf("unsupported WHATSAPP_PROVIDER value %q", rawProvider)
	}
}
