package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

// AuthHandler handles Google OAuth login and user profile requests.
type AuthHandler struct {
	db       *repository.DynamoDB
	userRepo *repository.UserRepo
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(db *repository.DynamoDB) *AuthHandler {
	return &AuthHandler{
		db:       db,
		userRepo: repository.NewUserRepo(db),
	}
}

// Handle routes API Gateway proxy requests to the correct sub-handler.
func (h *AuthHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == http.MethodPost && req.Path == "/auth/google":
		return h.googleLogin(ctx, req)
	case req.HTTPMethod == http.MethodGet && req.Path == "/auth/user":
		return h.getUser(ctx, req)
	case req.HTTPMethod == http.MethodPut && req.Path == "/auth/user":
		return h.updateUser(ctx, req)
	default:
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}
}

// googleLogin verifies a Google ID token and upserts the user, returning a JWT.
func (h *AuthHandler) googleLogin(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var authReq models.GoogleAuthRequest
	if err := json.Unmarshal([]byte(req.Body), &authReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}
	if err := utils.ValidateGoogleAuthRequest(&authReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.NewAppError(400, err.Error()))), nil
	}

	// Verify and decode the Google ID token.
	googleClaims, err := verifyGoogleIDToken(ctx, authReq.IDToken)
	if err != nil {
		utils.LogError("verify google id token", err)
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	// Look up existing user by Google subject ID.
	user, err := h.userRepo.GetByGoogleID(ctx, googleClaims.Subject)
	if err != nil {
		utils.LogError("get user by googleId", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	now := time.Now().UTC()
	if user == nil {
		// New user — create account.
		userID := uuid.NewString()
		user = &models.User{
			PK:             userID,
			SK:             "metadata",
			UserID:         userID,
			GoogleID:       googleClaims.Subject,
			Email:          googleClaims.Email,
			Name:           googleClaims.Name,
			ProfilePicture: googleClaims.Picture,
			Role:           models.UserRoleBuyer,
			Preferences: models.UserPreferences{
				FavoriteLocations: []string{},
				SearchHistory:     []string{},
				Notifications:     true,
			},
			Subscription: models.Subscription{
				Tier:                models.SubscriptionFree,
				MonthlyListingsUsed: 0,
			},
			CreatedAt:   now,
			LastLoginAt: now,
		}
	} else {
		user.LastLoginAt = now
		user.Name = googleClaims.Name
		user.ProfilePicture = googleClaims.Picture
	}

	if err := h.userRepo.Put(ctx, user); err != nil {
		utils.LogError("upsert user", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	token, err := utils.GenerateToken(user.UserID, user.Email)
	if err != nil {
		utils.LogError("generate token", err)
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	resp := models.AuthResponse{AccessToken: token, User: user}
	body, _ := json.Marshal(resp)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// getUser returns the profile of the currently authenticated user.
func (h *AuthHandler) getUser(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	body, _ := json.Marshal(user)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// updateUser applies a partial update to the authenticated user's profile.
func (h *AuthHandler) updateUser(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := extractUserID(req)
	if err != nil {
		return jsonResponse(http.StatusUnauthorized, utils.MarshalErrorResponse(utils.ErrUnauthorized)), nil
	}

	var updateReq models.UpdateUserRequest
	if err := json.Unmarshal([]byte(req.Body), &updateReq); err != nil {
		return jsonResponse(http.StatusBadRequest, utils.MarshalErrorResponse(utils.ErrBadRequest)), nil
	}

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return jsonResponse(http.StatusNotFound, utils.MarshalErrorResponse(utils.ErrNotFound)), nil
	}

	if updateReq.Phone != nil {
		user.Phone = *updateReq.Phone
	}
	if updateReq.Role != nil {
		user.Role = *updateReq.Role
	}
	if updateReq.Preferences != nil {
		user.Preferences = *updateReq.Preferences
	}

	if err := h.userRepo.Put(ctx, user); err != nil {
		return jsonResponse(http.StatusInternalServerError, utils.MarshalErrorResponse(utils.ErrInternal)), nil
	}

	body, _ := json.Marshal(user)
	return jsonResponse(http.StatusOK, string(body)), nil
}

// --- helpers ---

// extractUserID reads the Bearer token from the Authorization header and returns the user ID.
func extractUserID(req events.APIGatewayProxyRequest) (string, error) {
	authHeader := req.Headers["Authorization"]
	if authHeader == "" {
		authHeader = req.Headers["authorization"]
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", utils.ErrUnauthorized
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return "", utils.ErrUnauthorized
	}
	return claims.UserID, nil
}

// jsonResponse builds an API Gateway response with CORS headers and the given JSON body.
func jsonResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		},
		Body: body,
	}
}
