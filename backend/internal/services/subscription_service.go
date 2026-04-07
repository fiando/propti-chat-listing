package services

import (
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type TierEntitlements struct {
	Tier                  models.SubscriptionTier
	PriceIDR              int
	ActiveListingCap      int
	PhotoCapPerListing    int
	WhatsAppReadAllowed   bool
	WhatsAppCreateAllowed bool
	VoiceMinutesPerMonth  int
}

var tierEntitlements = map[models.SubscriptionTier]TierEntitlements{
	models.SubscriptionFree: {
		Tier:                  models.SubscriptionFree,
		PriceIDR:              0,
		ActiveListingCap:      3,
		PhotoCapPerListing:    3,
		WhatsAppReadAllowed:   false,
		WhatsAppCreateAllowed: true,
		VoiceMinutesPerMonth:  0,
	},
	models.SubscriptionBasic: {
		Tier:                  models.SubscriptionBasic,
		PriceIDR:              59000,
		ActiveListingCap:      6,
		PhotoCapPerListing:    8,
		WhatsAppReadAllowed:   true,
		WhatsAppCreateAllowed: true,
		VoiceMinutesPerMonth:  20,
	},
	models.SubscriptionPremium: {
		Tier:                  models.SubscriptionPremium,
		PriceIDR:              129000,
		ActiveListingCap:      20,
		PhotoCapPerListing:    15,
		WhatsAppReadAllowed:   true,
		WhatsAppCreateAllowed: true,
		VoiceMinutesPerMonth:  60,
	},
	models.SubscriptionPro: {
		Tier:                  models.SubscriptionPro,
		PriceIDR:              199000,
		ActiveListingCap:      50,
		PhotoCapPerListing:    20,
		WhatsAppReadAllowed:   true,
		WhatsAppCreateAllowed: true,
		VoiceMinutesPerMonth:  120,
	},
}

func TierEntitlementFor(tier models.SubscriptionTier) TierEntitlements {
	entitlement, ok := tierEntitlements[tier]
	if ok {
		return entitlement
	}
	return tierEntitlements[models.SubscriptionFree]
}

func IsPaidTier(tier models.SubscriptionTier) bool {
	return tier != models.SubscriptionFree
}

func IsSubscriptionEntitled(user *models.User, now time.Time) bool {
	if user == nil {
		return false
	}
	if !IsPaidTier(user.Subscription.Tier) {
		return false
	}
	if user.Subscription.RenewDate == nil {
		return false
	}
	return now.Before(*user.Subscription.RenewDate)
}

// IsPremiumEntitled is kept for backward compatibility; paid-tier entitlement is now tier-based.
func IsPremiumEntitled(user *models.User, now time.Time) bool {
	return IsSubscriptionEntitled(user, now)
}

// DeriveSubscriptionStatus derives the user-facing subscription status from user data.
// Returns "active", "expiring_soon", or "expired" for paid users; "" for free users.
func DeriveSubscriptionStatus(user *models.User, now time.Time) models.SubscriptionStatus {
	if user == nil || !IsPaidTier(user.Subscription.Tier) {
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
	if user == nil || !IsPaidTier(user.Subscription.Tier) {
		return false
	}
	status := DeriveSubscriptionStatus(user, now)
	return status == models.SubscriptionExpiringSoon || status == models.SubscriptionExpired
}
