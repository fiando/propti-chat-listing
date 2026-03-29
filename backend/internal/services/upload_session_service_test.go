package services

import (
	"context"
	"strings"
	"testing"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

func TestPrepareUploadRejectsFreeTierImageOverflow(t *testing.T) {
	ctx := context.Background()

	service := NewUploadSessionService(
		&fakeUploadSessionStore{},
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionFree,
				},
			},
		},
		&fakeListingStore{},
		&fakeMediaService{},
	)

	_, err := service.PrepareUpload(ctx, "user-1", &models.UploadPrepareRequest{
		RetainedImageCount: 1,
		FinalImageCount:    4,
		NewImages: []models.NewImageSpec{
			{ContentType: "image/jpeg", SizeBytes: 1024},
			{ContentType: "image/jpeg", SizeBytes: 1024},
			{ContentType: "image/jpeg", SizeBytes: 1024},
		},
	})
	if err == nil {
		t.Fatal("expected image limit validation error")
	}

	appErr, ok := err.(*utils.AppError)
	if !ok || appErr.Code != 400 {
		t.Fatalf("expected bad request app error, got %T %v", err, err)
	}
	if !strings.Contains(appErr.Message, "at most 3") {
		t.Fatalf("expected free tier limit message, got %q", appErr.Message)
	}
}

func TestPrepareUploadCreatesSingleUseSessionsAndPresignedSlots(t *testing.T) {
	ctx := context.Background()
	sessionStore := &fakeUploadSessionStore{}
	media := &fakeMediaService{
		presignedURLs: map[string]string{
			BuildStagingKey("user-1", "session-a", "upload.jpeg"): "https://upload.example/session-a",
			BuildStagingKey("user-1", "session-b", "upload.png"):  "https://upload.example/session-b",
		},
	}

	service := NewUploadSessionService(
		sessionStore,
		&fakeUserStore{
			user: &models.User{
				UserID: "user-1",
				Subscription: models.Subscription{
					Tier: models.SubscriptionBasic,
				},
			},
		},
		&fakeListingStore{},
		media,
	)
	service.idGenerator = func() string {
		if len(sessionStore.putCalls) == 0 {
			return "session-a"
		}
		return "session-b"
	}

	resp, err := service.PrepareUpload(ctx, "user-1", &models.UploadPrepareRequest{
		RetainedImageCount: 0,
		FinalImageCount:    2,
		NewImages: []models.NewImageSpec{
			{ContentType: "image/jpeg", SizeBytes: 1234},
			{ContentType: "image/png", SizeBytes: 4321},
		},
	})
	if err != nil {
		t.Fatalf("PrepareUpload returned error: %v", err)
	}
	if len(resp.Slots) != 2 {
		t.Fatalf("expected 2 upload slots, got %d", len(resp.Slots))
	}
	if resp.Slots[0].SessionID != "session-a" || resp.Slots[1].SessionID != "session-b" {
		t.Fatalf("unexpected session ordering: %#v", resp.Slots)
	}
	if len(sessionStore.putCalls) != 2 {
		t.Fatalf("expected sessions to be persisted, got %d", len(sessionStore.putCalls))
	}
	first := sessionStore.putCalls[0]
	if first.UserID != "user-1" {
		t.Fatalf("expected user binding to be persisted, got %q", first.UserID)
	}
	if first.ExpectedContentType != "image/jpeg" {
		t.Fatalf("expected content type to be stored, got %q", first.ExpectedContentType)
	}
	if first.ExpectedMaxSize != 1234 {
		t.Fatalf("expected size limit to be stored, got %d", first.ExpectedMaxSize)
	}
	if first.ExpiresAt.IsZero() {
		t.Fatal("expected expiry to be set")
	}
}
