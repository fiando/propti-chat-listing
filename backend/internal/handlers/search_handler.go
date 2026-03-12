package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

// LocationCatalogService is the interface for hierarchical Indonesia location lookup.
type LocationCatalogService interface {
	SearchProvinces(query string) []services.Province
	SearchCities(provinceID, query string) []services.City
	SearchDistricts(cityID, query string) []services.District
}

// SearchHandler handles nearby search and location autocomplete.
type SearchHandler struct {
	listingRepo     *repository.ListingRepo
	mapsService     *services.GoogleMapsService
	locationCatalog LocationCatalogService
}

// NewSearchHandler creates a SearchHandler.
func NewSearchHandler(listingRepo *repository.ListingRepo, mapsService *services.GoogleMapsService, locationCatalog LocationCatalogService) *SearchHandler {
	return &SearchHandler{
		listingRepo:     listingRepo,
		mapsService:     mapsService,
		locationCatalog: locationCatalog,
	}
}

// Handle routes API Gateway requests.
func (h *SearchHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == http.MethodOptions:
		return jsonResponse(http.StatusOK, ""), nil
	case req.HTTPMethod == http.MethodGet && req.Path == "/search/nearby":
		return h.searchNearby(ctx, req)
	case req.HTTPMethod == http.MethodGet && req.Path == "/locations/suggestions":
		return h.locationSuggestions(ctx, req)
	case req.HTTPMethod == http.MethodGet && req.Path == "/locations/provinces":
		return h.provinceSuggestions(ctx, req)
	case req.HTTPMethod == http.MethodGet && req.Path == "/locations/cities":
		return h.citySuggestions(ctx, req)
	case req.HTTPMethod == http.MethodGet && req.Path == "/locations/districts":
		return h.districtSuggestions(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

// searchNearby returns listings near a given lat/lng within an optional radius.
func (h *SearchHandler) searchNearby(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	latStr := req.QueryStringParameters["lat"]
	lngStr := req.QueryStringParameters["lng"]
	radiusStr := req.QueryStringParameters["radius"] // km

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || latStr == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "lat is required and must be a number"))), nil
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lngStr == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "lng is required and must be a number"))), nil
	}

	radius := 5.0 // default 5 km
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil && r > 0 {
			radius = r
		}
	}

	limit := int32(20)
	if l, err := strconv.Atoi(req.QueryStringParameters["limit"]); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = int32(l)
	}

	listings, err := h.listingRepo.ScanNearby(ctx, lat, lng, radius, limit)
	if err != nil {
		utils.LogError("scan nearby listings", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(map[string]any{
		"listings": listings,
		"total":    len(listings),
		"lat":      lat,
		"lng":      lng,
		"radius":   radius,
	})
	return jsonResponse(http.StatusOK, string(body)), nil
}

// locationSuggestions proxies the Google Places Autocomplete API.
func (h *SearchHandler) locationSuggestions(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	input := req.QueryStringParameters["q"]
	if input == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "q is required"))), nil
	}

	if h.mapsService == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "maps service unavailable"))), nil
	}

	suggestions, err := h.mapsService.GetPlaceSuggestions(ctx, input)
	if err != nil {
		utils.LogError("place suggestions", err, "input", input)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(map[string]any{"suggestions": suggestions})
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *SearchHandler) provinceSuggestions(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if h.locationCatalog == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "location catalog unavailable"))), nil
	}
	provinces := h.locationCatalog.SearchProvinces(req.QueryStringParameters["q"])
	body, _ := json.Marshal(map[string]any{"provinces": provinces})
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *SearchHandler) citySuggestions(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if h.locationCatalog == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "location catalog unavailable"))), nil
	}
	cities := h.locationCatalog.SearchCities(req.QueryStringParameters["provinceId"], req.QueryStringParameters["q"])
	body, _ := json.Marshal(map[string]any{"cities": cities})
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *SearchHandler) districtSuggestions(_ context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if h.locationCatalog == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "location catalog unavailable"))), nil
	}
	districts := h.locationCatalog.SearchDistricts(req.QueryStringParameters["cityId"], req.QueryStringParameters["q"])
	body, _ := json.Marshal(map[string]any{"districts": districts})
	return jsonResponse(http.StatusOK, string(body)), nil
}
