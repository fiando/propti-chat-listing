package services

import (
	"context"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

func TestListingMediaPresenterPublicSummaryUsesFeaturedThumbnail(t *testing.T) {
	presenter := NewListingMediaPresenter(&fakeMediaService{
		signedURLs: map[string]string{
			"thumbnails/listing-1/image-2": "https://signed.example/thumb-2",
		},
	})

	resp, err := presenter.PresentPublicSummary(context.Background(), &models.Listing{
		ListingID: "listing-1",
		Title:     "Rumah Depok",
		Images: models.ImageEntries{
			{ImageID: "image-1", ThumbnailKey: "thumbnails/listing-1/image-1"},
			{ImageID: "image-2", ThumbnailKey: "thumbnails/listing-1/image-2", IsFeatured: true},
		},
	})
	if err != nil {
		t.Fatalf("PresentPublicSummary returned error: %v", err)
	}
	if resp.FeaturedThumbnailURL != "https://signed.example/thumb-2" {
		t.Fatalf("expected featured thumbnail url, got %q", resp.FeaturedThumbnailURL)
	}
	if resp.Images != nil {
		t.Fatalf("expected summary response to omit structured gallery, got %#v", resp.Images)
	}
}

func TestListingMediaPresenterPublicDetailSignsThumbnailGallery(t *testing.T) {
	presenter := NewListingMediaPresenter(&fakeMediaService{
		signedURLs: map[string]string{
			"listings/listing-1/image-1":   "https://signed.example/image-1",
			"thumbnails/listing-1/image-1": "https://signed.example/thumb-1",
		},
	})

	resp, err := presenter.PresentPublicDetail(context.Background(), &models.Listing{
		ListingID: "listing-1",
		Images: models.ImageEntries{
			{
				ImageID:      "image-1",
				S3Key:        "listings/listing-1/image-1",
				ThumbnailKey: "thumbnails/listing-1/image-1",
				ContentType:  "image/jpeg",
				SizeBytes:    2048,
				IsFeatured:   true,
				UploadedAt:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			},
		},
	})
	if err != nil {
		t.Fatalf("PresentPublicDetail returned error: %v", err)
	}

	images, ok := resp.Images.([]models.ListingImageView)
	if !ok {
		t.Fatalf("expected typed listing gallery, got %#v", resp.Images)
	}
	if len(images) != 1 {
		t.Fatalf("expected one gallery image, got %d", len(images))
	}
	if images[0].ThumbnailURL != "https://signed.example/thumb-1" {
		t.Fatalf("expected signed thumbnail URL, got %q", images[0].ThumbnailURL)
	}
	if resp.FeaturedThumbnailURL != "https://signed.example/thumb-1" {
		t.Fatalf("expected featured thumbnail URL to reuse signed thumbnail, got %q", resp.FeaturedThumbnailURL)
	}
}

func TestListingMediaPresenterOwnerDetailSignsGalleryWithoutRawKeys(t *testing.T) {
	presenter := NewListingMediaPresenter(&fakeMediaService{
		signedURLs: map[string]string{
			"listings/listing-1/image-1": "https://signed.example/image-1",
		},
	})

	resp, err := presenter.PresentOwnerDetail(context.Background(), &models.Listing{
		ListingID: "listing-1",
		Images: models.ImageEntries{
			{
				ImageID:      "image-1",
				S3Key:        "listings/listing-1/image-1",
				ThumbnailKey: "thumbnails/listing-1/image-1",
				ContentType:  "image/jpeg",
				SizeBytes:    2048,
				IsFeatured:   true,
				UploadedAt:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
			},
		},
	})
	if err != nil {
		t.Fatalf("PresentOwnerDetail returned error: %v", err)
	}

	images, ok := resp.Images.([]models.ListingImageView)
	if !ok {
		t.Fatalf("expected typed listing gallery, got %#v", resp.Images)
	}
	if len(images) != 1 {
		t.Fatalf("expected one gallery image, got %d", len(images))
	}
	if images[0].URL != "https://signed.example/image-1" {
		t.Fatalf("expected signed URL, got %q", images[0].URL)
	}
	if images[0].S3Key != "" || images[0].ThumbnailKey != "" {
		t.Fatalf("expected raw keys to be hidden, got %#v", images[0])
	}
}

func TestListingMediaPresenterKeepsLegacyImagesReadable(t *testing.T) {
	presenter := NewListingMediaPresenter(&fakeMediaService{})

	resp, err := presenter.PresentPublicDetail(context.Background(), &models.Listing{
		ListingID: "listing-legacy",
		Images: models.ImageEntries{
			{LegacyValue: "data:image/jpeg;base64,ZmFrZQ=="},
			{LegacyValue: "https://example.com/legacy.jpg"},
		},
	})
	if err != nil {
		t.Fatalf("PresentPublicDetail returned error: %v", err)
	}

	legacyImages, ok := resp.Images.([]string)
	if !ok {
		t.Fatalf("expected legacy image list, got %#v", resp.Images)
	}
	if len(legacyImages) != 2 {
		t.Fatalf("expected legacy values to be preserved, got %#v", legacyImages)
	}
}
