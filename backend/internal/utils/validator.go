package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/fiando/propti/backend/internal/models"
)

// ValidateCreateListingRequest checks required fields and business rules.
func ValidateCreateListingRequest(req *models.CreateListingRequest) error {
	var errs []string

	if strings.TrimSpace(req.Title) == "" {
		errs = append(errs, "title is required")
	}
	if len(req.Title) > 200 {
		errs = append(errs, "title must be at most 200 characters")
	}
	if req.Price <= 0 {
		errs = append(errs, "price must be greater than 0")
	}
	if req.ListingType != models.ListingTypeSell && req.ListingType != models.ListingTypeRent {
		errs = append(errs, "listingType must be 'sell' or 'rent'")
	}
	if strings.TrimSpace(req.Location.Address) == "" {
		errs = append(errs, "location.address is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// ValidateGoogleAuthRequest checks that either a Google ID token or access token is present.
func ValidateGoogleAuthRequest(req *models.GoogleAuthRequest) error {
	if strings.TrimSpace(req.IDToken) == "" && strings.TrimSpace(req.AccessToken) == "" {
		return errors.New("idToken or accessToken is required")
	}
	return nil
}

// ValidateFeatureListingRequest validates a feature/promotion request.
func ValidateFeatureListingRequest(req *models.FeatureListingRequest) error {
	var errs []string

	if strings.TrimSpace(req.ListingID) == "" {
		errs = append(errs, "listingId is required")
	}
	if req.DurationDays <= 0 {
		errs = append(errs, "durationDays must be greater than 0")
	}
	if req.Type != "featured" && req.Type != "promotion" {
		errs = append(errs, "type must be 'featured' or 'promotion'")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// ValidatePagination clamps page and pageSize to sensible defaults/limits.
func ValidatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// ValidateMediaLimits checks media caps per subscription tier.
func ValidateMediaLimits(maxMediaItems int, tierLabel string, images, videos []string) error {
	total := len(images) + len(videos)
	if total > maxMediaItems {
		return fmt.Errorf("%s tier allows at most %d media items per listing (got %d)", tierLabel, maxMediaItems, total)
	}
	return nil
}

var digitsOnlyRegex = regexp.MustCompile(`^[0-9]+$`)

func NormalizeWhatsAppPhone(phone string) (string, error) {
	trimmed := strings.TrimSpace(phone)
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	if strings.HasPrefix(trimmed, "0") {
		trimmed = "+62" + strings.TrimPrefix(trimmed, "0")
	}
	if !strings.HasPrefix(trimmed, "+") {
		return "", errors.New("phone must use international format")
	}
	digits := strings.TrimPrefix(trimmed, "+")
	if !digitsOnlyRegex.MatchString(digits) {
		return "", errors.New("phone must contain digits only")
	}
	if len(digits) < 10 || len(digits) > 15 {
		return "", errors.New("phone must be 10 to 15 digits")
	}
	return "+" + digits, nil
}

func ValidateWhatsAppLinkIdentity(userID, phone string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user id is required")
	}
	if strings.TrimSpace(phone) == "" {
		return errors.New("phone is required")
	}
	return nil
}

func ValidateOTPCode(code string) error {
	trimmed := strings.TrimSpace(code)
	if len(trimmed) != 6 {
		return errors.New("otp must be 6 digits")
	}
	if !digitsOnlyRegex.MatchString(trimmed) {
		return errors.New("otp must contain digits only")
	}
	return nil
}

func ValidateOTPChallengeID(challengeID string) error {
	if strings.TrimSpace(challengeID) == "" {
		return errors.New("challenge id is required")
	}
	return nil
}
