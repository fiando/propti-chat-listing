package handlers

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakeListingHandlerService struct {
	publicListing *models.Listing
	ownerListing  *models.Listing
}

type fakeUploadPrepareService struct {
	resp *models.UploadPrepareResponse
	err  error
}

func (f *fakeListingHandlerService) CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) ListListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) ListMyListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) ListSavedListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) GetListing(ctx context.Context, listingID string) (*models.Listing, error) {
	return f.publicListing, nil
}

func (f *fakeListingHandlerService) GetOwnerListing(ctx context.Context, userID, listingID string) (*models.Listing, error) {
	return f.ownerListing, nil
}

func (f *fakeListingHandlerService) RecordListingView(ctx context.Context, listingID string) (*models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) RevealListingContact(ctx context.Context, viewerUserID, listingID string, channel models.ContactRevealChannel) (*models.ListingContactReveal, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) UpdateListing(ctx context.Context, userID, listingID string, req *models.UpdateListingRequest) (*models.Listing, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) DeleteListing(ctx context.Context, userID, listingID string) error {
	return nil
}

func (f *fakeListingHandlerService) SaveListing(ctx context.Context, userID, listingID string) error {
	return nil
}

func (f *fakeListingHandlerService) UnsaveListing(ctx context.Context, userID, listingID string) error {
	return nil
}

func (f *fakeListingHandlerService) ParseListingText(ctx context.Context, text string) (*models.ParseTextResponse, error) {
	return nil, nil
}

func (f *fakeListingHandlerService) GetUploadURL(ctx context.Context, userID, listingID, filename, contentType string) (string, string, error) {
	return "", "", nil
}

func (f *fakeUploadPrepareService) PrepareUpload(ctx context.Context, userID string, req *models.UploadPrepareRequest) (*models.UploadPrepareResponse, error) {
	return f.resp, f.err
}

func TestParseSearchParamsSupportsComprehensiveFilters(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"q":               "rumah keluarga",
			"province":        "DKI Jakarta",
			"city":            "Jakarta Selatan",
			"listingType":     "sell",
			"priceMin":        "500000000",
			"priceMax":        "2000000000",
			"bedrooms":        "3",
			"bathrooms":       "2",
			"buildingAreaMin": "90",
			"buildingAreaMax": "180",
			"landAreaMin":     "120",
			"landAreaMax":     "250",
			"legalStatus":     "SHM - Sertifikat Hak Milik",
			"amenities":       "carport,kolam_renang",
			"sortBy":          "popular",
			"page":            "2",
			"pageSize":        "12",
		},
	}

	params := parseSearchParams(req)

	if params.Query != "rumah keluarga" {
		t.Fatalf("expected query to be parsed, got %q", params.Query)
	}
	if params.Bathrooms != 2 {
		t.Fatalf("expected bathrooms=2, got %d", params.Bathrooms)
	}
	if params.BuildingAreaMin != 90 || params.BuildingAreaMax != 180 {
		t.Fatalf("expected building area min/max 90-180, got %v-%v", params.BuildingAreaMin, params.BuildingAreaMax)
	}
	if params.LandAreaMin != 120 || params.LandAreaMax != 250 {
		t.Fatalf("expected land area min/max 120-250, got %v-%v", params.LandAreaMin, params.LandAreaMax)
	}
	if params.LegalStatus != "SHM - Sertifikat Hak Milik" {
		t.Fatalf("expected legal status to be parsed, got %q", params.LegalStatus)
	}
	if !reflect.DeepEqual(params.Amenities, []string{"carport", "kolam_renang"}) {
		t.Fatalf("unexpected amenities: %#v", params.Amenities)
	}
}

func TestTemplateExposesListingViewAndContactRevealRoutes(t *testing.T) {
	template, err := os.ReadFile("../../template.yaml")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	content := string(template)

	if !strings.Contains(content, "Path: /listings/{id}/view") {
		t.Fatal("expected SAM template to expose POST /listings/{id}/view")
	}
	if !strings.Contains(content, "Path: /listings/{id}/contact-reveal") {
		t.Fatal("expected SAM template to expose POST /listings/{id}/contact-reveal")
	}
	if !strings.Contains(content, "Path: /listings/upload-prepare") {
		t.Fatal("expected SAM template to expose POST /listings/upload-prepare")
	}
	if !strings.Contains(content, "Path: /users/me/listings/{id}") {
		t.Fatal("expected SAM template to expose GET /users/me/listings/{id}")
	}
}

func TestListingHandlerPrepareUploadEndpointReturnsSlots(t *testing.T) {
	token, err := utils.GenerateToken("user-1", "user-1@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	handler := NewListingHandler(
		&fakeListingHandlerService{},
		&fakeUploadPrepareService{
			resp: &models.UploadPrepareResponse{
				Slots: []models.UploadSlot{
					{SessionID: "session-1", PresignedURL: "https://upload.example/1", StagingKey: "staging/user-1/session-1/upload.jpeg"},
				},
			},
		},
		services.NewListingMediaPresenter(&servicesNoopMedia{}),
	)

	body, _ := json.Marshal(models.UploadPrepareRequest{
		RetainedImageCount: 0,
		FinalImageCount:    1,
		NewImages:          []models.NewImageSpec{{ContentType: "image/jpeg", SizeBytes: 1024}},
	})
	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/listings/upload-prepare",
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
		Body: string(body),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 response, got %d body=%s", resp.StatusCode, resp.Body)
	}
	if !strings.Contains(resp.Body, "\"sessionId\":\"session-1\"") {
		t.Fatalf("expected response body to include slot payload, got %s", resp.Body)
	}
}

type servicesNoopMedia struct{}

func (s *servicesNoopMedia) GetPresignedUploadURL(ctx context.Context, key, contentType string) (string, error) {
	return "", nil
}

func (s *servicesNoopMedia) GetSignedDownloadURL(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (s *servicesNoopMedia) BuildPublicURL(key string) string {
	return ""
}

func (s *servicesNoopMedia) HeadObject(ctx context.Context, key string) (*services.MediaObjectHead, error) {
	return nil, nil
}

func (s *servicesNoopMedia) CopyObject(ctx context.Context, sourceKey, destinationKey string) error {
	return nil
}

func (s *servicesNoopMedia) DeleteObject(ctx context.Context, key string) error {
	return nil
}

func (s *servicesNoopMedia) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
