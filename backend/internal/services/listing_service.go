package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

const (
	freeTierMaxListings            = 3
	premiumTierMaxListings         = 15
	freeTierListingDurationDays    = 30
	premiumTierListingDurationDays = 90

	minLocationConfidence = 0.7
	contactRevealLimit    = 10
)

var contactRevealWindow = 10 * time.Minute

type ListingStore interface {
	Put(ctx context.Context, listing *models.Listing) error
	GetByID(ctx context.Context, userID, listingID string) (*models.Listing, error)
	GetByListingID(ctx context.Context, listingID string) (*models.Listing, error)
	ListByUserID(ctx context.Context, userID string, limit int32) ([]models.Listing, error)
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, userID, listingID string) error
	Scan(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error)
}

type UserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Put(ctx context.Context, user *models.User) error
}

type AIParseService interface {
	ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error)
}

type ContentModerator interface {
	ModerateContent(ctx context.Context, title, description string) (approved bool, reason string, flags []string, err error)
}

type ModerationEnqueuer interface {
	EnqueueListingModeration(ctx context.Context, listingID string, checkText bool, newImageIDs []string) error
}

type LocationNormalizer interface {
	NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion
}

type ListingService struct {
	listingRepo       ListingStore
	userRepo          UserStore
	aiService         AIParseService
	mediaStore        MediaStorage
	mapsService       *GoogleMapsService
	locationCatalog   LocationNormalizer
	moderationQueue   ModerationEnqueuer
	uploadSessionRepo UploadSessionStore
	idGenerator       func() string
	now               func() time.Time
}

func NewListingService(
	listingRepo ListingStore,
	userRepo UserStore,
	aiService AIParseService,
	mediaStore MediaStorage,
	mapsService *GoogleMapsService,
	locationCatalog LocationNormalizer,
) *ListingService {
	return &ListingService{
		listingRepo:     listingRepo,
		userRepo:        userRepo,
		aiService:       aiService,
		mediaStore:      mediaStore,
		mapsService:     mapsService,
		locationCatalog: locationCatalog,
		idGenerator:     uuid.NewString,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *ListingService) SetModerationEnqueuer(enqueuer ModerationEnqueuer) {
	s.moderationQueue = enqueuer
}

func (s *ListingService) SetUploadSessionStore(store UploadSessionStore) {
	s.uploadSessionRepo = store
}

var _ ListingStore = (*repository.ListingRepo)(nil)
var _ UserStore = (*repository.UserRepo)(nil)

func (s *ListingService) CreateListing(ctx context.Context, userID string, req *models.CreateListingRequest) (*models.Listing, error) {
	if err := utils.ValidateCreateListingRequest(req); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}
	if containsLegacyImagePayload(req.Images) {
		return nil, utils.NewAppError(400, "base64 image payloads are no longer accepted; use upload sessions")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, utils.ErrUnauthorized
	}

	isPremium := IsPremiumEntitled(user, time.Now())
	count, err := s.listingRepo.CountActiveByUserID(ctx, userID)
	if err != nil {
		utils.LogError("count active listings", err, "userId", userID)
		return nil, utils.ErrInternal
	}
	if count >= listingLimitForTier(isPremium) {
		return nil, listingLimitExceededError(isPremium)
	}

	if err := utils.ValidateMediaLimits(isPremium, req.NewImageUploadSessionIDs, req.Videos); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	listingID := s.idGenerator()
	images, err := s.resolveUploadImages(ctx, userID, listingID, req.NewImageUploadSessionIDs, req.FeaturedUploadSessionID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	expiresAt := now.AddDate(0, 0, listingDurationDaysForTier(isPremium))
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
		Images:           images,
		Videos:           req.Videos,
		ImageCount:       len(images),
		PremiumFeatures:  models.PremiumFeatures{IsPremium: isPremium},
		ModerationStatus: models.ModerationStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        &expiresAt,
	}

	if err := s.listingRepo.Put(ctx, listing); err != nil {
		utils.LogError("put listing", err, "listingId", listingID)
		return nil, utils.ErrInternal
	}

	if err := s.enqueueListingModeration(ctx, listing.ListingID, true, nil); err != nil {
		utils.LogError("enqueue listing moderation", err, "listingId", listing.ListingID)
		return nil, utils.ErrInternal
	}

	return listing, nil
}

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
	if listing.ModerationStatus == models.ModerationStatusRejected {
		return nil, utils.NewAppError(403, "listing telah ditolak moderasi dan tidak dapat diedit; hapus listing ini dan buat ulang jika diperlukan")
	}
	if containsLegacyImagePayload(req.Images) {
		return nil, utils.NewAppError(400, "base64 image payloads are no longer accepted; use upload sessions")
	}

	user, _ := s.userRepo.GetByID(ctx, userID)
	isPremium := user != nil && IsPremiumEntitled(user, time.Now())
	now := s.now()
	expired := listing.ExpiresAt != nil && !listing.ExpiresAt.After(now)
	listing, err = s.normalizeListingStateIfNeeded(ctx, listing)
	if err != nil {
		return nil, utils.ErrInternal
	}

	textChanged := req.Title != nil || req.Description != nil

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
		if !(expired && *req.Status == models.ListingStatusActive) {
			listing.Status = *req.Status
		}
	}
	if req.PropertyDetails != nil {
		listing.PropertyDetails = *req.PropertyDetails
	}
	if req.Location != nil {
		listing.Location = *req.Location
	}
	if req.Videos != nil {
		listing.Videos = req.Videos
	}

	newImageIDs, err := s.applyImageUpdate(ctx, listing, req, isPremium)
	if err != nil {
		return nil, err
	}

	needsModeration := textChanged || len(newImageIDs) > 0

	listing.UpdatedAt = s.now()

	if needsModeration {
		listing.ModerationStatus = models.ModerationStatusPending
		listing.ModerationReason = ""
	}

	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return nil, utils.ErrInternal
	}

	if needsModeration {
		if err := s.enqueueListingModeration(ctx, listing.ListingID, textChanged, newImageIDs); err != nil {
			utils.LogError("enqueue listing moderation", err, "listingId", listing.ListingID)
			return nil, utils.ErrInternal
		}
	}

	return listing, nil
}

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
	if seller, err := s.userRepo.GetByID(ctx, listing.UserID); err == nil && seller != nil {
		listing.SellerName = seller.Name
		listing.HasSellerPhone = strings.TrimSpace(seller.Phone) != ""
	}

	return listing, nil
}

func (s *ListingService) GetOwnerListing(ctx context.Context, userID, listingID string) (*models.Listing, error) {
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
	listing, err = s.normalizeListingStateIfNeeded(ctx, listing)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if seller, err := s.userRepo.GetByID(ctx, listing.UserID); err == nil && seller != nil {
		listing.SellerName = seller.Name
		listing.SellerPhone = seller.Phone
		listing.HasSellerPhone = strings.TrimSpace(seller.Phone) != ""
	}
	return listing, nil
}

func (s *ListingService) RevealListingContact(
	ctx context.Context,
	viewerUserID, listingID string,
	channel models.ContactRevealChannel,
) (*models.ListingContactReveal, error) {
	switch channel {
	case models.ContactRevealChannelWhatsApp, models.ContactRevealChannelPhone:
	default:
		return nil, utils.NewAppError(400, "channel is required")
	}

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

	viewer, err := s.userRepo.GetByID(ctx, viewerUserID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if viewer == nil {
		return nil, utils.ErrUnauthorized
	}

	seller, err := s.userRepo.GetByID(ctx, listing.UserID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if seller == nil || strings.TrimSpace(seller.Phone) == "" {
		return nil, utils.NewAppError(404, "seller contact is not available")
	}

	if err := s.consumeContactReveal(viewer); err != nil {
		return nil, err
	}
	if err := s.userRepo.Put(ctx, viewer); err != nil {
		return nil, utils.ErrInternal
	}

	utils.LogInfo(
		"listing contact revealed",
		"listingId", listingID,
		"viewerUserId", viewerUserID,
		"sellerUserId", seller.UserID,
		"channel", channel,
	)

	return &models.ListingContactReveal{
		SellerName:  seller.Name,
		SellerPhone: seller.Phone,
		Channel:     channel,
	}, nil
}

func (s *ListingService) consumeContactReveal(user *models.User) error {
	now := s.now()
	throttle := user.ContactRevealThrottle

	if throttle.WindowStartedAt.IsZero() || now.Sub(throttle.WindowStartedAt) >= contactRevealWindow {
		user.ContactRevealThrottle = models.ContactRevealThrottle{
			WindowStartedAt: now,
			RevealCount:     1,
		}
		return nil
	}

	if throttle.RevealCount >= contactRevealLimit {
		utils.LogWarn("listing contact reveal rate limited", "userId", user.UserID)
		return utils.NewAppError(429, "Terlalu banyak membuka kontak penjual. Coba lagi dalam 10 menit.")
	}

	throttle.RevealCount++
	user.ContactRevealThrottle = throttle
	return nil
}

func (s *ListingService) RecordListingView(ctx context.Context, listingID string) (*models.Listing, error) {
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

	listing.Views++
	listing.UpdatedAt = s.now()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return nil, utils.ErrInternal
	}

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

	return s.normalizeListingStateIfNeeded(ctx, current)
}

func listingLimitForTier(isPremium bool) int {
	if isPremium {
		return premiumTierMaxListings
	}
	return freeTierMaxListings
}

func listingDurationDaysForTier(isPremium bool) int {
	if isPremium {
		return premiumTierListingDurationDays
	}
	return freeTierListingDurationDays
}

func listingLimitExceededError(isPremium bool) error {
	if isPremium {
		return utils.NewAppError(403, fmt.Sprintf("premium tier allows at most %d active listing(s)", premiumTierMaxListings))
	}
	return utils.NewAppError(403, fmt.Sprintf("free tier allows at most %d active listing(s)", freeTierMaxListings))
}

func (s *ListingService) normalizeListingStateIfNeeded(ctx context.Context, listing *models.Listing) (*models.Listing, error) {
	if listing == nil {
		return nil, nil
	}

	now := s.now()
	changed := false

	if listing.Status == models.ListingStatusActive && listing.ExpiresAt != nil && !listing.ExpiresAt.After(now) {
		listing.Status = models.ListingStatusArchived
		changed = true
	}

	if listing.PremiumFeatures.IsFeatured && listing.PremiumFeatures.FeaturedUntil != nil && !listing.PremiumFeatures.FeaturedUntil.After(now) {
		listing.PremiumFeatures.IsFeatured = false
		listing.PremiumFeatures.FeaturedUntil = nil
		changed = true
	}

	if listing.PremiumFeatures.PromotionUntil != nil && !listing.PremiumFeatures.PromotionUntil.After(now) {
		listing.PremiumFeatures.PromotionUntil = nil
		changed = true
	}

	if !changed {
		return listing, nil
	}

	listing.UpdatedAt = now
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return nil, err
	}

	return listing, nil
}

func (s *ListingService) enqueueListingModeration(ctx context.Context, listingID string, checkText bool, newImageIDs []string) error {
	if s.moderationQueue == nil {
		return nil
	}
	return s.moderationQueue.EnqueueListingModeration(ctx, listingID, checkText, newImageIDs)
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

func (s *ListingService) ListListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	params.Page, params.PageSize = utils.ValidatePagination(params.Page, params.PageSize)
	listings, err := s.listingRepo.Scan(ctx, params)
	if err != nil {
		utils.LogError("scan listings", err)
		return nil, utils.ErrInternal
	}

	normalized := make([]models.Listing, 0, len(listings))
	for i := range listings {
		current, err := s.normalizeListingStateIfNeeded(ctx, &listings[i])
		if err != nil {
			utils.LogError("normalize listing state", err, "listingId", listings[i].ListingID)
			return nil, utils.ErrInternal
		}
		normalized = append(normalized, *current)
	}
	return normalized, nil
}

func (s *ListingService) SearchListings(ctx context.Context, params *models.ListingSearchParams) ([]models.Listing, error) {
	return s.ListListings(ctx, params)
}

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
		current, err = s.normalizeListingStateIfNeeded(ctx, current)
		if err != nil {
			utils.LogError("normalize user listing state", err, "userId", userID, "listingId", listing.ListingID)
			return nil, utils.ErrInternal
		}
		filtered = append(filtered, *current)
	}

	return filtered, nil
}

func (s *ListingService) ListSavedListings(ctx context.Context, userID string, params *models.ListingSearchParams) ([]models.Listing, error) {
	params.Page, params.PageSize = utils.ValidatePagination(params.Page, params.PageSize)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = len(user.SavedListingIDs)
	}

	listings := make([]models.Listing, 0, min(limit, len(user.SavedListingIDs)))
	for _, listingID := range user.SavedListingIDs {
		listing, err := s.getCurrentListingByListingID(ctx, listingID)
		if err != nil {
			utils.LogError("get saved listing", err, "userId", userID, "listingId", listingID)
			return nil, utils.ErrInternal
		}
		if listing == nil {
			continue
		}
		if listing.Status != models.ListingStatusActive || listing.ModerationStatus != models.ModerationStatusApproved {
			continue
		}
		listings = append(listings, *listing)
		if len(listings) >= limit {
			break
		}
	}

	return listings, nil
}

func (s *ListingService) SaveListing(ctx context.Context, userID, listingID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return utils.ErrInternal
	}
	if user == nil {
		return utils.ErrUnauthorized
	}

	listing, err := s.getCurrentListingByListingID(ctx, listingID)
	if err != nil {
		return utils.ErrInternal
	}
	if listing == nil || listing.Status != models.ListingStatusActive || listing.ModerationStatus != models.ModerationStatusApproved {
		return utils.ErrNotFound
	}

	for _, savedListingID := range user.SavedListingIDs {
		if savedListingID == listingID {
			return nil
		}
	}

	user.SavedListingIDs = append(user.SavedListingIDs, listingID)
	if err := s.userRepo.Put(ctx, user); err != nil {
		return utils.ErrInternal
	}

	listing.Saves++
	listing.UpdatedAt = s.now()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return utils.ErrInternal
	}

	return nil
}

func (s *ListingService) UnsaveListing(ctx context.Context, userID, listingID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return utils.ErrInternal
	}
	if user == nil {
		return utils.ErrUnauthorized
	}

	index := -1
	for i, savedListingID := range user.SavedListingIDs {
		if savedListingID == listingID {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}

	user.SavedListingIDs = append(user.SavedListingIDs[:index], user.SavedListingIDs[index+1:]...)
	if err := s.userRepo.Put(ctx, user); err != nil {
		return utils.ErrInternal
	}

	listing, err := s.getCurrentListingByListingID(ctx, listingID)
	if err != nil {
		return utils.ErrInternal
	}
	if listing == nil {
		return nil
	}

	if listing.Saves > 0 {
		listing.Saves--
	}
	listing.UpdatedAt = s.now()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return utils.ErrInternal
	}

	return nil
}

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

	if s.mediaStore != nil {
		for _, image := range listing.Images {
			if image.S3Key != "" {
				_ = s.mediaStore.DeleteObject(ctx, image.S3Key)
			}
			if image.ThumbnailKey != "" {
				_ = s.mediaStore.DeleteObject(ctx, image.ThumbnailKey)
			}
		}
	}
	return nil
}

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

	key := BuildPermanentKey(listingID, filename)
	uploadURL, err := s.mediaStore.GetPresignedUploadURL(ctx, key, contentType)
	if err != nil {
		return "", "", utils.ErrInternal
	}
	return uploadURL, key, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

	listing.UpdatedAt = s.now()
	if err := s.listingRepo.Put(ctx, listing); err != nil {
		return utils.ErrInternal
	}
	return nil
}

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
		suggestion.NormalizedAddress = parsed.LocationSuggestion.NormalizedAddress
		parsed.LocationSuggestion = suggestion
	}

	if parsed.LocationSuggestion.Confidence < minLocationConfidence {
		parsed.RequiresManualReview = true
	}

	return &models.ParseTextResponse{
		Parsed:             *parsed,
		RequiresCorrection: parsed.RequiresManualReview,
		Confidence:         parsed.Confidence,
	}, nil
}

// applyImageUpdate applies image changes from the request to the listing and returns the IDs of
// newly added images. Returns (nil, nil) if no image-related fields were provided.
func (s *ListingService) applyImageUpdate(ctx context.Context, listing *models.Listing, req *models.UpdateListingRequest, isPremium bool) ([]string, error) {
	imagesChanged := req.RetainedImageIDs != nil || req.NewImageUploadSessionIDs != nil || req.FeaturedImageID != nil || req.FeaturedUploadSessionID != nil
	if !imagesChanged {
		return nil, utils.ValidateMediaLimits(isPremium, make([]string, len(listing.Images)), listing.Videos)
	}

	currentByID := make(map[string]models.ImageEntry, len(listing.Images))
	for _, image := range listing.Images {
		if image.ImageID != "" {
			currentByID[image.ImageID] = image
		}
	}

	retainedIDs := req.RetainedImageIDs
	if retainedIDs == nil {
		retainedIDs = make([]string, 0, len(listing.Images))
		for _, image := range listing.Images {
			if image.ImageID != "" {
				retainedIDs = append(retainedIDs, image.ImageID)
			}
		}
	}

	retained := make(models.ImageEntries, 0, len(retainedIDs))
	retainedSet := make(map[string]struct{}, len(retainedIDs))
	for _, imageID := range retainedIDs {
		image, ok := currentByID[imageID]
		if !ok {
			return nil, utils.NewAppError(400, fmt.Sprintf("retained image %s was not found", imageID))
		}
		retained = append(retained, image)
		retainedSet[imageID] = struct{}{}
	}

	newFeaturedUploadSessionID := ""
	if req.FeaturedUploadSessionID != nil {
		newFeaturedUploadSessionID = *req.FeaturedUploadSessionID
	}
	newImages, err := s.resolveUploadImages(ctx, listing.UserID, listing.ListingID, req.NewImageUploadSessionIDs, newFeaturedUploadSessionID)
	if err != nil {
		return nil, err
	}

	combined := append(retained, newImages...)
	if err := utils.ValidateMediaLimits(isPremium, make([]string, len(combined)), listing.Videos); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	if req.FeaturedImageID != nil {
		found := false
		for index := range combined {
			combined[index].IsFeatured = combined[index].ImageID == *req.FeaturedImageID
			found = found || combined[index].IsFeatured
		}
		if !found && len(combined) > 0 {
			return nil, utils.NewAppError(400, "featuredImageId was not found in final image set")
		}
	} else if req.FeaturedUploadSessionID == nil {
		enforceSingleFeatured(combined, featuredImageID(listing.Images))
	}
	if !hasFeatured(combined) && len(combined) > 0 {
		combined[0].IsFeatured = true
	}

	for _, image := range listing.Images {
		if _, ok := retainedSet[image.ImageID]; ok {
			continue
		}
		if image.S3Key != "" && s.mediaStore != nil {
			if err := s.mediaStore.DeleteObject(ctx, image.S3Key); err != nil {
				return nil, utils.ErrInternal
			}
		}
		if image.ThumbnailKey != "" && s.mediaStore != nil {
			if err := s.mediaStore.DeleteObject(ctx, image.ThumbnailKey); err != nil {
				return nil, utils.ErrInternal
			}
		}
	}

	listing.Images = combined
	listing.ImageCount = len(combined)

	newIDs := make([]string, 0, len(newImages))
	for _, img := range newImages {
		newIDs = append(newIDs, img.ImageID)
	}
	return newIDs, nil
}

func (s *ListingService) resolveUploadImages(ctx context.Context, userID, listingID string, sessionIDs []string, featuredUploadSessionID string) (models.ImageEntries, error) {
	if len(sessionIDs) == 0 {
		return models.ImageEntries{}, nil
	}
	if s.uploadSessionRepo == nil {
		return nil, utils.ErrInternal
	}
	if s.mediaStore == nil {
		return nil, utils.ErrInternal
	}

	images := make(models.ImageEntries, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := s.uploadSessionRepo.GetBySessionID(ctx, sessionID)
		if err != nil {
			return nil, utils.ErrInternal
		}
		if session == nil || session.ExpiresAt.Before(s.now()) {
			return nil, utils.NewAppError(404, "upload session not found or expired")
		}
		if session.UserID != userID {
			return nil, utils.NewAppError(403, "upload session does not belong to the authenticated user")
		}
		if session.ConsumedAt != nil {
			return nil, utils.NewAppError(409, "upload session has already been consumed")
		}

		head, err := s.mediaStore.HeadObject(ctx, session.StagingKey)
		if err != nil {
			return nil, utils.NewAppError(404, "uploaded object was not found")
		}
		if head.ContentType != session.ExpectedContentType {
			return nil, utils.NewAppError(400, "uploaded object content type does not match prepared session")
		}
		if head.SizeBytes > session.ExpectedMaxSize {
			return nil, utils.NewAppError(400, "uploaded object size exceeds prepared session limit")
		}

		imageID := s.idGenerator()
		permanentKey := BuildPermanentKey(listingID, imageID)
		thumbnailKey := BuildThumbnailKey(listingID, imageID)
		if err := s.mediaStore.CopyObject(ctx, session.StagingKey, permanentKey); err != nil {
			return nil, utils.ErrInternal
		}
		if err := s.mediaStore.CopyObject(ctx, session.StagingKey, thumbnailKey); err != nil {
			return nil, utils.ErrInternal
		}
		if err := s.mediaStore.DeleteObject(ctx, session.StagingKey); err != nil {
			return nil, utils.ErrInternal
		}
		consumedAt := s.now()
		if err := s.uploadSessionRepo.Consume(ctx, sessionID, listingID, consumedAt); err != nil {
			return nil, utils.ErrInternal
		}

		images = append(images, models.ImageEntry{
			ImageID:      imageID,
			S3Key:        permanentKey,
			ThumbnailKey: thumbnailKey,
			ContentType:  head.ContentType,
			SizeBytes:    head.SizeBytes,
			IsFeatured:   featuredUploadSessionID != "" && sessionID == featuredUploadSessionID,
			UploadedAt:   consumedAt,
		})
	}

	if !hasFeatured(images) && len(images) > 0 {
		images[0].IsFeatured = true
	}
	return images, nil
}

func enforceSingleFeatured(images models.ImageEntries, preferredID string) {
	if len(images) == 0 {
		return
	}
	if preferredID == "" {
		for index := range images {
			images[index].IsFeatured = index == 0
		}
		return
	}

	found := false
	for index := range images {
		images[index].IsFeatured = images[index].ImageID == preferredID
		found = found || images[index].IsFeatured
	}
	if !found {
		images[0].IsFeatured = true
	}
}

func featuredImageID(images models.ImageEntries) string {
	for _, image := range images {
		if image.IsFeatured {
			return image.ImageID
		}
	}
	return ""
}

func hasFeatured(images models.ImageEntries) bool {
	for _, image := range images {
		if image.IsFeatured {
			return true
		}
	}
	return false
}
