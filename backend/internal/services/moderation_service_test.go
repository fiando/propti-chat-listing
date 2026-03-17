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
}

func (f *fakeImageModerator) ModerateImages(ctx context.Context, images []string) (bool, string, []string, error) {
	f.calls++
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
	store := &fakeListingStore{
		listingsByID: map[string]*models.Listing{
			"listing-1": {
				ListingID:        "listing-1",
				UserID:           "user-1",
				Title:            "Rumah siap huni",
				Description:      "Dekat tol dan sekolah",
				Status:           models.ListingStatusActive,
				ModerationStatus: models.ModerationStatusPending,
				Images:           []string{"data:image/jpeg;base64,ZmFrZQ=="},
				UpdatedAt:        time.Now().UTC(),
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
					Images:           []string{"data:image/jpeg;base64,ZmFrZQ=="},
					UpdatedAt:        time.Now().UTC(),
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

	service := NewModerationService(textModerator, imageModerator, moderationStore, store)

	listing, err := service.ModerateListing(ctx, "listing-1")
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
}

func TestLooksLikePropertyRequiresRelevantLabels(t *testing.T) {
	propertyLabels := []string{"Airplane", "Poster", "Vehicle"}
	if looksLikeProperty(buildLabels(propertyLabels...)) {
		t.Fatalf("expected non-property labels to be rejected")
	}

	if !looksLikeProperty(buildLabels("House", "Building", "Window")) {
		t.Fatalf("expected property-related labels to be accepted")
	}
}

func buildLabels(names ...string) []rekognitiontypes.Label {
	labels := make([]rekognitiontypes.Label, 0, len(names))
	for _, name := range names {
		value := name
		labels = append(labels, rekognitiontypes.Label{Name: &value})
	}
	return labels
}
