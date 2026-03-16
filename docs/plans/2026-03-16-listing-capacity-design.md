# Listing Capacity Expansion Design

## Problem

Subscribed sellers need a way to post more property listings than the current fixed cap allows. Today the data model and business rules are too narrow for that:

- subscription state is binary-oriented (`free` vs `premium`)
- usage tracking only exposes `monthlyListingsUsed`
- the current rules do not model purchased quota, quota packages, or agency-style limits

This makes growth difficult because every future change would otherwise require hardcoding new exceptions into the listing creation flow.

## Goals

- let paid users increase listing capacity without replacing the entire subscription system each time
- keep the base free/premium story understandable
- preserve a path for temporary capacity boosts and larger professional accounts
- design for provider-independent billing products so payment provider changes do not force quota redesign

## Non-goals

- implementing new quota rules now
- changing payment provider behavior in this document
- defining final pricing numbers in code

## Recommended approach

Use a hybrid quota model built around three layers:

1. **Base plan quota**
   - every subscription tier has a configurable included monthly listing allowance
   - example: Free = 1, Premium = 5

2. **Purchased add-on quota**
   - users can buy extra posting capacity in packs
   - packs can represent either posting credits or temporary active-slot expansions

3. **Manual/custom quota override**
   - support an internal override for agencies, promotions, or enterprise users
   - this avoids forcing every special case into public pricing immediately

## Product behaviors

### A. Quota types

There are two reasonable quota semantics:

- **Posting credits**: one unit is consumed when a new listing is created
- **Active slot capacity**: the user can keep up to N active listings live at once

Recommendation: begin with **posting credits** if the goal is simple upsell and low operational friction. Use **active slots** only if business wants stricter inventory control.

### B. Effective quota calculation

At runtime the product should reason about a single effective allowance:

`effective_capacity = base_plan_quota + active_add_on_quota + manual_override`

The listing creation policy should then compare current usage against this effective value instead of branching only on free/premium status.

### C. Reset and expiry policy

Recommended default:

- base plan quota resets monthly with the subscription cycle
- add-on quota does **not** reset with the same monthly cycle automatically
- instead, each add-on product defines its own expiration rule, for example:
  - 30-day extra pack
  - 90-day seasonal pack
  - non-expiring enterprise allocation

This makes the product more flexible and easier to explain in sales campaigns.

## Suggested data model evolution

### Current model constraints

Current subscription fields are too limited:

- `tier`
- `monthlyListingsUsed`
- `renewDate`

### Proposed direction

Keep the existing subscription object, but extend the account model with explicit quota fields.

Example conceptual shape:

```ts
subscription: {
  tier: 'free' | 'premium' | 'agency'
  baseMonthlyListingQuota: number
  monthlyListingsUsed: number
  renewDate?: string
}
listingCapacity: {
  addOnQuotaRemaining: number
  addOnQuotaExpiresAt?: string
  manualQuotaOverride?: number
  effectiveQuota: number
}
```

Alternative: track add-ons as purchase records and compute quota server-side rather than storing all derived fields on the user directly.

Recommendation: store purchase records as the source of truth, and cache derived capacity fields on the user only if needed for performance.

## API design ideas

### Read APIs

Potential new read surfaces:

- `GET /users/me/listing-capacity`
- include capacity summary in `GET /auth/user`

The UI needs a clean answer to:

- how many listings are included
- how many have been used
- how many extra are available
- when any extra quota expires

### Purchase APIs

Potential new purchase surfaces:

- `POST /premium/listing-capacity/purchase`
- `GET /premium/listing-capacity/products`

These should be provider-agnostic at the application layer, similar to the existing wrapper direction used for payments.

## UX ideas

### Profile / Settings

Add a dedicated capacity card showing:

- included quota
- used quota
- remaining quota
- active add-on packs
- a CTA such as `Tambah Kapasitas Iklan`

### Listing creation

If the seller is near or above the limit:

- show current remaining capacity before submit
- explain why posting is blocked
- give a direct upgrade/add-on path instead of a generic premium error

### Pricing / Upsell

The key UX principle is to distinguish clearly between:

- the base subscription plan
- extra capacity packs
- special/custom agency allocations

## Operational notes

- promotions can be modeled as bonus quota packs instead of bespoke code paths
- enterprise sales can use manual override before a full agency product exists
- finance/reporting should track purchased quota separately from subscription revenue

## Risks

- unclear quota semantics can confuse users
- if posting credits and active slots are mixed too early, support burden rises quickly
- if add-on expiry rules are inconsistent, disputes become likely

## Recommended rollout path

1. keep Free and Premium base plans
2. define one included premium quota number
3. launch one extra-capacity add-on pack
4. expose a capacity summary in profile/settings
5. add an internal manual override path for special accounts
6. only after that, decide whether a formal Agency tier is needed

## What to implement later

When you want to build this, the implementation should start with:

- refactoring listing-capacity logic away from simple free/premium branching
- defining a source-of-truth quota model in backend data
- exposing that capacity in auth/profile responses
- then updating purchase UX and seller messaging

This sequence keeps the system flexible without hardcoding another temporary cap exception.
