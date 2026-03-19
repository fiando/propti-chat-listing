package services

import (
	"context"

	"github.com/fiando/propti/backend/internal/models"
)

type ListingMediaPresenter struct {
	mediaStore MediaStorage
}

func NewListingMediaPresenter(mediaStore MediaStorage) *ListingMediaPresenter {
	return &ListingMediaPresenter{mediaStore: mediaStore}
}

func (p *ListingMediaPresenter) PresentPublicSummary(ctx context.Context, listing *models.Listing) (*models.ListingResponse, error) {
	resp := baseListingResponse(listing)
	legacy := legacyImageValues(listing.Images)
	if len(legacy) > 0 {
		resp.Images = legacy
		return resp, nil
	}

	featured := featuredImage(listing.Images)
	if featured != nil && featured.ThumbnailKey != "" && p.mediaStore != nil {
		signedURL, err := p.mediaStore.GetSignedDownloadURL(ctx, featured.ThumbnailKey)
		if err != nil {
			return nil, err
		}
		resp.FeaturedThumbnailURL = signedURL
	}
	return resp, nil
}

func (p *ListingMediaPresenter) PresentPublicDetail(ctx context.Context, listing *models.Listing) (*models.ListingResponse, error) {
	return p.presentDetail(ctx, listing, false)
}

func (p *ListingMediaPresenter) PresentOwnerDetail(ctx context.Context, listing *models.Listing) (*models.ListingResponse, error) {
	return p.presentDetail(ctx, listing, true)
}

func (p *ListingMediaPresenter) PresentSummaryCollection(ctx context.Context, listings []models.Listing) ([]models.ListingResponse, error) {
	responses := make([]models.ListingResponse, 0, len(listings))
	for i := range listings {
		resp, err := p.PresentPublicSummary(ctx, &listings[i])
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}

func (p *ListingMediaPresenter) presentDetail(ctx context.Context, listing *models.Listing, ownerView bool) (*models.ListingResponse, error) {
	resp := baseListingResponse(listing)
	legacy := legacyImageValues(listing.Images)
	if len(legacy) > 0 {
		resp.Images = legacy
		return resp, nil
	}

	views := make([]models.ListingImageView, 0, len(listing.Images))
	for _, image := range listing.Images {
		if image.IsLegacy() {
			continue
		}

		url := ""
		if p.mediaStore != nil && image.S3Key != "" {
			signedURL, err := p.mediaStore.GetSignedDownloadURL(ctx, image.S3Key)
			if err != nil {
				return nil, err
			}
			url = signedURL
		}

		thumbnailURL := ""
		if image.ThumbnailKey != "" && p.mediaStore != nil {
			signedURL, err := p.mediaStore.GetSignedDownloadURL(ctx, image.ThumbnailKey)
			if err != nil {
				return nil, err
			}
			thumbnailURL = signedURL
		}

		views = append(views, models.ListingImageView{
			ImageID:      image.ImageID,
			URL:          url,
			ThumbnailURL: thumbnailURL,
			ContentType:  image.ContentType,
			SizeBytes:    image.SizeBytes,
			IsFeatured:   image.IsFeatured,
			UploadedAt:   image.UploadedAt,
		})
		if image.IsFeatured && resp.FeaturedThumbnailURL == "" && thumbnailURL != "" {
			resp.FeaturedThumbnailURL = thumbnailURL
		}
	}
	resp.Images = views
	return resp, nil
}

func baseListingResponse(listing *models.Listing) *models.ListingResponse {
	return &models.ListingResponse{
		ListingID:        listing.ListingID,
		UserID:           listing.UserID,
		Title:            listing.Title,
		Description:      listing.Description,
		Price:            listing.Price,
		PriceUnit:        listing.PriceUnit,
		ListingType:      listing.ListingType,
		Status:           listing.Status,
		PropertyDetails:  listing.PropertyDetails,
		Location:         listing.Location,
		Videos:           listing.Videos,
		ImageCount:       listing.ImageCount,
		PremiumFeatures:  listing.PremiumFeatures,
		SellerName:       listing.SellerName,
		SellerPhone:      listing.SellerPhone,
		HasSellerPhone:   listing.HasSellerPhone,
		Views:            listing.Views,
		Saves:            listing.Saves,
		ModerationStatus: listing.ModerationStatus,
		ModerationReason: listing.ModerationReason,
		CreatedAt:        listing.CreatedAt,
		UpdatedAt:        listing.UpdatedAt,
	}
}

func featuredImage(images models.ImageEntries) *models.ImageEntry {
	for i := range images {
		if images[i].IsFeatured {
			return &images[i]
		}
	}
	if len(images) == 0 {
		return nil
	}
	return &images[0]
}
