package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

type fakeWhatsAppIdentityUserStore struct {
	usersByID           map[string]*models.User
	userByWhatsAppPhone map[string]*models.User
	putCalls            int
}

func (f *fakeWhatsAppIdentityUserStore) GetByID(_ context.Context, userID string) (*models.User, error) {
	if f.usersByID == nil {
		return nil, nil
	}
	user := f.usersByID[userID]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (f *fakeWhatsAppIdentityUserStore) Put(_ context.Context, user *models.User) error {
	f.putCalls++
	if f.usersByID == nil {
		f.usersByID = map[string]*models.User{}
	}
	copy := *user
	f.usersByID[user.UserID] = &copy
	if f.userByWhatsAppPhone != nil {
		for phone, existing := range f.userByWhatsAppPhone {
			if existing != nil && existing.UserID == copy.UserID {
				delete(f.userByWhatsAppPhone, phone)
			}
		}
	}
	if copy.WhatsAppLinkedPhone != "" && copy.WhatsAppVerifiedAt != nil {
		if f.userByWhatsAppPhone == nil {
			f.userByWhatsAppPhone = map[string]*models.User{}
		}
		f.userByWhatsAppPhone[copy.WhatsAppLinkedPhone] = &copy
	}
	return nil
}

func (f *fakeWhatsAppIdentityUserStore) GetByWhatsAppPhone(_ context.Context, phone string) (*models.User, error) {
	if f.userByWhatsAppPhone == nil {
		return nil, nil
	}
	user := f.userByWhatsAppPhone[phone]
	if user == nil || user.WhatsAppVerifiedAt == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

type fakeWhatsAppOTPStore struct {
	challenges map[string]*repository.OTPChallenge
}

type fakeWhatsAppOTPSender struct {
	calls []struct {
		phone string
		code  string
	}
	err error
}

func (f *fakeWhatsAppOTPSender) Send(_ context.Context, phone, code string) error {
	f.calls = append(f.calls, struct {
		phone string
		code  string
	}{phone: phone, code: code})
	return f.err
}

func (f *fakeWhatsAppOTPStore) Put(_ context.Context, challenge *repository.OTPChallenge) error {
	if f.challenges == nil {
		f.challenges = map[string]*repository.OTPChallenge{}
	}
	copy := *challenge
	f.challenges[challenge.ChallengeID] = &copy
	return nil
}

func (f *fakeWhatsAppOTPStore) GetByID(_ context.Context, challengeID string) (*repository.OTPChallenge, error) {
	challenge := f.challenges[challengeID]
	if challenge == nil {
		return nil, nil
	}
	copy := *challenge
	return &copy, nil
}

func (f *fakeWhatsAppOTPStore) GetLatestByUser(_ context.Context, userID string) (*repository.OTPChallenge, error) {
	var latest *repository.OTPChallenge
	for _, item := range f.challenges {
		if item.UserID != userID {
			continue
		}
		if latest == nil || item.CreatedAt.After(latest.CreatedAt) {
			copy := *item
			latest = &copy
		}
	}
	return latest, nil
}

func (f *fakeWhatsAppOTPStore) GetLatestByPhone(_ context.Context, phone string) (*repository.OTPChallenge, error) {
	var latest *repository.OTPChallenge
	for _, item := range f.challenges {
		if item.Phone != phone {
			continue
		}
		if latest == nil || item.CreatedAt.After(latest.CreatedAt) {
			copy := *item
			latest = &copy
		}
	}
	return latest, nil
}

func TestWhatsAppIdentityStartLinkCreatesChallengeWithExpiry(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}}}
	otpStore := &fakeWhatsAppOTPStore{}

	svc, err := NewWhatsAppIdentityService(users, otpStore, WhatsAppIdentityOptions{
		Now:          func() time.Time { return now },
		OTPGenerator: func() (string, error) { return "123456", nil },
		OTPExpiry:    5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	challenge, err := svc.StartLink(context.Background(), "user-1", "+628123456789")
	if err != nil {
		t.Fatalf("StartLink returned error: %v", err)
	}
	if challenge.ChallengeID == "" {
		t.Fatal("expected challenge id")
	}
	if !challenge.ExpiresAt.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("expected expiry %s, got %s", now.Add(5*time.Minute), challenge.ExpiresAt)
	}

	stored, err := otpStore.GetByID(context.Background(), challenge.ChallengeID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if stored == nil {
		t.Fatal("expected stored challenge")
	}
	if stored.OTPCode != "123456" {
		t.Fatalf("expected otp code 123456, got %q", stored.OTPCode)
	}
	if stored.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", stored.RetryCount)
	}
	if users.putCalls != 0 {
		t.Fatalf("expected start link to avoid mutating user before verification, got %d put calls", users.putCalls)
	}
	if users.usersByID["user-1"].WhatsAppLinkedPhone != "" {
		t.Fatalf("expected whatsapp phone to stay unset before verification, got %q", users.usersByID["user-1"].WhatsAppLinkedPhone)
	}
}

func TestWhatsAppIdentityStartLinkReturnsInboundChallengeMessageAndLink(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}}}
	otpStore := &fakeWhatsAppOTPStore{}

	svc, err := NewWhatsAppIdentityService(users, otpStore, WhatsAppIdentityOptions{
		Now:                   func() time.Time { return now },
		OTPGenerator:          func() (string, error) { return "123456", nil },
		WhatsAppMessageTarget: "whatsapp:+14155238886",
	})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	challenge, err := svc.StartLink(context.Background(), "user-1", "+628123456789")
	if err != nil {
		t.Fatalf("StartLink returned error: %v", err)
	}
	if challenge.MessageText != "PROPTI LINK 123456" {
		t.Fatalf("expected challenge message text, got %q", challenge.MessageText)
	}
	if challenge.MessageLink == "" {
		t.Fatal("expected challenge message link to be present")
	}
}

func TestWhatsAppIdentityVerifyLinkFromInboundSetsVerifiedIdentity(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}}}
	otpStore := &fakeWhatsAppOTPStore{
		challenges: map[string]*repository.OTPChallenge{
			"challenge-1": {
				ChallengeID: "challenge-1",
				UserID:      "user-1",
				Phone:       "+628123456789",
				OTPCode:     "123456",
				ExpiresAt:   now.Add(5 * time.Minute),
				CreatedAt:   now,
			},
		},
	}
	svc, err := NewWhatsAppIdentityService(users, otpStore, WhatsAppIdentityOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	verified, err := svc.VerifyLinkFromInbound(context.Background(), "+628123456789", "PROPTI LINK 123456")
	if err != nil {
		t.Fatalf("VerifyLinkFromInbound returned error: %v", err)
	}
	if !verified {
		t.Fatal("expected inbound verification to succeed")
	}
	user := users.usersByID["user-1"]
	if user.WhatsAppLinkedPhone != "+628123456789" || user.WhatsAppVerifiedAt == nil {
		t.Fatalf("expected user whatsapp identity to be linked+verified, got %#v", user)
	}
}

func TestWhatsAppIdentityStartLinkRejectsAlreadyVerifiedPhoneByAnotherUser(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-time.Hour)
	users := &fakeWhatsAppIdentityUserStore{
		usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}},
		userByWhatsAppPhone: map[string]*models.User{
			"+628123456789": {
				UserID:              "user-2",
				WhatsAppLinkedPhone: "+628123456789",
				WhatsAppVerifiedAt:  &verifiedAt,
			},
		},
	}

	svc, err := NewWhatsAppIdentityService(users, &fakeWhatsAppOTPStore{}, WhatsAppIdentityOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	_, err = svc.StartLink(context.Background(), "user-1", "+628123456789")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if appCode(t, err) != 409 {
		t.Fatalf("expected status 409, got %d (%v)", appCode(t, err), err)
	}
}

func TestWhatsAppIdentityStartLinkAllowsPhoneHeldOnlyByUnverifiedUser(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{
		usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}},
		userByWhatsAppPhone: map[string]*models.User{
			"+628123456789": {
				UserID:              "user-2",
				WhatsAppLinkedPhone: "+628123456789",
			},
		},
	}

	svc, err := NewWhatsAppIdentityService(users, &fakeWhatsAppOTPStore{}, WhatsAppIdentityOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	challenge, err := svc.StartLink(context.Background(), "user-1", "+628123456789")
	if err != nil {
		t.Fatalf("expected unverified linked phone to remain reusable, got %v", err)
	}
	if challenge == nil {
		t.Fatal("expected challenge to be created")
	}
}

func TestWhatsAppIdentityStartLinkAppliesRetryGuard(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}}}
	otpStore := &fakeWhatsAppOTPStore{challenges: map[string]*repository.OTPChallenge{
		"existing": {
			ChallengeID: "existing",
			UserID:      "user-1",
			Phone:       "+628123456789",
			RetryCount:  3,
			ExpiresAt:   now.Add(2 * time.Minute),
			CreatedAt:   now.Add(-time.Minute),
		},
	}}

	svc, err := NewWhatsAppIdentityService(users, otpStore, WhatsAppIdentityOptions{Now: func() time.Time { return now }, MaxChallengeRetries: 3})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	_, err = svc.StartLink(context.Background(), "user-1", "+628123456789")
	if err == nil {
		t.Fatal("expected retry-guard error")
	}
	if appCode(t, err) != 429 {
		t.Fatalf("expected status 429, got %d (%v)", appCode(t, err), err)
	}
}

func TestWhatsAppIdentityVerifyLinkHandlesAttemptsAndVerification(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{"user-1": {UserID: "user-1"}}}
	otpStore := &fakeWhatsAppOTPStore{challenges: map[string]*repository.OTPChallenge{
		"challenge-1": {
			ChallengeID: "challenge-1",
			UserID:      "user-1",
			Phone:       "+628123456789",
			OTPCode:     "123456",
			ExpiresAt:   now.Add(5 * time.Minute),
			CreatedAt:   now,
		},
	}}

	svc, err := NewWhatsAppIdentityService(users, otpStore, WhatsAppIdentityOptions{Now: func() time.Time { return now }, MaxVerificationAttempts: 2})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	_, err = svc.VerifyLink(context.Background(), "user-1", "challenge-1", "000000")
	if err == nil || appCode(t, err) != 401 {
		t.Fatalf("expected first invalid-otp 401 error, got %v", err)
	}

	_, err = svc.VerifyLink(context.Background(), "user-1", "challenge-1", "000000")
	if err == nil || appCode(t, err) != 429 {
		t.Fatalf("expected second invalid-otp to lock with 429, got %v", err)
	}

	otpStore.challenges["challenge-2"] = &repository.OTPChallenge{
		ChallengeID: "challenge-2",
		UserID:      "user-1",
		Phone:       "+628123456789",
		OTPCode:     "654321",
		ExpiresAt:   now.Add(5 * time.Minute),
		CreatedAt:   now.Add(time.Minute),
	}

	eligibility, err := svc.VerifyLink(context.Background(), "user-1", "challenge-2", "654321")
	if err != nil {
		t.Fatalf("VerifyLink returned error: %v", err)
	}
	if !eligibility.Eligible {
		t.Fatalf("expected eligible after verification, got %#v", eligibility)
	}
	if users.usersByID["user-1"].WhatsAppVerifiedAt == nil {
		t.Fatal("expected user verifiedAt to be set")
	}
}

func TestWhatsAppIdentityWriteEligibilityDependsOnVerifiedLinkage(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-5 * time.Minute)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{
		"user-unverified": {UserID: "user-unverified", WhatsAppLinkedPhone: "+628100000001"},
		"user-verified": {
			UserID:              "user-verified",
			WhatsAppLinkedPhone: "+628100000002",
			WhatsAppVerifiedAt:  &verifiedAt,
		},
	}}

	svc, err := NewWhatsAppIdentityService(users, &fakeWhatsAppOTPStore{}, WhatsAppIdentityOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	unverified, err := svc.GetWriteEligibility(context.Background(), "user-unverified")
	if err != nil {
		t.Fatalf("GetWriteEligibility returned error: %v", err)
	}
	if unverified.Eligible {
		t.Fatalf("expected unverified user to be ineligible, got %#v", unverified)
	}

	verified, err := svc.GetWriteEligibility(context.Background(), "user-verified")
	if err != nil {
		t.Fatalf("GetWriteEligibility returned error: %v", err)
	}
	if !verified.Eligible {
		t.Fatalf("expected verified user to be eligible, got %#v", verified)
	}
}

func TestWhatsAppIdentityDisconnectLinkClearsWhatsAppIdentityState(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	linkedAt := now.Add(-2 * time.Hour)
	verifiedAt := now.Add(-time.Hour)
	users := &fakeWhatsAppIdentityUserStore{usersByID: map[string]*models.User{
		"user-1": {
			UserID:              "user-1",
			WhatsAppLinkedPhone: "+628123456789",
			WhatsAppLinkedAt:    &linkedAt,
			WhatsAppVerifiedAt:  &verifiedAt,
		},
	}}

	svc, err := NewWhatsAppIdentityService(users, &fakeWhatsAppOTPStore{}, WhatsAppIdentityOptions{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("NewWhatsAppIdentityService returned error: %v", err)
	}

	eligibility, err := svc.DisconnectLink(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("DisconnectLink returned error: %v", err)
	}
	if eligibility.Eligible {
		t.Fatalf("expected user to become ineligible after disconnect, got %#v", eligibility)
	}
	if users.usersByID["user-1"].WhatsAppLinkedPhone != "" {
		t.Fatalf("expected linked phone to be cleared, got %q", users.usersByID["user-1"].WhatsAppLinkedPhone)
	}
	if users.usersByID["user-1"].WhatsAppLinkedAt != nil {
		t.Fatal("expected linkedAt to be cleared")
	}
	if users.usersByID["user-1"].WhatsAppVerifiedAt != nil {
		t.Fatal("expected verifiedAt to be cleared")
	}
}

func appCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T (%v)", err, err)
	}
	return appErr.Code
}
