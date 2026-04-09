package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type leadService interface {
	CreateLead(ctx context.Context, ownerUserID string, req *models.CreateLeadRequest) (*models.Lead, error)
	ListLeads(ctx context.Context, ownerUserID string, stage string, dueOnly bool) ([]models.Lead, error)
	ListLeadsPaged(ctx context.Context, ownerUserID string, stage string, dueOnly bool, limit int32, cursor string) (*models.LeadListResponse, error)
	GetLead(ctx context.Context, ownerUserID, leadID string) (*models.Lead, error)
	UpdateLeadStage(ctx context.Context, ownerUserID, leadID string, req *models.UpdateLeadStageRequest) (*models.Lead, error)
	AddLeadNote(ctx context.Context, ownerUserID, leadID string, req *models.AddLeadNoteRequest) (*models.Lead, error)
	CompleteFollowUpTask(ctx context.Context, ownerUserID, leadID, taskID string, req *models.CompleteFollowUpTaskRequest) (*models.Lead, error)
	Analytics(ctx context.Context, ownerUserID string) (*models.AgentAnalyticsResponse, error)
}

type LeadHandler struct {
	leadService leadService
}

func NewLeadHandler(leadService leadService) *LeadHandler {
	return &LeadHandler{leadService: leadService}
}

func (h *LeadHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	method := req.HTTPMethod

	switch {
	case method == http.MethodOptions:
		return jsonResponse(http.StatusOK, ""), nil
	case method == http.MethodGet && path == "/leads/analytics":
		return h.getAnalytics(ctx, req)
	case method == http.MethodPost && path == "/leads":
		return h.createLead(ctx, req)
	case method == http.MethodGet && path == "/leads":
		return h.listLeads(ctx, req)
	case method == http.MethodGet && isLeadByIDPath(path):
		return h.getLead(ctx, req)
	case method == http.MethodPost && isLeadStagePath(path):
		return h.updateLeadStage(ctx, req)
	case method == http.MethodPost && isLeadNotePath(path):
		return h.addLeadNote(ctx, req)
	case method == http.MethodPost && isLeadFollowUpPath(path):
		return h.completeFollowUpTask(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

func (h *LeadHandler) createLead(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	var payload models.CreateLeadRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	lead, err := h.leadService.CreateLead(ctx, userID, &payload)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(lead)
	return jsonResponse(http.StatusCreated, string(body)), nil
}

func (h *LeadHandler) listLeads(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	stage := strings.TrimSpace(req.QueryStringParameters["stage"])
	dueOnly := strings.EqualFold(strings.TrimSpace(req.QueryStringParameters["dueOnly"]), "true")
	cursor := strings.TrimSpace(req.QueryStringParameters["cursor"])
	var limit int32 = 50
	if l := strings.TrimSpace(req.QueryStringParameters["limit"]); l != "" {
		if parsed, err := parseLimit(l); err == nil {
			limit = parsed
		}
	}
	result, err := h.leadService.ListLeadsPaged(ctx, userID, stage, dueOnly, limit, cursor)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(result)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *LeadHandler) getLead(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	leadID := extractLeadID(req)
	if leadID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	lead, err := h.leadService.GetLead(ctx, userID, leadID)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(lead)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *LeadHandler) updateLeadStage(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	leadID := extractLeadID(req)
	if leadID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	var payload models.UpdateLeadStageRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	lead, err := h.leadService.UpdateLeadStage(ctx, userID, leadID, &payload)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(lead)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *LeadHandler) addLeadNote(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	leadID := extractLeadID(req)
	if leadID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	var payload models.AddLeadNoteRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	lead, err := h.leadService.AddLeadNote(ctx, userID, leadID, &payload)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(lead)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *LeadHandler) completeFollowUpTask(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	leadID := extractLeadID(req)
	taskID := extractTaskID(req)
	if leadID == "" || taskID == "" {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	var payload models.CompleteFollowUpTaskRequest
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	lead, err := h.leadService.CompleteFollowUpTask(ctx, userID, leadID, taskID, &payload)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(lead)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func (h *LeadHandler) getAnalytics(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}
	analytics, err := h.leadService.Analytics(ctx, userID)
	if err != nil {
		return appErrorResponse(err), nil
	}
	body, _ := json.Marshal(analytics)
	return jsonResponse(http.StatusOK, string(body)), nil
}

func isLeadByIDPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 && parts[0] == "leads" && parts[1] != ""
}

func isLeadStagePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "leads" && parts[1] != "" && parts[2] == "stage"
}

func isLeadNotePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 3 && parts[0] == "leads" && parts[1] != "" && parts[2] == "notes"
}

func isLeadFollowUpPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 5 && parts[0] == "leads" && parts[1] != "" && parts[2] == "followups" && parts[3] != "" && parts[4] == "complete"
}

func extractLeadID(req events.APIGatewayProxyRequest) string {
	if id, ok := req.PathParameters["id"]; ok && id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(req.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "leads" {
		return parts[1]
	}
	return ""
}

func extractTaskID(req events.APIGatewayProxyRequest) string {
	if taskID, ok := req.PathParameters["taskId"]; ok && taskID != "" {
		return taskID
	}
	parts := strings.Split(strings.Trim(req.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "leads" && parts[2] == "followups" {
		return parts[3]
	}
	return ""
}

func parseLimit(s string) (int32, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if v <= 0 {
		v = 50
	}
	if v > 200 {
		v = 200
	}
	// Safe: v is bounded to [1, 200] which fits in int32.
	return int32(v), nil //nolint:gosec
}
