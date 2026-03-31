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
	WhatsAppCommandIntentUnknown            WhatsAppCommandIntent = "unknown"
	WhatsAppCommandIntentListingCreate      WhatsAppCommandIntent = "listing_create"
	WhatsAppCommandIntentListingEdit        WhatsAppCommandIntent = "listing_edit"
	WhatsAppCommandIntentListingDelete      WhatsAppCommandIntent = "listing_delete"
	WhatsAppCommandIntentListingRead        WhatsAppCommandIntent = "listing_read"
	WhatsAppCommandIntentSearch             WhatsAppCommandIntent = "search"
	WhatsAppCommandIntentSubscriptionStatus WhatsAppCommandIntent = "subscription_status"
)

type WhatsAppCommandRequest struct {
	UserID string
	Text   string
}

type WhatsAppSubscriptionSummary struct {
	Tier              models.SubscriptionTier   `json:"tier"`
	Status            models.SubscriptionStatus `json:"status"`
	UsedListings      int                       `json:"usedListings"`
	RemainingListings int                       `json:"remainingListings"`
	LimitListings     int                       `json:"limitListings"`
	UpgradeGuidance   string                    `json:"upgradeGuidance,omitempty"`
}

type WhatsAppCommandResponse struct {
	Intent       WhatsAppCommandIntent        `json:"intent"`
	Message      string                       `json:"message,omitempty"`
	Listing      *models.Listing              `json:"listing,omitempty"`
	Listings     []models.Listing             `json:"listings,omitempty"`
	SearchIntent *models.SearchIntentResponse `json:"searchIntent,omitempty"`
	WebDeepLink  string                       `json:"webDeepLink,omitempty"`
	Subscription *WhatsAppSubscriptionSummary `json:"subscription,omitempty"`
}

type WhatsAppOrchestratorListingService interface {
	ParseListingText(ctx context.Context, text string) (*models.ParseTextResponse, error)
	CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error)
	UpdateListing(ctx context.Context, userID, listingID string, req *models.UpdateListingRequest) (*models.Listing, error)
	DeleteListing(ctx context.Context, userID, listingID string) error
	ListMyListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error)
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
	RecordUpgradeIntent(ctx context.Context, metric WhatsAppUpgradeIntentMetric) error
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
	user, tier, err := o.resolveUserAndTier(ctx, req.UserID)
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
	case WhatsAppCommandIntentListingEdit:
		if err := o.enforceTierGate(tier, intent); err != nil {
			return nil, err
		}
		if err := o.requireWriteEligible(ctx, req.UserID); err != nil {
			return nil, err
		}
		resp, err := o.handleEdit(ctx, req.UserID, payload)
		if err != nil {
			return nil, err
		}
		o.recordCompletionMetrics(ctx, req.UserID, tier, intent)
		return resp, nil
	case WhatsAppCommandIntentListingDelete:
		if err := o.enforceTierGate(tier, intent); err != nil {
			return nil, err
		}
		if err := o.requireWriteEligible(ctx, req.UserID); err != nil {
			return nil, err
		}
		resp, err := o.handleDelete(ctx, req.UserID, payload)
		if err != nil {
			return nil, err
		}
		o.recordCompletionMetrics(ctx, req.UserID, tier, intent)
		return resp, nil
	case WhatsAppCommandIntentListingRead:
		if err := o.enforceTierGate(tier, intent); err != nil {
			return nil, err
		}
		resp, err := o.handleList(ctx, req.UserID)
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
	case WhatsAppCommandIntentSubscriptionStatus:
		resp, err := o.handleSubscription(ctx, user, tier)
		if err != nil {
			return nil, err
		}
		if resp.Subscription != nil && strings.TrimSpace(resp.Subscription.UpgradeGuidance) != "" {
			o.recordUpgradeIntentMetric(ctx, req.UserID, tier, resp.Subscription.UpgradeGuidance)
		}
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

func (o *WhatsAppCommandOrchestrator) recordUpgradeIntentMetric(ctx context.Context, userID string, tier models.SubscriptionTier, hint string) {
	if o.metricsSink == nil {
		return
	}
	_ = o.metricsSink.RecordUpgradeIntent(ctx, WhatsAppUpgradeIntentMetric{
		UserID:      userID,
		Tier:        tier,
		UpgradeHint: hint,
		Source:      string(WhatsAppCommandIntentSubscriptionStatus),
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
	case WhatsAppCommandIntentListingRead, WhatsAppCommandIntentSearch:
		if entitlement.WhatsAppReadAllowed {
			return nil
		}
	case WhatsAppCommandIntentListingEdit:
		if entitlement.WhatsAppEditAllowed {
			return nil
		}
	case WhatsAppCommandIntentListingDelete:
		if entitlement.WhatsAppDeleteAllowed {
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

func (o *WhatsAppCommandOrchestrator) handleEdit(ctx context.Context, userID, payload string) (*WhatsAppCommandResponse, error) {
	parts := strings.SplitN(payload, "|", 2)
	listingID := strings.TrimSpace(parts[0])
	if listingID == "" {
		return nil, utils.NewAppError(400, "listing id is required for edit")
	}
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return nil, utils.NewAppError(400, "edit content is required")
	}

	parsedResp, err := o.listingService.ParseListingText(ctx, strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}
	parsed := parsedResp.Parsed

	updateReq := &models.UpdateListingRequest{}
	if parsed.Title != "" {
		title := parsed.Title
		updateReq.Title = &title
	}
	if parsed.Description != "" {
		description := parsed.Description
		updateReq.Description = &description
	}
	if parsed.Price > 0 {
		price := parsed.Price
		updateReq.Price = &price
	}
	if parsed.PriceUnit != "" {
		priceUnit := parsed.PriceUnit
		updateReq.PriceUnit = &priceUnit
	}
	if parsed.Address != "" || parsed.LocationSuggestion.City != "" || parsed.LocationSuggestion.Province != "" || parsed.LocationSuggestion.District != "" {
		location := models.Location{
			Address:  parsed.Address,
			Province: parsed.LocationSuggestion.Province,
			City:     parsed.LocationSuggestion.City,
			District: parsed.LocationSuggestion.District,
		}
		updateReq.Location = &location
	}

	listing, err := o.listingService.UpdateListing(ctx, userID, listingID, updateReq)
	if err != nil {
		return nil, err
	}
	link := o.buildListingLink(listing.ListingID)
	msg := fmt.Sprintf("Listing berhasil diperbarui. Lihat hasilnya di: %s", link)
	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingEdit, Listing: listing, Message: msg}, nil
}

func (o *WhatsAppCommandOrchestrator) handleDelete(ctx context.Context, userID, listingID string) (*WhatsAppCommandResponse, error) {
	listingID = strings.TrimSpace(listingID)
	if listingID == "" {
		return nil, utils.NewAppError(400, "listing id is required for delete")
	}
	if err := o.listingService.DeleteListing(ctx, userID, listingID); err != nil {
		return nil, err
	}
	listingsLink := o.webBaseURL + "/listings"
	msg := fmt.Sprintf("Listing berhasil dihapus. Kelola iklan Anda di: %s", listingsLink)
	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingDelete, Message: msg}, nil
}

func (o *WhatsAppCommandOrchestrator) handleList(ctx context.Context, userID string) (*WhatsAppCommandResponse, error) {
	listings, err := o.listingService.ListMyListings(ctx, userID, &models.ListingSearchParams{Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	listingsLink := o.webBaseURL + "/listings"
	msg := fmt.Sprintf("Kelola semua iklan Anda di: %s", listingsLink)
	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentListingRead, Listings: listings, Message: msg}, nil
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

func (o *WhatsAppCommandOrchestrator) handleSubscription(ctx context.Context, user *models.User, effectiveTier models.SubscriptionTier) (*WhatsAppCommandResponse, error) {
	listings, err := o.listingService.ListMyListings(ctx, user.UserID, &models.ListingSearchParams{Page: 1, PageSize: 100})
	if err != nil {
		return nil, err
	}

	entitlements := TierEntitlementFor(effectiveTier)
	limit := entitlements.ActiveListingCap
	used := len(listings)
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}

	summary := &WhatsAppSubscriptionSummary{
		Tier:              effectiveTier,
		Status:            DeriveSubscriptionStatus(user, o.now()),
		UsedListings:      used,
		RemainingListings: remaining,
		LimitListings:     limit,
		UpgradeGuidance:   o.upgradeGuidanceForTier(effectiveTier),
	}

	return &WhatsAppCommandResponse{Intent: WhatsAppCommandIntentSubscriptionStatus, Subscription: summary}, nil
}

func (o *WhatsAppCommandOrchestrator) upgradeGuidanceForTier(tier models.SubscriptionTier) string {
	switch tier {
	case models.SubscriptionFree:
		return "Upgrade ke Basic atau lebih untuk membaca dan mencari listing via WhatsApp."
	case models.SubscriptionBasic:
		return "Upgrade ke Premium atau Pro untuk mengedit dan menghapus listing via WhatsApp."
	case models.SubscriptionPremium:
		return "Upgrade ke Pro untuk batas listing lebih tinggi dan kuota voice lebih banyak."
	default:
		return "Paket Anda sudah mendukung semua fitur perintah WhatsApp."
	}
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
	case strings.HasPrefix(lower, "status paket"), strings.HasPrefix(lower, "cek paket"), strings.HasPrefix(lower, "info paket"),
		strings.HasPrefix(lower, "status subscription"), strings.HasPrefix(lower, "subscription status"):
		return WhatsAppCommandIntentSubscriptionStatus, ""
	case strings.HasPrefix(lower, "listing saya"), strings.HasPrefix(lower, "iklan saya"), strings.HasPrefix(lower, "daftar iklan"),
		strings.HasPrefix(lower, "list listing"), strings.HasPrefix(lower, "my listings"):
		return WhatsAppCommandIntentListingRead, ""
	case strings.HasPrefix(lower, "cari "), strings.HasPrefix(lower, "search "), strings.HasPrefix(lower, "find "):
		return WhatsAppCommandIntentSearch, trimmed
	case strings.HasPrefix(lower, "ubah "), strings.HasPrefix(lower, "edit "), strings.HasPrefix(lower, "perbarui "), strings.HasPrefix(lower, "update "):
		if idx := strings.IndexByte(trimmed, ' '); idx >= 0 {
			return WhatsAppCommandIntentListingEdit, strings.TrimSpace(trimmed[idx+1:])
		}
		return WhatsAppCommandIntentListingEdit, ""
	case strings.HasPrefix(lower, "hapus "), strings.HasPrefix(lower, "delete "):
		if idx := strings.IndexByte(trimmed, ' '); idx >= 0 {
			return WhatsAppCommandIntentListingDelete, strings.TrimSpace(trimmed[idx+1:])
		}
		return WhatsAppCommandIntentListingDelete, ""
	default:
		return WhatsAppCommandIntentListingCreate, trimmed
	}
}
