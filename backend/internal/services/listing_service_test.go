package services

import (
	"context"
	"strings"
	"testing"

	"github.com/fiando/propti/backend/internal/models"
)

// --- fakes ---

type fakeListingStore struct{}

func (f *fakeListingStore) Put(ctx context.Context, listing *models.Listing) error { return nil }
func (f *fakeListingStore) GetByID(ctx context.Context, userID, listingID string) (*models.Listing, error) {
	return nil, nil
}
func (f *fakeListingStore) GetByListingID(ctx context.Context, listingID string) (*models.Listing, error) {
	return nil, nil
}
func (f *fakeListingStore) CountMonthlyByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
func (f *fakeListingStore) Delete(ctx context.Context, userID, listingID string) error { return nil }
func (f *fakeListingStore) Scan(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	return nil, nil
}

type fakeUserStore struct {
	user *models.User
}

func (f *fakeUserStore) GetByID(ctx context.Context, userID string) (*models.User, error) {
	return f.user, nil
}

type fakeAIService struct {
	parseResult *models.ParsedListing
	err         error
}

func (f *fakeAIService) ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error) {
	return f.parseResult, f.err
}

// --- tests ---

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
