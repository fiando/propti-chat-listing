package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
)

type fakeLocationCatalogService struct {
	provinces []services.Province
	cities    []services.City
	districts []services.District
}

type fakeSearchIntentService struct {
	resp *models.SearchIntentResponse
	err  error
}

func (f *fakeLocationCatalogService) SearchProvinces(_ string) []services.Province {
	return f.provinces
}

func (f *fakeLocationCatalogService) SearchCities(_, _ string) []services.City {
	return f.cities
}

func (f *fakeLocationCatalogService) SearchDistricts(_, _ string) []services.District {
	return f.districts
}

func (f *fakeSearchIntentService) ParseIntent(_ context.Context, query string) (*models.SearchIntentResponse, error) {
	if f.resp != nil {
		return f.resp, nil
	}

	return &models.SearchIntentResponse{
		SearchParams: models.ListingSearchParams{
			Province: "DI Yogyakarta",
			City:     "Gunungkidul",
			SortBy:   "price_asc",
		},
		Normalized: models.SearchIntent{
			Query:    query,
			Province: "DI Yogyakarta",
			City:     "Gunungkidul",
			SortBy:   "price_asc",
		},
	}, f.err
}

func TestSearchHandlerReturnsCitiesFromLocalCatalog(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{
		cities: []services.City{{ID: "3276", Name: "Depok", ProvinceID: "32"}},
	}, nil)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/locations/cities",
		QueryStringParameters: map[string]string{
			"provinceId": "32",
			"q":          "dep",
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := body["cities"]; !ok {
		t.Fatal("response missing 'cities' key")
	}
}

func TestSearchHandlerCitiesMissingProvinceId(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{}, nil)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		Path:                  "/locations/cities",
		QueryStringParameters: map[string]string{"q": "dep"},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSearchHandlerReturnsProvincesFromLocalCatalog(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{
		provinces: []services.Province{{ID: "32", Name: "Jawa Barat"}},
	}, nil)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		Path:                  "/locations/provinces",
		QueryStringParameters: map[string]string{"q": "jawa"},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	provinces, ok := body["provinces"]
	if !ok || provinces == nil {
		t.Fatal("response missing 'provinces' key")
	}
}

func TestSearchHandlerReturnsDistrictsFromLocalCatalog(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{
		districts: []services.District{{ID: "3276010", Name: "Beji", CityID: "3276"}},
	}, nil)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		Path:                  "/locations/districts",
		QueryStringParameters: map[string]string{"cityId": "3276", "q": "bej"},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := body["districts"]; !ok {
		t.Fatal("response missing 'districts' key")
	}
}

func TestSearchHandlerDistrictsMissingCityId(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{}, nil)

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		Path:                  "/locations/districts",
		QueryStringParameters: map[string]string{"q": "bej"},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSearchHandlerCatalogUnavailable(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, nil, nil)

	for _, path := range []string{"/locations/provinces", "/locations/cities", "/locations/districts"} {
		resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Path:       path,
		})
		if err != nil {
			t.Fatalf("%s: Handle returned error: %v", path, err)
		}
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("%s: expected 503, got %d", path, resp.StatusCode)
		}
	}
}

func TestSearchHandlerParsesSmartSearchIntent(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(nil, nil, &fakeLocationCatalogService{}, &fakeSearchIntentService{})

	resp, err := handler.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/search/parse-intent",
		Body:       `{"query":"rumah dijual di jogja harga termurah"}`,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, resp.Body)
	}

	var body struct {
		SearchParams map[string]any `json:"searchParams"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.SearchParams["province"] != "DI Yogyakarta" {
		t.Fatalf("expected normalized province, got %#v", body.SearchParams["province"])
	}
	if body.SearchParams["sortBy"] != "price_asc" {
		t.Fatalf("expected normalized sort, got %#v", body.SearchParams["sortBy"])
	}
}
