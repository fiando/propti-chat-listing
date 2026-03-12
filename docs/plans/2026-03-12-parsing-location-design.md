# Parsing, Wording, and Location Suggestion Design

## Problem

Current listing parsing already helps remove dirty language and organize messy WhatsApp text, but the experience still has several product gaps:

- User-facing wording around "AI parsing" is still too technical for non-technical users.
- Parsing can extract listing details, but location handling is still incomplete:
  - city choices are limited,
  - district selection is still manual,
  - full address often remains abbreviated and messy.
- The product needs a lower-cost GPT-5 option that still performs well for Indonesian property listing parsing.
- Location search must scale to complete Indonesia coverage without hurting performance.

There is also a follow-up issue to address later: moderation currently appears to work during parsing/image flow, but not when the user manually re-enters offensive text before final save.

## Goals

- Make parsing-related wording feel friendly and clear for end users.
- Improve listing parsing quality while keeping model cost low.
- Let AI suggest city, district, and cleaner full-address values without removing user control.
- Support complete Indonesia province/city/district coverage.
- Keep search and suggestion performance fast and deterministic.

## Non-Goals

- Fully autonomous AI location decisions without user confirmation.
- Making Google Maps the primary source of location truth.
- Solving the manual-edit moderation bug in this design; that is tracked as follow-up work.

## Recommended Approach

Use a hybrid architecture:

- `gpt-5-mini` handles the main listing parsing flow.
- A complete Indonesia administrative dataset provides province, city/kabupaten, and district/kecamatan options.
- AI generates location and address suggestions, but the user confirms the final values.
- Google Maps remains optional for geocoding/enrichment, not the primary source of official administrative data.

This balances cost, parsing quality, reliability, and performance better than an AI-heavy or ultra-rule-based design.

## Wording Design

### Principles

- Avoid technical terms such as `parsing`, `parser`, and preferably also `ekstraksi` in user-facing copy.
- Prefer wording that describes the user outcome:
  - `Rapikan otomatis`
  - `Bantu isi detail properti`
  - `Ubah chat jadi data properti`
  - `Susun detail properti`
- If AI is mentioned, position it as support, not the headline concept.

### Suggested Copy Direction

- `Parse dengan AI` -> `Rapikan Otomatis`
- `Mengekstrak informasi dari teks` -> `Sedang merapikan detail dari chat Anda`
- `AI WhatsApp Parser` -> `Ubah Chat WhatsApp Jadi Data Properti`
- Descriptions using `AI parsing` should be rewritten to emphasize outcomes, for example:
  - `Bantu ubah chat properti yang berantakan jadi detail yang rapi dan siap dicek`
- Copy using `ekstrak semua detail` should shift toward:
  - `bantu isi detail properti secara otomatis`

### Consistency Rules

- CTA labels should use clear action verbs such as `Rapikan`, `Isi`, `Ubah`, or `Susun`.
- Descriptions should lead with the result, not the technology.
- Loading, success, and error states should sound helpful and calming.

## Architecture

### Parsing Flow

1. User pastes WhatsApp listing text.
2. `gpt-5-mini` parses the main listing fields:
   - title
   - property type
   - price
   - bedrooms/bathrooms
   - land/building area
   - raw address
   - optional warnings/confidence
3. The backend runs location normalization using the Indonesia administrative dataset.
4. The system prepares suggestion outputs for:
   - city/kabupaten
   - district/kecamatan
   - cleaned full address
5. The UI shows:
   - detected listing details
   - suggested location improvements
6. User confirms or edits before final save.

### Model Choice

- Primary recommendation: `gpt-5-mini`
- Why:
  - better reasoning than the cheapest variants for messy property chats,
  - still cost-efficient,
  - suitable for structured extraction tasks.

`gpt-5-nano` may still be useful later for very narrow sub-tasks, but it is not the recommended main parser for this workflow.

## Components

### Backend

- Update AI parsing service to use `gpt-5-mini`.
- Keep structured output / strict JSON-style response handling.
- Add a location normalization layer after parsing.
- Add backend endpoints to serve hierarchical location suggestions from a local Indonesia dataset.

### Frontend

- Replace technical parsing wording with user-friendly outcome-focused labels.
- Update parse result UI to include a `Saran perbaikan lokasi` section.
- Let users review and confirm city, district, and cleaned address suggestions.
- Replace hardcoded city lists with backend-driven location search and selection.

## Indonesia Location Data Strategy

### Source Strategy

Use a complete Indonesia administrative dataset sourced from an official or frequently refreshed source, then snapshot it into the application as local data for runtime lookup.

Recommended operational pattern:

- pull/update from a reliable Indonesia region data source,
- normalize into app-friendly structure,
- serve locally from backend storage/cache.

### Data Shape

Hierarchical structure:

- province
- city/kabupaten
- district/kecamatan

The data model should support:

- official IDs/codes where available,
- normalized names,
- searchable aliases if needed later,
- parent-child relationships for fast cascading lookups.

## Search and Performance Design

### Runtime Strategy

- Do not depend on third-party region APIs during live user typing.
- Use local backend data and indexed/normalized lookups.
- Separate endpoints by hierarchy level:
  - provinces
  - cities by province
  - districts by city
- Return only a small suggestion window, such as top 10-20 matches.

### Why This Approach

- complete Indonesia coverage,
- predictable latency,
- near-zero runtime data cost,
- stable UX independent of external API availability.

Google Maps can still be used for geocoding or map enrichment, but not for the main location option list.

## UX Design

### Parsing Output

The parse result should present two distinct blocks:

- `Hasil yang terdeteksi`
- `Saran perbaikan lokasi`

Suggested location fields should remain editable.

### Confidence Handling

When location confidence is low:

- do not auto-select a final city/district,
- show the best suggestions,
- ask the user to confirm from the official list.

### Address Normalization

AI should help expand and clean common shorthand found in chats such as:

- `jl`
- `gg`
- informal landmarks
- abbreviated neighborhood references

The normalized address remains editable before save.

## Error Handling

- If parsing succeeds but location confidence is weak, preserve parsed raw address and show suggestions instead of forcing a value.
- If location normalization fails, the user can still continue with manual selection.
- If official location lookup returns no match, surface that clearly and keep the raw parsed text for manual correction.
- Avoid silent fallback behavior that hides uncertainty from the user.

## Testing Considerations

- Validate structured parsing quality on realistic Indonesian WhatsApp property text.
- Test abbreviation-heavy addresses and ambiguous city/district combinations.
- Test hierarchical search performance across full Indonesia data.
- Verify low-confidence flows do not auto-commit incorrect location values.
- Verify user-friendly wording across CTA, loading, description, and result states.

## Trade-offs Considered

### Alternative 1: AI-heavy

Pros:

- more flexible for messy text

Cons:

- higher cost,
- less deterministic location handling,
- harder to validate.

### Alternative 2: Ultra-hemat / rules-first

Pros:

- cheapest runtime,
- simplest cost control

Cons:

- weaker at interpreting ambiguous shorthand,
- lower parsing quality for real-world WhatsApp chats.

## Decision

Proceed with the hybrid design:

- user-friendly wording,
- `gpt-5-mini` for parsing,
- AI suggestions rather than automatic final location selection,
- complete Indonesia local location dataset,
- fast deterministic lookup-backed search.

## Follow-up Backlog

- Investigate and fix content moderation so manual edits are re-checked before final save, not only the initial parsing/media flow.
