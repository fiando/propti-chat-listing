package utils

import (
	"errors"
	"fmt"
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

// ValidateGoogleAuthRequest checks that an ID token is present.
func ValidateGoogleAuthRequest(req *models.GoogleAuthRequest) error {
	if strings.TrimSpace(req.IDToken) == "" {
		return errors.New("idToken is required")
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

// ValidateMediaLimits checks the free-tier media cap (3 items per listing).
func ValidateMediaLimits(isPremium bool, images, videos []string) error {
	if isPremium {
		return nil
	}
	total := len(images) + len(videos)
	if total > 3 {
		return fmt.Errorf("free tier allows at most 3 media items per listing (got %d)", total)
	}
	return nil
}
