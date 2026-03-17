import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(
  new URL('../app/(app)/profile/page.tsx', import.meta.url),
  'utf8'
);

test('profile page shows active listing count sourced from the current listings query', () => {
  assert.match(profilePage, /useMyListings/);
  assert.match(profilePage, /Iklan Aktif/);
  assert.match(profilePage, /getActiveListingCount\(myListingsData\?\.items\)/);
  assert.doesNotMatch(profilePage, /monthlyListingsUsed/);
  assert.doesNotMatch(profilePage, /Iklan Dipasang Bulan Ini/);
});
