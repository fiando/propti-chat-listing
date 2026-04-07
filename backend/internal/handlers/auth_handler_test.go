package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type fakeActiveListingCounter struct {
	count   int
	err     error
	userIDs []string
}

func (f *fakeActiveListingCounter) CountActiveByUserID(_ context.Context, userID string) (int, error) {
	f.userIDs = append(f.userIDs, userID)
	if f.err != nil {
		return 0, f.err
	}
	return f.count, nil
}

func TestPopulateActiveListingCountSetsComputedCount(t *testing.T) {
	user := &models.User{
		UserID: "user-1",
		Subscription: models.Subscription{
			Tier: models.SubscriptionFree,
		},
	}
	counter := &fakeActiveListingCounter{count: 3}

	if err := populateActiveListingCount(context.Background(), counter, user); err != nil {
		t.Fatalf("populateActiveListingCount returned error: %v", err)
	}

	if got := user.Subscription.ActiveListingsCount; got != 3 {
		t.Fatalf("expected activeListingsCount 3, got %d", got)
	}

	if len(counter.userIDs) != 1 || counter.userIDs[0] != "user-1" {
		t.Fatalf("expected counter to be called for user-1, got %#v", counter.userIDs)
	}
}

func TestPopulateActiveListingCountReturnsCounterError(t *testing.T) {
	user := &models.User{UserID: "user-1"}
	expectedErr := errors.New("boom")

	err := populateActiveListingCount(context.Background(), &fakeActiveListingCounter{err: expectedErr}, user)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestApplyUserProfileUpdateKeepsWhatsAppIdentityWhenSavingPhone(t *testing.T) {
	linkedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	verifiedAt := linkedAt.Add(5 * time.Minute)
	user := &models.User{
		UserID:              "user-1",
		Phone:               "081111111111",
		Role:                models.UserRoleBuyer,
		WhatsAppLinkedPhone: "+628123456789",
		WhatsAppLinkedAt:    &linkedAt,
		WhatsAppVerifiedAt:  &verifiedAt,
		Preferences: models.UserPreferences{
			FavoriteLocations: []string{"sleman"},
			SearchHistory:     []string{"rumah"},
			Notifications:     true,
		},
	}
	newPhone := "082222222222"

	applyUserProfileUpdate(user, models.UpdateUserRequest{Phone: &newPhone})

	if user.Phone != newPhone {
		t.Fatalf("expected phone to update to %q, got %q", newPhone, user.Phone)
	}
	if user.WhatsAppLinkedPhone != "+628123456789" {
		t.Fatalf("expected linked whatsapp phone to stay unchanged, got %q", user.WhatsAppLinkedPhone)
	}
	if user.WhatsAppLinkedAt != &linkedAt {
		t.Fatalf("expected linked timestamp pointer to stay unchanged")
	}
	if user.WhatsAppVerifiedAt != &verifiedAt {
		t.Fatalf("expected verified timestamp pointer to stay unchanged")
	}
}
