package services

import (
	"context"
	"strings"
	"testing"
	"time"

	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/fiando/propti/backend/internal/models"
)

type fakeImageModerator struct {
	calls    int
	approved bool
	reason   string
	flags    []string
	err      error
	bytes    [][]byte
}

func (f *fakeImageModerator) ModerateImages(ctx context.Context, images [][]byte) (bool, string, []string, error) {
	f.calls++
	f.bytes = append(f.bytes, images...)
	return f.approved, f.reason, f.flags, f.err
}

type fakeModerationStore struct {
	records []*models.Moderation
}

func (f *fakeModerationStore) Put(ctx context.Context, moderation *models.Moderation) error {
	copy := *moderation
	f.records = append(f.records, &copy)
	return nil
}

func TestModerationServiceRejectsListingWhenImageIsNotPropertyRelated(t *testing.T) {
	ctx := context.Background()
	media := &fakeMediaService{
		bytesByKey: map[string][]byte{
			"listings/listing-1/image-1": []byte("image-bytes"),
		},
	}
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah siap huni",
				Description:      "Dekat tol dan sekolah",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusPending,
				Images: models.ImageEntries{
					{ImageID: "image-1", S3Key: "listings/listing-1/image-1", ThumbnailKey: "thumbnails/listing-1/image-1", IsFeatured: true},
				},
				UpdatedAt: time.Now().UTC(),
			},
		},
		listingsByUser: map[string]map[string]*models.Listing{
			"user-1": {
				"listing-1": {
					ListingID:        "listing-1",
					UserID:           "user-1",
					Title:            "Rumah siap huni",
					Description:      "Dekat tol dan sekolah",
					Status:           models.ListingStatusActive,
					ModerationStatus: models.ModerationStatusPending,
					Images: models.ImageEntries{
						{ImageID: "image-1", S3Key: "listings/listing-1/image-1", ThumbnailKey: "thumbnails/listing-1/image-1", IsFeatured: true},
					},
					UpdatedAt: time.Now().UTC(),
				},
			},
		},
	}
	textModerator := &fakeAIService{
		moderateOK: true,
	}
	imageModerator := &fakeImageModerator{
		approved: false,
		reason:   "Gambar 1 tidak terlihat relevan dengan properti yang dijual atau disewakan",
		flags:    []string{"aircraft", "poster"},
	}
	moderationStore := &fakeModerationStore{}

	service := NewModerationService(textModerator, imageModerator, moderationStore, store, media)

	listing, err := service.ModerateListing(ctx, "listing-1", true, nil)
	if err != nil {
		t.Fatalf("ModerateListing returned error: %v", err)
	}
	if listing.ModerationStatus != models.ModerationStatusRejected {
		t.Fatalf("expected listing to be rejected, got %q", listing.ModerationStatus)
	}
	if listing.Status != models.ListingStatusArchived {
		t.Fatalf("expected rejected listing to be archived, got %q", listing.Status)
	}
	if !strings.Contains(listing.ModerationReason, "tidak terlihat relevan") {
		t.Fatalf("expected rejection reason to mention non-property image, got %q", listing.ModerationReason)
	}
	if len(moderationStore.records) != 2 {
		t.Fatalf("expected 2 moderation records (text + media), got %d", len(moderationStore.records))
	}
	if moderationStore.records[1].Type != models.ModerationTypeMedia {
		t.Fatalf("expected second moderation record to be media check, got %q", moderationStore.records[1].Type)
	}
	if len(imageModerator.bytes) != 1 || string(imageModerator.bytes[0]) != "image-bytes" {
		t.Fatalf("expected moderation to use downloaded S3 bytes, got %#v", imageModerator.bytes)
	}
	if listing.Images[0].S3Key != "rejected/listing-1/image-1" {
		t.Fatalf("expected rejected image key to be moved, got %#v", listing.Images[0])
	}
	if listing.Images[0].ThumbnailKey != "" {
		t.Fatalf("expected rejected listing to clear thumbnail exposure, got %#v", listing.Images[0])
	}
	if len(media.copies) == 0 {
		t.Fatal("expected rejected moderation to move objects into rejected prefix")
	}
}

func TestLooksLikePropertyRequiresRelevantLabels(t *testing.T) {
	if looksLikeProperty(buildLabels(90.0, "Airplane", "Poster", "Vehicle")) {
		t.Fatalf("expected non-property labels to be rejected")
	}

	if !looksLikeProperty(buildLabels(90.0, "House", "Building", "Window")) {
		t.Fatalf("expected property-related labels to be accepted")
	}

	// Labels below confidence threshold must not count as property.
	if looksLikeProperty(buildLabels(50.0, "House", "Building", "Window")) {
		t.Fatalf("expected low-confidence property labels to be rejected")
	}

	// Parent/alias expansion must NOT bypass the check (game screenshot scenario).
	gameLabels := buildLabels(90.0, "Video Game", "Controller", "Screen")
	// Manually add a parent "Indoor" to simulate Rekognition's taxonomy expansion.
	indoor := "Indoor"
	gameLabels[0].Parents = []rekognitiontypes.Parent{{Name: &indoor}}
	if looksLikeProperty(gameLabels) {
		t.Fatalf("expected game labels with indoor parent to be rejected (no parent expansion)")
	}
}

func buildLabels(confidence float32, names ...string) []rekognitiontypes.Label {
	labels := make([]rekognitiontypes.Label, 0, len(names))
	for _, name := range names {
		value := name
		conf := confidence
		labels = append(labels, rekognitiontypes.Label{Name: &value, Confidence: &conf})
	}
	return labels
}
