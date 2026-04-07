package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

const (
	defaultOTPExpiry             = 5 * time.Minute
	defaultMaxChallengeRetries   = 3
	defaultMaxVerificationErrors = 3
)

type WhatsAppIdentityUserStore interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Put(ctx context.Context, user *models.User) error
	GetByWhatsAppPhone(ctx context.Context, phone string) (*models.User, error)
}

type WhatsAppIdentityOptions struct {
	Now                     func() time.Time
	IDGenerator             func() string
	OTPGenerator            func() (string, error)
	SendOTP                 func(ctx context.Context, phone, otpCode string) error
	OTPExpiry               time.Duration
	MaxChallengeRetries     int
	MaxVerificationAttempts int
	WhatsAppMessageTarget   string
}

type WhatsAppLinkChallenge struct {
	ChallengeID string    `json:"challengeId"`
	Phone       string    `json:"phone"`
	ExpiresAt   time.Time `json:"expiresAt"`
	RetryCount  int       `json:"retryCount"`
	MessageText string    `json:"messageText,omitempty"`
	MessageLink string    `json:"messageLink,omitempty"`
}

type WhatsAppWriteEligibility struct {
	Eligible    bool       `json:"eligible"`
	Reason      string     `json:"reason,omitempty"`
	LinkedPhone string     `json:"linkedPhone,omitempty"`
	VerifiedAt  *time.Time `json:"verifiedAt,omitempty"`
}

type WhatsAppIdentityService struct {
	userStore               WhatsAppIdentityUserStore
	otpStore                repository.OTPStore
	now                     func() time.Time
	idGenerator             func() string
	otpGenerator            func() (string, error)
	sendOTP                 func(ctx context.Context, phone, otpCode string) error
	otpExpiry               time.Duration
	maxChallengeRetries     int
	maxVerificationAttempts int
	whatsAppMessageTarget   string
}

func NewWhatsAppIdentityService(userStore WhatsAppIdentityUserStore, otpStore repository.OTPStore, opts WhatsAppIdentityOptions) (*WhatsAppIdentityService, error) {
	if userStore == nil {
		return nil, fmt.Errorf("user store is required")
	}
	if otpStore == nil {
		return nil, fmt.Errorf("otp store is required")
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}
	idGen := opts.IDGenerator
	if idGen == nil {
		idGen = uuid.NewString
	}
	otpGen := opts.OTPGenerator
	if otpGen == nil {
		otpGen = generateOTPCode
	}
	expiry := opts.OTPExpiry
	if expiry <= 0 {
		expiry = defaultOTPExpiry
	}
	maxRetries := opts.MaxChallengeRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxChallengeRetries
	}
	maxAttempts := opts.MaxVerificationAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxVerificationErrors
	}

	return &WhatsAppIdentityService{
		userStore:               userStore,
		otpStore:                otpStore,
		now:                     nowFn,
		idGenerator:             idGen,
		otpGenerator:            otpGen,
		sendOTP:                 opts.SendOTP,
		otpExpiry:               expiry,
		maxChallengeRetries:     maxRetries,
		maxVerificationAttempts: maxAttempts,
		whatsAppMessageTarget:   strings.TrimSpace(opts.WhatsAppMessageTarget),
	}, nil
}

func (s *WhatsAppIdentityService) StartLink(ctx context.Context, userID, phone string) (*WhatsAppLinkChallenge, error) {
	normalizedPhone, err := utils.NormalizeWhatsAppPhone(phone)
	if err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}
	if err := utils.ValidateWhatsAppLinkIdentity(userID, normalizedPhone); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	if err := s.ensureUniquePhone(ctx, userID, normalizedPhone); err != nil {
		return nil, err
	}

	currentTime := s.now()
	retryCount := 1
	latest, err := s.otpStore.GetLatestByUser(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if latest != nil && latest.Phone == normalizedPhone && latest.VerifiedAt == nil && latest.ExpiresAt.After(currentTime) {
		if latest.RetryCount >= s.maxChallengeRetries {
			return nil, utils.NewAppError(429, "otp retry limit reached")
		}
		retryCount = latest.RetryCount + 1
	}

	otpCode, err := s.otpGenerator()
	if err != nil {
		return nil, utils.ErrInternal
	}

	challenge := &repository.OTPChallenge{
		ChallengeID:  s.idGenerator(),
		UserID:       userID,
		Phone:        normalizedPhone,
		OTPCode:      otpCode,
		RetryCount:   retryCount,
		AttemptCount: 0,
		MaxAttempts:  s.maxVerificationAttempts,
		ExpiresAt:    currentTime.Add(s.otpExpiry),
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
	}

	if err := s.otpStore.Put(ctx, challenge); err != nil {
		return nil, utils.ErrInternal
	}
	messageText := buildWhatsAppLinkMessage(challenge.OTPCode)
	messageLink := buildWhatsAppMessageLink(s.whatsAppMessageTarget, messageText)

	return &WhatsAppLinkChallenge{
		ChallengeID: challenge.ChallengeID,
		Phone:       challenge.Phone,
		ExpiresAt:   challenge.ExpiresAt,
		RetryCount:  challenge.RetryCount,
		MessageText: messageText,
		MessageLink: messageLink,
	}, nil
}

func (s *WhatsAppIdentityService) VerifyLink(ctx context.Context, userID, challengeID, otpCode string) (*WhatsAppWriteEligibility, error) {
	if err := utils.ValidateOTPChallengeID(challengeID); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}
	if err := utils.ValidateOTPCode(otpCode); err != nil {
		return nil, utils.WrapError(utils.ErrBadRequest, err)
	}

	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	challenge, err := s.otpStore.GetByID(ctx, challengeID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if challenge == nil {
		return nil, utils.ErrNotFound
	}
	if challenge.UserID != userID {
		return nil, utils.ErrForbidden
	}

	currentTime := s.now()
	maxAttempts := challenge.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = s.maxVerificationAttempts
		challenge.MaxAttempts = maxAttempts
	}
	if challenge.VerifiedAt != nil {
		return nil, utils.NewAppError(400, "otp challenge already verified")
	}
	if !challenge.ExpiresAt.After(currentTime) {
		return nil, utils.NewAppError(410, "otp challenge has expired")
	}
	if challenge.AttemptCount >= maxAttempts {
		return nil, utils.NewAppError(429, "otp verification attempts exceeded")
	}

	if challenge.OTPCode != strings.TrimSpace(otpCode) {
		challenge.AttemptCount++
		challenge.UpdatedAt = currentTime
		if err := s.otpStore.Put(ctx, challenge); err != nil {
			return nil, utils.ErrInternal
		}
		if challenge.AttemptCount >= maxAttempts {
			return nil, utils.NewAppError(429, "otp verification attempts exceeded")
		}
		return nil, utils.NewAppError(401, "invalid otp code")
	}

	if err := s.ensureUniquePhone(ctx, userID, challenge.Phone); err != nil {
		return nil, err
	}

	previousLinkedPhone := strings.TrimSpace(user.WhatsAppLinkedPhone)
	user.WhatsAppLinkedPhone = challenge.Phone
	if previousLinkedPhone == "" || previousLinkedPhone != challenge.Phone || user.WhatsAppLinkedAt == nil {
		user.WhatsAppLinkedAt = &currentTime
	}
	user.WhatsAppVerifiedAt = &currentTime

	if err := s.userStore.Put(ctx, user); err != nil {
		return nil, utils.ErrInternal
	}

	challenge.VerifiedAt = &currentTime
	challenge.UpdatedAt = currentTime
	if err := s.otpStore.Put(ctx, challenge); err != nil {
		return nil, utils.ErrInternal
	}

	return s.GetWriteEligibility(ctx, userID)
}

func (s *WhatsAppIdentityService) GetWriteEligibility(ctx context.Context, userID string) (*WhatsAppWriteEligibility, error) {
	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	if user.IsWhatsAppWriteEligible() {
		return &WhatsAppWriteEligibility{
			Eligible:    true,
			LinkedPhone: user.WhatsAppLinkedPhone,
			VerifiedAt:  user.WhatsAppVerifiedAt,
		}, nil
	}

	reason := "whatsapp identity is not verified"
	if strings.TrimSpace(user.WhatsAppLinkedPhone) == "" {
		reason = "whatsapp phone is not linked"
	}
	return &WhatsAppWriteEligibility{
		Eligible:    false,
		Reason:      reason,
		LinkedPhone: user.WhatsAppLinkedPhone,
		VerifiedAt:  user.WhatsAppVerifiedAt,
	}, nil
}

func (s *WhatsAppIdentityService) RequireWriteEligible(ctx context.Context, userID string) error {
	eligibility, err := s.GetWriteEligibility(ctx, userID)
	if err != nil {
		return err
	}
	if eligibility.Eligible {
		return nil
	}
	if eligibility.Reason != "" {
		return utils.NewAppError(403, eligibility.Reason)
	}
	return utils.NewAppError(403, "whatsapp identity verification is required")
}

func (s *WhatsAppIdentityService) DisconnectLink(ctx context.Context, userID string) (*WhatsAppWriteEligibility, error) {
	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	user.WhatsAppLinkedPhone = ""
	user.WhatsAppLinkedAt = nil
	user.WhatsAppVerifiedAt = nil
	if err := s.userStore.Put(ctx, user); err != nil {
		return nil, utils.ErrInternal
	}

	return s.GetWriteEligibility(ctx, userID)
}

func (s *WhatsAppIdentityService) VerifyLinkFromInbound(ctx context.Context, fromPhone, text string) (bool, error) {
	normalizedPhone, err := utils.NormalizeWhatsAppPhone(fromPhone)
	if err != nil {
		return false, nil
	}

	challenge, err := s.otpStore.GetLatestByPhone(ctx, normalizedPhone)
	if err != nil {
		return false, utils.ErrInternal
	}
	if challenge == nil || challenge.VerifiedAt != nil {
		return false, nil
	}

	currentTime := s.now()
	if !challenge.ExpiresAt.After(currentTime) {
		return false, nil
	}

	expectedMessage := buildWhatsAppLinkMessage(challenge.OTPCode)
	if strings.TrimSpace(text) != expectedMessage {
		return false, nil
	}

	user, err := s.userStore.GetByID(ctx, challenge.UserID)
	if err != nil {
		return false, utils.ErrInternal
	}
	if user == nil {
		return false, utils.ErrUnauthorized
	}

	if err := s.ensureUniquePhone(ctx, user.UserID, normalizedPhone); err != nil {
		return false, err
	}

	previousLinkedPhone := strings.TrimSpace(user.WhatsAppLinkedPhone)
	user.WhatsAppLinkedPhone = normalizedPhone
	if previousLinkedPhone == "" || previousLinkedPhone != normalizedPhone || user.WhatsAppLinkedAt == nil {
		user.WhatsAppLinkedAt = &currentTime
	}
	user.WhatsAppVerifiedAt = &currentTime
	if err := s.userStore.Put(ctx, user); err != nil {
		return false, utils.ErrInternal
	}

	challenge.VerifiedAt = &currentTime
	challenge.UpdatedAt = currentTime
	if err := s.otpStore.Put(ctx, challenge); err != nil {
		return false, utils.ErrInternal
	}

	return true, nil
}

func (s *WhatsAppIdentityService) ensureUniquePhone(ctx context.Context, userID, phone string) error {
	existing, err := s.userStore.GetByWhatsAppPhone(ctx, phone)
	if err != nil {
		return utils.ErrInternal
	}
	if existing != nil && existing.UserID != userID {
		return utils.NewAppError(409, "whatsapp phone is already linked to another user")
	}
	return nil
}

func generateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func buildWhatsAppLinkMessage(code string) string {
	return fmt.Sprintf("PROPTI LINK %s", strings.TrimSpace(code))
}

func buildWhatsAppMessageLink(target, text string) string {
	target = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(target)), "whatsapp:"))
	target = strings.TrimPrefix(target, "+")
	if target == "" || strings.TrimSpace(text) == "" {
		return ""
	}
	return fmt.Sprintf("https://wa.me/%s?text=%s", target, url.QueryEscape(text))
}
