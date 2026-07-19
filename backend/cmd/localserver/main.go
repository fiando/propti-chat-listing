package main

// localserver is a thin HTTP wrapper around the Lambda handlers for local development.
// It avoids the need for Docker/SAM local and works with rootless Podman or any environment.
//
// Usage:
//
//	go run ./cmd/localserver
//	PORT=3001 go run ./cmd/localserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fiando/propti/backend/internal/data"
	"github.com/fiando/propti/backend/internal/handlers"
	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/payments"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type lambdaHandlerFunc func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// ── Local-only WhatsApp sinks ────────────────────────────────────────────────

type localDeliveryStatusSink struct{}

func (l *localDeliveryStatusSink) HandleDeliveryStatus(_ context.Context, event models.WhatsAppDeliveryStatusEvent) error {
	utils.LogInfo("whatsapp delivery status",
		"provider", event.Provider,
		"status", event.Status,
	)
	return nil
}

type localMetricsLogger struct{}

func (l *localMetricsLogger) Record(_ context.Context, metric models.WhatsAppMetric) error {
	utils.LogInfo("whatsapp metric", "event", metric.Event, "userId", metric.UserID)
	return nil
}

type localMediaDownloader struct {
	twilioAccountSID string
	twilioAuthToken  string
}

func (d *localMediaDownloader) DownloadMedia(ctx context.Context, _ models.WhatsAppMessageEnvelope, media models.WhatsAppMediaItem) ([]byte, error) {
	if strings.TrimSpace(media.URL) == "" {
		return nil, fmt.Errorf("voice media url is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, media.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build voice media request: %w", err)
	}
	if d.twilioAccountSID != "" && d.twilioAuthToken != "" && strings.Contains(strings.ToLower(media.URL), "twilio.com") {
		req.SetBasicAuth(d.twilioAccountSID, d.twilioAuthToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download voice media: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read voice media: %w", err)
	}
	return body, nil
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	ctx := context.Background()

	// ── Shared DynamoDB ──────────────────────────────────────────────────────
	db, err := repository.NewDynamoDB(ctx)
	if err != nil {
		log.Fatalf("init dynamodb: %v", err)
	}

	listingRepo := repository.NewListingRepo(db)
	userRepo := repository.NewUserRepo(db)
	leadRepo := repository.NewLeadRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	otpRepo := repository.NewOTPRepo(db)
	uploadSessionRepo := repository.NewUploadSessionRepo(db)
	sessionRepo := repository.NewWhatsAppSessionRepo(db)
	moderationRepo := repository.NewModerationRepo(db)

	// ── Optional AI & S3 ────────────────────────────────────────────────────
	var aiSvc *services.AIService
	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		aiSvc = services.NewAIService(key)
	}

	var s3Svc services.MediaStorage
	if bucket := strings.TrimSpace(os.Getenv("S3_MEDIA_BUCKET")); bucket != "" {
		s3Svc, err = services.NewS3Service(ctx, bucket)
		if err != nil {
			utils.LogError("init s3 service (non-fatal in local dev)", err)
		}
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()

	locationCatalog, err := services.NewLocationCatalogFromReader(bytes.NewReader(data.IndonesiaLocationsJSON))
	if err != nil {
		log.Fatalf("init location catalog: %v", err)
	}

	var aiParseService services.AIParseService
	if aiSvc != nil {
		aiParseService = aiSvc
	}

	// ── Listing & Search services ────────────────────────────────────────────
	listingSvc := services.NewListingService(listingRepo, userRepo, aiParseService, s3Svc, mapsSvc, locationCatalog)
	listingSvc.SetUploadSessionStore(uploadSessionRepo)
	uploadSessionSvc := services.NewUploadSessionService(uploadSessionRepo, userRepo, listingRepo, s3Svc)
	searchIntentSvc := services.NewSearchIntentService(aiSvc, locationCatalog)
	leadSvc := services.NewLeadService(leadRepo)

	// Optional image moderation
	var imageModerator services.ImageModerator
	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		imageModerator, err = services.NewImageModerator(key)
		if err != nil {
			utils.LogError("init image moderator (non-fatal)", err)
		}
	}
	var textModerator services.ContentModerator
	if aiSvc != nil {
		textModerator = aiSvc
	}
	_ = services.NewModerationService(textModerator, imageModerator, moderationRepo, listingRepo, s3Svc)

	// ── WhatsApp Identity ────────────────────────────────────────────────────
	identitySvc, err := services.NewWhatsAppIdentityService(userRepo, otpRepo, services.WhatsAppIdentityOptions{
		OTPExpiry:             10 * time.Minute,
		WhatsAppMessageTarget: strings.TrimSpace(os.Getenv("WHATSAPP_MESSAGE_TARGET")),
	})
	if err != nil {
		utils.LogError("init whatsapp identity (non-fatal)", err)
	}
	if identitySvc != nil {
		listingSvc.SetWriteEligibilityGuard(identitySvc)
	}

	// ── Payment ──────────────────────────────────────────────────────────────
	dokuBaseURL := "https://api-sandbox.doku.com"
	if os.Getenv("DOKU_ENV") == "production" {
		dokuBaseURL = "https://api.doku.com"
	}
	paymentProvider := payments.NewDOKUProvider(payments.DOKUConfig{
		ClientID:  os.Getenv("DOKU_CLIENT_ID"),
		SecretKey: os.Getenv("DOKU_SECRET_KEY"),
		BaseURL:   dokuBaseURL,
	})

	// ── Handler wiring ───────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(db)
	listingHandler := handlers.NewListingHandler(listingSvc, uploadSessionSvc, services.NewListingMediaPresenter(s3Svc))
	searchHandler := handlers.NewSearchHandler(listingRepo, mapsSvc, locationCatalog, searchIntentSvc)
	leadHandler := handlers.NewLeadHandler(leadSvc)
	premiumHandler := handlers.NewPremiumHandler(userRepo, transactionRepo, listingSvc, paymentProvider)

	// WhatsApp handler (optional — only if provider is configured)
	var whatsappHandler *handlers.WhatsAppHandler
	if waProvider, metaVerifyToken, waErr := buildWhatsAppProvider(); waErr == nil {
		policy, polErr := services.NewWhatsAppPolicy(sessionRepo, services.WhatsAppPolicyOptions{})
		if polErr != nil {
			utils.LogError("init whatsapp policy (non-fatal)", polErr)
		}
		metricsSvc, _ := services.NewWhatsAppMetricsService(&localMetricsLogger{}, services.WhatsAppMetricsServiceOptions{})
		commandOrchestrator, orchErr := services.NewWhatsAppCommandOrchestrator(
			listingSvc,
			searchIntentSvc,
			userRepo,
			identitySvc,
			services.WhatsAppCommandOrchestratorOptions{
				MetricsSink: metricsSvc,
				WebBaseURL:  strings.TrimSpace(os.Getenv("WEB_BASE_URL")),
			},
		)
		if orchErr != nil {
			utils.LogError("init whatsapp command orchestrator (non-fatal)", orchErr)
		} else {
			var voiceSvc *services.WhatsAppVoiceService
			if aiSvc != nil {
				voiceSvc, _ = services.NewWhatsAppVoiceService(
					userRepo,
					&localMediaDownloader{
						twilioAccountSID: strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID")),
						twilioAuthToken:  strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN")),
					},
					aiSvc,
					commandOrchestrator,
					services.WhatsAppVoiceServiceOptions{MetricsSink: metricsSvc},
				)
			}

			whatsappHandler = handlers.NewWhatsAppHandler(handlers.WhatsAppHandlerDependencies{
				Provider:            waProvider,
				UserStore:           userRepo,
				CommandOrchestrator: commandOrchestrator,
				VoiceService:        voiceSvc,
				Policy:              policy,
				IdentityVerifier:    identitySvc,
				ConfirmSender:       waProvider,
				CommandReplySender:  waProvider,
				StatusSink:          &localDeliveryStatusSink{},
				MetaVerifyToken:     metaVerifyToken,
			})
		}
	}

	// ── HTTP router ──────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	mux.Handle("/auth/", wrap(authHandler.Handle))
	mux.Handle("/premium/", wrap(premiumHandler.Handle))

	if whatsappHandler != nil {
		mux.Handle("/whatsapp/", wrap(whatsappHandler.Handle))
	}

	// All remaining routes → combined listings/search/leads handler
	mux.Handle("/", wrap(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handlers.CombinedListingHandler(ctx, req, listingHandler, searchHandler, leadHandler)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Propti local dev server on http://localhost:%s (no Docker required)", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func wrap(h lambdaHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[strings.ToLower(k)] = v[0]
			}
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			headers["Authorization"] = auth
		}

		queryParams := make(map[string]string)
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				queryParams[k] = v[0]
			}
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod:            r.Method,
			Path:                  r.URL.Path,
			Headers:               headers,
			QueryStringParameters: queryParams,
			PathParameters:        extractPathParams(r.URL.Path),
			Body:                  string(body),
		}

		resp, err := h(r.Context(), req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"message":"internal server error","error":%q}`, err.Error())
			return
		}

		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, resp.Body)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func extractPathParams(urlPath string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")

	switch {
	case len(parts) >= 2 && parts[0] == "listings":
		nonID := map[string]bool{"parse-text": true, "upload-prepare": true, "": true}
		if !nonID[parts[1]] {
			params["id"] = parts[1]
		}
	case len(parts) >= 2 && parts[0] == "leads":
		if parts[1] != "" && parts[1] != "analytics" {
			params["id"] = parts[1]
		}
		if len(parts) >= 5 && parts[2] == "followups" {
			params["taskId"] = parts[3]
		}
	case len(parts) >= 4 && parts[0] == "users" && parts[1] == "me" && parts[2] == "listings":
		if parts[3] != "" {
			params["id"] = parts[3]
		}
	}
	return params
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
		return nil, "", fmt.Errorf("unsupported WHATSAPP_PROVIDER: %q", rawProvider)
	}
}

// Ensure json is used (satisfies import check).
var _ = json.Marshal
