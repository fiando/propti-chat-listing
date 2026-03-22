import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const apiFile = readFileSync(new URL('./api.ts', import.meta.url), 'utf8');
const hooksFile = readFileSync(new URL('../hooks/useListings.ts', import.meta.url), 'utf8');
const myListingsPageFile = readFileSync(
  new URL('../app/(app)/listings/page.tsx', import.meta.url),
  'utf8'
);
const listingGridFile = readFileSync(
  new URL('../components/listings/ListingGrid.tsx', import.meta.url),
  'utf8'
);
const listingCardFile = readFileSync(
  new URL('../components/listings/ListingCard.tsx', import.meta.url),
  'utf8'
);
const listingDetailPageFile = readFileSync(
  new URL('../app/(app)/listings/[id]/page.tsx', import.meta.url),
  'utf8'
);
const listingDetailFile = readFileSync(
  new URL('../components/listings/ListingDetail.tsx', import.meta.url),
  'utf8'
);

test('frontend API exposes owner relist endpoint', () => {
  assert.match(apiFile, /export async function relistListing/);
  assert.match(apiFile, /\/users\/me\/listings\/\$\{id\}\/relist/);
});

test('frontend hooks expose relist mutation and refresh owner listing caches', () => {
  assert.match(hooksFile, /relistListing/);
  assert.match(hooksFile, /export function useRelistListing/);
  assert.match(hooksFile, /invalidateQueries\(\{ queryKey: \['my-listings'\] \}\)/);
  assert.match(hooksFile, /invalidateQueries\(\{ queryKey: \['my-listing-quota-summary'\] \}\)/);
});

test('my listings page wires relist action into the owner listing grid', () => {
  assert.match(myListingsPageFile, /useRelistListing/);
  assert.match(myListingsPageFile, /onRelist=/);
});

test('listing grid and card expose relist CTA for archived owner listings', () => {
  assert.match(listingGridFile, /onRelist\?: \(id: string\) => void/);
  assert.match(listingCardFile, /onRelist\?: \(id: string\) => void/);
  assert.match(listingCardFile, /Relist Iklan/);
  assert.match(listingCardFile, /shouldShowRelistAction/);
});

test('owner detail page wires relist action for archived listings', () => {
  assert.match(listingDetailPageFile, /useRelistListing/);
  assert.match(listingDetailPageFile, /onRelist=/);
  assert.match(listingDetailFile, /Relist Iklan/);
});

test('owner listing surfaces show expiry feedback to explain active-until or archived state', () => {
  assert.match(listingCardFile, /getListingExpiryInfo/);
  assert.match(listingDetailFile, /getListingExpiryInfo/);
});
