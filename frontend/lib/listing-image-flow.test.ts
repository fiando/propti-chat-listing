import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const typesFile = readFileSync(new URL('../types/index.ts', import.meta.url), 'utf8');
const apiFile = readFileSync(new URL('./api.ts', import.meta.url), 'utf8');
const hooksFile = readFileSync(new URL('../hooks/useListings.ts', import.meta.url), 'utf8');
const imageUploadFile = readFileSync(
  new URL('../components/listings/ImageUpload.tsx', import.meta.url),
  'utf8'
);
const createClientFile = readFileSync(
  new URL('../components/listings/CreateListingClient.tsx', import.meta.url),
  'utf8'
);
const editPageFile = readFileSync(
  new URL('../app/(app)/listings/[id]/edit/page.tsx', import.meta.url),
  'utf8'
);
const listingCardFile = readFileSync(
  new URL('../components/listings/ListingCard.tsx', import.meta.url),
  'utf8'
);
const listingDetailFile = readFileSync(
  new URL('../components/listings/ListingDetail.tsx', import.meta.url),
  'utf8'
);

test('frontend listing types model structured media and staged uploads', () => {
  assert.match(typesFile, /export interface ListingImageView/);
  assert.match(typesFile, /featuredThumbnailUrl\??:\s*string/);
  assert.match(typesFile, /newImageUploadSessionIds\??:\s*string\[]/);
  assert.match(typesFile, /featuredUploadSessionId\??:\s*string/);
  assert.match(typesFile, /retainedImageIds\??:\s*string\[]/);
  assert.match(typesFile, /featuredImageId\??:\s*string/);
});

test('frontend API exposes upload prepare and owner listing endpoints', () => {
  assert.match(apiFile, /\/listings\/upload-prepare/);
  assert.match(apiFile, /\/users\/me\/listings\/\$\{id\}/);
  assert.match(apiFile, /export async function prepareListingUpload/);
  assert.match(apiFile, /export async function getOwnerListing/);
});

test('frontend hooks expose owner listing reads for edit flow', () => {
  assert.match(hooksFile, /getOwnerListing/);
  assert.match(hooksFile, /export function useOwnerListing/);
});

test('image upload no longer stores base64 payloads and exposes featured selection', () => {
  assert.doesNotMatch(imageUploadFile, /readAsDataURL/);
  assert.doesNotMatch(imageUploadFile, /toDataURL/);
  assert.match(imageUploadFile, /URL\.createObjectURL/);
  assert.match(imageUploadFile, /Jadikan foto utama|Foto utama/);
});

test('listing create and edit flows upload files before submitting session ids', () => {
  assert.match(createClientFile, /prepareListingUpload|uploadListingImages|uploadPendingListingImages/);
  assert.match(createClientFile, /newImageUploadSessionIds/);
  assert.doesNotMatch(createClientFile, /images:\s*data\.images/);

  assert.match(editPageFile, /useOwnerListing/);
  assert.match(editPageFile, /retainedImageIds/);
  assert.match(editPageFile, /newImageUploadSessionIds/);
});

test('listing cards and detail pages read the normalized media helpers instead of raw string arrays', () => {
  assert.doesNotMatch(listingCardFile, /listing\.images\?\.\[0\]/);
  assert.doesNotMatch(listingDetailFile, /const images = listing\.images\?\.length > 0 \? listing\.images : \[\]/);
  assert.match(listingCardFile, /featuredThumbnailUrl|getListingCardImage|normalizeListingMedia/);
  assert.match(listingDetailFile, /getListingGalleryImages|normalizeListingMedia/);
});
