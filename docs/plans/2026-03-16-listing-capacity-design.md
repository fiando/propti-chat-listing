# Listing Capacity Expansion Design

## Updated direction

The public pricing model should remain exactly:

- **Free**
- **Premium**

No third public subscription tier should be introduced.

Extra paid listing capacity should exist only as a **Premium-only expansion mechanism**, not as a replacement for Premium.

## Core product principle

Premium must feel valuable on its own.

That means users should subscribe because Premium already offers a clearly better seller experience. Extra credits should only appear after that point, when a Premium user needs more volume than their included allowance.

In other words:

- Free is the entry point
- Premium is the main monetized seller plan
- extra capacity is a usage expansion inside Premium

## Problem

Today the model is too rigid:

- subscription state is mostly binary (`free` vs `premium`)
- usage tracking only exposes `monthlyListingsUsed`
- listing creation rules do not support flexible purchased capacity

At the same time, opening credits to everyone would damage Premium positioning. If a free user can simply buy more quota directly, Premium risks feeling optional or weak.

## Recommended revenue model

### 1. Free plan

Free should stay intentionally constrained:

- low listing quota
- basic media and seller tools
- strong upgrade path into Premium

Its role is activation, not scale.

### 2. Premium plan

Premium should be the real seller plan.

It should include:

- a meaningful monthly included listing allowance
- stronger listing presentation value
- better media limits
- access to seller-focused perks
- access to purchase extra capacity packs

The key is that Premium must already feel worth buying *even if the user never buys extra credits*.

### 3. Premium-only extra capacity packs

Extra capacity should be purchasable only by Premium users.

This protects Premium because:

- top-ups become an upsell after subscription
- top-ups cannot be used as a subscription substitute
- high-volume sellers expand revenue without forcing a new public tier

## Why this is the best approach

This model creates a strong revenue ladder:

1. a free seller hits natural product limits
2. they upgrade to Premium to unlock the full seller experience
3. if they become more active, they buy extra capacity inside Premium

That sequence is much healthier than letting free users bypass Premium with credits.

## What Premium must uniquely own

To avoid Premium feeling useless, Premium should own multiple advantages that credits alone cannot replace.

Recommended Premium-only value pillars:

- included monthly listing quota
- higher media and richer listing presentation
- discounted or bundled featured-listing promotions
- better seller analytics and lead insights
- faster moderation or priority support path
- trust/status signaling such as a premium seller badge
- access to extra listing packs

These make Premium the strategic product, not just a gate to another purchase.

## Economic design principles

### A. Best unit economics belong to Premium

Premium should already be the cheapest cost-per-listing path for any seller who posts regularly.

That means:

- the monthly subscription should feel cheaper than repeatedly trying to buy around it
- extra packs should be priced as convenience scale for Premium users, not as a cheaper alternative to Premium itself

### B. Top-ups should feel optional, not required

If Premium users immediately feel forced to buy top-ups, Premium will feel underpowered.

So the included Premium allowance should cover the typical active seller. Top-ups should mainly serve:

- seasonal spikes
- campaign periods
- fast-growing sellers
- agencies or teams before a deeper account model exists

### C. Retention rewards should live inside Premium

To make Premium more attractive over time, reward continued subscription.

Examples:

- bonus listing credits after each renewal
- small rollover of unused included quota, capped to prevent abuse
- better top-up pricing after consecutive paid months
- occasional campaign bundles for long-term Premium users

This improves retention without creating another subscription plan.

## Quota semantics

Recommendation: use **posting credits** rather than active-slot expansion for the first version.

Why:

- easier for users to understand
- simpler backend accounting
- cleaner purchase story
- easier to run promotions around

Suggested mental model:

- Premium includes a monthly posting allowance
- extra packs add more posting capacity when needed
- the system calculates a single effective posting balance

## Conceptual data model direction

Keep the public tier model simple, but make quota state more expressive.

Example conceptual shape:

```ts
subscription: {
  tier: 'free' | 'premium'
  baseMonthlyListingQuota: number
  monthlyListingsUsed: number
  renewDate?: string
  loyaltyTierMonths?: number
}
listingCapacity: {
  premiumTopUpCreditsRemaining: number
  premiumTopUpCreditsExpiresAt?: string
  effectiveAvailableCredits: number
}
```

Important product rule:

- if `tier !== premium`, top-up purchase is not allowed

## API design ideas

### Read surfaces

The client should receive a clear capacity summary in profile/auth responses:

- included monthly quota
- used quota
- remaining base quota
- remaining premium top-up credits
- next reset date
- any bonus/loyalty credit status

### Purchase surfaces

Potential API shape:

- `GET /premium/listing-capacity/products`
- `POST /premium/listing-capacity/purchase`

These should remain payment-provider independent.

## UX design ideas

### Pricing page

The pricing page should clearly communicate:

- Free is for trying the platform
- Premium is for active sellers
- extra capacity is available after becoming Premium

### Profile / Settings

Add a capacity section that shows:

- plan quota included this month
- usage so far
- extra credits remaining
- CTA: `Tambah Kapasitas Iklan`

### Listing creation

When a Premium user is close to the limit:

- show remaining capacity early
- explain available top-up options before they fail at submit time

When a Free user hits the cap:

- route them to Premium, not to credits

## Risks to avoid

- making Premium too weak, so top-ups feel mandatory
- letting free users buy around Premium and weakening subscriptions
- creating too many quota types at once
- mixing posting credits and active slots too early

## Recommended refined rollout

1. keep only Free and Premium publicly
2. increase Premium value beyond raw quota alone
3. make extra capacity purchasable only by Premium users
4. expose clear capacity tracking in profile/settings
5. add retention mechanics that reward ongoing Premium membership
6. reserve internal manual overrides for exceptional accounts, but do not market them as a public third plan

## Final recommendation

The best approach for attracting users while protecting Premium is:

- make Premium the only path to serious selling
- make Premium valuable even without top-ups
- sell extra capacity only to Premium users
- give long-term Premium members better economics and loyalty advantages

That combination improves recurring revenue, preserves plan clarity, and keeps Premium from feeling diluted by extra credits.
