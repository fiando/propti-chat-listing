package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

// --- fakes ---

type fakeListingStore struct{}

func (f *fakeListingStore) Put(_ context.Context, _ *models.Listing) error { return nil }
func (f *fakeListingStore) GetByID(_ context.Context, _, _ string) (*models.Listing, error) {
	return nil, nil
}
func (f *fakeListingStore) GetByListingID(_ context.Context, _ string) (*models.Listing, error) {
	return nil, nil
}
func (f *fakeListingStore) CountMonthlyByUserID(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (f *fakeListingStore) Delete(_ context.Context, _, _ string) error { return nil }
func (f *fakeListingStore) Scan(_ context.Context, _ *models.ListingSearchParams) ([]models.Listing, error) {
	return nil, nil
}

type fakeUserStore struct{}

func (f *fakeUserStore) GetByID(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
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
	svc := services.NewListingService(
		&fakeListingStore{},
		&fakeUserStore{},
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
