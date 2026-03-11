package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

// ListingHandler handles listing CRUD, text parsing and media upload URLs.
type ListingHandler struct {
	listingService *services.ListingService
	userRepo       *repository.UserRepo
}

// NewListingHandler creates a ListingHandler.
func NewListingHandler(listingService *services.ListingService, userRepo *repository.UserRepo) *ListingHandler {
	return &ListingHandler{
		listingService: listingService,
		userRepo:       userRepo,
	}
}

// Handle routes API Gateway requests to the appropriate sub-handler.
func (h *ListingHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	method := req.HTTPMethod

	switch {
	case method == http.MethodOptions:
		return jsonResponse(http.StatusOK, ""), nil

	case method == http.MethodPost && path == "/listings":
		return h.createListing(ctx, req)

	case method == http.MethodGet && path == "/listings":
		return h.listListings(ctx, req)

	case method == http.MethodGet && isListingPath(path):
		return h.getListing(ctx, req)

	case method == http.MethodPut && isListingPath(path):
		return h.updateListing(ctx, req)

	case method == http.MethodDelete && isListingPath(path):
		return h.deleteListing(ctx, req)

	case method == http.MethodPost && path == "/listings/parse-text":
		return h.parseText(ctx, req)

	case method == http.MethodPost && strings.HasSuffix(path, "/upload-url"):
		return h.getUploadURL(ctx, req)

	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

func (h *ListingHandler) createListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var createReq models.CreateListingRequest
	if err := json.Unmarshal([]byte(req.Body), &createReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	listing, err := h.listingService.CreateListing(ctx, userID, &createReq)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(listing)
	return jsonResponse(http.StatusCreated, string(body)), nil
}

func (h *ListingHandler) listListings(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	params := parseSearchParams(req)
	listings, err := h.listingService.ListListings(ctx, params)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(map[string]any{
		"listings": listings,
		"total":    len(listings),
		"page":     params.Page,
		"pageSize": params.PageSize,
	})
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) getListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	listing, err := h.listingService.GetListing(ctx, listingID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(listing)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) updateListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	var updateReq models.UpdateListingRequest
	if err := json.Unmarshal([]byte(req.Body), &updateReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	listing, err := h.listingService.UpdateListing(ctx, userID, listingID, &updateReq)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(listing)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) deleteListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	if err := h.listingService.DeleteListing(ctx, userID, listingID); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	return jsonResponse(http.StatusNoContent, ""), nil
}

func (h *ListingHandler) parseText(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse-text is available to authenticated users only.
	if _, err := extractUserID(req); err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var parseReq models.ParseTextRequest
	if err := json.Unmarshal([]byte(req.Body), &parseReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if strings.TrimSpace(parseReq.Text) == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, "text is required"))), nil
	}

	result, err := h.listingService.ParseListingText(ctx, parseReq.Text)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(result)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) getUploadURL(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	var body struct {
		Filename    string `json:"filename"`
		ContentType string `json:"contentType"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.Filename == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if body.ContentType == "" {
		body.ContentType = "application/octet-stream"
	}

	uploadURL, key, err := h.listingService.GetUploadURL(ctx, userID, listingID, body.Filename, body.ContentType)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr)), nil
		}
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp, _ := json.Marshal(map[string]string{"uploadUrl": uploadURL, "key": key})
	return jsonResponse(http.StatusOK, string(resp)), nil
}

// --- path helpers ---

func isListingPath(path string) bool {
	// Matches /listings/{id} but NOT /listings/parse-text
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 && parts[0] == "listings" && parts[1] != "parse-text"
}

func extractListingID(req events.APIGatewayProxyRequest) string {
	if id, ok := req.PathParameters["id"]; ok {
		return id
	}
	// Fallback: parse from path.
	parts := strings.Split(strings.Trim(req.Path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func parseSearchParams(req events.APIGatewayProxyRequest) *models.ListingSearchParams {
	p := &models.ListingSearchParams{
		Query:    req.QueryStringParameters["q"],
		City:     req.QueryStringParameters["city"],
		SortBy:   req.QueryStringParameters["sortBy"],
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["priceMin"], 64); err == nil {
		p.PriceMin = v
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["priceMax"], 64); err == nil {
		p.PriceMax = v
	}
	if v, err := strconv.Atoi(req.QueryStringParameters["bedrooms"]); err == nil {
		p.Bedrooms = v
	}
	if v, err := strconv.Atoi(req.QueryStringParameters["page"]); err == nil {
		p.Page = v
	}
	if v, err := strconv.Atoi(req.QueryStringParameters["pageSize"]); err == nil {
		p.PageSize = v
	}
	return p
}
