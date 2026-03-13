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

const (
	freeTierMaxListingsPerMonth = 1
	freeTierMaxMedia            = 3
	premiumTierMaxMedia         = 30

	minLocationConfidence = 0.7
)

// ListingStore is the storage interface for listing persistence.
type ListingStore interface {
	Put(ctx context.Context, listing *models.Listing) error
	GetByID(ctx context.Context, userID, listingID string) (*models.Listing, error)
	GetByListingID(ctx context.Context, listingID string) (*models.Listing, error)
	ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Listing, error)
	CountMonthlyByUserID(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, userID, listingID string) error
	Scan(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error)
}

// UserStore is the storage interface for user persistence.
type UserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
}

// AIParseService is the interface for AI-based listing text parsing.
type AIParseService interface {
	ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error)
}

type aiContentModerator interface {
	ModerateContent(ctx context.Context, title, description string) (approved bool, reason string, flags []string, err error)
}

// LocationNormalizer validates and normalizes AI-suggested location data against a catalog.
type LocationNormalizer interface {
	NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion
}

// ListingService orchestrates listing lifecycle operations.
type ListingService struct {
	listingRepo     ListingStore
	userRepo        UserStore
	aiService       AIParseService
	s3Service       *S3Service
	mapsService     *GoogleMapsService
	locationCatalog LocationNormalizer
}

// NewListingService creates a fully-wired ListingService.
func NewListingService(
	listingRepo ListingStore,
	userRepo UserStore,
	aiService AIParseService,
	s3Service *S3Service,
	mapsService *GoogleMapsService,
	locationCatalog LocationNormalizer,
) *ListingService {
	return &ListingService{
		listingRepo:     listingRepo,
		userRepo:        userRepo,
		aiService:       aiService,
		s3Service:       s3Service,
		mapsService:     mapsService,
		locationCatalog: locationCatalog,
	}
}

// ensure concrete repository types satisfy the service interfaces.
var _ ListingStore = (*repository.ListingRepo)(nil)
var _ UserStore = (*repository.UserRepo)(nil)

// CreateListing validates limits and persists a new listing.
func (s *ListingService) CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error) {
	if err := utils.ValidateCreateListingRequest(req); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, utils.ErrUnauthorized
	}

	isPremium := user.Subscription.Tier == models.SubscriptionPremium

	// Enforce free-tier listing limit.
	if !isPremium {
		count, err := s.listingRepo.CountMonthlyByUserID(ctx, userID)
		if err != nil {
			utils.LogError("count monthly listings", err, "userId", userID)
			return nil, utils.ErrInternal
		}
		if count >= freeTierMaxListingsPerMonth {
			return nil, utils.NewAppError(403, fmt.Sprintf("free tier allows at most %d listing(s) per month", freeTierMaxListingsPerMonth))
		}
	}

	// Enforce free-tier media limit.
	if err := utils.ValidateMediaLimits(isPremium, req.Images, req.Videos); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	listingID := uuid.NewString()
	now := time.Now().UTC()

	listing := &models.Listing{
		PK:               fmt.Sprintf("%s#%s", userID, listingID),
		SK:               listingID,
		ListingID:        listingID,
		UserID:           userID,
		Title:            req.Title,
		Description:      req.Description,
		Price:            req.Price,
		PriceUnit:        req.PriceUnit,
		ListingType:      req.ListingType,
		Status:           models.ListingStatusActive,
		PropertyDetails:  req.PropertyDetails,
		Location:         req.Location,
		Images:           req.Images,
		Videos:           req.Videos,
		ImageCount:       len(req.Images),
		PremiumFeatures:  models.PremiumFeatures{IsPremium: isPremium},
		ModerationStatus: models.ModerationStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.listingRepo.Put(ctx, listing); err != nil {
		utils.LogError("put listing", err, "listingId", listingID)
		return nil, utils.ErrInternal
	}

	s.applyContentModeration(ctx, listing)
	return listing, nil
}

// UpdateListing applies partial updates to an existing listing.
func (s *ListingService) UpdateListing(ctx context.Context, userID, listingID string, req *models.UpdateListingRequest) (*models.Listing, error) {
	listing, err := s.listingRepo.GetByID(ctx, userID, listingID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if listing == nil {
		return nil, utils.ErrNotFound
	}
	if listing.UserID != userID {
		return nil, utils.ErrForbidden
	}

	user, _ := s.userRepo.GetByID(ctx, userID)
	isPremium := user != nil && user.Subscription.Tier == models.SubscriptionPremium

	if req.Title != nil {
		listing.Title = *req.Title
	}
	if req.Description != nil {
		listing.Description = *req.Description
	}
	if req.Price != nil {
		listing.Price = *req.Price
	}
	if req.PriceUnit != nil {
		listing.PriceUnit = *req.PriceUnit
	}
	if req.Status != nil {
		listing.Status = *req.Status
	}
	if req.PropertyDetails != nil {
		listing.PropertyDetails = *req.PropertyDetails
	}
	if req.Location != nil {
		listing.Location = *req.Location
	}
	if req.Images != nil || req.Videos != nil {
		newImages := req.Images
		if newImages == nil {
			newImages = listing.Images
		}
		newVideos := req.Videos
		if newVideos == nil {
			newVideos = listing.Videos
		}
		if err := utils.ValidateMediaLimits(isPremium, newImages, newVideos); err != nil {
			return nil, utils.WrapError(utils.ErrBadRequest, err)
		}
		if req.Images != nil {
			listing.Images = req.Images
			listing.ImageCount = len(req.Images)
		}
		if req.Videos != nil {
			listing.Videos = req.Videos
		}
	}

	listing.UpdatedAt = time.Now().UTC()
	// Re-queue for moderation when content changes.
	listing.ModerationStatus = models.ModerationStatusPending

	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return nil, utils.ErrInternal
	}

	s.applyContentModeration(ctx, listing)
	return listing, nil
}

// GetListing fetches a listing by ID. It increments the view counter.
func (s *ListingService) GetListing(ctx context.Context, listingID string) (*models.Listing, error) {
	listing, err := s.getCurrentListingByListingID(ctx, listingID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if listing == nil {
		return nil, utils.ErrNotFound
	}
	if listing.Status != models.ListingStatusActive || listing.ModerationStatus != models.ModerationStatusApproved {
		return nil, utils.ErrNotFound
	}

	// Best-effort view increment.
	listing.Views++
	listing.UpdatedAt = time.Now().UTC()
	_ = s.listingRepo.Put(ctx, listing)

	return listing, nil
}

func (s *ListingService) getCurrentListingByListingID(ctx context.Context, listingID string) (*models.Listing, error) {
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil || listing == nil {
		return listing, err
	}

	current, err := s.listingRepo.GetByID(ctx, listing.UserID, listing.ListingID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	return current, nil
}

func (s *ListingService) applyContentModeration(ctx context.Context, listing *models.Listing) {
	if s.aiService == nil || listing == nil {
		return
	}

	moderator, ok := s.aiService.(aiContentModerator)
	if !ok {
		return
	}

	approved, reason, flags, err := moderator.ModerateContent(ctx, listing.Title, listing.Description)
	if err != nil {
		utils.LogError("moderate listing content", err, "listingId", listing.ListingID)
		return
	}

	if approved {
		listing.ModerationStatus = models.ModerationStatusApproved
		listing.Status = models.ListingStatusActive
	} else {
		listing.ModerationStatus = models.ModerationStatusRejected
		listing.Status = models.ListingStatusArchived
	}
	listing.ModerationReason = joinModerationReason(reason, flags)
	listing.UpdatedAt = time.Now().UTC()

	if err := s.listingRepo.Put(ctx, listing); err != nil {
		utils.LogError("persist moderated listing", err, "listingId", listing.ListingID)
	}
}

func joinModerationReason(reason string, flags []string) string {
	flagsText := ""
	for i, flag := range flags {
		if i > 0 {
			flagsText += "; "
		}
		flagsText += flag
	}

	if reason != "" && flagsText != "" {
		return reason + " | flags: " + flagsText
	}
	if flagsText != "" {
		return "flags: " + flagsText
	}
	return reason
}

// ListListings returns paginated listings with optional filters.
func (s *ListingService) ListListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	params.Page, params.PageSize = utils.ValidatePagination(params.Page, params.PageSize)
	listings, err := s.listingRepo.Scan(ctx, params)
	if err != nil {
		utils.LogError("scan listings", err)
		return nil, utils.ErrInternal
	}
	return listings, nil
}

// SearchListings delegates to ListListings (uses the same DynamoDB scan with filters).
func (s *ListingService) SearchListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	return s.ListListings(ctx, params)
}

// ListMyListings returns the authenticated user's listings.
func (s *ListingService) ListMyListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error) {
	params.Page, params.PageSize = utils.ValidatePagination(params.Page, params.PageSize)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	listings, err := s.listingRepo.ListByUserID(ctx, userID, int32(params.PageSize))
	if err != nil {
		utils.LogError("list user listings", err, "userId", userID)
		return nil, utils.ErrInternal
	}

	filtered := make([]models.Listing, 0, len(listings))
	for _, listing := range listings {
		current, err := s.listingRepo.GetByID(ctx, userID, listing.ListingID)
		if err != nil {
			utils.LogError("get current listing after user index query", err, "userId", userID, "listingId", listing.ListingID)
			return nil, utils.ErrInternal
		}
		if current == nil {
			continue
		}
		if current.ModerationStatus == models.ModerationStatusRejected {
			continue
		}
		filtered = append(filtered, *current)
	}

	return filtered, nil
}

// DeleteListing removes a listing after ownership verification.
func (s *ListingService) DeleteListing(ctx context.Context, userID, listingID string) error {
	listing, err := s.listingRepo.GetByID(ctx, userID, listingID)
	if err != nil {
		return utils.ErrInternal
	}
	if listing == nil {
		return utils.ErrNotFound
	}
	if listing.UserID != userID {
		return utils.ErrForbidden
	}

	if err := s.listingRepo.Delete(ctx, userID, listingID); err != nil {
		return utils.ErrInternal
	}

	// Best-effort cleanup of associated S3 objects.
	if s.s3Service != nil {
		keys, _ := s.s3Service.ListObjects(ctx, fmt.Sprintf("listings/%s/", listingID))
		for _, k := range keys {
			_ = s.s3Service.DeleteObject(ctx, k)
		}
	}
	return nil
}

// GetUploadURL generates a presigned S3 PUT URL for listing media after verifying ownership.
func (s *ListingService) GetUploadURL(ctx context.Context, userID, listingID, filename, contentType string) (string, string, error) {
	listing, err := s.getCurrentListingByListingID(ctx, listingID)
	if err != nil {
		return "", "", utils.ErrInternal
	}
	if listing == nil {
		return "", "", utils.ErrNotFound
	}
	if listing.UserID != userID {
		return "", "", utils.ErrForbidden
	}

	key := listingMediaKey(listingID, filename)
	uploadURL, err := s.s3Service.GetPresignedUploadURL(ctx, key, contentType)
	if err != nil {
		return "", "", utils.ErrInternal
	}
	return uploadURL, key, nil
}

// listingMediaKey builds the S3 key for listing media.
func listingMediaKey(listingID, filename string) string {
	return fmt.Sprintf("listings/%s/%s", listingID, filename)
}

// FeatureListing marks a listing as featured/promoted until the given expiry.
func (s *ListingService) FeatureListing(ctx context.Context, userID, listingID, featureType string, until time.Time) error {
	listing, err := s.listingRepo.GetByID(ctx, userID, listingID)
	if err != nil {
		return utils.ErrInternal
	}
	if listing == nil {
		return utils.ErrNotFound
	}
	if listing.UserID != userID {
		return utils.ErrForbidden
	}

	switch featureType {
	case "featured":
		listing.PremiumFeatures.IsFeatured = true
		listing.PremiumFeatures.FeaturedUntil = &until
	case "promotion":
		listing.PremiumFeatures.PromotionUntil = &until
	default:
		return utils.NewAppError(400, "unknown feature type")
	}

	listing.UpdatedAt = time.Now().UTC()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return utils.ErrInternal
	}
	return nil
}

// ParseListingText delegates to the AI service and enriches the result with
// validated location suggestions from the catalog.
func (s *ListingService) ParseListingText(ctx context.Context, text string) (*models.ParseTextResponse, error) {
	if s.aiService == nil {
		return nil, utils.NewAppError(503, "AI service unavailable")
	}

	parsed, err := s.aiService.ParseListingText(ctx, text)
	if err != nil {
		utils.LogError("parse listing text", err)
		return nil, utils.ErrInternal
	}

	if s.locationCatalog != nil {
		suggestion := s.locationCatalog.NormalizeSuggestion(
			parsed.LocationSuggestion.Province,
			parsed.LocationSuggestion.City,
			parsed.LocationSuggestion.District,
		)
		// NormalizedAddress from the AI is intentionally preserved; catalog
		// normalization only updates the structured Province/City/District fields.
		suggestion.NormalizedAddress = parsed.LocationSuggestion.NormalizedAddress
		parsed.LocationSuggestion = suggestion
	}

	if parsed.LocationSuggestion.Confidence < minLocationConfidence {
		parsed.RequiresManualReview = true
	}

	// Confidence mirrors the AI's overall parse confidence; RequiresCorrection
	// can additionally be driven by low location-validation confidence
	// independently of the AI score.
	return &models.ParseTextResponse{
		Parsed:             *parsed,
		RequiresCorrection: parsed.RequiresManualReview,
		Confidence:         parsed.Confidence,
	}, nil
}
