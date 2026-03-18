import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';

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
const cacheSourcePath = new URL('./revealed-contact-cache.ts', import.meta.url);
const cacheSource = existsSync(cacheSourcePath) ? readFileSync(cacheSourcePath, 'utf8') : '';

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

test('listing detail reuses a temporary contact cache and keeps button copy static', () => {
  assert.match(cacheSource, /export async function getOrLoadRevealedContact/);
  assert.match(cacheSource, /export function getCachedRevealedContact/);
  assert.match(listingDetail, /getCachedRevealedContact/);
  assert.match(listingDetail, /getOrLoadRevealedContact/);
  assert.doesNotMatch(listingDetail, /Membuka kontak\.\.\./);
  assert.doesNotMatch(listingDetail, /disabled=\{isRevealingContact\}/);
});

test('listing detail still surfaces reveal-contact errors through the toast message', () => {
  assert.match(
    listingDetail,
    /toast\(error instanceof Error \? error\.message : 'Gagal membuka kontak penjual\.', 'error'\)/
  );
});
