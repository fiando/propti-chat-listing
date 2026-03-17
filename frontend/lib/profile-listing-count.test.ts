import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  getActiveListingCount,
  isCreateListingLimitReached,
} from './create-listing-errors.ts';

const profilePage = readFileSync(
  new URL('../app/(app)/profile/page.tsx', import.meta.url),
  'utf8'
);
const createPage = readFileSync(
  new URL('../app/(app)/listings/create/page.tsx', import.meta.url),
  'utf8'
);
const textParseForm = readFileSync(
  new URL('../components/listings/TextParseForm.tsx', import.meta.url),
  'utf8'
);
const useListingsHook = readFileSync(
  new URL('../hooks/useListings.ts', import.meta.url),
  'utf8'
);

test('profile page uses my listings count instead of stale monthlyListingsUsed subscription field', () => {
  assert.match(profilePage, /useMyListings/);
  assert.doesNotMatch(profilePage, /monthlyListingsUsed/);
  assert.match(profilePage, /Iklan Aktif/);
  assert.match(profilePage, /getActiveListingCount\(myListingsData\?\.items\)/);
  assert.match(profilePage, /keepPreviousData:\s*false/);
  assert.match(profilePage, /userId:\s*profile\?\.userId\s*\?\?\s*null/);
});

test('counts only active listings toward the free-tier create limit', () => {
  const listings = [
    { status: 'active' },
    { status: 'active' },
    { status: 'archived' },
    { status: 'sold' },
  ];

  assert.equal(getActiveListingCount(listings), 2);
  assert.equal(isCreateListingLimitReached({ isPremium: false, listings }), false);
  assert.equal(
    isCreateListingLimitReached({
      isPremium: false,
      listings: [{ status: 'active' }, { status: 'active' }, { status: 'active' }, { status: 'archived' }],
    }),
    true
  );
  assert.equal(
    isCreateListingLimitReached({
      isPremium: true,
      listings: [{ status: 'active' }, { status: 'active' }, { status: 'active' }, { status: 'active' }],
    }),
    false
  );
});

test('create listing page blocks free-tier sellers early using current active listings count', () => {
  assert.match(createPage, /getCreateListingAccessState/);
  assert.match(createPage, /activeListingsCount = createAccessState\.activeListingsCount/);
  assert.match(createPage, /setStep\('choose'\)/);
  assert.doesNotMatch(createPage, /monthlyListingsUsed/);
  assert.match(createPage, /Upgrade ke Premium/);
  assert.match(createPage, /Kembali ke iklan saya/);
  assert.match(createPage, /useMyListingQuotaSummary/);
  assert.match(createPage, /activeLimit:\s*FREE_TIER_LISTING_LIMIT/);
  assert.match(createPage, /userId:\s*profile\?\.userId\s*\?\?\s*null/);
  assert.match(createPage, /Coba lagi/);
  assert.match(useListingsHook, /queryKey:\s*\['my-listing-quota-summary',\s*options\?\.userId\s*\?\?\s*null,\s*options\?\.activeLimit\s*\?\?\s*null\]/);
  assert.match(useListingsHook, /invalidateQueries\(\{ queryKey: \['my-listing-quota-summary'\] \}\)/);
});

test('parsed-result handoff is explicit and scrolls final review to the top', () => {
  assert.match(textParseForm, /Gunakan hasil parsing/);
  assert.match(createPage, /window\.scrollTo\(\{ top: 0, behavior: 'smooth' \}\)/);
});
