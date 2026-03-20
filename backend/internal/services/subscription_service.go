package services

import (
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

// IsPremiumEntitled returns true only if the user has an active premium subscription
// that has not yet expired.
func IsPremiumEntitled(user *models.User, now time.Time) bool {
	if user == nil {
		return false
	}
	if user.Subscription.Tier != models.SubscriptionPremium {
		return false
	}
	if user.Subscription.RenewDate == nil {
		return false
	}
	return now.Before(*user.Subscription.RenewDate)
}

// DeriveSubscriptionStatus derives the user-facing subscription status from user data.
// Returns "active", "expiring_soon", or "expired" for premium users; "" for free users.
func DeriveSubscriptionStatus(user *models.User, now time.Time) models.SubscriptionStatus {
	if user == nil || user.Subscription.Tier != models.SubscriptionPremium {
		return ""
	}

	if user.Subscription.RenewDate == nil {
		return models.SubscriptionExpired
	}

	remaining := user.Subscription.RenewDate.Sub(now)

	if remaining < 0 {
		return models.SubscriptionExpired
	}
	if remaining <= 7*24*time.Hour {
		return models.SubscriptionExpiringSoon
	}
	return models.SubscriptionActive
}

// CanInitiateRenewal returns true when the user is allowed to pay for a renewal.
// Renewal is open when the subscription is expiring_soon (within 7 days) or expired.
func CanInitiateRenewal(user *models.User, now time.Time) bool {
	if user == nil || user.Subscription.Tier != models.SubscriptionPremium {
		return false
	}
	status := DeriveSubscriptionStatus(user, now)
	return status == models.SubscriptionExpiringSoon || status == models.SubscriptionExpired
}
