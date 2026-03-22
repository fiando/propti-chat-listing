# Demand Acquisition Plan: Solving the Ghost Listing Problem

## Problem

Propti is a new marketplace. Sellers who consider listing here face a rational concern:

> "Belum ada userbase yang siap cari properti di sini — listing saya bisa jadi listing hantu."

This is the classic chicken-and-egg marketplace problem. The plan below breaks the cycle without requiring either side to act on faith. Instead of waiting for organic mass adoption, Propti generates real buyer intent from day one by piggybacking on existing buyer behavior — primarily Google search and WhatsApp sharing — rather than requiring buyers to discover Propti directly.

---

## Core Strategy: Bypass the Chicken-and-Egg Problem

Most marketplace demand strategies try to build both sides simultaneously. That is slow and expensive. Propti's approach is different:

**Buyers don't need to know Propti exists. They just need to find the listing.**

This means:
- Every Propti listing must be indexable and rankable in Google
- Every seller must be able to share their listing to their own WhatsApp audience
- Propti's growth in buyer traffic is a byproduct of seller distribution, not the precondition for it

This inverts the typical problem: instead of needing buyers before sellers come, Propti lets sellers bring buyers with them.

---

## Phase 1 — Search-First Visibility (Month 1–2)

### Goal

Ensure every published listing can be discovered via Google organic search without any Propti brand recognition.

### Implementation

#### Listing page SEO

Each listing URL at `/listings/[id]` must include:

- `<title>` tag: `{title} — {harga} | Propti`
- `<meta name="description">`: first 2 sentences of description with price and location
- Open Graph tags for WhatsApp/social link previews (title, price, main image)
- Structured data using `schema.org/RealEstateListing` or `Product` for rich results
- Canonical URL to prevent duplicate content from filtering variants

#### Sitemap and indexing

- Generate `sitemap.xml` with all active listing URLs
- Submit sitemap to Google Search Console
- Set `robots.txt` to allow full crawling of `/listings/` paths
- Avoid `noindex` on listing pages; only apply it to internal search result pages if needed

#### Slug-based URLs

Replace opaque IDs with descriptive paths:

- Current: `/listings/abc123`
- Target: `/listings/rumah-dijual-3kt-di-cibubur-abc123`

Slug format: `{tipe}-{transaksi}-{kamar}kt-di-{kelurahan-or-kecamatan}-{short-id}`

This improves CTR in search results and matches buyer query patterns.

#### Location and type landing pages

Create static or SSR pages for common search combinations:

- `/jual/rumah/jakarta-selatan`
- `/sewa/apartemen/bandung`
- `/jual/tanah/surabaya`

These pages act as SEO aggregators that feed traffic into filtered listing results.

### Success metrics

- Listings indexed in Google within 14 days of publishing
- At least one organic search session per active listing per month
- Landing page impressions in Google Search Console for location + property type queries

---

## Phase 2 — Seller-Led Distribution (Month 1–3)

### Goal

Turn every seller into a demand channel. Sellers already share their listings in WhatsApp groups, neighborhood chats, and social media. Propti should be the tool that makes their listing the best-looking thing in that conversation.

### Implementation

#### WhatsApp-optimized link preview

When a seller shares a Propti listing URL in WhatsApp, the preview must show:

- Property title
- Price formatted in Rupiah (e.g., `Rp 850 jt`)
- Main listing photo
- Location (kota / kecamatan)

This requires correct Open Graph meta tags on every listing page. Buyers who tap the link land directly on the listing and see a CTA to reveal the seller's contact.

#### Share button in seller dashboard

In the seller's listing management view (`/iklan-saya`), add a share button per listing that:

- Copies the listing URL to clipboard
- Opens WhatsApp with a pre-filled message: `Cek properti ini di Propti: [link]`
- Optionally generates a summary card image (Phase 2 extension)

#### Share prompt after listing goes live

After moderation approves a listing, show the seller an explicit share prompt:

> "Listing Anda sudah tayang! Bagikan ke grup WhatsApp Anda untuk lebih banyak calon pembeli."

Include one-tap share buttons: WhatsApp, copy link.

### Success metrics

- Percentage of new listings that receive at least one external referral session within 7 days
- Click-through rate on WhatsApp-previewed listing links
- Referral traffic from WhatsApp and direct (typed URLs) in analytics

---

## Phase 3 — Seed Supply from Quality Sellers (Month 1–6)

### Goal

Build a critical mass of high-quality listings before organic buyer traffic reaches scale. A curated, well-photographed, accurately-priced listing set makes Propti worth revisiting for buyers even at low volume.

### Implementation

#### Agent and developer outreach program

Target small-to-mid agents (1–10 agents in one area) and small property developers with:

- Offer: free Premium access for first 3 months in exchange for listing active inventory
- Requirement: minimum 5 listings, photos, accurate pricing
- Benefit framing: "Anda dapat listing profesional gratis, lebih terlihat dari OLX, dan analytics yang OLX tidak kasih"

Agents list everywhere anyway — Propti gets a copy of their inventory at no extra work for them.

#### Targeted city-by-city rollout

Instead of spreading thin across all of Indonesia, concentrate supply and demand efforts in 1–2 cities at a time:

1. Start: Jabodetabek (highest property search volume)
2. Expand: Bandung, Surabaya
3. Later: Tier-2 cities

Concentrated supply in one city makes Propti genuinely useful for buyers in that city before it needs national scale.

#### Minimum viable inventory target

Define a threshold per city that makes the search experience non-empty:

- Target: 50+ active listings per major city before promotional push
- Listings should span at least 3 property types (rumah, apartemen, tanah) and 3 price bands

### Success metrics

- Active listing count per city
- Photo coverage rate (listings with 3+ photos)
- Price accuracy rate (confirmed by agent partner review)

---

## Phase 4 — Buyer Retention and Return Loops (Month 3–6)

### Goal

Once buyers visit via search or seller sharing, give them a reason to come back directly rather than only arriving via external links.

### Implementation

#### Saved search and listing alerts

Allow buyers to save a search (e.g., "rumah 3 kamar di Depok, max Rp 1M") and receive a notification when a new listing matches. This creates pull — buyers return to Propti for new listings without needing a new marketing touchpoint.

Initial implementation: email notification for new matches once per day.

#### Listing save / wishlist

Allow authenticated buyers to save listings to a shortlist. Sellers see their listing's save count in analytics. This creates value for both sides:

- Buyers: easier comparison and shortlisting
- Sellers: visible signal that there is real interest even before contact reveal

#### Verified listing badge

Listings where the seller's identity is confirmed (phone verified, listing photos reviewed) get a "Terverifikasi" badge. This gives buyers confidence to initiate contact and differentiates Propti listings from unverified OLX posts.

### Success metrics

- Percentage of first-visit buyers who return within 30 days
- Save-to-reveal conversion rate per listing
- Email alert click-through rate

---

## Phase 5 — Proof and Conversion Messaging (Month 2–6)

### Goal

Give sellers tangible evidence that Propti delivers real buyer intent, so they renew Premium and refer other sellers.

### Implementation

#### Listing analytics dashboard

For each listing, show:

- Total views
- Number of contact reveals
- Number of WhatsApp clicks
- Number of call clicks
- Number of saves

This data turns abstract "platform value" into a concrete ROI number. A seller who sees "3 nomor dibuka, 2 WhatsApp diklik" understands that those 3 people were serious buyers — a signal that is invisible on OLX.

#### Seller success stories (internal)

Track which listings on Propti result in completed transactions (via seller self-report after contact). Use anonymized data to:

- Show median time-to-first-contact for listings with photos vs without
- Show median time-to-first-contact for Premium vs Free
- Surface this in onboarding and Premium upsell copy

#### Referral program

Active Premium sellers who refer another seller who converts to Premium receive one month of Premium credit. This converts satisfied sellers into a distribution channel.

### Success metrics

- Premium renewal rate
- Seller-to-seller referral conversion
- Time-to-first-contact for new listings

---

## Addressing the Objection Directly: Messaging Framework

When sellers raise the "listing hantu" concern, the product and support responses should be:

### At listing creation

Show expected reach based on location:

> "Di Depok, ada [X] pencarian properti aktif bulan ini. Listing Anda akan muncul di hasil pencarian ini."

_(Initially this is estimated from SEO impression data; later from actual platform search sessions.)_

### In seller onboarding

Reframe the value proposition:

> "Propti bukan hanya tempat listing. Ini alat untuk bikin listing terbaik kamu, terus bagikan ke mana-mana. Nomor kamu aman, dan kamu tahu siapa yang benar-benar minat."

### In premium upsell

Show ROI calculation:

> "Satu pembeli serius yang masuk karena listing Premium Propti = nilai lebih dari berbulan-bulan biaya langganan."

---

## Timeline Summary

| Phase | Period | Focus |
|---|---|---|
| Phase 1 | Month 1–2 | SEO, sitemap, slug URLs, location landing pages |
| Phase 2 | Month 1–3 | Share buttons, WhatsApp preview, post-publish prompt |
| Phase 3 | Month 1–6 | Agent outreach, city-by-city concentration, inventory seeding |
| Phase 4 | Month 3–6 | Saved search alerts, listing saves, verified badge |
| Phase 5 | Month 2–6 | Analytics dashboard, success stories, referral program |

Phases 1 and 2 have the highest leverage because they generate demand without requiring Propti to build its own audience. They should be prioritized immediately after the core listing flow is stable.

---

## Related Plans

- [Competitive Positioning Design](2026-03-22-competitive-positioning-design.md) — why Propti beats free platforms on quality, not on volume
- [Trust-First Product Cleanup](2026-03-17-trust-first-product-cleanup-design.md) — product changes that build seller and buyer trust
- [Listing Capacity Options](2026-03-16-listing-capacity-options.md) — how capacity tiers support seller growth within Propti
