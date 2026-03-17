package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type ImageModerator interface {
	ModerateImages(ctx context.Context, images []string) (approved bool, reason string, flags []string, err error)
}

type ModerationRecordStore interface {
	Put(ctx context.Context, moderation *models.Moderation) error
}

// ModerationService runs asynchronous moderation checks and persists the results.
type ModerationService struct {
	textModerator  ContentModerator
	imageModerator ImageModerator
	moderationRepo ModerationRecordStore
	listingRepo    ListingStore
}

func NewModerationService(
	textModerator ContentModerator,
	imageModerator ImageModerator,
	moderationRepo ModerationRecordStore,
	listingRepo ListingStore,
) *ModerationService {
	return &ModerationService{
		textModerator:  textModerator,
		imageModerator: imageModerator,
		moderationRepo: moderationRepo,
		listingRepo:    listingRepo,
	}
}

func (s *ModerationService) ModerateListing(ctx context.Context, listingID string) (*models.Listing, error) {
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("get listing by listing id: %w", err)
	}
	if listing == nil {
		return nil, utils.ErrNotFound
	}

	current, err := s.listingRepo.GetByID(ctx, listing.UserID, listing.ListingID)
	if err != nil {
		return nil, fmt.Errorf("get current listing: %w", err)
	}
	if current == nil {
		return nil, utils.ErrNotFound
	}

	var rejectionReasons []string

	if err := s.moderateText(ctx, current, &rejectionReasons); err != nil {
		return nil, err
	}
	if err := s.moderateImages(ctx, current, &rejectionReasons); err != nil {
		return nil, err
	}

	current.UpdatedAt = time.Now().UTC()
	if len(rejectionReasons) > 0 {
		current.ModerationStatus = models.ModerationStatusRejected
		current.Status = models.ListingStatusArchived
		current.ModerationReason = strings.Join(rejectionReasons, " | ")
	} else {
		current.ModerationStatus = models.ModerationStatusApproved
		current.Status = models.ListingStatusActive
		current.ModerationReason = ""
	}

	if err := s.listingRepo.Put(ctx, current); err != nil {
		return nil, fmt.Errorf("persist moderated listing: %w", err)
	}

	return current, nil
}

func (s *ModerationService) moderateText(ctx context.Context, listing *models.Listing, rejectionReasons *[]string) error {
	if strings.TrimSpace(listing.Title) == "" && strings.TrimSpace(listing.Description) == "" {
		return nil
	}
	if s.textModerator == nil {
		return fmt.Errorf("text moderator not configured")
	}

	approved, reason, flags, err := s.textModerator.ModerateContent(ctx, listing.Title, listing.Description)
	if err != nil {
		return fmt.Errorf("moderate listing text: %w", err)
	}

	joinedReason := joinModerationReason(reason, flags)
	if err := s.saveModerationRecord(ctx, listing, models.ModerationTypeContent, approved, joinedReason); err != nil {
		utils.LogError("save text moderation record", err, "listingId", listing.ListingID)
	}
	if !approved && joinedReason != "" {
		*rejectionReasons = append(*rejectionReasons, joinedReason)
	}

	return nil
}

func (s *ModerationService) moderateImages(ctx context.Context, listing *models.Listing, rejectionReasons *[]string) error {
	if len(listing.Images) == 0 {
		return nil
	}
	if s.imageModerator == nil {
		return fmt.Errorf("image moderator not configured")
	}

	approved, reason, flags, err := s.imageModerator.ModerateImages(ctx, listing.Images)
	if err != nil {
		return fmt.Errorf("moderate listing images: %w", err)
	}

	joinedReason := joinModerationReason(reason, flags)
	if err := s.saveModerationRecord(ctx, listing, models.ModerationTypeMedia, approved, joinedReason); err != nil {
		utils.LogError("save image moderation record", err, "listingId", listing.ListingID)
	}
	if !approved && joinedReason != "" {
		*rejectionReasons = append(*rejectionReasons, joinedReason)
	}

	return nil
}

func (s *ModerationService) saveModerationRecord(ctx context.Context, listing *models.Listing, moderationType models.ModerationType, approved bool, reason string) error {
	if s.moderationRepo == nil {
		return nil
	}

	result := models.ModerationResultApproved
	if !approved {
		result = models.ModerationResultRejected
	}

	now := time.Now().UTC()
	return s.moderationRepo.Put(ctx, &models.Moderation{
		PK:           uuid.NewString(),
		SK:           now.Format(time.RFC3339Nano),
		ModerationID: uuid.NewString(),
		ListingID:    listing.ListingID,
		UserID:       listing.UserID,
		Type:         moderationType,
		Result:       result,
		Reason:       reason,
		Moderator:    models.ModeratorAI,
		Timestamp:    now,
	})
}
