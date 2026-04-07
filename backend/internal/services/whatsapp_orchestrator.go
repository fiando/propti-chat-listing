package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type WhatsAppCommandIntent string

const (
	WhatsAppCommandIntentUnknown       WhatsAppCommandIntent = "unknown"
	WhatsAppCommandIntentListingCreate WhatsAppCommandIntent = "listing_create"
	WhatsAppCommandIntentSearch        WhatsAppCommandIntent = "search"
)

type WhatsAppCommandRequest struct {
	UserID string
	Text   string
}

type WhatsAppCommandResponse struct {
	Intent       WhatsAppCommandIntent        `json:"intent"`
	Message      string                       `json:"message,omitempty"`
	Listing      *models.Listing              `json:"listing,omitempty"`
	Listings     []models.Listing             `json:"listings,omitempty"`
	SearchIntent *models.SearchIntentResponse `json:"searchIntent,omitempty"`
	WebDeepLink  string                       `json:"webDeepLink,omitempty"`
}

type WhatsAppOrchestratorListingService interface {
	ParseListingText(ctx context.Context, text string) (*models.ParseTextResponse, error)
	CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error)
	SearchListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error)
}

type WhatsAppOrchestratorSearchIntentService interface {
	ParseIntent(ctx context.Context, query string) (*models.SearchIntentResponse, error)
}

type WhatsAppOrchestratorUserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
}

type WhatsAppOrchestratorWriteEligibilityGuard interface {
	RequireWriteEligible(ctx context.Context, userID string) error
}

type WhatsAppCommandOrchestratorOptions struct {
	Now          func() time.Time
	WebSearchURL string
	WebBaseURL   string
	MetricsSink  WhatsAppCommandMetricsSink
}

type WhatsAppCommandMetricsSink interface {
	RecordChatFirstCompletion(ctx context.Context, metric WhatsAppChatFirstCompletionMetric) error
	RecordZeroContextSwitchCompletion(ctx context.Context, metric WhatsAppZeroContextSwitchCompletionMetric) error
}

type WhatsAppCommandOrchestrator struct {
	listingService        WhatsAppOrchestratorListingService
	searchIntentService   WhatsAppOrchestratorSearchIntentService
	userStore             WhatsAppOrchestratorUserStore
	writeEligibilityGuard WhatsAppOrchestratorWriteEligibilityGuard
	now                   func() time.Time
	webSearchURL          string
	webBaseURL            string
	metricsSink           WhatsAppCommandMetricsSink
}

func NewWhatsAppCommandOrchestrator(
	listingService WhatsAppOrchestratorListingService,
	searchIntentService WhatsAppOrchestratorSearchIntentService,
	userStore WhatsAppOrchestratorUserStore,
	writeEligibilityGuard WhatsAppOrchestratorWriteEligibilityGuard,
	opts WhatsAppCommandOrchestratorOptions,
) (*WhatsAppCommandOrchestrator, error) {
	if listingService == nil {
		return nil, fmt.Errorf("listing service is required")
	}
	if searchIntentService == nil {
		return nil, fmt.Errorf("search intent service is required")
	}
	if userStore == nil {
		return nil, fmt.Errorf("user store is required")
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}
	webURL := strings.TrimSpace(opts.WebSearchURL)
	if webURL == "" {
		webURL = "https://propti.id/search"
	}
	webBase := strings.TrimRight(strings.TrimSpace(opts.WebBaseURL), "/")
	if webBase == "" {
		webBase = "https://propti.id"
	}

	return &WhatsAppCommandOrchestrator{
		listingService:        listingService,
		searchIntentService:   searchIntentService,
		userStore:             userStore,
		writeEligibilityGuard: writeEligibilityGuard,
		now:                   nowFn,
		webSearchURL:          webURL,
		webBaseURL:            webBase,
		metricsSink:           opts.MetricsSink,
	}, nil
}

func (o *WhatsAppCommandOrchestrator) HandleText(ctx context.Context, req WhatsAppCommandRequest) (*WhatsAppCommandResponse, error) {
	_, tier, err := o.resolveUserAndTier(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	intent, payload := detectWhatsAppIntent(req.Text)
	switch intent {
	case WhatsAppCommandIntentListingCreate:
		if err := o.enforceTierGate(tier, intent); err != nil {
			return nil, err
		}
		if err := o.requireWriteEligible(ctx, req.UserID); err != nil {
			return nil, err
		}
		resp, err := o.handleCreate(ctx, req.UserID, payload)
		if err != nil {
			return nil, err
		}
		o.recordCompletionMetrics(ctx, req.UserID, tier, intent)
		return resp, nil
	case WhatsAppCommandIntentSearch:
		if err := o.enforceTierGate(tier, intent); err != nil {
			return nil, err
		}
		resp, err := o.handleSearch(ctx, strings.TrimSpace(req.Text))
		if err != nil {
			return nil, err
		}
		o.recordCompletionMetrics(ctx, req.UserID, tier, intent)
		return resp, nil
	default:
		return nil, utils.NewAppError(400, "unknown whatsapp command")
	}
}

func (o *WhatsAppCommandOrchestrator) recordCompletionMetrics(ctx context.Context, userID string, tier models.SubscriptionTier, intent WhatsAppCommandIntent) {
	if o.metricsSink == nil {
		return
	}
	_ = o.metricsSink.RecordChatFirstCompletion(ctx, WhatsAppChatFirstCompletionMetric{
		UserID:            userID,
		Tier:              tier,
		Intent:            string(intent),
		Source:            "whatsapp_text",
		ZeroContextSwitch: true,
	})
	_ = o.metricsSink.RecordZeroContextSwitchCompletion(ctx, WhatsAppZeroContextSwitchCompletionMetric{
		UserID: userID,
		Tier:   tier,
		Intent: string(intent),
	})
}

func (o *WhatsAppCommandOrchestrator) resolveUserAndTier(ctx context.Context, userID string) (*models.User, models.SubscriptionTier, error) {
	user, err := o.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, "", utils.ErrInternal
	}
	if user == nil {
		return nil, "", utils.ErrUnauthorized
	}
	return user, effectiveTierForUser(user, o.now()), nil
}

func (o *WhatsAppCommandOrchestrator) requireWriteEligible(ctx context.Context, userID string) error {
	if o.writeEligibilityGuard == nil {
		return nil
	}
	return o.writeEligibilityGuard.RequireWriteEligible(ctx, userID)
}

func (o *WhatsAppCommandOrchestrator) enforceTierGate(tier models.SubscriptionTier, intent WhatsAppCommandIntent) error {
	entitlement := TierEntitlementFor(tier)
	switch intent {
	case WhatsAppCommandIntentListingCreate:
		if entitlement.WhatsAppCreateAllowed {
			return nil
		}
	case WhatsAppCommandIntentSearch:
		if entitlement.WhatsAppReadAllowed {
			return nil
		}
	default:
		return nil
	}
	return utils.NewAppError(403, fmt.Sprintf("%s tier does not allow %s via WhatsApp", tier, intent))
}

func (o *WhatsAppCommandOrchestrator) handleCreate(ctx context.Context, userID, text string) (*WhatsAppCommandResponse, error) {
	parsedResp, err := o.listingService.ParseListingText(ctx, text)
	if err != nil {
		return nil, err
	}

	parsed := parsedResp.Parsed
	createReq := &models.CreateListingRequest{
		Title:       parsed.Title,
		Description: parsed.Description,
		Price:       parsed.Price,
		PriceUnit:   parsed.PriceUnit,
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  parsed.Address,
			Province: parsed.LocationSuggestion.Province,
			City:     parsed.LocationSuggestion.City,
			District: parsed.LocationSuggestion.District,
		},
		PropertyDetails: parsed.PropertyDetails,
	}
	listing, err := o.listingService.CreateListing(ctx, userID, createReq)
	if err != nil {
		return nil, err
	}

	link := o.buildListingLink(listing.ListingID)
	msg := fmt.Sprintf("Listing berhasil dibuat. Tinjau dan lengkapi iklan Anda di: %s", link)
	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingCreate, Listing: listing, Message: msg}, nil
}

func (o *WhatsAppCommandOrchestrator) handleSearch(ctx context.Context, query string) (*WhatsAppCommandResponse, error) {
	intentResp, err := o.searchIntentService.ParseIntent(ctx, query)
	if err != nil {
		return nil, err
	}

	listings, err := o.listingService.SearchListings(ctx, &intentResp.SearchParams)
	if err != nil {
		return nil, err
	}

	return &WhatsAppCommandResponse{
		Intent:       WhatsAppCommandIntentSearch,
		Listings:     listings,
		SearchIntent: intentResp,
		WebDeepLink:  o.buildWebSearchDeepLink(query),
	}, nil
}

func (o *WhatsAppCommandOrchestrator) buildListingLink(listingID string) string {
	return o.webBaseURL + "/listings/" + listingID
}

func (o *WhatsAppCommandOrchestrator) buildWebSearchDeepLink(query string) string {
	base := strings.TrimRight(o.webSearchURL, "?")
	return base + "?q=" + url.QueryEscape(strings.TrimSpace(query))
}

func detectWhatsAppIntent(text string) (WhatsAppCommandIntent, string) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)

	switch {
	case strings.HasPrefix(lower, "cari "), strings.HasPrefix(lower, "search "), strings.HasPrefix(lower, "find "):
		return WhatsAppCommandIntentSearch, trimmed
	default:
		return WhatsAppCommandIntentListingCreate, trimmed
	}
}
