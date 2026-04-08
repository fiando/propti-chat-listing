package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakeWhatsAppOrchestratorListingService struct {
	parseResp     *models.ParseTextResponse
	parseErr      error
	createResp    *models.Listing
	createErr     error
	searchResp    []models.Listing
	searchErr     error
	parseCalls    int
	createCalls   int
	searchCalls   int
	lastCreateReq *models.CreateListingRequest
	lastSearch    *models.ListingSearchParams
}

func (f *fakeWhatsAppOrchestratorListingService) ParseListingText(_ context.Context, text string) (*models.ParseTextResponse, error) {
	f.parseCalls++
	if f.parseErr != nil {
		return nil, f.parseErr
	}
	if f.parseResp != nil {
		resp := *f.parseResp
		return &resp, nil
	}
	return &models.ParseTextResponse{}, nil
}

func (f *fakeWhatsAppOrchestratorListingService) CreateListing(_ context.Context, _ string, req *models.CreateListingRequest) (*models.Listing, error) {
	f.createCalls++
	f.lastCreateReq = req
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createResp != nil {
		listing := *f.createResp
		return &listing, nil
	}
	return &models.Listing{ListingID: "listing-created"}, nil
}

func (f *fakeWhatsAppOrchestratorListingService) SearchListings(_ context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	f.searchCalls++
	f.lastSearch = params
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return append([]models.Listing(nil), f.searchResp...), nil
}

type fakeWhatsAppOrchestratorSearchIntentService struct {
	resp  *models.SearchIntentResponse
	err   error
	calls int
}

func (f *fakeWhatsAppOrchestratorSearchIntentService) ParseIntent(_ context.Context, _ string) (*models.SearchIntentResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if f.resp != nil {
		resp := *f.resp
		return &resp, nil
	}
	return &models.SearchIntentResponse{}, nil
}

type fakeWhatsAppOrchestratorUserStore struct {
	user *models.User
	err  error
}

func (f *fakeWhatsAppOrchestratorUserStore) GetByID(_ context.Context, _ string) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.user == nil {
		return nil, nil
	}
	user := *f.user
	return &user, nil
}

type fakeWhatsAppOrchestratorEligibilityGuard struct {
	err   error
	calls int
}

func (f *fakeWhatsAppOrchestratorEligibilityGuard) RequireWriteEligible(_ context.Context, _ string) error {
	f.calls++
	return f.err
}

func TestWhatsAppOrchestratorFreeTierAllowsCreate(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp:  &models.ParseTextResponse{Parsed: models.ParsedListing{Title: "Rumah", Description: "Bagus", Price: 1200000000, PriceUnit: "total", Address: "Jl. Malioboro"}},
		createResp: &models.Listing{ListingID: "listing-1"},
	}
	guard := &fakeWhatsAppOrchestratorEligibilityGuard{}
	orchestrator := mustNewWhatsAppOrchestrator(t, listingSvc, &fakeWhatsAppOrchestratorSearchIntentService{}, &fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionFree}}}, guard)

	created, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "jual rumah di jogja harga 1.2M"})
	if err != nil {
		t.Fatalf("HandleText create returned error: %v", err)
	}
	if created.Intent != WhatsAppCommandIntentListingCreate {
		t.Fatalf("expected create intent, got %q", created.Intent)
	}
	if listingSvc.createCalls != 1 {
		t.Fatalf("expected create to be called once, got %d", listingSvc.createCalls)
	}
	if guard.calls != 1 {
		t.Fatalf("expected write eligibility guard to be called, got %d", guard.calls)
	}
}

func TestWhatsAppOrchestratorFreeTierBlocksSearch(t *testing.T) {
	t.Parallel()

	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		&fakeWhatsAppOrchestratorListingService{},
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionFree}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	_, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "cari rumah jogja"})
	if err == nil {
		t.Fatal("expected free tier search to be blocked")
	}
	if appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected 403 for blocked search, got %d (%v)", appCodeFromErr(t, err), err)
	}
}

func TestWhatsAppOrchestratorSearchBuildsDeepLinkContinuation(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{searchResp: []models.Listing{{ListingID: "listing-1"}}}
	searchSvc := &fakeWhatsAppOrchestratorSearchIntentService{resp: &models.SearchIntentResponse{SearchParams: models.ListingSearchParams{ListingType: models.ListingTypeSell, City: "Yogyakarta"}}}
	renewDate := time.Now().UTC().Add(24 * time.Hour)
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		searchSvc,
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renewDate}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "cari rumah dijual di jogja"})
	if err != nil {
		t.Fatalf("HandleText search returned error: %v", err)
	}
	if resp.Intent != WhatsAppCommandIntentSearch {
		t.Fatalf("expected search intent, got %q", resp.Intent)
	}
	if searchSvc.calls != 1 || listingSvc.searchCalls != 1 {
		t.Fatalf("expected search services to be called once each, parse=%d search=%d", searchSvc.calls, listingSvc.searchCalls)
	}
	if !strings.Contains(resp.Message, "Ditemukan 1 listing") {
		t.Fatalf("expected search reply message to summarize results, got %q", resp.Message)
	}
	if !strings.Contains(resp.WebDeepLink, "https://propti.id/search") || !strings.Contains(resp.WebDeepLink, "q=cari+rumah+dijual+di+jogja") {
		t.Fatalf("expected deep-link continuation to include encoded query, got %q", resp.WebDeepLink)
	}
	if !strings.Contains(resp.Message, resp.WebDeepLink) {
		t.Fatalf("expected search reply message to include deep-link continuation, got %q", resp.Message)
	}
}

func TestWhatsAppOrchestratorWriteEligibilityGuardBlocksWriteCommands(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{}
	guard := &fakeWhatsAppOrchestratorEligibilityGuard{err: utils.NewAppError(403, "whatsapp identity is not verified")}
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionFree}}},
		guard,
	)

	_, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "jual rumah minimalis"})
	if err == nil {
		t.Fatal("expected guard error")
	}
	if appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected 403 from eligibility guard, got %d (%v)", appCodeFromErr(t, err), err)
	}
	if listingSvc.createCalls != 0 {
		t.Fatalf("expected create not to be called when guard blocks, got %d", listingSvc.createCalls)
	}
}

func TestWhatsAppOrchestratorExpiredPaidPlanFallsBackToFreeGate(t *testing.T) {
	t.Parallel()

	expired := time.Now().UTC().Add(-24 * time.Hour)
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		&fakeWhatsAppOrchestratorListingService{},
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: &expired}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	// Expired premium falls back to free tier, which cannot search via WhatsApp.
	_, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "cari rumah jogja"})
	if err == nil {
		t.Fatal("expected search to be blocked after premium expiry")
	}
	if appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected 403 for expired paid tier gate, got %d (%v)", appCodeFromErr(t, err), err)
	}
}

func mustNewWhatsAppOrchestrator(
	t *testing.T,
	listing WhatsAppOrchestratorListingService,
	search WhatsAppOrchestratorSearchIntentService,
	users WhatsAppOrchestratorUserStore,
	guard WhatsAppOrchestratorWriteEligibilityGuard,
) *WhatsAppCommandOrchestrator {
	t.Helper()
	orchestrator, err := NewWhatsAppCommandOrchestrator(listing, search, users, guard, WhatsAppCommandOrchestratorOptions{})
	if err != nil {
		t.Fatalf("NewWhatsAppCommandOrchestrator returned error: %v", err)
	}
	return orchestrator
}

func appCodeFromErr(t *testing.T, err error) int {
	t.Helper()
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T (%v)", err, err)
	}
	return appErr.Code
}

func TestWhatsAppOrchestratorCreateWithRequiresCorrectionSavesDraft(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp: &models.ParseTextResponse{
			Parsed:             models.ParsedListing{Title: "Rumah", Description: "Bagus", Price: 1200000000, PriceUnit: "total", Address: "Jl. Malioboro"},
			RequiresCorrection: true,
		},
		createResp: &models.Listing{ListingID: "draft-listing-1", ModerationStatus: models.ModerationStatusDraft},
	}
	renewDate := time.Now().UTC().Add(24 * time.Hour)
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renewDate}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "jual rumah"})
	if err != nil {
		t.Fatalf("HandleText returned error: %v", err)
	}
	if resp.Intent != WhatsAppCommandIntentListingCreate {
		t.Fatalf("expected create intent, got %q", resp.Intent)
	}
	if listingSvc.lastCreateReq == nil || !listingSvc.lastCreateReq.IsDraft {
		t.Fatal("expected IsDraft=true when RequiresCorrection=true")
	}
	if !strings.Contains(resp.Message, "draft") {
		t.Fatalf("expected draft message, got %q", resp.Message)
	}
}

func TestWhatsAppOrchestratorCreateWithoutRequiresCorrectionSubmitsPending(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp: &models.ParseTextResponse{
			Parsed:             models.ParsedListing{Title: "Rumah", Description: "Bagus", Price: 1200000000, PriceUnit: "total", Address: "Jl. Malioboro"},
			RequiresCorrection: false,
		},
		createResp: &models.Listing{ListingID: "listing-pending-1", ModerationStatus: models.ModerationStatusPending},
	}
	renewDate := time.Now().UTC().Add(24 * time.Hour)
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renewDate}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "jual rumah lengkap di bandung harga 2M 3KT 2KM LT120 LB90"})
	if err != nil {
		t.Fatalf("HandleText returned error: %v", err)
	}
	if listingSvc.lastCreateReq == nil || listingSvc.lastCreateReq.IsDraft {
		t.Fatal("expected IsDraft=false when RequiresCorrection=false")
	}
	if !strings.Contains(resp.Message, "berhasil") {
		t.Fatalf("expected success message, got %q", resp.Message)
	}
}
