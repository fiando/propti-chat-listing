package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

// --- fakes ---

type fakeListingStore struct {
	putCalls            int
	lastListing         *models.Listing
	listingsByID        map[string]*models.Listing
	listingsByUser      map[string]map[string]*models.Listing
	listingsByUserIndex map[string][]models.Listing
	scanListings        []models.Listing
	now                 func() time.Time
}

func (f *fakeListingStore) Put(ctx context.Context, listing *models.Listing) error {
	f.putCalls++
	if listing == nil {
		return nil
	}
	copy := *listing
	f.lastListing = &copy
	if f.listingsByID == nil {
		f.listingsByID = make(map[string]*models.Listing)
	}
	f.listingsByID[listing.ListingID] = &copy
	if f.listingsByUser == nil {
		f.listingsByUser = make(map[string]map[string]*models.Listing)
	}
	if f.listingsByUser[listing.UserID] == nil {
		f.listingsByUser[listing.UserID] = make(map[string]*models.Listing)
	}
	f.listingsByUser[listing.UserID][listing.ListingID] = &copy
	return nil
}
func (f *fakeListingStore) GetByID(ctx context.Context, userID, listingID string) (*models.Listing, error) {
	if f.listingsByUser == nil || f.listingsByUser[userID] == nil {
		return nil, nil
	}
	listing := f.listingsByUser[userID][listingID]
	if listing == nil {
		return nil, nil
	}
	copy := *listing
	return &copy, nil
}
func (f *fakeListingStore) GetByListingID(ctx context.Context, listingID string) (*models.Listing, error) {
	if f.listingsByID == nil {
		return nil, nil
	}
	listing := f.listingsByID[listingID]
	if listing == nil {
		return nil, nil
	}
	copy := *listing
	return &copy, nil
	return nil, nil
}
func (f *fakeListingStore) ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Listing, error) {
	if f.listingsByUserIndex != nil {
		listings := f.listingsByUserIndex[userID]
		result := make([]models.Listing, len(listings))
		copy(result, listings)
		return result, nil
	}
	if f.listingsByUser == nil || f.listingsByUser[userID] == nil {
		return nil, nil
	}
	result := make([]models.Listing, 0, len(f.listingsByUser[userID]))
	for _, listing := range f.listingsByUser[userID] {
		copy := *listing
		result = append(result, copy)
	}
	return result, nil
}
func (f *fakeListingStore) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	if f.listingsByUser == nil || f.listingsByUser[userID] == nil {
		return 0, nil
	}

	now := time.Now().UTC()
	if f.now != nil {
		now = f.now()
	}
	count := 0
	for _, listing := range f.listingsByUser[userID] {
		if listing.Status != models.ListingStatusActive {
			continue
		}
		if listing.ExpiresAt != nil && !listing.ExpiresAt.After(now) {
			continue
		}
		count++
	}

	return count, nil
}
func (f *fakeListingStore) Delete(ctx context.Context, userID, listingID string) error { return nil }
func (f *fakeListingStore) Scan(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	if f.scanListings == nil {
		return nil, nil
	}
	result := make([]models.Listing, len(f.scanListings))
	copy(result, f.scanListings)
	return result, nil
}

type fakeUserStore struct {
	user      *models.User
	usersByID map[string]*models.User
}

func (f *fakeUserStore) GetByID(ctx context.Context, userID string) (*models.User, error) {
	if f.usersByID != nil {
		user := f.usersByID[userID]
		if user == nil {
			return nil, nil
		}
		copy := *user
		return &copy, nil
	}
	return f.user, nil
}

func (f *fakeUserStore) Put(ctx context.Context, user *models.User) error {
	if user == nil {
		f.user = nil
		return nil
	}
	copy := *user
	f.user = &copy
	if f.usersByID != nil {
		f.usersByID[user.UserID] = &copy
	}
	return nil
}

type fakeAIService struct {
	parseResult     *models.ParsedListing
	err             error
	moderateCalls   int
	moderateOK      bool
	moderateReason  string
	moderateFlags   []string
	moderationError error
}

func (f *fakeAIService) ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error) {
	return f.parseResult, f.err
}

func (f *fakeAIService) ModerateContent(ctx context.Context, title, description string) (bool, string, []string, error) {
	f.moderateCalls++
	return f.moderateOK, f.moderateReason, f.moderateFlags, f.moderationError
}

type fakeModerationEnqueuer struct {
	listingIDs []string
	err        error
}

func (f *fakeModerationEnqueuer) EnqueueListingModeration(ctx context.Context, listingID string, checkText bool, newImageIDs []string) error {
	if f.err != nil {
		return f.err
	}
	f.listingIDs = append(f.listingIDs, listingID)
	return nil
}

type countingRoundTripper struct {
	calls int
}

func (c *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("unexpected outbound maps request")
}

// --- tests ---

func TestCreateListingDoesNotGeocodeDuringSubmit(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{}
	transport := &countingRoundTripper{}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		nil,
		&GoogleMapsService{
			client: &http.Client{Transport: transport},
		},
		nil,
	)

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah minimalis dekat tol Depok",
		Description: "Rumah siap huni dengan akses cepat ke tol dan stasiun.",
		Price:       850000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  "Jl. Margonda Raya No. 1",
			Province: "Jawa Barat",
			City:     "Depok",
			District: "Beji",
		},
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if transport.calls != 0 {
		t.Fatalf("expected create listing to avoid outbound geocoding, got %d call(s)", transport.calls)
	}
	if store.putCalls != 1 {
		t.Fatalf("expected listing to be persisted once, got %d", store.putCalls)
	}
	if store.lastListing == nil {
		t.Fatalf("expected persisted listing to be captured")
	}
	if store.lastListing.Location.City != "Depok" {
		t.Fatalf("expected location city to be preserved, got %q", store.lastListing.Location.City)
	}
	if listing.Location.Latitude != 0 || listing.Location.Longitude != 0 {
		t.Fatalf("expected submit path to leave coordinates untouched, got lat=%v lng=%v", listing.Location.Latitude, listing.Location.Longitude)
	}
}

func TestCreateListingFreeTierIgnoresArchivedListingsInQuota(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		now: func() time.Time { return now },
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
				"listing-2": {
					ListingID:        "listing-2",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusPending,
				},
				"listing-3": {
					ListingID:        "listing-3",
					UserID:           "user-1",
					Status:           models.ListingStatusArchived,
					ModerationStatus: models.ModerationStatusRejected,
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah minimalis dekat tol Depok",
		Description: "Rumah siap huni dengan akses cepat ke tol dan stasiun.",
		Price:       850000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  "Jl. Margonda Raya No. 1",
			Province: "Jawa Barat",
			City:     "Depok",
			District: "Beji",
		},
	})
	if err != nil {
		t.Fatalf("expected archived listing to not consume free tier slot, got error: %v", err)
	}
}

func TestCreateListingPremiumTierRejectsWhenActiveQuotaIsReached(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	activeListings := make(map[string]*models.Listing, 15)
	for i := 0; i < 15; i++ {
		listingID := fmt.Sprintf("listing-%d", i+1)
		expiresAt := now.Add(24 * time.Hour)
		activeListings[listingID] = &models.Listing{
			ListingID:        listingID,
			UserID:           "user-1",
			Status:           models.ListingStatusActive,
			ModerationStatus: models.ModerationStatusApproved,
			ExpiresAt:        &expiresAt,
		}
	}

	service := NewListingService(
		&fakeListingStore{
			listingsByUser: map[string]map[string]*models.Listing{
				"user-1": activeListings,
			},
		},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: func() *time.Time { t := now.Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah premium",
		Description: "Siap tayang",
		Price:       1500000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Premium No. 15",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "premium tier allows at most 15 active listing") {
		t.Fatalf("expected premium quota error, got %v", err)
	}
}

func TestCreateListingPremiumTierSetsNinetyDayExpiry(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{}
	fixedNow := time.Date(2026, time.March, 19, 7, 30, 0, 0, time.UTC)

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: func() *time.Time { t := fixedNow.Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	service.now = func() time.Time { return fixedNow }

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah premium",
		Description: "Siap tayang",
		Price:       1500000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Premium No. 15",
		},
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if listing.ExpiresAt == nil {
		t.Fatal("expected premium listing expiry to be set")
	}
	want := fixedNow.AddDate(0, 0, 90)
	if !listing.ExpiresAt.Equal(want) {
		t.Fatalf("expected premium expiry %s, got %s", want.Format(time.RFC3339), listing.ExpiresAt.Format(time.RFC3339))
	}
	if !listing.PremiumFeatures.IsPremium {
		t.Fatal("expected premium listing flag to be set")
	}
}

func TestGetListingReturnsNotFoundWhenIndexEntryIsStale(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
		},
		&fakeUserStore{},
		nil,
		nil,
		nil,
		nil,
	)

	listing, err := service.GetListing(ctx, "listing-1")
	if err != utils.ErrNotFound {
		t.Fatalf("expected not found for stale index entry, got listing=%v err=%v", listing, err)
	}
}

func TestGetListingDoesNotIncrementViewsOnRead(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah Depok",
					Views:            41,
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
			listingsByUser: map[string]map[string]*models.Listing{
				"user-1": {
					"listing-1": {
						ListingID:        "listing-1",
						UserID:           "user-1",
						Title:            "Rumah Depok",
						Views:            41,
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{},
		nil,
		nil,
		nil,
		nil,
	)

	listing, err := service.GetListing(ctx, "listing-1")
	if err != nil {
		t.Fatalf("GetListing returned error: %v", err)
	}
	if listing.Views != 41 {
		t.Fatalf("expected read path to leave views unchanged, got %d", listing.Views)
	}
}

func TestGetListingReturnsNotFoundWhenListingHasExpired(t *testing.T) {
	ctx := context.Background()
	expiredAt := time.Now().UTC().Add(-time.Hour)
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah Depok",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				ExpiresAt:        &expiredAt,
				PremiumFeatures: models.PremiumFeatures{
					IsFeatured:     true,
					FeaturedUntil:  &expiredAt,
					PromotionUntil: &expiredAt,
				},
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &expiredAt,
					PremiumFeatures: models.PremiumFeatures{
						IsFeatured:     true,
						FeaturedUntil:  &expiredAt,
						PromotionUntil: &expiredAt,
					},
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{},
		nil,
		nil,
		nil,
		nil,
	)

	listing, err := service.GetListing(ctx, "listing-1")
	if err != utils.ErrNotFound {
		t.Fatalf("expected not found for expired listing, got listing=%v err=%v", listing, err)
	}
	if store.lastListing == nil || store.lastListing.Status != models.ListingStatusArchived {
		t.Fatalf("expected expired listing to be archived, got %#v", store.lastListing)
	}
	if store.lastListing.PremiumFeatures.IsFeatured {
		t.Fatal("expected expired featured flag to be cleared")
	}
	if store.lastListing.PremiumFeatures.FeaturedUntil != nil {
		t.Fatal("expected expired featured timestamp to be cleared")
	}
	if store.lastListing.PremiumFeatures.PromotionUntil != nil {
		t.Fatal("expected expired promotion timestamp to be cleared")
	}
}

func TestListListingsArchivesExpiredListingsAndClearsExpiredPremiumFlags(t *testing.T) {
	ctx := context.Background()
	expiredAt := time.Now().UTC().Add(-time.Hour)
	store := &fakeListingStore{
		scanListings: []models.Listing{
			{
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah Depok",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				ExpiresAt:        &expiredAt,
				PremiumFeatures: models.PremiumFeatures{
					IsPremium:      true,
					IsFeatured:     true,
					FeaturedUntil:  &expiredAt,
					PromotionUntil: &expiredAt,
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{},
		nil,
		nil,
		nil,
		nil,
	)

	listings, err := service.ListListings(ctx, &models.ListingSearchParams{})
	if err != nil {
		t.Fatalf("ListListings returned error: %v", err)
	}
	if len(listings) != 0 {
		t.Fatalf("expected expired listing to be removed from public results, got %#v", listings)
	}
	if store.lastListing == nil || store.lastListing.Status != models.ListingStatusArchived {
		t.Fatalf("expected expired listing to be persisted as archived, got %#v", store.lastListing)
	}
	if store.lastListing.PremiumFeatures.IsFeatured {
		t.Fatal("expected expired featured flag to be cleared when archiving")
	}
	if store.lastListing.PremiumFeatures.FeaturedUntil != nil {
		t.Fatal("expected expired featured timestamp to be cleared when archiving")
	}
	if store.lastListing.PremiumFeatures.PromotionUntil != nil {
		t.Fatal("expected expired promotion timestamp to be cleared when archiving")
	}
}

func TestRecordListingViewIncrementsViewsForApprovedActiveListings(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah Depok",
				Views:            8,
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				UpdatedAt:        time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah Depok",
					Views:            8,
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					UpdatedAt:        time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{},
		nil,
		nil,
		nil,
		nil,
	)

	listing, err := service.RecordListingView(ctx, "listing-1")
	if err != nil {
		t.Fatalf("RecordListingView returned error: %v", err)
	}
	if listing.Views != 9 {
		t.Fatalf("expected incremented view count, got %d", listing.Views)
	}
	if store.lastListing == nil || store.lastListing.Views != 9 {
		t.Fatalf("expected incremented listing to be persisted, got %#v", store.lastListing)
	}
}

func TestGetListingIncludesSellerNameWithoutSellerPhone(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
			listingsByUser: map[string]map[string]*models.Listing{
				"user-1": {
					"listing-1": {
						ListingID:        "listing-1",
						UserID:           "user-1",
						Title:            "Rumah Depok",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Name:   "Budi Hartono",
				Phone:  "081234567890",
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	listing, err := service.GetListing(ctx, "listing-1")
	if err != nil {
		t.Fatalf("GetListing returned error: %v", err)
	}
	if listing.SellerName != "Budi Hartono" {
		t.Fatalf("expected seller name to be exposed, got %q", listing.SellerName)
	}
	if listing.SellerPhone != "" {
		t.Fatalf("expected seller phone to stay hidden on public detail, got %q", listing.SellerPhone)
	}
}

func TestRevealListingContactReturnsSellerPhoneWhenAvailable(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "seller-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
			listingsByUser: map[string]map[string]*models.Listing{
				"seller-1": {
					"listing-1": {
						ListingID:        "listing-1",
						UserID:           "seller-1",
						Title:            "Rumah Depok",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{
			user: &models.User{
				UserID: "seller-1",
				Name:   "Budi Hartono",
				Phone:  "081234567890",
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	contact, err := service.RevealListingContact(ctx, "viewer-1", "listing-1", models.ContactRevealChannelWhatsApp)
	if err != nil {
		t.Fatalf("RevealListingContact returned error: %v", err)
	}
	if contact.SellerName != "Budi Hartono" {
		t.Fatalf("expected seller name in reveal payload, got %q", contact.SellerName)
	}
	if contact.SellerPhone != "081234567890" {
		t.Fatalf("expected seller phone in reveal payload, got %q", contact.SellerPhone)
	}
	if contact.Channel != models.ContactRevealChannelWhatsApp {
		t.Fatalf("expected contact reveal channel to round-trip, got %q", contact.Channel)
	}
}

func TestRevealListingContactDoesNotConsumeQuotaWhenSellerPhoneUnavailable(t *testing.T) {
	ctx := context.Background()

	userStore := &fakeUserStore{
		usersByID: map[string]*models.User{
			"viewer-1": {
				UserID: "viewer-1",
				Name:   "Pembeli Aktif",
				ContactRevealThrottle: models.ContactRevealThrottle{
					WindowStartedAt: time.Now().UTC(),
					RevealCount:     2,
				},
			},
			"seller-1": {
				UserID: "seller-1",
				Name:   "Budi Hartono",
				Phone:  "   ",
			},
		},
	}

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "seller-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
			listingsByUser: map[string]map[string]*models.Listing{
				"seller-1": {
					"listing-1": {
						ListingID:        "listing-1",
						UserID:           "seller-1",
						Title:            "Rumah Depok",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		userStore,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.RevealListingContact(ctx, "viewer-1", "listing-1", models.ContactRevealChannelWhatsApp)
	if err == nil {
		t.Fatal("expected reveal to fail when seller phone is unavailable")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing seller contact, got %d", appErr.Code)
	}

	if got := userStore.usersByID["viewer-1"].ContactRevealThrottle.RevealCount; got != 2 {
		t.Fatalf("expected failed reveal to keep quota usage unchanged, got %d", got)
	}
}

func TestRevealListingContactRateLimitsExcessiveReveals(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByID: map[string]*models.Listing{
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "seller-1",
					Title:            "Rumah Depok",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
			listingsByUser: map[string]map[string]*models.Listing{
				"seller-1": {
					"listing-1": {
						ListingID:        "listing-1",
						UserID:           "seller-1",
						Title:            "Rumah Depok",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{
			usersByID: map[string]*models.User{
				"viewer-1": {
					UserID: "viewer-1",
					Name:   "Pembeli Aktif",
					ContactRevealThrottle: models.ContactRevealThrottle{
						WindowStartedAt: time.Now().UTC(),
						RevealCount:     contactRevealLimit,
					},
				},
				"seller-1": {
					UserID: "seller-1",
					Name:   "Budi Hartono",
					Phone:  "081234567890",
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.RevealListingContact(ctx, "viewer-1", "listing-1", models.ContactRevealChannelWhatsApp)
	if err == nil {
		t.Fatal("expected rate-limited contact reveal to fail")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 rate limit, got %d", appErr.Code)
	}
	if !strings.Contains(appErr.Message, "Terlalu banyak membuka kontak penjual") {
		t.Fatalf("unexpected rate limit message: %q", appErr.Message)
	}
}

func TestListMyListingsSkipsDeletedListingsWhenUserIndexIsStale(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByUser: map[string]map[string]*models.Listing{
				"user-1": {
					"listing-live": {
						ListingID:        "listing-live",
						UserID:           "user-1",
						Title:            "Rumah Aktif",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
			listingsByUserIndex: map[string][]models.Listing{
				"user-1": {
					{
						ListingID:        "listing-deleted",
						UserID:           "user-1",
						Title:            "Rumah Terhapus",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
					{
						ListingID:        "listing-live",
						UserID:           "user-1",
						Title:            "Rumah Aktif",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{
			user: &models.User{UserID: "user-1"},
		},
		nil,
		nil,
		nil,
		nil,
	)

	listings, err := service.ListMyListings(ctx, "user-1", &models.ListingSearchParams{})
	if err != nil {
		t.Fatalf("ListMyListings returned error: %v", err)
	}
	if len(listings) != 1 {
		t.Fatalf("expected only live listing after filtering stale index results, got %d", len(listings))
	}
	if listings[0].ListingID != "listing-live" {
		t.Fatalf("expected listing-live, got %q", listings[0].ListingID)
	}
}

func TestListMyListingsIncludesRejectedListingsForOwner(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{
			listingsByUser: map[string]map[string]*models.Listing{
				"user-1": {
					"listing-approved": {
						ListingID:        "listing-approved",
						UserID:           "user-1",
						Title:            "Rumah Aman",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
					"listing-rejected": {
						ListingID:        "listing-rejected",
						UserID:           "user-1",
						Title:            "Rumah Ditolak",
						Status:           models.ListingStatusArchived,
						ModerationStatus: models.ModerationStatusRejected,
					},
				},
			},
			listingsByUserIndex: map[string][]models.Listing{
				"user-1": {
					{
						ListingID:        "listing-rejected",
						UserID:           "user-1",
						Title:            "Rumah Ditolak",
						Status:           models.ListingStatusArchived,
						ModerationStatus: models.ModerationStatusRejected,
					},
					{
						ListingID:        "listing-approved",
						UserID:           "user-1",
						Title:            "Rumah Aman",
						Status:           models.ListingStatusActive,
						ModerationStatus: models.ModerationStatusApproved,
					},
				},
			},
		},
		&fakeUserStore{
			user: &models.User{UserID: "user-1"},
		},
		nil,
		nil,
		nil,
		nil,
	)

	listings, err := service.ListMyListings(ctx, "user-1", &models.ListingSearchParams{})
	if err != nil {
		t.Fatalf("ListMyListings returned error: %v", err)
	}
	if len(listings) != 2 {
		t.Fatalf("expected approved and rejected owner listings to remain visible, got %d", len(listings))
	}
	if listings[0].ListingID != "listing-rejected" {
		t.Fatalf("expected newest rejected listing first, got %q", listings[0].ListingID)
	}
	if listings[1].ListingID != "listing-approved" {
		t.Fatalf("expected approved listing second, got %q", listings[1].ListingID)
	}
}

func TestCreateListingDefersModerationWork(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{}
	ai := &fakeAIService{
		moderateOK:     false,
		moderateReason: "offensive language",
		moderateFlags:  []string{"hate"},
	}
	queue := &fakeModerationEnqueuer{}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		ai,
		nil,
		nil,
		nil,
	)
	service.SetModerationEnqueuer(queue)

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Offensive title",
		Description: "contains offensive language",
		Price:       850000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  "Jl. Margonda Raya No. 1",
			Province: "Jawa Barat",
			City:     "Depok",
			District: "Beji",
		},
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if ai.moderateCalls != 0 {
		t.Fatalf("expected moderation to be deferred, got %d inline call(s)", ai.moderateCalls)
	}
	if len(queue.listingIDs) != 1 || queue.listingIDs[0] != listing.ListingID {
		t.Fatalf("expected moderation queue to receive listing %q, got %#v", listing.ListingID, queue.listingIDs)
	}
	if listing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected returned listing moderation status pending before async moderation, got %q", listing.ModerationStatus)
	}
	if listing.Status != models.ListingStatusActive {
		t.Fatalf("expected returned listing status to remain active before async moderation, got %q", listing.Status)
	}
	if store.lastListing == nil {
		t.Fatalf("expected persisted listing to be captured")
	}
	if store.lastListing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected persisted listing moderation status pending before async moderation, got %q", store.lastListing.ModerationStatus)
	}
	if store.lastListing.Status != models.ListingStatusActive {
		t.Fatalf("expected persisted listing status to remain active before async moderation, got %q", store.lastListing.Status)
	}
	if store.lastListing.ModerationReason != "" {
		t.Fatalf("expected moderation reason to stay empty until worker runs, got %q", store.lastListing.ModerationReason)
	}
}

func TestCreateListingReturnsImmediatelyWithoutInlineModeration(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{}
	ai := &fakeAIService{
		moderationError: errors.New("openai unavailable"),
	}
	queue := &fakeModerationEnqueuer{}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		ai,
		nil,
		nil,
		nil,
	)
	service.SetModerationEnqueuer(queue)

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah minimalis dekat tol Depok",
		Description: "Rumah siap huni dengan akses cepat ke tol dan stasiun.",
		Price:       850000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  "Jl. Margonda Raya No. 1",
			Province: "Jawa Barat",
			City:     "Depok",
			District: "Beji",
		},
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if ai.moderateCalls != 0 {
		t.Fatalf("expected no inline moderation calls, got %d", ai.moderateCalls)
	}
	if len(queue.listingIDs) != 1 || queue.listingIDs[0] != listing.ListingID {
		t.Fatalf("expected moderation queue to receive listing %q, got %#v", listing.ListingID, queue.listingIDs)
	}
	if listing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected listing to be pending until async moderation runs, got %q", listing.ModerationStatus)
	}
	if store.lastListing == nil {
		t.Fatalf("expected persisted listing to be captured")
	}
	if store.lastListing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected persisted listing to be pending until async moderation runs, got %q", store.lastListing.ModerationStatus)
	}
}

func TestCreateListingNormalizesEmptyImagesBeforePersistence(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{}
	queue := &fakeModerationEnqueuer{}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		&fakeAIService{},
		nil,
		nil,
		nil,
	)
	service.SetModerationEnqueuer(queue)

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah tanpa foto",
		Description: "Listing baru tanpa upload foto.",
		Price:       850000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address:  "Jl. Margonda Raya No. 1",
			Province: "Jawa Barat",
			City:     "Depok",
			District: "Beji",
		},
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if listing.Images == nil {
		t.Fatalf("expected returned listing images to normalize to an empty slice, got nil")
	}
	if len(listing.Images) != 0 {
		t.Fatalf("expected returned listing images to be empty, got %d", len(listing.Images))
	}
	if store.lastListing == nil {
		t.Fatalf("expected persisted listing to be captured")
	}
	if store.lastListing.Images == nil {
		t.Fatalf("expected persisted listing images to normalize to an empty slice, got nil")
	}
	if len(store.lastListing.Images) != 0 {
		t.Fatalf("expected persisted listing images to be empty, got %d", len(store.lastListing.Images))
	}
}

func TestUpdateListingRequeuesModerationWithoutInlineExecution(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah lama",
				Description:      "deskripsi lama",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				ModerationReason: "",
				Location:         models.Location{City: "Depok"},
				PropertyDetails:  models.PropertyDetails{},
				PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah lama",
					Description:      "deskripsi lama",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					Location:         models.Location{City: "Depok"},
					PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
				},
			},
		},
	}
	ai := &fakeAIService{
		moderateOK:     false,
		moderateReason: "profanity detected",
	}
	queue := &fakeModerationEnqueuer{}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		ai,
		nil,
		nil,
		nil,
	)
	service.SetModerationEnqueuer(queue)

	description := "contains profanity"
	listing, err := service.UpdateListing(ctx, "user-1", "listing-1", &models.UpdateListingRequest{
		Description: &description,
	})
	if err != nil {
		t.Fatalf("UpdateListing returned error: %v", err)
	}
	if ai.moderateCalls != 0 {
		t.Fatalf("expected moderation to be deferred, got %d inline call(s)", ai.moderateCalls)
	}
	if len(queue.listingIDs) != 1 || queue.listingIDs[0] != listing.ListingID {
		t.Fatalf("expected moderation queue to receive listing %q, got %#v", listing.ListingID, queue.listingIDs)
	}
	if listing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected returned listing moderation status pending after requeue, got %q", listing.ModerationStatus)
	}
	if listing.Status != models.ListingStatusActive {
		t.Fatalf("expected returned listing status to remain active before worker runs, got %q", listing.Status)
	}
	if store.lastListing == nil {
		t.Fatalf("expected persisted listing to be captured")
	}
	if store.lastListing.ModerationStatus != models.ModerationStatusPending {
		t.Fatalf("expected persisted listing moderation status pending after requeue, got %q", store.lastListing.ModerationStatus)
	}
	if store.lastListing.Status != models.ListingStatusActive {
		t.Fatalf("expected persisted listing status to remain active before worker runs, got %q", store.lastListing.Status)
	}
}

func TestUpdateListingNonModeratableChangeKeepsModerationStatus(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
				},
			},
		},
	}
	queue := &fakeModerationEnqueuer{}
	newPrice := float64(1500000000)
	service := NewListingService(
		store,
		&fakeUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }()}}},
		nil, nil, nil, nil,
	)
	service.SetModerationEnqueuer(queue)

	listing, err := service.UpdateListing(ctx, "user-1", "listing-1", &models.UpdateListingRequest{
		Price: &newPrice,
	})
	if err != nil {
		t.Fatalf("UpdateListing returned error: %v", err)
	}
	if len(queue.listingIDs) != 0 {
		t.Fatalf("expected no moderation enqueue for price-only change, got %d", len(queue.listingIDs))
	}
	if listing.ModerationStatus != models.ModerationStatusApproved {
		t.Fatalf("expected moderation status to remain approved, got %q", listing.ModerationStatus)
	}
}

func TestUpdateListingRejectedListingCannotBeEdited(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Status:           models.ListingStatusArchived,
				ModerationStatus: models.ModerationStatusRejected,
				ModerationReason: "inappropriate content",
				PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Status:           models.ListingStatusArchived,
					ModerationStatus: models.ModerationStatusRejected,
					ModerationReason: "inappropriate content",
					PremiumFeatures:  models.PremiumFeatures{IsPremium: true},
				},
			},
		},
	}
	queue := &fakeModerationEnqueuer{}
	newPrice := float64(1200000000)
	service := NewListingService(
		store,
		&fakeUserStore{user: &models.User{UserID: "user-1", Subscription: models.Subscription{Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }()}}},
		nil, nil, nil, nil,
	)
	service.SetModerationEnqueuer(queue)

	_, err := service.UpdateListing(ctx, "user-1", "listing-1", &models.UpdateListingRequest{
		Price: &newPrice,
	})
	if err == nil {
		t.Fatal("expected error when editing a rejected listing, got nil")
	}
	if len(queue.listingIDs) != 0 {
		t.Fatalf("expected no moderation enqueue for rejected listing edit attempt, got %d", len(queue.listingIDs))
	}
}

func TestParseListingTextReturnsLocationSuggestionsWithoutAutoFinalizing(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{user: &models.User{UserID: "user-1"}},
		&fakeAIService{
			parseResult: &models.ParsedListing{
				Address: "Depok Beji dkt tol",
				LocationSuggestion: models.ParsedLocationSuggestion{
					Province:          "Jawa Barat",
					City:              "Depok",
					District:          "Beji",
					NormalizedAddress: "Jalan Margonda Raya, Beji, Depok",
					Confidence:        0.81,
				},
			},
		},
		nil, nil, nil,
	)

	parsed, err := service.ParseListingText(ctx, "Depok Beji dkt tol")
	if err != nil {
		t.Fatalf("ParseListingText returned error: %v", err)
	}
	if parsed.Parsed.LocationSuggestion.City != "Depok" {
		t.Fatalf("expected city suggestion Depok, got %q", parsed.Parsed.LocationSuggestion.City)
	}
	if parsed.Parsed.Address != "Depok Beji dkt tol" {
		t.Fatalf("expected raw address to remain editable, got %q", parsed.Parsed.Address)
	}
}

func TestParseListingTextSetsManualReviewWhenLowConfidence(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{user: &models.User{UserID: "user-1"}},
		&fakeAIService{
			parseResult: &models.ParsedListing{
				Address: "somewhere unclear",
				LocationSuggestion: models.ParsedLocationSuggestion{
					Province:   "",
					City:       "",
					District:   "",
					Confidence: 0.5,
				},
			},
		},
		nil, nil, nil,
	)

	result, err := service.ParseListingText(ctx, "somewhere unclear")
	if err != nil {
		t.Fatalf("ParseListingText returned error: %v", err)
	}
	if !result.RequiresCorrection {
		t.Fatalf("expected RequiresCorrection=true for low confidence, got false")
	}
}

func TestParseListingTextNormalizesViaCatalog(t *testing.T) {
	ctx := context.Background()

	catalog, err := newTestCatalog()
	if err != nil {
		t.Fatalf("build test catalog: %v", err)
	}

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{user: &models.User{UserID: "user-1"}},
		&fakeAIService{
			parseResult: &models.ParsedListing{
				Address: "Depok Beji dkt tol",
				LocationSuggestion: models.ParsedLocationSuggestion{
					Province:   "jawa barat",
					City:       "depok",
					District:   "beji",
					Confidence: 0.5,
				},
			},
		},
		nil, nil, catalog,
	)

	result, err := service.ParseListingText(ctx, "Depok Beji dkt tol")
	if err != nil {
		t.Fatalf("ParseListingText returned error: %v", err)
	}
	if result.Parsed.LocationSuggestion.Province != "Jawa Barat" {
		t.Fatalf("expected normalized province 'Jawa Barat', got %q", result.Parsed.LocationSuggestion.Province)
	}
	if result.Parsed.LocationSuggestion.City != "Depok" {
		t.Fatalf("expected normalized city 'Depok', got %q", result.Parsed.LocationSuggestion.City)
	}
	if result.Parsed.LocationSuggestion.District != "Beji" {
		t.Fatalf("expected normalized district 'Beji', got %q", result.Parsed.LocationSuggestion.District)
	}
	if result.Parsed.LocationSuggestion.Confidence < 0.99 {
		t.Fatalf("expected full confidence when all fields resolve, got %v", result.Parsed.LocationSuggestion.Confidence)
	}
}

func newTestCatalog() (*LocationCatalog, error) {
	data := `{
		"provinces":[{"id":"32","name":"Jawa Barat"}],
		"cities":[{"id":"3276","provinceId":"32","name":"Depok"}],
		"districts":[{"id":"3276010","cityId":"3276","name":"Beji"}]
	}`
	return NewLocationCatalogFromReader(strings.NewReader(data))
}

func TestCreateListingRejectsLegacyBase64Payloads(t *testing.T) {
	ctx := context.Background()

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah siap huni",
		Description: "Dekat tol",
		Price:       900000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Margonda Raya",
		},
		Images: []string{"data:image/jpeg;base64,ZmFrZQ=="},
	})
	if err == nil {
		t.Fatal("expected base64 payload to be rejected")
	}
	if appErr, ok := err.(*utils.AppError); !ok || appErr.Code != 400 {
		t.Fatalf("expected 400 app error, got %T %v", err, err)
	}
}

func TestCreateListingPromotesUploadedImagesIntoStructuredMedia(t *testing.T) {
	ctx := context.Background()
	media := &fakeMediaService{
		heads: map[string]*MediaObjectHead{
			"staging/user-1/session-1/upload.jpeg": {ContentType: "image/jpeg", SizeBytes: 1234},
			"staging/user-1/session-2/upload.png":  {ContentType: "image/png", SizeBytes: 2345},
		},
	}
	sessionStore := &fakeUploadSessionStore{
		sessions: map[string]*models.UploadSession{
			"session-1": {
				SessionID:           "session-1",
				UserID:              "user-1",
				StagingKey:          "staging/user-1/session-1/upload.jpeg",
				ExpectedContentType: "image/jpeg",
				ExpectedMaxSize:     2000,
				ExpiresAt:           time.Now().Add(time.Hour),
			},
			"session-2": {
				SessionID:           "session-2",
				UserID:              "user-1",
				StagingKey:          "staging/user-1/session-2/upload.png",
				ExpectedContentType: "image/png",
				ExpectedMaxSize:     3000,
				ExpiresAt:           time.Now().Add(time.Hour),
			},
		},
	}

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		media,
		nil,
		nil,
	)
	service.SetUploadSessionStore(sessionStore)
	service.idGenerator = sequentialIDGenerator("listing-1", "image-1", "image-2")

	listing, err := service.CreateListing(ctx, "user-1", &models.CreateListingRequest{
		Title:       "Rumah siap huni",
		Description: "Dekat tol",
		Price:       900000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Margonda Raya",
		},
		NewImageUploadSessionIDs: []string{"session-1", "session-2"},
		FeaturedUploadSessionID:  "session-2",
	})
	if err != nil {
		t.Fatalf("CreateListing returned error: %v", err)
	}
	if len(listing.Images) != 2 {
		t.Fatalf("expected 2 structured images, got %d", len(listing.Images))
	}
	if listing.Images[0].ImageID != "image-1" || listing.Images[1].ImageID != "image-2" {
		t.Fatalf("expected image IDs to be generated in order, got %#v", listing.Images)
	}
	if listing.Images[1].S3Key != "listings/listing-1/image-2" {
		t.Fatalf("expected permanent key to be stored, got %q", listing.Images[1].S3Key)
	}
	if !listing.Images[1].IsFeatured {
		t.Fatal("expected featured image to be selectable from upload session ID")
	}
	if len(media.copies) != 4 {
		t.Fatalf("expected original + thumbnail copy operations, got %#v", media.copies)
	}
	if len(sessionStore.consumeCalls) != 2 {
		t.Fatalf("expected upload sessions to be consumed, got %#v", sessionStore.consumeCalls)
	}
}

func TestUpdateListingRetainsOrderAndDeletesRemovedImages(t *testing.T) {
	ctx := context.Background()
	store := &fakeListingStore{
		now: func() time.Time { return now },
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:   "listing-1",
					UserID:      "user-1",
					Title:       "Rumah",
					Price:       900000000,
					PriceUnit:   "total",
					ListingType: models.ListingTypeSell,
					Location: models.Location{
						Address: "Jl. Margonda Raya",
					},
					Images: models.ImageEntries{
						{ImageID: "image-1", S3Key: "listings/listing-1/image-1", ThumbnailKey: "thumbnails/listing-1/image-1"},
						{ImageID: "image-2", S3Key: "listings/listing-1/image-2", ThumbnailKey: "thumbnails/listing-1/image-2", IsFeatured: true},
					},
					ImageCount:       2,
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
				},
			},
		},
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
			},
		},
	}
	media := &fakeMediaService{
		heads: map[string]*MediaObjectHead{
			"staging/user-1/session-3/upload.jpeg": {ContentType: "image/jpeg", SizeBytes: 1400},
		},
	}
	sessionStore := &fakeUploadSessionStore{
		sessions: map[string]*models.UploadSession{
			"session-3": {
				SessionID:           "session-3",
				UserID:              "user-1",
				StagingKey:          "staging/user-1/session-3/upload.jpeg",
				ExpectedContentType: "image/jpeg",
				ExpectedMaxSize:     2000,
				ExpiresAt:           time.Now().Add(time.Hour),
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionPremium, RenewDate: func() *time.Time { t := time.Now().UTC().Add(30 * 24 * time.Hour); return &t }(),
				},
			},
		},
		nil,
		media,
		nil,
		nil,
	)
	service.SetUploadSessionStore(sessionStore)
	service.idGenerator = sequentialIDGenerator("image-3")

	updated, err := service.UpdateListing(ctx, "user-1", "listing-1", &models.UpdateListingRequest{
		RetainedImageIDs:         []string{"image-2"},
		NewImageUploadSessionIDs: []string{"session-3"},
		FeaturedImageID:          stringPtr("image-2"),
	})
	if err != nil {
		t.Fatalf("UpdateListing returned error: %v", err)
	}
	if len(updated.Images) != 2 {
		t.Fatalf("expected retained + new image, got %#v", updated.Images)
	}
	if updated.Images[0].ImageID != "image-2" || updated.Images[1].ImageID != "image-3" {
		t.Fatalf("expected retained image order to be preserved, got %#v", updated.Images)
	}
	if !containsString(media.deletedKeys, "listings/listing-1/image-1") || !containsString(media.deletedKeys, "thumbnails/listing-1/image-1") {
		t.Fatalf("expected removed image original and thumbnail deletion, got %#v", media.deletedKeys)
	}
}

func TestCreateListingBlocksExpiredPremiumUser(t *testing.T) {
	ctx := context.Background()
	expiredDate := time.Now().UTC().Add(-24 * time.Hour)

	service := NewListingService(
		&fakeListingStore{},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-expired",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: &expiredDate,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.CreateListing(ctx, "user-expired", &models.CreateListingRequest{
		Title:       "Rumah expired",
		Description: "Tidak bisa tayang",
		Price:       500000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Expired No. 1",
		},
	})
	// Expired premium user is treated as free tier; free tier allows 3 active listings.
	// With 0 active listings they can still create, but they should be treated as free user
	// (not premium). We verify that the user is treated as free by checking the limit applies
	// as the free tier (3 listings) rather than premium (15 listings).
	// With 0 active, creation succeeds but under free-tier limits.
	// This test creates a second scenario where the free quota is full.
	_ = err // creation with 0 listings is still allowed under free tier

	// Now fill up the free quota and ensure the expired premium user hits it
	activeListings := make(map[string]*models.Listing, 3)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("listing-%d", i+1)
		exp := now.Add(24 * time.Hour)
		activeListings[id] = &models.Listing{
			ListingID:        id,
			UserID:           "user-expired",
			Status:           models.ListingStatusActive,
			ModerationStatus: models.ModerationStatusApproved,
			ExpiresAt:        &exp,
		}
	}

	service2 := NewListingService(
		&fakeListingStore{
			listingsByUser: map[string]map[string]*models.Listing{
				"user-expired": activeListings,
			},
		},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-expired",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionPremium,
					RenewDate: &expiredDate,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)

	_, err2 := service2.CreateListing(ctx, "user-expired", &models.CreateListingRequest{
		Title:       "Rumah 4th",
		Description: "Harusnya ditolak",
		Price:       500000000,
		PriceUnit:   "total",
		ListingType: models.ListingTypeSell,
		Location: models.Location{
			Address: "Jl. Test No. 4",
		},
	})
	if err2 == nil || !strings.Contains(err2.Error(), "free tier allows at most 3 active listing") {
		t.Fatalf("expected free tier quota error for expired premium user, got %v", err2)
	}
}

func TestRelistListingReactivatesExpiredArchivedListingWithFreshExpiry(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.March, 21, 10, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-24 * time.Hour)

	store := &fakeListingStore{
		now: func() time.Time { return now },
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah siap huni",
					Status:           models.ListingStatusArchived,
					ModerationStatus: models.ModerationStatusApproved,
					CreatedAt:        now.AddDate(0, 0, -45),
					UpdatedAt:        now.AddDate(0, 0, -5),
					ExpiresAt:        &expiredAt,
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier:      models.SubscriptionFree,
					RenewDate: nil,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	service.now = func() time.Time { return now }

	relisted, err := service.RelistListing(ctx, "user-1", "listing-1")
	if err != nil {
		t.Fatalf("RelistListing returned error: %v", err)
	}
	if relisted.Status != models.ListingStatusActive {
		t.Fatalf("expected active listing after relist, got %q", relisted.Status)
	}
	if relisted.ExpiresAt == nil {
		t.Fatal("expected relisted listing expiry to be set")
	}
	wantExpiry := now.AddDate(0, 0, freeTierListingDurationDays)
	if !relisted.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expected relisted expiry %s, got %s", wantExpiry.Format(time.RFC3339), relisted.ExpiresAt.Format(time.RFC3339))
	}
	if !relisted.CreatedAt.Equal(now) {
		t.Fatalf("expected createdAt to refresh on relist, got %s", relisted.CreatedAt.Format(time.RFC3339))
	}
	if store.lastListing == nil || store.lastListing.Status != models.ListingStatusActive {
		t.Fatalf("expected relisted listing to be persisted as active, got %#v", store.lastListing)
	}
}

func TestRelistListingRejectsWhenFreeTierQuotaIsFull(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.March, 21, 10, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-time.Hour)
	activeExpiry := now.Add(24 * time.Hour)

	store := &fakeListingStore{
		now: func() time.Time { return now },
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Status:           models.ListingStatusArchived,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &expiredAt,
				},
				"listing-2": {
					ListingID:        "listing-2",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &activeExpiry,
				},
				"listing-3": {
					ListingID:        "listing-3",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &activeExpiry,
				},
				"listing-4": {
					ListingID:        "listing-4",
					UserID:           "user-1",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &activeExpiry,
				},
			},
		},
	}

	service := NewListingService(
		store,
		&fakeUserStore{
			user: &models.User{
				UserID:       "user-1",
				Subscription: models.Subscription{Tier: models.SubscriptionFree},
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	service.now = func() time.Time { return now }

	_, err := service.RelistListing(ctx, "user-1", "listing-1")
	if err == nil || !strings.Contains(err.Error(), "free tier allows at most 3 active listing") {
		t.Fatalf("expected free-tier quota error for relist, got %v", err)
	}
}

func TestListListingsExcludesEntriesThatBecomeArchivedDuringNormalization(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.March, 22, 14, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-time.Minute)

	store := &fakeListingStore{
		now: func() time.Time { return now },
		scanListings: []models.Listing{
			{
				ListingID:        "listing-expired",
				UserID:           "user-1",
				Title:            "Listing expired",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
				ExpiresAt:        &expiredAt,
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-expired": {
					ListingID:        "listing-expired",
					UserID:           "user-1",
					Title:            "Listing expired",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					ExpiresAt:        &expiredAt,
				},
			},
		},
	}

	service := NewListingService(store, &fakeUserStore{}, nil, nil, nil, nil)
	service.now = func() time.Time { return now }

	listings, err := service.ListListings(ctx, &models.ListingSearchParams{})
	if err != nil {
		t.Fatalf("ListListings returned error: %v", err)
	}
	if len(listings) != 0 {
		t.Fatalf("expected expired listing to be removed from public results, got %#v", listings)
	}
	if store.lastListing == nil || store.lastListing.Status != models.ListingStatusArchived {
		t.Fatalf("expected expired listing to be archived during normalization, got %#v", store.lastListing)
	}
}

func sequentialIDGenerator(values ...string) func() string {
	index := 0
	return func() string {
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}

func stringPtr(value string) *string {
	return &value
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
