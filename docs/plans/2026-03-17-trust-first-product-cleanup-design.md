# Trust-First Product Cleanup Design

## Problem

The current product experience mixes promising product behavior with trust-risky messaging and late-stage friction:

- the landing page uses hardcoded scale/satisfaction stats and testimonial-style content that is not a good fit for a new app
- `Boost / Iklan Unggulan` exists technically, but its value and payment model are not explained clearly enough
- seller phone numbers are exposed too directly on listing detail pages
- anti-abuse controls for contact access are not strong enough
- the create-listing flow informs users about listing limits too late
- several UI labels and interactions still feel rough (`fasilitas`, location suggestion wording, parsed-result review flow, broken `Cara Kerja` connector)

The goal is to make the product feel more credible, safer, and clearer without overclaiming marketplace scale.

## Design Direction

Use a trust-first redesign:

- replace social-proof claims with product proof
- keep monetization real, but move it out of the landing page until it is explained better
- protect seller contact details behind authenticated, tracked actions
- surface seller limits early instead of failing at the end
- tighten language and interaction details across the create and listing detail flows

## Approved Sections

### 1. Landing page trust-first redesign

Replace hardcoded scale claims and testimonials with product-proof sections.

#### Changes

- remove hardcoded hero stats like `10.000+ Properti`, `5.000+ Pengguna`, and `99% Kepuasan`
- remove fake or placeholder testimonials
- keep and elevate the AI parsing demo as the main proof block
- hide `Boost / Iklan Unggulan` from the landing page for now
- rewrite the features area to focus on:
  - AI parser
  - moderation and safety
  - clean, searchable property presentation
  - simple listing insight
- replace testimonial space with one of:
  - early-seller value proposition
  - a clearer explanation of what happens after pasting ad text

#### Copy direction

- avoid scale claims that a new app cannot defend
- emphasize honesty and low friction:
  - `gratis 3 listing pertama`
  - `review dulu sebelum tayang`
  - `AI bantu rapikan detail dari chat`
  - `setiap listing melewati moderasi`

#### Visual fix

- repair the broken green `Cara Kerja` connector by replacing the fragile single absolute line with a more robust per-step connector/layout approach

### 2. Contact privacy, seller identity, and anti-spam

Move from passive public exposure to gated, tracked reveal.

#### Public listing detail behavior

- do not expose the full seller phone number in the public listing-detail payload
- always show seller name on the listing detail page
- show either masked contact or a login-gated contact state

#### Contact flow

1. user opens listing detail
2. seller name is visible
3. unauthenticated users see `Login untuk lihat nomor & hubungi`
4. authenticated users click `Lihat nomor penjual`
5. frontend calls a protected reveal endpoint
6. backend logs the reveal event and returns contact data
7. UI reveals the number and enables `WhatsApp` / `Telepon`

#### Recommended API shape

- `POST /listings/:id/contact/reveal`
  - auth required
  - logs reveal attempt
  - returns seller name, phone, and a contact event/session identifier
- `POST /listings/:id/contact/click`
  - auth required
  - body indicates `whatsapp` or `call`
  - logs actual intent before redirect/open action

#### Anti-abuse baseline

- require login for reveal
- rate limit by `userId + IP`
- recommended starting limits:
  - 10 reveals per hour
  - 30 reveals per day
  - 5 rapid attempts per minute
- block self-contact on own listings
- log suspicious behavior for review

#### Important note

This materially reduces easy scraping, but does not make scraping impossible for determined logged-in attackers. It is still the correct first practical layer.

### 3. Create-listing clarity and form polish

Fix late-stage surprises and rough UI wording in the seller flow.

#### Early quota check

- check listing eligibility immediately on `/listings/create` or from the first create CTA
- if the user already reached the listing limit, block from the start instead of only failing at submit time
- show a clear blocked state with:
  - `Kuota listing Anda sudah habis`
  - `Paket gratis hanya mendukung hingga 3 listing aktif.`
  - CTA: `Upgrade ke Premium`
  - secondary CTA: `Kembali ke iklan saya`

#### Parsing/result review polish

- after clicking `Gunakan hasil parsing`, scroll to the top of the form so users naturally review before submit

#### Copy and label fixes

- rename `Saran perbaikan lokasi` to `Saran lokasi`
- never show underscored facility keys in UI
- always map amenity ids to clean display labels in:
  - parse preview
  - selected chips
  - listing detail page
- if an unknown amenity id appears, use a safe formatter fallback

#### UX consistency

- if quota is already exhausted, disable both parse and manual create options on the start screen
- keep the explanation consistent across inline messaging, toasts, and upgrade/paywall messaging

### 4. Boost/payment clarity and analytics tweaks

Keep the functionality, but clarify its scope and stop advertising it too early.

#### Product reality

`Boost / Iklan Unggulan` is already backed by a paid feature-listing flow and featured listings are prioritized in homepage ordering. The problem is clarity, not existence.

#### Recommended positioning

- remove boost from the landing page
- keep boost in seller-facing surfaces:
  - `Iklan Saya`
  - owner listing detail
  - premium/monetization modal

#### Required explanation

Use plain language:

- `Sorot listing ini agar tampil lebih menonjol`
- `Berbayar dan berlaku untuk listing ini saja`

#### Suggested owner flow

- provide a clear owner CTA like `Jadikan Unggulan`
- before payment, show:
  - benefit
  - duration
  - price
  - one-listing scope
- after payment success, show:
  - `Iklan unggulan aktif sampai ...`

#### Analytics reframing

Avoid broad “analytics” language for a new app. Reframe as simple listing insight:

- `Dilihat`
- `Disimpan`
- `Nomor dibuka`
- `Klik WhatsApp`
- `Klik Telepon`

Recommended copy direction:

- from `Analitik Gratis`
- to `Insight Iklan`

and:

- from `Data real-time untuk semua pengguna`
- to `Lihat respons dasar terhadap iklan Anda`

## Implementation Priorities

### Priority 1

- landing page trust rewrite
- remove fake stats and testimonials
- hide boost from landing page
- fix broken `Cara Kerja` connector

### Priority 2

- gated contact reveal
- seller name exposure
- tracked contact events
- baseline rate limiting

### Priority 3

- early listing-limit block
- create-flow scroll-to-top after parse
- location and amenity label cleanup

### Priority 4

- seller-facing boost clarity
- analytics terminology and metrics cleanup

## Success Criteria

- landing page no longer depends on exaggerated social proof
- boost remains a real monetization feature but is explained clearly where relevant
- seller phone is not publicly exposed by default
- contact intent becomes measurable
- users understand listing limits before investing effort
- facility/location wording feels polished and intentional
