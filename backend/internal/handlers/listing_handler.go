package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

type listingService interface {
	CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error)
	ListListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error)
	ListMyListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error)
	ListSavedListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error)
	GetListing(ctx context.Context, listingID string) (*models.Listing, error)
	GetOwnerListing(ctx context.Context, userID, listingID string) (*models.Listing, error)
	RecordListingView(ctx context.Context, listingID string) (*models.Listing, error)
	RevealListingContact(ctx context.Context, viewerUserID, listingID string, channel models.ContactRevealChannel) (*models.ListingContactReveal, error)
	UpdateListing(ctx context.Context, userID, listingID string, req *models.UpdateListingRequest) (*models.Listing, error)
	DeleteListing(ctx context.Context, userID, listingID string) error
	SaveListing(ctx context.Context, userID, listingID string) error
	UnsaveListing(ctx context.Context, userID, listingID string) error
	ParseListingText(ctx context.Context, text string) (*models.ParseTextResponse, error)
	GetUploadURL(ctx context.Context, userID, listingID, filename, contentType string) (string, string, error)
}

type uploadPrepareService interface {
	PrepareUpload(ctx context.Context, userID string, req *models.UploadPrepareRequest) (*models.UploadPrepareResponse, error)
}

type ListingHandler struct {
	listingService listingService
	uploadService  uploadPrepareService
	presenter      *services.ListingMediaPresenter
}

func NewListingHandler(listingService listingService, uploadService uploadPrepareService, presenter *services.ListingMediaPresenter) *ListingHandler {
	return &ListingHandler{
		listingService: listingService,
		uploadService:  uploadService,
		presenter:      presenter,
	}
}

func (h *ListingHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	method := req.HTTPMethod

	switch {
	case method == http.MethodOptions:
		return jsonResponse(http.StatusOK, ""), nil
	case method == http.MethodPost && path == "/listings":
		return h.createListing(ctx, req)
	case method == http.MethodPost && path == "/listings/upload-prepare":
		return h.prepareUpload(ctx, req)
	case method == http.MethodGet && path == "/listings":
		return h.listListings(ctx, req)
	case method == http.MethodGet && path == "/users/me/listings":
		return h.listMyListings(ctx, req)
	case method == http.MethodGet && path == "/users/me/saved":
		return h.listSavedListings(ctx, req)
	case method == http.MethodGet && isOwnerListingPath(path):
		return h.getOwnerListing(ctx, req)
	case method == http.MethodGet && isListingPath(path):
		return h.getListing(ctx, req)
	case method == http.MethodPost && isListingViewPath(path):
		return h.recordListingView(ctx, req)
	case method == http.MethodPost && isListingContactRevealPath(path):
		return h.revealListingContact(ctx, req)
	case method == http.MethodPut && isListingPath(path):
		return h.updateListing(ctx, req)
	case method == http.MethodDelete && isListingPath(path):
		return h.deleteListing(ctx, req)
	case method == http.MethodPost && isListingSavePath(path):
		return h.saveListing(ctx, req)
	case method == http.MethodDelete && isListingSavePath(path):
		return h.unsaveListing(ctx, req)
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
		return appErrorResponse(err), nil
	}
	return h.ownerListingResponse(ctx, http.StatusCreated, listing)
}

func (h *ListingHandler) prepareUpload(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	if h.uploadService == nil {
		return jsonResponse(http.StatusServiceUnavailable, utils.MarshalErrorResponse(utils.NewAppError(503, "upload prepare unavailable"))), nil
	}

	var uploadReq models.UploadPrepareRequest
	if err := json.Unmarshal([]byte(req.Body), &uploadReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	resp, err := h.uploadService.PrepareUpload(ctx, userID, &uploadReq)
	if err != nil {
		return appErrorResponse(err), nil
	}

	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) listListings(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	params := parseSearchParams(req)
	listings, err := h.listingService.ListListings(ctx, params)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	respListings, err := h.presenter.PresentSummaryCollection(ctx, listings)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(map[string]any{
		"listings": respListings,
		"total":    len(respListings),
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
		return appErrorResponse(err), nil
	}

	resp, err := h.presenter.PresentPublicDetail(ctx, listing)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) getOwnerListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	listing, err := h.listingService.GetOwnerListing(ctx, userID, listingID)
	if err != nil {
		return appErrorResponse(err), nil
	}
	return h.ownerListingResponse(ctx, http.StatusOK, listing)
}

func (h *ListingHandler) recordListingView(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	listing, err := h.listingService.RecordListingView(ctx, listingID)
	if err != nil {
		return appErrorResponse(err), nil
	}

	resp, err := h.presenter.PresentPublicSummary(ctx, listing)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) revealListingContact(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	var revealReq models.RevealListingContactRequest
	if err := json.Unmarshal([]byte(req.Body), &revealReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	contact, err := h.listingService.RevealListingContact(ctx, userID, listingID, revealReq.Channel)
	if err != nil {
		return appErrorResponse(err), nil
	}

	body, _ := json.Marshal(contact)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) listMyListings(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	params := parseSearchParams(req)
	listings, err := h.listingService.ListMyListings(ctx, userID, params)
	if err != nil {
		return appErrorResponse(err), nil
	}

	respListings, err := h.presenter.PresentSummaryCollection(ctx, listings)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(map[string]any{
		"listings": respListings,
		"total":    len(respListings),
		"page":     params.Page,
		"pageSize": params.PageSize,
	})
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *ListingHandler) listSavedListings(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	params := parseSearchParams(req)
	listings, err := h.listingService.ListSavedListings(ctx, userID, params)
	if err != nil {
		return appErrorResponse(err), nil
	}

	respListings, err := h.presenter.PresentSummaryCollection(ctx, listings)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(map[string]any{
		"listings": respListings,
		"total":    len(respListings),
		"page":     params.Page,
		"pageSize": params.PageSize,
	})
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
		return appErrorResponse(err), nil
	}
	return h.ownerListingResponse(ctx, http.StatusOK, listing)
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
		return appErrorResponse(err), nil
	}
	return jsonResponse(http.StatusNoContent, ""), nil
}

func (h *ListingHandler) saveListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	if err := h.listingService.SaveListing(ctx, userID, listingID); err != nil {
		return appErrorResponse(err), nil
	}
	return jsonResponse(http.StatusNoContent, ""), nil
}

func (h *ListingHandler) unsaveListing(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	listingID := extractListingID(req)
	if listingID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	if err := h.listingService.UnsaveListing(ctx, userID, listingID); err != nil {
		return appErrorResponse(err), nil
	}
	return jsonResponse(http.StatusNoContent, ""), nil
}

func (h *ListingHandler) parseText(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
		return appErrorResponse(err), nil
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
		return appErrorResponse(err), nil
	}

	resp, _ := json.Marshal(map[string]string{"uploadUrl": uploadURL, "key": key})
	return jsonResponse(http.StatusOK, string(resp)), nil
}

func (h *ListingHandler) ownerListingResponse(ctx context.Context, statusCode int, listing *models.Listing) (events.APIGatewayProxyResponse, error) {
	resp, err := h.presenter.PresentOwnerDetail(ctx, listing)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}
	body, _ := json.Marshal(resp)
	return jsonResponse(statusCode, string(body)), nil
}

func appErrorResponse(err error) events.APIGatewayProxyResponse {
	if appErr, ok := err.(*utils.AppError); ok {
		return jsonResponse(appErr.Code, utils.MarshalErrorResponse(appErr))
	}
	return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal))
}

func isListingPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 && parts[0] == "listings" && parts[1] != "parse-text" && parts[1] != "upload-prepare"
}

func isOwnerListingPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 4 && parts[0] == "users" && parts[1] == "me" && parts[2] == "listings" && parts[3] != ""
}

func isListingSavePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "listings" && parts[1] != "" && parts[2] == "save"
}

func isListingViewPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "listings" && parts[1] != "" && parts[2] == "view"
}

func isListingContactRevealPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "listings" && parts[1] != "" && parts[2] == "contact-reveal"
}

func extractListingID(req events.APIGatewayProxyRequest) string {
	if id, ok := req.PathParameters["id"]; ok {
		return id
	}
	parts := strings.Split(strings.Trim(req.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "listings" {
		return parts[1]
	}
	if len(parts) >= 4 && parts[0] == "users" && parts[2] == "listings" {
		return parts[3]
	}
	return ""
}

func parseSearchParams(req events.APIGatewayProxyRequest) *models.ListingSearchParams {
	p := &models.ListingSearchParams{
		Query:       req.QueryStringParameters["q"],
		Province:    req.QueryStringParameters["province"],
		City:        req.QueryStringParameters["city"],
		ListingType: models.ListingType(req.QueryStringParameters["listingType"]),
		LegalStatus: req.QueryStringParameters["legalStatus"],
		SortBy:      req.QueryStringParameters["sortBy"],
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
	if v, err := strconv.Atoi(req.QueryStringParameters["bathrooms"]); err == nil {
		p.Bathrooms = v
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["buildingAreaMin"], 64); err == nil {
		p.BuildingAreaMin = v
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["buildingAreaMax"], 64); err == nil {
		p.BuildingAreaMax = v
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["landAreaMin"], 64); err == nil {
		p.LandAreaMin = v
	}
	if v, err := strconv.ParseFloat(req.QueryStringParameters["landAreaMax"], 64); err == nil {
		p.LandAreaMax = v
	}
	if v, err := strconv.Atoi(req.QueryStringParameters["page"]); err == nil {
		p.Page = v
	}
	if v, err := strconv.Atoi(req.QueryStringParameters["pageSize"]); err == nil {
		p.PageSize = v
	}
	if amenities := strings.TrimSpace(req.QueryStringParameters["amenities"]); amenities != "" {
		p.Amenities = strings.Split(amenities, ",")
	}
	return p
}
