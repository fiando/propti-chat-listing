================================================================================
IMPLEMENTATION PLANNING DOCUMENT - SUBSCRIPTION RENEWAL & EXPIRY ENFORCEMENT
Specification: /home/bobby/Development/IdeaProjects/saas/propti/docs/superpowers/specs/2026-03-18-subscription-renewal-design.md
================================================================================

EXECUTIVE SUMMARY
─────────────────
This plan maps exact file paths for implementing manual renewal + expiry enforcement.
Focus: 7-day & 1-day reminder windows, subscription status derivation, and renewal gating.

KEY IMPLEMENTATION PRINCIPLES
────────────────────────────
1. Premium entitlement = tier=="premium" AND renewDate!=null AND renewDate>now
2. Three UI states: "active" (>7 days remaining), "expiring_soon" (0-7 days), "expired" (past)
3. Renewal allowed ONLY when in "expiring_soon" or "expired" state
4. If renewing before expiry: extend from existing date (no lost days)
5. If renewing after expiry: start from payment success time
6. All premium checks use centralized IsPremiumEntitled() service

═════════════════════════════════════════════════════════════════════════════════
BACKEND CHANGES (13 files)
═════════════════════════════════════════════════════════════════════════════════

PRIORITY 0 (CORE - MUST DO FIRST)
─────────────────────────────────

[NEW] /backend/internal/services/subscription_service.go
  - IsPremiumEntitled(user, now) → bool
  - DeriveSubscriptionStatus(user, now) → "active"|"expiring_soon"|"expired"
  - CanInitiateRenewal(user, now) → bool

[UPDATE] /backend/internal/models/user.go (lines 23-29)
  - Add SubscriptionStatus type const: "active", "expiring_soon", "expired"

[UPDATE] /backend/internal/services/listing_service.go (line ~115)
  - Replace: `isPremium := user.Subscription.Tier == SubscriptionPremium`
  - Use: `IsPremiumEntitled(user, time.Now())`

[UPDATE] /backend/internal/services/payment_reconciler.go (line ~302)
  - If user.Subscription.RenewDate > now (renewing early):
      newDate = user.Subscription.RenewDate.AddDate(0, 1, 0)  // extend
  - Else:
      newDate = time.Now().AddDate(0, 1, 0)  // start fresh

[UPDATE] /backend/internal/handlers/premium_handler.go
  - upgradeToPremium() line 86: Add CanInitiateRenewal() check
  - Return 400 "Renewal opens 7 days before expiry" if too early

[UPDATE] /backend/internal/handlers/auth_handler.go
  - Profile endpoints: Include "subscriptionStatus" in response JSON

[NEW] /backend/internal/services/subscription_service_test.go
  - Test IsPremiumEntitled with: active, expiring, expired, null renewDate
  - Test DeriveSubscriptionStatus edge cases (7-day boundary)
  - Test CanInitiateRenewal for all user states

[EXTEND] /backend/internal/handlers/premium_handler_test.go
  - Test renewal rejected when not in window
  - Test renewal succeeds with date extension

[EXTEND] /backend/internal/services/listing_service_test.go
  - Update premium checks to use IsPremiumEntitled()
  - Add test: expired premium blocked from listings

[EXTEND] /backend/internal/services/payment_reconciler_test.go
  - Test renewal extends from existing renewDate
  - Test renewal after expiry starts from payment time

PRIORITY 1 (SUPPORTING)
──────────────────────

[UPDATE] /backend/internal/utils/validator.go
  - Update callers to use IsPremiumEntitled()

═════════════════════════════════════════════════════════════════════════════════
FRONTEND CHANGES (14 files)
═════════════════════════════════════════════════════════════════════════════════

PRIORITY 0 (CORE - MUST DO FIRST)
─────────────────────────────────

[UPDATE] /frontend/types/index.ts
  - User.subscription: add "status?: 'active'|'expiring_soon'|'expired'"
  - Add type: SubscriptionStatus = 'active'|'expiring_soon'|'expired'|'free'|'loading'

[UPDATE] /frontend/lib/subscription-status.ts (line ~30)
  - Change return: 'premium'|'free' → 'active'|'expiring_soon'|'expired'|'free'
  - Add helpers: getDaysUntilExpiry(), getExpiryMessage()

[NEW] /frontend/lib/premium-renewal.ts
  - shouldShowRenewalCTA(status) → show button if expiring_soon|expired
  - getRenewalUXCopy(status, renewDate) → localized headings, body, CTA text

[UPDATE] /frontend/components/premium/PremiumUpgradeModal.tsx
  - Support mode prop: 'upgrade' (free→premium) vs 'renew' (extend)
  - Show "Perpanjang Premium" copy, expiry date when renewing
  - Update button: "Upgrade" vs "Perpanjang"

[NEW] /frontend/components/premium/SubscriptionStatusBadge.tsx
  - Display: "Premium aktif sampai 25 Apr 2026" OR
            "Premium berakhir dalam 7 hari" OR
            "Premium telah berakhir"

[UPDATE] /frontend/components/profile/ProfilePageClient.tsx (line ~34)
  - Replace: isPremium = tier === 'premium'
  - Use: subscription status from API
  - Add expiry date display
  - Add SubscriptionStatusBadge component
  - Show renewal CTA if expiring_soon|expired

[UPDATE] /frontend/app/(app)/settings/page.tsx (line ~164)
  - Enhance "Status Akun" section
  - Add expiry date
  - Add "Perpanjang Premium" button if eligible

PRIORITY 1 (SUPPORTING)
──────────────────────

[UPDATE] /frontend/lib/api.ts
  - Add renewPremium() or verify upgradePremium() handles both modes

[EXTEND] /frontend/lib/subscription-status.test.ts
  - Test: expiry 0-7 days → 'expiring_soon'
  - Test: expiry in past → 'expired'
  - Test: expiry 8+ days → 'active'

[NEW] /frontend/lib/premium-renewal.test.ts
  - Test shouldShowRenewalCTA() logic
  - Test getRenewalUXCopy() for each status

[UPDATE] /frontend/app/(app)/profile/page.tsx
  - Ensure ProfilePageClient receives status prop

═════════════════════════════════════════════════════════════════════════════════
API CONTRACTS
═════════════════════════════════════════════════════════════════════════════════

MODIFIED ENDPOINT: GET /users/me
  Backend Handler: /backend/internal/handlers/auth_handler.go
  
  Response (ADD field):
  {
    "user": {
      "userId": "...",
      "subscription": {
        "tier": "premium",
        "renewDate": "2026-04-25T00:00:00Z",
        "activeListingsCount": 5,
        "monthlyListingsUsed": 0
      }
    },
    "subscriptionStatus": "active|expiring_soon|expired|free"
  }

REUSED ENDPOINT: POST /premium/upgrade
  Used for both upgrade (free→premium) and renewal (premium→premium)
  Backend validates CanInitiateRenewal() for existing premium users
  Error response if renewal not yet allowed:
  {
    "error": "Renewal opens 7 days before expiry",
    "code": 400
  }

═════════════════════════════════════════════════════════════════════════════════
BUILD & TEST COMMANDS
═════════════════════════════════════════════════════════════════════════════════

BACKEND TESTS (from /backend directory):
  make test                    # All tests
  make lint                    # Static analysis
  go test ./internal/services/subscription_service_test.go -v
  go test ./internal/services -run Premium -v

FRONTEND TESTS (from /frontend directory):
  npm run lint
  npm run build
  npm test                     # If configured (check package.json)

═════════════════════════════════════════════════════════════════════════════════
MIGRATIONS & SCHEDULED JOBS
═════════════════════════════════════════════════════════════════════════════════

PRE-ROLLOUT DATA AUDIT (ONE-TIME)
  What: Scan all premium users for inconsistent renewDate (missing or past)
  Where: Create audit utility in /backend/internal/utils/subscription_audit.go
  Action: Log inconsistencies for manual review
  Done: Before deploying this feature to production

PHASE 1 (CURRENT)
  Reminder method: In-app status badge ("Premium berakhir dalam 7 hari")
  No scheduled jobs required
  User sees expiry and renewal CTA in profile/settings

PHASE 2 (FUTURE - OUT OF SCOPE)
  Scheduled reminder job:
    - Location: /backend/cmd/reminders/main.go (new Lambda)
    - Trigger: CloudWatch Events daily
    - Logic: Scan premium users, create notification for 7-day & 1-day windows
    - Channel: In-app notifications (push notifications in later phase)

═════════════════════════════════════════════════════════════════════════════════
CRITICAL BOUNDARY CONDITIONS & EDGE CASES
═════════════════════════════════════════════════════════════════════════════════

7-DAY BOUNDARY CALCULATION
  "expiring_soon" if: (renewDate - now) >= 0 AND (renewDate - now) <= 7 days
  Exactly 7 days until expiry: status = "expiring_soon" (renewal opens)
  Exactly 0 days (expiry date = today): status = "expiring_soon" (still active)
  1 second past renewDate: status = "expired"

EXPIRED PREMIUM USER BEHAVIOR
  ✓ Cannot create new listings (blocked by IsPremiumEntitled check)
  ✓ Sees "Premium telah berakhir" badge
  ✓ Sees "Perpanjang Premium" button (renewal allowed anytime after expiry)
  ✓ Existing listings remain published but count against free tier limits

RENEWAL TIMING RULES
  Before expiry: newRenewDate = oldRenewDate + 1 month (extends period, no lost days)
  After expiry:  newRenewDate = paymentSuccessTime + 1 month (restart)
  Late payment callback: Reactivates from callback time if previously expired

MISSING/INVALID RENEWDATE
  Missing renewDate on premium user: Treat as expired (not premium)
  Past renewDate on premium user: Treat as expired
  Null renewDate on free user: OK, expected
  Log: All inconsistencies for audit trail

CONCURRENT ACCESS
  Payment callback arrives while user checking expiry: No race condition
  All reads use consistent timestamp (now = time.Now().UTC() at start)
  Payment updates atomically via DynamoDB Put()

═════════════════════════════════════════════════════════════════════════════════
PRIORITY SEQUENCING
═════════════════════════════════════════════════════════════════════════════════

PHASE 1A (BACKEND CORE):
  1. subscription_service.go [NEW] - 3 core functions
  2. models/user.go - Add SubscriptionStatus type
  3. listing_service.go - Update to use IsPremiumEntitled()
  4. payment_reconciler.go - Fix renewal timing logic
  5. premium_handler.go - Add renewal window validation
  6. auth_handler.go - Include status in response

PHASE 1B (BACKEND TESTS):
  7. subscription_service_test.go [NEW]
  8. premium_handler_test.go [EXTEND]
  9. listing_service_test.go [EXTEND]
  10. payment_reconciler_test.go [EXTEND]

PHASE 2A (FRONTEND CORE):
  11. types/index.ts - Add status field/type
  12. subscription-status.ts - Support 3 new statuses
  13. premium-renewal.ts [NEW] - Renewal helpers
  14. PremiumUpgradeModal.tsx - Support renewal mode
  15. SubscriptionStatusBadge.tsx [NEW] - Display status

PHASE 2B (FRONTEND UI):
  16. ProfilePageClient.tsx - Show status & renewal CTA
  17. settings/page.tsx - Enhance status section

PHASE 2C (FRONTEND TESTS):
  18. subscription-status.test.ts [EXTEND]
  19. premium-renewal.test.ts [NEW]

OPTIONAL:
  20. validator.go - Update media limit callers
  21. Pre-rollout data audit utility

═════════════════════════════════════════════════════════════════════════════════
ACCEPTANCE CRITERIA
═════════════════════════════════════════════════════════════════════════════════

✓ Expired premium user cannot create more listings
✓ Expiring premium user sees "Premium berakhir dalam 7 hari"
✓ Renewal button only appears for expiring/expired state
✓ Early renewal (before 7-day window) returns 400 error
✓ Renewal payment extends from existing expiry (no lost days)
✓ Late renewal (after expiry) starts new 30-day period
✓ All API responses include subscriptionStatus derived field
✓ No premium tier without valid future renewDate
✓ Unit tests cover all boundary cases (exact 7-day edge)
✓ Integration test: full expiry→renewal→reactivation flow

═════════════════════════════════════════════════════════════════════════════════
