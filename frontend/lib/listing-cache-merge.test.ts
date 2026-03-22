import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const hooksFile = readFileSync(new URL('../hooks/useListings.ts', import.meta.url), 'utf8');

test('track listing view merges summary payload into existing detail cache instead of replacing seller metadata', () => {
  assert.match(hooksFile, /queryClient\.setQueryData\(\['listing', listing\.listingId\], \(previous: Listing \| undefined\) =>/);
  assert.match(hooksFile, /sellerName: previous\?\.sellerName \?\? listing\.sellerName/);
  assert.match(hooksFile, /hasSellerPhone: previous\?\.hasSellerPhone \?\? listing\.hasSellerPhone/);
});
