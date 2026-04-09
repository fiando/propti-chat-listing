package services

import (
	"testing"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

func premiumUser(renewDate time.Time) *models.User {
	return &models.User{
		Subscription: models.Subscription{
			Tier:      models.SubscriptionPremium,
			RenewDate: &renewDate,
		},
	}
}

func basicUser(renewDate time.Time) *models.User {
	return &models.User{
		Subscription: models.Subscription{
			Tier:      models.SubscriptionBasic,
			RenewDate: &renewDate,
		},
	}
}

func freeUser() *models.User {
	return &models.User{
		Subscription: models.Subscription{Tier: models.SubscriptionFree},
	}
}

var now = time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

// ─── IsPremiumEntitled ────────────────────────────────────────────────────────

func TestIsPremiumEntitled_ActivePremium(t *testing.T) {
	user := premiumUser(now.Add(30 * 24 * time.Hour))
	if !IsPremiumEntitled(user, now) {
		t.Fatal("expected entitled for active premium")
	}
}

func TestIsPremiumEntitled_ExpiredPremium(t *testing.T) {
	user := premiumUser(now.Add(-1 * time.Second))
	if IsPremiumEntitled(user, now) {
		t.Fatal("expected not entitled for expired premium")
	}
}

func TestIsPremiumEntitled_NilRenewDate(t *testing.T) {
	user := &models.User{Subscription: models.Subscription{Tier: models.SubscriptionPremium}}
	if IsPremiumEntitled(user, now) {
		t.Fatal("expected not entitled when renewDate is nil")
	}
}

func TestIsPremiumEntitled_FreeUser(t *testing.T) {
	if IsPremiumEntitled(freeUser(), now) {
		t.Fatal("expected not entitled for free user")
	}
}

func TestIsPremiumEntitled_NilUser(t *testing.T) {
	if IsPremiumEntitled(nil, now) {
		t.Fatal("expected not entitled for nil user")
	}
}

func TestIsSubscriptionEntitled_ActiveBasic(t *testing.T) {
	user := basicUser(now.Add(20 * 24 * time.Hour))
	if !IsSubscriptionEntitled(user, now) {
		t.Fatal("expected entitled for active basic")
	}
}

// ─── DeriveSubscriptionStatus ─────────────────────────────────────────────────

func TestDeriveSubscriptionStatus_Active(t *testing.T) {
	user := premiumUser(now.Add(8 * 24 * time.Hour))
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionActive {
		t.Fatalf("expected active, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_ExpiringSoon_7Days(t *testing.T) {
	// Exactly 7 days: boundary is inclusive
	user := premiumUser(now.Add(7 * 24 * time.Hour))
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpiringSoon {
		t.Fatalf("expected expiring_soon at 7-day boundary, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_ExpiringSoon_1Day(t *testing.T) {
	user := premiumUser(now.Add(24 * time.Hour))
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpiringSoon {
		t.Fatalf("expected expiring_soon, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_ExpiryIsToday(t *testing.T) {
	// renewDate == now: 0 remaining → expiring_soon (still active until the second)
	user := premiumUser(now)
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpiringSoon {
		t.Fatalf("expected expiring_soon when renewDate==now, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_Expired(t *testing.T) {
	user := premiumUser(now.Add(-1 * time.Second))
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpired {
		t.Fatalf("expected expired, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_NilRenewDate(t *testing.T) {
	user := &models.User{Subscription: models.Subscription{Tier: models.SubscriptionPremium}}
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpired {
		t.Fatalf("expected expired for nil renewDate, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_FreeUser(t *testing.T) {
	if got := DeriveSubscriptionStatus(freeUser(), now); got != "" {
		t.Fatalf("expected empty status for free user, got %q", got)
	}
}

func TestDeriveSubscriptionStatus_BasicUsesPaidStatusRules(t *testing.T) {
	user := basicUser(now.Add(2 * 24 * time.Hour))
	if got := DeriveSubscriptionStatus(user, now); got != models.SubscriptionExpiringSoon {
		t.Fatalf("expected expiring_soon for basic user near renew date, got %q", got)
	}
}

// ─── CanInitiateRenewal ───────────────────────────────────────────────────────

func TestCanInitiateRenewal_ExpiringSoon(t *testing.T) {
	user := premiumUser(now.Add(3 * 24 * time.Hour))
	if !CanInitiateRenewal(user, now) {
		t.Fatal("expected renewal allowed when expiring_soon")
	}
}

func TestCanInitiateRenewal_Expired(t *testing.T) {
	user := premiumUser(now.Add(-24 * time.Hour))
	if !CanInitiateRenewal(user, now) {
		t.Fatal("expected renewal allowed when expired")
	}
}

func TestCanInitiateRenewal_Active(t *testing.T) {
	user := premiumUser(now.Add(30 * 24 * time.Hour))
	if CanInitiateRenewal(user, now) {
		t.Fatal("expected renewal not allowed when active (>7 days)")
	}
}

func TestCanInitiateRenewal_FreeUser(t *testing.T) {
	if CanInitiateRenewal(freeUser(), now) {
		t.Fatal("expected renewal not allowed for free user")
	}
}

func TestTierEntitlementFor_KnownTiers(t *testing.T) {
	cases := []struct {
		tier         models.SubscriptionTier
		price        int
		activeCap    int
		photoCap     int
		voiceMinutes int
		waRead       bool
		waCreate     bool
	}{
		{tier: models.SubscriptionFree, price: 0, activeCap: 3, photoCap: 3, voiceMinutes: 0, waRead: false, waCreate: true},
		{tier: models.SubscriptionBasic, price: 59000, activeCap: 8, photoCap: 8, voiceMinutes: 20, waRead: true, waCreate: true},
		{tier: models.SubscriptionPremium, price: 129000, activeCap: 20, photoCap: 15, voiceMinutes: 60, waRead: true, waCreate: true},
		{tier: models.SubscriptionPro, price: 199000, activeCap: 50, photoCap: 25, voiceMinutes: 120, waRead: true, waCreate: true},
	}

	for _, tc := range cases {
		entitlement := TierEntitlementFor(tc.tier)
		if entitlement.PriceIDR != tc.price {
			t.Fatalf("tier %s expected price %d got %d", tc.tier, tc.price, entitlement.PriceIDR)
		}
		if entitlement.ActiveListingCap != tc.activeCap {
			t.Fatalf("tier %s expected active cap %d got %d", tc.tier, tc.activeCap, entitlement.ActiveListingCap)
		}
		if entitlement.PhotoCapPerListing != tc.photoCap {
			t.Fatalf("tier %s expected photo cap %d got %d", tc.tier, tc.photoCap, entitlement.PhotoCapPerListing)
		}
		if entitlement.VoiceMinutesPerMonth != tc.voiceMinutes {
			t.Fatalf("tier %s expected voice %d got %d", tc.tier, tc.voiceMinutes, entitlement.VoiceMinutesPerMonth)
		}
		if entitlement.WhatsAppReadAllowed != tc.waRead || entitlement.WhatsAppCreateAllowed != tc.waCreate {
			t.Fatalf("tier %s WA entitlement mismatch: %#v", tc.tier, entitlement)
		}
	}
}
