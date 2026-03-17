import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const listingDetail = readFileSync(
  new URL('../components/listings/ListingDetail.tsx', import.meta.url),
  'utf8'
);
const apiSource = readFileSync(
  new URL('./api.ts', import.meta.url),
  'utf8'
);
const hooksSource = readFileSync(
  new URL('../hooks/useListings.ts', import.meta.url),
  'utf8'
);

test('listing detail gates seller phone behind an explicit contact reveal flow', () => {
  assert.match(listingDetail, /sellerName/);
  assert.match(listingDetail, /useRevealListingContact/);
  assert.match(listingDetail, /Masuk untuk melihat nomor/);
  assert.match(listingDetail, /revealListingContact/);
  assert.doesNotMatch(listingDetail, /buildListingContactLinks\(listing\.sellerPhone/);
});

test('frontend api exposes a dedicated revealListingContact call and hook', () => {
  assert.match(apiSource, /export async function revealListingContact/);
  assert.match(apiSource, /\/listings\/\$\{id\}\/contact-reveal/);
  assert.match(hooksSource, /export function useRevealListingContact/);
});
