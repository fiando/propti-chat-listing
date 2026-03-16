# Listing Capacity Options

## Current state

The current product model is simple but rigid:

- user subscription only stores `tier`, `monthlyListingsUsed`, and `renewDate`
- listing creation logic explicitly blocks free users after `1` listing per month
- premium is treated mostly as a binary access state instead of a configurable quota product
- there is no concept of purchased quota, carry-over balance, or custom seller plans

This makes Premium easy to explain, but it also makes growth offers hard to model. If a subscribed seller wants more capacity, the product currently has no clean path besides changing the single Premium rule itself.

## Option 1 — Multiple subscription tiers

Examples:

- Free: 1 listing/month
- Premium Basic: 5 listings/month
- Premium Plus: 20 listings/month
- Agency: 100 listings/month

### Pros

- very easy for users to understand
- pricing page is straightforward
- quota reset logic stays subscription-based
- good for predictable recurring revenue

### Cons

- users must upgrade the whole plan even when they only need a temporary increase
- tier migration logic becomes important
- not very flexible for seasonal demand spikes

### Best for

A product that wants a clean SaaS pricing ladder with low operational complexity.

## Option 2 — Pay-per-extra-listing credits

Examples:

- Free or Premium users buy extra listing credits in packs
- each new listing consumes one credit after the included plan quota is exhausted

### Pros

- highly flexible for occasional sellers
- easier upsell for users who do not want a larger recurring plan
- good monetization for demand spikes

### Cons

- balance accounting becomes more complex
- users may be confused about whether they are buying a subscription or prepaid quota
- credit expiry rules need to be clear

### Best for

A marketplace with bursty seller behavior and many occasional payers.

## Option 3 — Hybrid model (recommended)

Examples:

- Free: 1 listing/month
- Premium: included monthly quota, e.g. 5 active listings/month
- Add-on packs: +5, +10, +25 temporary listing slots or posting credits
- Agency / enterprise override: custom negotiated quota

### Why this is the strongest fit

It preserves the simplicity of a recurring Premium product while unlocking flexibility for power sellers. Users can stay on Premium for baseline value, then add extra capacity only when they actually need it. The same model also leaves room for agent teams or agencies later.

### Pros

- flexible without abandoning the existing Premium concept
- easy upgrade story: subscription first, add-ons second
- supports both recurring sellers and seasonal spikes
- gives room for enterprise or admin-managed quota later

### Cons

- most complex option to model correctly
- requires clear UI around included quota vs purchased add-ons
- needs explicit policy decisions for expiry, refunds, and reset behavior

## Recommendation

Use the hybrid model as the main design direction.

Start with a simple structure:

- keep Free and Premium as base plans
- define a configurable included monthly quota per plan
- add one purchasable extra-capacity product type
- reserve a manual custom-quota override for internal/admin use

That gives Propti a path from small sellers to serious agents without needing to rebuild the subscription model each time pricing evolves.
