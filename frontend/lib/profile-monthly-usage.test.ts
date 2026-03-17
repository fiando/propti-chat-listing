import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(
  new URL('../app/(app)/profile/page.tsx', import.meta.url),
  'utf8'
);

test('profile page shows monthly listing usage instead of active listing count', () => {
  assert.match(profilePage, /monthlyListingsUsed/);
  assert.match(profilePage, /Iklan Dipasang Bulan Ini/);
  assert.doesNotMatch(profilePage, /useMyListings/);
  assert.doesNotMatch(profilePage, /Iklan Aktif/);
});
