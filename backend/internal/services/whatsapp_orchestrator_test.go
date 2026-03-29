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
	parseResp      *models.ParseTextResponse
	parseErr       error
	createResp     *models.Listing
	createErr      error
	updateResp     *models.Listing
	updateErr      error
	deleteErr      error
	listResp       []models.Listing
	listErr        error
	searchResp     []models.Listing
	searchErr      error
	parseCalls     int
	createCalls    int
	updateCalls    int
	deleteCalls    int
	listCalls      int
	searchCalls    int
	lastCreateReq  *models.CreateListingRequest
	lastUpdateReq  *models.UpdateListingRequest
	lastDeleteID   string
	lastSearch     *models.ListingSearchParams
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

func (f *fakeWhatsAppOrchestratorListingService) UpdateListing(_ context.Context, _, _ string, req *models.UpdateListingRequest) (*models.Listing, error) {
	f.updateCalls++
	f.lastUpdateReq = req
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateResp != nil {
		listing := *f.updateResp
		return &listing, nil
	}
	return &models.Listing{ListingID: "listing-updated"}, nil
}

func (f *fakeWhatsAppOrchestratorListingService) DeleteListing(_ context.Context, _, listingID string) error {
	f.deleteCalls++
	f.lastDeleteID = listingID
	return f.deleteErr
}

func (f *fakeWhatsAppOrchestratorListingService) ListMyListings(_ context.Context, _ string, _ *models.ListingSearchParams) ([]models.Listing, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]models.Listing(nil), f.listResp...), nil
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

func TestWhatsAppOrchestratorFreeTierAllowsCreateOnly(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp: &models.ParseTextResponse{Parsed: models.ParsedListing{Title: "Rumah", Description: "Bagus", Price: 1200000000, PriceUnit: "total", Address: "Jl. Malioboro"}},
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

	_, err = orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "listing saya"})
	if err == nil {
		t.Fatal("expected free tier read to be blocked")
	}
	if appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected 403 for blocked read, got %d (%v)", appCodeFromErr(t, err), err)
	}
}

func TestWhatsAppOrchestratorBasicTierBlocksEditAndDelete(t *testing.T) {
	t.Parallel()

	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		&fakeWhatsAppOrchestratorListingService{},
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	_, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "edit listing-1 | ubah deskripsi"})
	if err == nil || appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected edit blocked for basic tier with 403, got %v", err)
	}

	_, err = orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "delete listing-1"})
	if err == nil || appCodeFromErr(t, err) != 403 {
		t.Fatalf("expected delete blocked for basic tier with 403, got %v", err)
	}
}

func TestWhatsAppOrchestratorPremiumAllowsDeleteAndRoutesToListingService(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{}
	guard := &fakeWhatsAppOrchestratorEligibilityGuard{}
	renewDate := time.Now().UTC().Add(24 * time.Hour)
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: &renewDate}}},
		guard,
	)

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "hapus listing-9"})
	if err != nil {
		t.Fatalf("HandleText delete returned error: %v", err)
	}
	if resp.Intent != WhatsAppCommandIntentListingDelete {
		t.Fatalf("expected delete intent, got %q", resp.Intent)
	}
	if listingSvc.deleteCalls != 1 || listingSvc.lastDeleteID != "listing-9" {
		t.Fatalf("expected delete listing-9 to be routed, got calls=%d id=%q", listingSvc.deleteCalls, listingSvc.lastDeleteID)
	}
	if guard.calls != 1 {
		t.Fatalf("expected write guard to run for delete, got %d", guard.calls)
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
	if !strings.Contains(resp.WebDeepLink, "https://propti.id/search") || !strings.Contains(resp.WebDeepLink, "q=cari+rumah+dijual+di+jogja") {
		t.Fatalf("expected deep-link continuation to include encoded query, got %q", resp.WebDeepLink)
	}
}

func TestWhatsAppOrchestratorSubscriptionIncludesRemainingAndUpgradeGuidance(t *testing.T) {
	t.Parallel()

	renewDate := time.Now().UTC().Add(5 * 24 * time.Hour)
	listingSvc := &fakeWhatsAppOrchestratorListingService{listResp: []models.Listing{{ListingID: "a"}, {ListingID: "b"}}}
	orchestrator := mustNewWhatsAppOrchestrator(
		t,
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionBasic, RenewDate: &renewDate}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
	)

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "status subscription"})
	if err != nil {
		t.Fatalf("HandleText subscription returned error: %v", err)
	}
	if resp.Intent != WhatsAppCommandIntentSubscriptionStatus {
		t.Fatalf("expected subscription intent, got %q", resp.Intent)
	}
	if resp.Subscription == nil {
		t.Fatal("expected subscription payload")
	}
	if resp.Subscription.RemainingListings != 4 {
		t.Fatalf("expected remaining listing cap 4 for basic with 2 used, got %d", resp.Subscription.RemainingListings)
	}
	if resp.Subscription.Status != models.SubscriptionExpiringSoon {
		t.Fatalf("expected expiring soon status, got %q", resp.Subscription.Status)
	}
	if resp.Subscription.UpgradeGuidance == "" {
		t.Fatal("expected upgrade guidance to be populated")
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

	_, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "cari rumah jogja"})
	if err == nil {
		t.Fatal("expected search read to be blocked after premium expiry")
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

func TestWhatsAppOrchestratorCreatePopulatesListingDeepLink(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp:  &models.ParseTextResponse{Parsed: models.ParsedListing{Title: "Rumah Bagus", Price: 500000000, PriceUnit: "total"}},
		createResp: &models.Listing{ListingID: "listing-abc"},
	}
	orchestrator, err := NewWhatsAppCommandOrchestrator(
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionFree}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
		WhatsAppCommandOrchestratorOptions{WebBaseURL: "https://propti.id"},
	)
	if err != nil {
		t.Fatalf("NewWhatsAppCommandOrchestrator error: %v", err)
	}

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "jual rumah bagus"})
	if err != nil {
		t.Fatalf("HandleText error: %v", err)
	}
	if resp.WebDeepLink != "https://propti.id/listings/listing-abc" {
		t.Fatalf("expected listing deep link https://propti.id/listings/listing-abc, got %q", resp.WebDeepLink)
	}
}

func TestWhatsAppOrchestratorEditPopulatesListingDeepLink(t *testing.T) {
	t.Parallel()

	listingSvc := &fakeWhatsAppOrchestratorListingService{
		parseResp:  &models.ParseTextResponse{Parsed: models.ParsedListing{Title: "Rumah Diperbarui"}},
		updateResp: &models.Listing{ListingID: "listing-xyz"},
	}
	renewDate := time.Now().UTC().Add(24 * time.Hour)
	orchestrator, err := NewWhatsAppCommandOrchestrator(
		listingSvc,
		&fakeWhatsAppOrchestratorSearchIntentService{},
		&fakeWhatsAppOrchestratorUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: &renewDate}}},
		&fakeWhatsAppOrchestratorEligibilityGuard{},
		WhatsAppCommandOrchestratorOptions{WebBaseURL: "https://propti.id"},
	)
	if err != nil {
		t.Fatalf("NewWhatsAppCommandOrchestrator error: %v", err)
	}

	resp, err := orchestrator.HandleText(context.Background(), WhatsAppCommandRequest{UserID: "user-1", Text: "edit listing-xyz | ubah judul"})
	if err != nil {
		t.Fatalf("HandleText edit error: %v", err)
	}
	if resp.WebDeepLink != "https://propti.id/listings/listing-xyz" {
		t.Fatalf("expected listing deep link https://propti.id/listings/listing-xyz, got %q", resp.WebDeepLink)
	}
}
