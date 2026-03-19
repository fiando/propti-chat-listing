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
	ModerateImages(ctx context.Context, images [][]byte) (approved bool, reason string, flags []string, err error)
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
	mediaStore     MediaStorage
}

func NewModerationService(
	textModerator ContentModerator,
	imageModerator ImageModerator,
	moderationRepo ModerationRecordStore,
	listingRepo ListingStore,
	mediaStore MediaStorage,
) *ModerationService {
	return &ModerationService{
		textModerator:  textModerator,
		imageModerator: imageModerator,
		moderationRepo: moderationRepo,
		listingRepo:    listingRepo,
		mediaStore:     mediaStore,
	}
}

// ModerateListing runs content checks on a listing.
// checkText controls whether the title/description is re-evaluated.
// newImageIDs, when non-nil, limits image moderation to only those IDs; nil means check all images.
func (s *ModerationService) ModerateListing(ctx context.Context, listingID string, checkText bool, newImageIDs []string) (*models.Listing, error) {
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

	if checkText {
		if err := s.moderateText(ctx, current, &rejectionReasons); err != nil {
			return nil, err
		}
	}
	if err := s.moderateImages(ctx, current, newImageIDs, &rejectionReasons); err != nil {
		return nil, err
	}

	current.UpdatedAt = time.Now().UTC()
	if len(rejectionReasons) > 0 {
		if err := s.moveRejectedMedia(ctx, current); err != nil {
			return nil, err
		}
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

// moderateImages checks images for inappropriate or non-property content.
// imageIDsFilter, when non-nil, restricts checks to only images with those IDs; nil checks all images.
func (s *ModerationService) moderateImages(ctx context.Context, listing *models.Listing, imageIDsFilter []string, rejectionReasons *[]string) error {
	var imagesToCheck models.ImageEntries
	if imageIDsFilter == nil {
		imagesToCheck = listing.Images
	} else {
		filterSet := make(map[string]struct{}, len(imageIDsFilter))
		for _, id := range imageIDsFilter {
			filterSet[id] = struct{}{}
		}
		for _, img := range listing.Images {
			if _, ok := filterSet[img.ImageID]; ok {
				imagesToCheck = append(imagesToCheck, img)
			}
		}
	}

	if len(imagesToCheck) == 0 {
		return nil
	}
	if s.imageModerator == nil {
		return fmt.Errorf("image moderator not configured")
	}

	images, err := s.loadListingImages(ctx, imagesToCheck)
	if err != nil {
		return fmt.Errorf("load listing images: %w", err)
	}
	if len(images) == 0 {
		return nil
	}

	approved, reason, flags, err := s.imageModerator.ModerateImages(ctx, images)
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

func (s *ModerationService) loadListingImages(ctx context.Context, images models.ImageEntries) ([][]byte, error) {
	result := make([][]byte, 0, len(images))
	for _, image := range images {
		switch {
		case image.S3Key != "":
			if s.mediaStore == nil {
				return nil, fmt.Errorf("media store not configured")
			}
			body, err := s.mediaStore.GetObjectBytes(ctx, image.S3Key)
			if err != nil {
				return nil, err
			}
			result = append(result, body)
		case image.LegacyValue != "":
			body, err := decodeListingImage(image.LegacyValue)
			if err != nil {
				return nil, err
			}
			result = append(result, body)
		}
	}
	return result, nil
}

func (s *ModerationService) moveRejectedMedia(ctx context.Context, listing *models.Listing) error {
	if s.mediaStore == nil {
		return nil
	}

	for index := range listing.Images {
		image := &listing.Images[index]
		if image.IsLegacy() || image.S3Key == "" {
			continue
		}

		rejectedKey := BuildRejectedKey(listing.ListingID, image.ImageID)
		if err := s.mediaStore.CopyObject(ctx, image.S3Key, rejectedKey); err != nil {
			return err
		}
		if err := s.mediaStore.DeleteObject(ctx, image.S3Key); err != nil {
			return err
		}
		if image.ThumbnailKey != "" {
			rejectedThumbnailKey := BuildRejectedKey(listing.ListingID, image.ImageID+"-thumbnail")
			if err := s.mediaStore.CopyObject(ctx, image.ThumbnailKey, rejectedThumbnailKey); err != nil {
				return err
			}
			if err := s.mediaStore.DeleteObject(ctx, image.ThumbnailKey); err != nil {
				return err
			}
		}

		image.S3Key = rejectedKey
		image.ThumbnailKey = ""
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
