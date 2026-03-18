# Listing image storage redesign

## Problem

Listing images are currently handled as inline string payloads in the listing flow, which is a poor fit for production:

- base64-style payloads bloat request bodies and stored records
- validation and moderation are coupled to string decoding instead of object storage
- the current codebase already has partial S3 presigned upload support, but the main listing flow still stores image strings directly
- free and premium image limits are inconsistent across the codebase and must be unified

The target state is a secure S3-backed image flow that preserves strict server-side enforcement for free and premium users, keeps image validation intact, and avoids draft listings for brand-new listings.

## Approved constraints

- Free tier: maximum `3` images per listing
- Premium tier: maximum `15` images per listing
- Limit scope: images only
- Compression: always enabled before upload; larger files may use stronger compression, but every image is normalized before storage
- Stored representation: S3 object keys only
- Public access model:
  - public surfaces may expose the featured image thumbnail
  - listing detail pages receive temporary signed view URLs for non-featured gallery images
- Featured image: explicitly selectable by the user and validated by the backend
- Existing base64 listing data: ignore; no migration required
- New base64 payloads: explicitly rejected after rollout
- New listing flow: keep single submit, no pre-created draft listing

## Goals

- Remove base64 image storage from new listing create/update flows
- Store only stable S3 object keys and related metadata in persisted listing records
- Keep tier enforcement authoritative on the backend
- Preserve or improve current moderation and safety checks
- Keep uploads efficient by sending file bytes directly from the client to S3
- Support featured-image selection without exposing the full gallery publicly

## Non-goals

- Migrating legacy base64 listings
- Introducing draft listings for the create flow
- Expanding limits beyond `3` free and `15` premium images
- Redesigning video handling in this project phase

## Recommended approach

Use a staged single-submit upload flow:

1. The client compresses and normalizes selected images before upload.
2. The client requests upload slots from the backend.
3. The backend validates the request and returns signed PUT URLs plus short-lived upload tokens.
4. The client uploads image bytes directly to S3 staging keys.
5. The client submits the listing once, including the issued upload tokens and featured-image choice.
6. The backend validates the submitted tokens and staged objects, promotes approved files into permanent listing keys, stores only object keys in the listing, and rejects base64 payloads.

This approach matches the requirement to avoid draft listings while keeping server-side security and tier enforcement.

## Architecture

### Data model

Persist listing images as structured image references rather than raw image strings. The record should store:

- stable S3 object key
- content type
- file size
- width and height if available
- thumbnail/original variant role
- featured-image flag or featured-image pointer

The canonical persisted value is the S3 object key, not a URL and not a signed URL.

### Bucket organization

Use separate logical prefixes:

- staging uploads: user-scoped, temporary
- permanent listing originals: listing-scoped
- permanent listing thumbnails: listing-scoped and suitable for public exposure if desired

Staging objects must never be treated as valid listing images unless they are promoted during final create/update.

### Read model

- Public list/search/card pages expose only the featured thumbnail URL
- Listing detail responses generate temporary signed GET URLs for gallery images at read time
- Signed GET URLs are not persisted and are refreshed by re-fetching listing detail when expired

## Upload and submit flow

### 1. Client-side preprocessing

When a user selects images:

- validate basic count and supported formats for immediate UX feedback
- always compress/normalize every image before upload
- use stronger compression for large inputs such as files over `5 MB`
- keep output visually acceptable while reducing storage and transfer size

The client-side step improves UX and efficiency, but it is not the trust boundary.

### 2. Upload slot issuance

The backend exposes an authenticated endpoint that accepts the requested upload count and basic file descriptors. It must:

- identify the user and current subscription tier
- enforce the canonical image cap (`3` free, `15` premium)
- validate allowed MIME types
- validate file count per request and total intended count
- issue one signed PUT URL and one upload token per approved image
- bind each token to:
  - user identity
  - staging object key
  - expected content type
  - expected max size
  - expiry time
  - single-use semantics

The response must not grant broader write access than the specific staged object key.

### 3. Direct S3 upload

The client uploads each normalized image directly to the provided signed PUT URL. This keeps the API out of the hot path for large file transfer and avoids Lambda/API Gateway payload pressure.

### 4. Final listing create/update

On create or update, the client submits:

- normal listing fields
- upload tokens for any newly uploaded images
- the final ordered image set for the listing
- featured-image choice

For updates, the final image set may contain a mix of:

- existing persisted image keys the user chose to keep
- newly uploaded images referenced by upload tokens

Any previously stored image not included in the final submitted set is treated as removed.

The backend must then:

- reject any new base64 image payloads
- resolve each token and verify ownership, expiry, and unused status
- verify the referenced S3 staged objects exist
- verify any retained existing image keys already belong to that listing
- re-check the final image count against the user tier
- verify the featured image belongs to the final submitted image set, whether retained or newly uploaded
- promote/copy staged objects into permanent listing keys
- create thumbnail artifacts if needed
- store only the final object keys and related metadata in the listing record
- mark tokens as consumed so they cannot be replayed
- delete or schedule deletion for removed permanent objects that are no longer referenced by the listing

If create/update fails, staged objects remain temporary and are cleaned up later.

## Security model

### Trust boundaries

- The frontend may help with UX, but the backend is the authority for tier enforcement and accepted image sets
- The database stores object keys only
- Signed PUT and GET URLs are temporary capabilities, not persistent identifiers

### Required backend validation

The backend must enforce all of the following regardless of frontend behavior:

- authenticated user ownership
- canonical image-count limit for the user tier
- MIME type allowlist
- extension/content-type consistency where practical
- post-compression per-file size cap
- upload token expiry
- upload token single-use enforcement
- uploaded object presence in the expected staging key
- featured-image membership in the submitted image set
- explicit rejection of base64 image payloads in create/update APIs

### Public exposure

Only the featured thumbnail is allowed to be broadly exposed on public pages. The rest of the gallery remains protected behind listing detail reads that mint temporary signed GET URLs.

## Moderation and validation

Current moderation logic decodes base64 image strings. In the new design, moderation must operate on bytes read from S3 objects instead.

The moderation pipeline should preserve current behavior:

- unsafe-content detection
- property-relevance checks
- rejection or pending moderation outcomes using the existing listing moderation fields

Moderation can run against staged objects before promotion or against newly promoted permanent objects, but the implementation plan should choose one consistent point in the workflow and keep cleanup behavior clear.

## Limits and consistency

The rollout must unify all image-limit messaging and validation to one canonical rule:

- free: `3` images
- premium: `15` images

Any current mismatch in UI copy, frontend props, or backend validator messages must be corrected so the product promise and enforcement match.

## Failure handling

- Expired upload tokens: require the client to request fresh upload slots
- Expired signed gallery URLs: refresh by re-fetching listing detail
- Submit with missing staged objects: reject with a clear validation error
- Token replay: reject and log
- Moderation failure: keep current listing moderation semantics and clean up or quarantine related objects according to the chosen implementation detail
- Abandoned uploads: rely on staged-prefix cleanup

## Cleanup strategy

Use bucket lifecycle rules or equivalent scheduled cleanup for abandoned staging uploads. The cleanup window should exceed normal user form-completion time but still prevent long-lived orphaned objects.

## Testing requirements

### Backend

Add tests for:

- upload slot issuance respects free and premium caps
- upload slot issuance rejects unsupported file types
- create/update rejects base64 image payloads
- expired upload token rejection
- upload token replay rejection
- featured-image validation
- final image-count enforcement for create and update
- staged object verification before promotion
- signed detail-view URL generation

### Frontend

Add tests for:

- always-on compression path
- stronger compression behavior for large files
- upload UX when limits are exceeded
- featured-image selection behavior
- expired gallery URL refresh path
- error messaging when backend rejects invalid upload tokens or counts

### Operational verification

Monitor after release:

- growth of staging prefixes
- create/update validation failures
- moderation failure counts
- image upload failure rates

## Open implementation decisions for planning

These are intentionally deferred to implementation planning, not design approval blockers:

- exact metadata shape for persisted image references
- whether thumbnails are generated client-side, server-side, or both
- whether moderation runs before or after promotion into permanent keys
- whether the featured thumbnail uses a public bucket path, CDN rule, or another controlled public delivery layer

## Summary

The approved design replaces base64 image persistence with a secure staged S3 upload flow that:

- keeps listing create/update as a single submit flow
- stores only S3 keys in DynamoDB
- exposes only the featured thumbnail publicly
- serves the rest of the gallery with temporary signed GET URLs
- preserves server-side enforcement of `3` free and `15` premium images
- rejects new base64 payloads
- keeps image moderation and validation authoritative on the backend
