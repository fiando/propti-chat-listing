import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(
  new URL('../app/(app)/profile/page.tsx', import.meta.url),
  'utf8'
);

test('profile page uses my listings count instead of stale monthlyListingsUsed subscription field', () => {
  assert.match(profilePage, /useMyListings/);
  assert.doesNotMatch(profilePage, /monthlyListingsUsed/);
  assert.match(profilePage, /Iklan Aktif/);
});
