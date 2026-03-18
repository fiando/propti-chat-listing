# Subscription expiry and manual renewal design

## Problem

The current premium flow stores a renewal date, but premium entitlement is still treated like a permanent flag in key parts of the system. That creates two user-facing problems:

- users can look permanently premium even after the paid period should end
- users do not have a clear path to continue their subscription when the period is close to ending or already expired

This project currently uses DOKU with manual-style payment methods such as virtual account, Indomaret, and e-wallets. Because there is no credit-card recurring flow in production, the subscription design should assume manual renewal rather than auto-renew.

## Goal

Make premium behave as a fixed-term prepaid plan with:

- clear expiry visibility
- reminder notifications before expiry
- a clear `Perpanjang Premium` renewal path
- strict downgrade to free entitlement after expiry when no new payment succeeds

## Chosen approach

Use **manual renewal with expiry enforcement**.

The user pays for a premium period, sees when it ends, receives reminders before it ends, and can manually renew through the existing payment flow. If the premium period expires without a successful renewal payment, the backend must immediately stop granting premium access.

## Product behavior

### Premium lifecycle

1. User completes a premium payment.
2. Backend sets the premium end date on the user subscription.
3. Frontend shows the current status and expiry date in profile and settings.
4. System sends reminder notifications 7 days before expiry and 1 day before expiry.
5. User can renew at any time by tapping `Perpanjang Premium`.
6. If expiry passes without a successful renewal payment, entitlement falls back to free.
7. If the user renews after expiry, the new premium period starts from payment success time.

### User-visible states

- **Active**
  - copy example: `Premium aktif sampai 25 Apr 2026`
  - show renewal CTA
- **Expiring soon**
  - copy example: `Premium berakhir dalam 7 hari`
  - emphasize renewal CTA
- **Expired**
  - copy example: `Premium telah berakhir`
  - show renewal CTA and free-tier restrictions

## Architecture and components

### Backend entitlement model

The backend must treat premium access as a time-bounded entitlement, not just a tier label.

The entitlement decision should be centralized behind shared logic that checks:

- `subscription.tier`
- `subscription.renewDate` (or a renamed `expiresAt` field if the team later chooses to make naming clearer)
- current time

A user is only premium-entitled when:

- tier is premium
- renewal date exists
- renewal date is in the future

Any existing premium checks that only compare `tier == premium` must be updated to use the shared entitlement rule.

### Derived subscription status

API responses that return current-user or profile subscription information should also return a derived state for presentation:

- `active`
- `expiring_soon`
- `expired`

This avoids duplicating time-window logic in multiple frontend components and keeps user messaging consistent.

### Frontend surfaces

The following surfaces should reflect subscription state consistently:

- profile page
- settings page
- premium upgrade or renewal modal
- any premium gate shown during listing creation or upgrade prompts

Each surface should show:

- current plan
- expiry date when available
- current subscription state
- primary renewal CTA

### Renewal entry point

The primary CTA should be `Perpanjang Premium`.

It should route the user into the existing premium purchase flow, reusing the current payment initiation path rather than creating a separate billing system.

### Copy updates

Remove or replace wording that implies unsupported behavior. In particular, copy such as `Batalkan kapan saja` should be removed unless a real cancellation flow is implemented.

Recommended replacement is copy that reflects the real model, for example:

- `Premium berlaku selama 30 hari sejak pembayaran berhasil`
- `Perpanjang kapan saja sebelum atau sesudah masa aktif berakhir`

## Notification design

### Reminder schedule

Send reminders at:

- 7 days before expiry
- 1 day before expiry

### Delivery priority

Phase 1 should support in-app notification first.

Additional channels such as email or WhatsApp can be added later, but they are not required to fix the current broken subscription flow.

### Notification purpose

Notifications are reminders only. They do not affect entitlement. If notification delivery fails, expiry enforcement still occurs based on subscription time state.

## Payment and renewal rules

### Supported payment model

This design assumes manual renewal using currently supported payment methods:

- virtual account
- Indomaret / convenience-store flow
- e-wallets

True auto-renew is out of scope because the current live payment setup does not support a reliable recurring merchant-initiated charge flow across the active payment methods.

### Renewal timing rule

For the selected flow, renewal after expiry starts a new premium period from payment success time.

If early renewal is allowed later, the team can decide whether to:

- start from payment success time, or
- extend from the current expiry

That behavior is intentionally out of scope for this phase because the chosen product direction is simply manual renewal with expiry enforcement.

## Edge cases

### User has more than free-tier listings after expiry

Do not automatically delete or unpublish listings.

Instead:

- block creating new listings while the account is expired and still above free-tier limits
- allow renewal to restore premium entitlement
- optionally guide the user to reduce listings if they want to remain on free tier

### Missing or invalid premium expiry data

If a user is marked as premium but has no expiry date, or the expiry date is already past, the backend should treat the user as not entitled to premium features and log the inconsistency for investigation.

### Late payment callback

If a valid successful payment callback arrives after the previous premium period already expired, premium should reactivate from the successful payment time.

## Error handling principles

- premium entitlement should fail safe toward free access, not premium access
- payment success is the only event that can create or extend premium entitlement
- data inconsistencies should be logged and observable rather than silently corrected in a way that hides problems

## Testing strategy

### Backend

Add tests for:

- active premium user
- premium user expiring within reminder window
- expired premium user
- premium tier with missing expiry date
- free-tier enforcement after expiry

### Frontend

Add tests for:

- active state UI
- expiring-soon state UI
- expired state UI
- renewal CTA visibility and wording

### Integration

Add an end-to-end or integration-level test confirming:

- an expired user cannot bypass free-tier limits
- the expired user sees a renewal path
- a successful renewal payment restores premium entitlement

## Out of scope

The following are intentionally not included in this design:

- auto-renew with credit cards or tokenized payments
- cancellation flow
- grace period after expiry
- billing history portal
- refunds or proration
- cross-channel notification orchestration beyond a first notification channel

## Expected outcome

After this design is implemented:

- premium no longer appears permanent
- users always know when premium ends
- users receive reminders before expiry
- users can continue their subscription through a clear manual renewal path
- expired users no longer retain premium access indefinitely
