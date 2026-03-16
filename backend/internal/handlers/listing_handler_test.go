package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

// --- fakes ---

type fakeListingStore struct {
	byUserID    map[string][]models.Listing
	byListingID map[string]*models.Listing
}

func (f *fakeListingStore) Put(_ context.Context, listing *models.Listing) error {
	if f.byListingID == nil {
		f.byListingID = map[string]*models.Listing{}
	}
	if listing != nil {
		copy := *listing
		f.byListingID[listing.ListingID] = &copy
		if f.byUserID == nil {
			f.byUserID = map[string][]models.Listing{}
		}
		userListings := f.byUserID[listing.UserID]
		replaced := false
		for i := range userListings {
			if userListings[i].ListingID == listing.ListingID {
				userListings[i] = copy
				replaced = true
				break
			}
		}
		if !replaced {
			userListings = append(userListings, copy)
		}
		f.byUserID[listing.UserID] = userListings
	}
	return nil
}
func (f *fakeListingStore) GetByID(_ context.Context, userID, listingID string) (*models.Listing, error) {
	for _, listing := range f.byUserID[userID] {
		if listing.ListingID == listingID {
			copy := listing
			return &copy, nil
		}
	}
	return nil, nil
}
func (f *fakeListingStore) GetByListingID(_ context.Context, listingID string) (*models.Listing, error) {
	if listing, ok := f.byListingID[listingID]; ok {
		copy := *listing
		return &copy, nil
	}
	return nil, nil
}
func (f *fakeListingStore) CountMonthlyByUserID(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (f *fakeListingStore) Delete(_ context.Context, _, _ string) error { return nil }
func (f *fakeListingStore) Scan(_ context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	if len(f.byListingID) == 0 {
		return nil, nil
	}

	paramsValue := reflect.ValueOf(params)
	if !paramsValue.IsValid() || paramsValue.IsNil() {
		return nil, nil
	}

	var province string
	if provinceField := paramsValue.Elem().FieldByName("Province"); provinceField.IsValid() && provinceField.Kind() == reflect.String {
		province = provinceField.String()
	}

	cityField := paramsValue.Elem().FieldByName("City")
	city := ""
	if cityField.IsValid() && cityField.Kind() == reflect.String {
		city = cityField.String()
	}

	filtered := make([]models.Listing, 0, len(f.byListingID))
	for _, listing := range f.byListingID {
		if listing == nil {
			continue
		}
		if province != "" && !strings.EqualFold(listing.Location.Province, province) {
			continue
		}
		if city != "" && !strings.EqualFold(listing.Location.City, city) {
			continue
		}
		copy := *listing
		filtered = append(filtered, copy)
	}

	return filtered, nil
}

func (f *fakeListingStore) ListByUserID(_ context.Context, userID string, _ int32) ([]models.Listing, error) {
	listings := f.byUserID[userID]
	result := make([]models.Listing, len(listings))
	copy(result, listings)
	return result, nil
}

type fakeUserStore struct {
	byID map[string]*models.User
}

func (f *fakeUserStore) GetByID(_ context.Context, userID string) (*models.User, error) {
	if f.byID == nil {
		return nil, nil
	}
	if user, ok := f.byID[userID]; ok {
		copy := *user
		return &copy, nil
	}
	return nil, nil
}

func (f *fakeUserStore) Put(_ context.Context, user *models.User) error {
	if f.byID == nil {
		f.byID = map[string]*models.User{}
	}
	if user == nil {
		return nil
	}
	copy := *user
	f.byID[user.UserID] = &copy
	return nil
}

type fakeAIParseService struct {
	result *models.ParsedListing
}

func (f *fakeAIParseService) ParseListingText(_ context.Context, _ string) (*models.ParsedListing, error) {
	return f.result, nil
}

type fakeLocationNormalizer struct{}

func (f *fakeLocationNormalizer) NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion {
	return models.ParsedLocationSuggestion{
		Province:          province,
		City:              city,
		District:          district,
		NormalizedAddress: "Jalan Margonda Raya, Beji, Depok, Jawa Barat",
		Confidence:        0.9,
	}
}

// newTestListingHandler wires a ListingHandler with in-memory fakes.
func newTestListingHandler(ai services.AIParseService, loc services.LocationNormalizer) *ListingHandler {
	return newTestListingHandlerWithStores(&fakeListingStore{}, &fakeUserStore{}, ai, loc)
}

func newTestListingHandlerWithStores(
	listingStore *fakeListingStore,
	userStore *fakeUserStore,
	ai services.AIParseService,
	loc services.LocationNormalizer,
) *ListingHandler {
	svc := services.NewListingService(
		listingStore,
		userStore,
		ai,
		nil, // s3 not needed for parse-text
		nil, // maps not needed for parse-text
		loc,
	)
	return NewListingHandler(svc, nil)
}

// authHeader generates a valid Bearer token for use in test requests.
func authHeader(t *testing.T, userID string) map[string]string {
	t.Helper()
	token, err := utils.GenerateToken(userID, "test@example.com")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

// --- tests ---

func TestParseTextHandlerReturnsStructuredResponse(t *testing.T) {
	t.Parallel()

	ai := &fakeAIParseService{
		result: &models.ParsedListing{
			Title:       "Rumah Dijual Depok",
			Description: "Rumah 3KT 2KM strategis dekat tol",
			Price:       1500000000,
			PriceUnit:   "total",
			Address:     "Jl. Margonda Raya No. 1, Depok",
			PropertyDetails: models.PropertyDetails{
				Bedrooms:  3,
				Bathrooms: 2,
				LandArea:  120,
			},
			LocationSuggestion: models.ParsedLocationSuggestion{
				Province:          "Jawa Barat",
				City:              "Depok",
				District:          "Beji",
				NormalizedAddress: "Jalan Margonda Raya, Beji, Depok",
				Confidence:        0.9,
			},
			Confidence:           0.88,
			RequiresManualReview: false,
		},
	}

	handler := newTestListingHandler(ai, &fakeLocationNormalizer{})

	body, _ := json.Marshal(models.ParseTextRequest{Text: "Jual rumah Depok Beji 3KT 2KM LT120 1.5M"})
	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/parse-text",
		Headers:    authHeader(t, "user-test-1"),
		Body:       string(body),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", resp.StatusCode, resp.Body)
	}

	var parsed models.ParseTextResponse
	if err := json.Unmarshal([]byte(resp.Body), &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if parsed.Parsed.Title == "" {
		t.Error("expected non-empty parsed.title")
	}
	if parsed.Parsed.LocationSuggestion.City == "" {
		t.Error("expected non-empty parsed.locationSuggestion.city")
	}
	if parsed.Parsed.LocationSuggestion.NormalizedAddress == "" {
		t.Error("expected non-empty parsed.locationSuggestion.normalizedAddress")
	}
	if parsed.Confidence <= 0 {
		t.Errorf("expected positive confidence, got %f", parsed.Confidence)
	}
}

func TestParseTextHandlerRequiresAuth(t *testing.T) {
	t.Parallel()

	handler := newTestListingHandler(&fakeAIParseService{result: &models.ParsedListing{}}, nil)

	body, _ := json.Marshal(models.ParseTextRequest{Text: "some text"})
	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/parse-text",
		Body:       string(body),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestParseTextHandlerRejectsMissingText(t *testing.T) {
	t.Parallel()

	handler := newTestListingHandler(&fakeAIParseService{result: &models.ParsedListing{}}, nil)

	body, _ := json.Marshal(models.ParseTextRequest{Text: "   "})
	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/parse-text",
		Headers:    authHeader(t, "user-test-2"),
		Body:       string(body),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestParseTextHandlerWithoutLocationCatalog(t *testing.T) {
	t.Parallel()

	ai := &fakeAIParseService{
		result: &models.ParsedListing{
			Title:   "Rumah Bandung",
			Address: "Jl. Dago No. 5",
			LocationSuggestion: models.ParsedLocationSuggestion{
				Province:   "Jawa Barat",
				City:       "Bandung",
				Confidence: 0.6,
			},
			Confidence: 0.75,
		},
	}

	handler := newTestListingHandler(ai, nil)

	body, _ := json.Marshal(models.ParseTextRequest{Text: "Jual rumah Bandung Dago"})
	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/parse-text",
		Headers:    authHeader(t, "user-test-3"),
		Body:       string(body),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var parsed models.ParseTextResponse
	if err := json.Unmarshal([]byte(resp.Body), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.Parsed.LocationSuggestion.City != "Bandung" {
		t.Errorf("expected city Bandung, got %q", parsed.Parsed.LocationSuggestion.City)
	}
	if !parsed.RequiresCorrection {
		t.Error("expected requiresCorrection=true when location confidence < 0.7")
	}
}

func TestMyListingsHandlerReturnsAuthenticatedUsersListings(t *testing.T) {
	t.Parallel()

	handler := newTestListingHandlerWithStores(
		&fakeListingStore{
			byUserID: map[string][]models.Listing{
				"user-test-4": {
					{
						ListingID: "listing-1",
						UserID:    "user-test-4",
						Title:     "Rumah Depok",
						Status:    models.ListingStatusActive,
					},
				},
			},
		},
		&fakeUserStore{
			byID: map[string]*models.User{
				"user-test-4": {UserID: "user-test-4"},
			},
		},
		&fakeAIParseService{result: &models.ParsedListing{}},
		nil,
	)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/users/me/listings",
		Headers:    authHeader(t, "user-test-4"),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", resp.StatusCode, resp.Body)
	}

	var body struct {
		Listings []models.Listing `json:"listings"`
		Total    int              `json:"total"`
		Page     int              `json:"page"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if len(body.Listings) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(body.Listings))
	}
	if body.Listings[0].ListingID != "listing-1" {
		t.Fatalf("expected listing-1, got %q", body.Listings[0].ListingID)
	}
}

func TestSaveListingHandlerPersistsSavedListing(t *testing.T) {
	t.Parallel()

	listingStore := &fakeListingStore{
		byUserID: map[string][]models.Listing{
			"owner-1": {{
				ListingID:        "listing-save-1",
				UserID:           "owner-1",
				Title:            "Rumah Cinere",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
			}},
		},
		byListingID: map[string]*models.Listing{
			"listing-save-1": {
				ListingID:        "listing-save-1",
				UserID:           "owner-1",
				Title:            "Rumah Cinere",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
			},
		},
	}
	userStore := &fakeUserStore{byID: map[string]*models.User{
		"user-test-save": {UserID: "user-test-save"},
	}}
	handler := newTestListingHandlerWithStores(listingStore, userStore, &fakeAIParseService{result: &models.ParsedListing{}}, nil)

	saveResp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/listing-save-1/save",
		Headers:    authHeader(t, "user-test-save"),
	})
	if err != nil {
		t.Fatalf("save Handle returned error: %v", err)
	}
	if saveResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 from save, got %d body=%s", saveResp.StatusCode, saveResp.Body)
	}

	listResp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/users/me/saved",
		Headers:    authHeader(t, "user-test-save"),
	})
	if err != nil {
		t.Fatalf("saved list Handle returned error: %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from saved list, got %d body=%s", listResp.StatusCode, listResp.Body)
	}

	var body struct {
		Listings []models.Listing `json:"listings"`
	}
	if err := json.Unmarshal([]byte(listResp.Body), &body); err != nil {
		t.Fatalf("saved list response is not valid JSON: %v", err)
	}
	if len(body.Listings) != 1 || body.Listings[0].ListingID != "listing-save-1" {
		t.Fatalf("expected saved listings to include listing-save-1, got %#v", body.Listings)
	}
}

func TestUnsaveListingHandlerRemovesSavedListing(t *testing.T) {
	t.Parallel()

	listingStore := &fakeListingStore{
		byUserID: map[string][]models.Listing{
			"owner-2": {{
				ListingID:        "listing-save-2",
				UserID:           "owner-2",
				Title:            "Apartemen Blok M",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
			}},
		},
		byListingID: map[string]*models.Listing{
			"listing-save-2": {
				ListingID:        "listing-save-2",
				UserID:           "owner-2",
				Title:            "Apartemen Blok M",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusApproved,
			},
		},
	}
	userStore := &fakeUserStore{byID: map[string]*models.User{
		"user-test-unsave": {UserID: "user-test-unsave"},
	}}
	handler := newTestListingHandlerWithStores(listingStore, userStore, &fakeAIParseService{result: &models.ParsedListing{}}, nil)

	_, _ = handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/listings/listing-save-2/save",
		Headers:    authHeader(t, "user-test-unsave"),
	})

	unsaveResp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodDelete,
		Path:       "/listings/listing-save-2/save",
		Headers:    authHeader(t, "user-test-unsave"),
	})
	if err != nil {
		t.Fatalf("unsave Handle returned error: %v", err)
	}
	if unsaveResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 from unsave, got %d body=%s", unsaveResp.StatusCode, unsaveResp.Body)
	}

	listResp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/users/me/saved",
		Headers:    authHeader(t, "user-test-unsave"),
	})
	if err != nil {
		t.Fatalf("saved list Handle returned error: %v", err)
	}

	var body struct {
		Listings []models.Listing `json:"listings"`
	}
	if err := json.Unmarshal([]byte(listResp.Body), &body); err != nil {
		t.Fatalf("saved list response is not valid JSON: %v", err)
	}
	if len(body.Listings) != 0 {
		t.Fatalf("expected no saved listings after unsave, got %#v", body.Listings)
	}
}

func TestGetListingHandlerRejectsPendingListings(t *testing.T) {
	t.Parallel()

	handler := newTestListingHandlerWithStores(
		&fakeListingStore{
			byListingID: map[string]*models.Listing{
				"listing-pending": {
					ListingID:        "listing-pending",
					UserID:           "user-test-5",
					Title:            "Pending Listing",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusPending,
				},
			},
		},
		&fakeUserStore{},
		&fakeAIParseService{result: &models.ParsedListing{}},
		nil,
	)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/listings/listing-pending",
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for pending listing, got %d — body: %s", resp.StatusCode, resp.Body)
	}
}

func TestListListingsHandlerFiltersByProvince(t *testing.T) {
	t.Parallel()

	handler := newTestListingHandlerWithStores(
		&fakeListingStore{
			byListingID: map[string]*models.Listing{
				"listing-jambi": {
					ListingID:        "listing-jambi",
					UserID:           "user-jambi",
					Title:            "Rumah Jambi",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					Location: models.Location{
						Province: "Jambi",
						City:     "Kota Jambi",
					},
				},
				"listing-jabar": {
					ListingID:        "listing-jabar",
					UserID:           "user-jabar",
					Title:            "Rumah Bekasi",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusApproved,
					Location: models.Location{
						Province: "Jawa Barat",
						City:     "Bekasi",
					},
				},
			},
		},
		&fakeUserStore{},
		&fakeAIParseService{result: &models.ParsedListing{}},
		nil,
	)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/listings",
		QueryStringParameters: map[string]string{
			"province": "Jambi",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", resp.StatusCode, resp.Body)
	}

	var body struct {
		Listings []models.Listing `json:"listings"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if len(body.Listings) != 1 {
		t.Fatalf("expected 1 listing for province filter, got %d", len(body.Listings))
	}
	if body.Listings[0].ListingID != "listing-jambi" {
		t.Fatalf("expected listing-jambi, got %q", body.Listings[0].ListingID)
	}
}
