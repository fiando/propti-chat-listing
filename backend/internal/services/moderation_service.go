package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

// ModerationService runs AI-based moderation checks and persists the results.
type ModerationService struct {
	aiService      *AIService
	moderationRepo *repository.ModerationRepo
	listingRepo    *repository.ListingRepo
}

// NewModerationService creates a ModerationService.
func NewModerationService(
	aiService *AIService,
	moderationRepo *repository.ModerationRepo,
	listingRepo *repository.ListingRepo,
) *ModerationService {
	return &ModerationService{
		aiService:      aiService,
		moderationRepo: moderationRepo,
		listingRepo:    listingRepo,
	}
}

// ModerateListingContent runs an AI moderation check on a listing and persists the decision.
// It also updates the listing's moderationStatus accordingly.
func (s *ModerationService) ModerateListingContent(ctx context.Context, listing *models.Listing) (*models.Moderation, error) {
	approved, reason, flags, err := s.aiService.ModerateContent(ctx, listing.Title, listing.Description)
	if err != nil {
		utils.LogError("ai moderation failed", err, "listingId", listing.ListingID)
		// On AI error, default to pending so a human can review.
		return nil, fmt.Errorf("moderation ai call: %w", err)
	}

	result := models.ModerationResultApproved
	if !approved {
		result = models.ModerationResultRejected
	}

	flagsStr := ""
	for i, f := range flags {
		if i > 0 {
			flagsStr += "; "
		}
		flagsStr += f
	}
	if reason != "" && flagsStr != "" {
		reason = reason + " | flags: " + flagsStr
	} else if flagsStr != "" {
		reason = "flags: " + flagsStr
	}

	moderation := &models.Moderation{
		PK:           fmt.Sprintf("mod#%s", uuid.NewString()),
		SK:           time.Now().UTC().Format(time.RFC3339),
		ModerationID: uuid.NewString(),
		ListingID:    listing.ListingID,
		UserID:       listing.UserID,
		Type:         models.ModerationTypeContent,
		Result:       result,
		Reason:       reason,
		Moderator:    models.ModeratorAI,
		Timestamp:    time.Now().UTC(),
	}
	moderation.PK = moderation.ModerationID

	if err := s.moderationRepo.Put(ctx, moderation); err != nil {
		utils.LogError("save moderation record", err, "listingId", listing.ListingID)
	}

	// Update listing status based on moderation result.
	moderationStatus := models.ModerationStatusApproved
	if !approved {
		moderationStatus = models.ModerationStatusRejected
	}
	listing.ModerationStatus = moderationStatus
	listing.ModerationReason = reason
	listing.UpdatedAt = time.Now().UTC()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		utils.LogError("update listing moderation status", err, "listingId", listing.ListingID)
	}

	return moderation, nil
}
