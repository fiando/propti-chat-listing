package services

import (
	"context"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type UploadSessionStore interface {
	Put(ctx context.Context, session *models.UploadSession) error
	GetBySessionID(ctx context.Context, sessionID string) (*models.UploadSession, error)
	Consume(ctx context.Context, sessionID, listingID string, consumedAt time.Time) error
}

type UploadSessionService struct {
	sessionStore UploadSessionStore
	userStore    UserStore
	listingStore ListingStore
	mediaStore   MediaStorage
	idGenerator  func() string
	now          func() time.Time
}

func NewUploadSessionService(sessionStore UploadSessionStore, userStore UserStore, listingStore ListingStore, mediaStore MediaStorage) *UploadSessionService {
	return &UploadSessionService{
		sessionStore: sessionStore,
		userStore:    userStore,
		listingStore: listingStore,
		mediaStore:   mediaStore,
		idGenerator:  uuid.NewString,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *UploadSessionService) PrepareUpload(ctx context.Context, userID string, req *models.UploadPrepareRequest) (*models.UploadPrepareResponse, error) {
	if req.FinalImageCount != req.RetainedImageCount+len(req.NewImages) {
		return nil, utils.NewAppError(400, "finalImageCount must equal retainedImageCount + len(newImages)")
	}

	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	effectiveTier := effectiveTierForUser(user, s.now())
	maxImages := TierEntitlementFor(effectiveTier).PhotoCapPerListing
	if req.FinalImageCount > maxImages {
		return nil, utils.NewAppError(400, fmt.Sprintf("subscription tier allows at most %d images per listing", maxImages))
	}

	if req.ListingID != "" {
		listing, err := s.listingStore.GetByListingID(ctx, req.ListingID)
		if err != nil {
			return nil, utils.ErrInternal
		}
		if listing == nil {
			return nil, utils.ErrNotFound
		}
		if listing.UserID != userID {
			return nil, utils.ErrForbidden
		}
	}

	now := s.now()
	response := &models.UploadPrepareResponse{
		Slots: make([]models.UploadSlot, 0, len(req.NewImages)),
	}

	for _, image := range req.NewImages {
		if !strings.HasPrefix(image.ContentType, "image/") {
			return nil, utils.NewAppError(400, "newImages contentType must be image/*")
		}
		if image.SizeBytes <= 0 {
			return nil, utils.NewAppError(400, "newImages sizeBytes must be greater than 0")
		}

		sessionID := s.idGenerator()
		filename := sessionFilenameFromContentType(image.ContentType)
		stagingKey := BuildStagingKey(userID, sessionID, filename)
		presignedURL, err := s.mediaStore.GetPresignedUploadURL(ctx, stagingKey, image.ContentType)
		if err != nil {
			return nil, utils.ErrInternal
		}

		session := &models.UploadSession{
			SessionID:           sessionID,
			UserID:              userID,
			ListingID:           req.ListingID,
			StagingKey:          stagingKey,
			ExpectedContentType: image.ContentType,
			ExpectedMaxSize:     image.SizeBytes,
			ExpiresAt:           now.Add(15 * time.Minute),
			CreatedAt:           now,
		}
		if err := s.sessionStore.Put(ctx, session); err != nil {
			return nil, utils.ErrInternal
		}

		response.Slots = append(response.Slots, models.UploadSlot{
			SessionID:    sessionID,
			PresignedURL: presignedURL,
			StagingKey:   stagingKey,
			ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func sessionFilenameFromContentType(contentType string) string {
	extensions, _ := mime.ExtensionsByType(contentType)
	if len(extensions) > 0 {
		return "upload" + extensions[0]
	}
	return "upload.bin"
}
