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
const profileClient = readFileSync(
  new URL('../components/profile/ProfilePageClient.tsx', import.meta.url),
  'utf8'
);
const createPage = readFileSync(
  new URL('../app/(app)/listings/create/page.tsx', import.meta.url),
  'utf8'
);
const createClient = readFileSync(
  new URL('../components/listings/CreateListingClient.tsx', import.meta.url),
  'utf8'
);
const header = readFileSync(
  new URL('../components/common/Header.tsx', import.meta.url),
  'utf8'
);
const textParseForm = readFileSync(
  new URL('../components/listings/TextParseForm.tsx', import.meta.url),
  'utf8'
);
const mobileNav = readFileSync(
  new URL('../components/common/MobileNav.tsx', import.meta.url),
  'utf8'
);
const homePage = readFileSync(
  new URL('../app/(app)/page.tsx', import.meta.url),
  'utf8'
);
const savedPage = readFileSync(
  new URL('../app/(app)/saved/page.tsx', import.meta.url),
  'utf8'
);
const typesFile = readFileSync(
  new URL('../types/index.ts', import.meta.url),
  'utf8'
);
const listingForm = readFileSync(
  new URL('../components/listings/ListingForm.tsx', import.meta.url),
  'utf8'
);

test('profile page is server-first and reads active listing count from subscription payload', () => {
  assert.doesNotMatch(profilePage, /'use client'/);
  assert.doesNotMatch(profilePage, /useSession/);
  assert.doesNotMatch(profilePage, /useMyListings/);
  assert.doesNotMatch(profilePage, /monthlyListingsUsed/);
  assert.match(profileClient, /subscription\.activeListingsCount/);
  assert.match(profileClient, /subscription\.tier/);
});

test('counts only active listings toward the free-tier create limit', () => {
  const listings = [
    { status: 'active' },
    { status: 'active' },
    { status: 'archived' },
    { status: 'sold' },
  ];

  assert.equal(getActiveListingCount(listings), 2);
  assert.equal(isCreateListingLimitReached({ tier: 'free', listings }), false);
  assert.equal(
    isCreateListingLimitReached({
      tier: 'free',
      listings: [{ status: 'active' }, { status: 'active' }, { status: 'active' }, { status: 'archived' }],
    }),
    true
  );
  assert.equal(
    isCreateListingLimitReached({
      tier: 'premium',
      listings: Array.from({ length: 19 }, () => ({ status: 'active' })),
    }),
    false
  );
  assert.equal(
    isCreateListingLimitReached({
      tier: 'premium',
      listings: Array.from({ length: 20 }, () => ({ status: 'active' })),
    }),
    true
  );
});

test('create listing page is server-first and passes initial access state to client shell', () => {
  assert.doesNotMatch(createPage, /'use client'/);
  assert.match(createPage, /CreateListingClient/);
  assert.match(createPage, /initialCreateAccessState/);
  assert.doesNotMatch(createPage, /Memeriksa slot listing gratis/);
  assert.doesNotMatch(createClient, /isProfileFetchedAfterMount/);
  assert.doesNotMatch(createClient, /isProfileFetching/);
  assert.doesNotMatch(createClient, /refetchOnMount:\s*'always'/);
  assert.match(typesFile, /activeListingsCount\??:\s*number/);
});

test('listing form reads image caps from the shared ImageLimits contract', () => {
  assert.match(typesFile, /export const ImageLimits = \{/);
  assert.match(listingForm, /ImageLimits/);
  assert.doesNotMatch(listingForm, /maxImages=\{isPremium \? 15 : 3\}/);
});

test('header keeps one mobile trigger and moves account links into profile dropdown', () => {
  assert.match(header, /\{\/\* Mobile menu button \*\//);
  assert.match(header, /Menu/);
  assert.match(header, /href="\/listings"/);
  assert.match(header, /href="\/saved"/);
  assert.doesNotMatch(header, /<Heart className="w-4 h-4"\/>\s*\n\s*Tersimpan\s*\n\s*<\/Link>\s*\n\s*<\/>(?:.|\n)*\{\/\* Desktop nav \*\}/);
});

test('profile and create routes expose route-level loading skeletons', () => {
  const profileLoading = readFileSync(
    new URL('../app/(app)/profile/loading.tsx', import.meta.url),
    'utf8'
  );
  const createLoading = readFileSync(
    new URL('../app/(app)/listings/create/loading.tsx', import.meta.url),
    'utf8'
  );

  assert.match(profileLoading, /animate-pulse/);
  assert.match(createLoading, /animate-pulse/);
  assert.match(profileLoading, /Memuat profil/);
  assert.match(createLoading, /Menyiapkan form iklan/);
});

test('header gives immediate feedback when navigating to slow SSR pages', () => {
  assert.match(header, /pendingRoute/);
  assert.match(header, /startRoutePending\('/);
  assert.match(header, /Memuat\.\.\./);
});

test('header mobile menu keeps saved/profile visible for guests via login callback links', () => {
  assert.match(header, /getAuthHref\('\/profile'\)/);
  assert.match(header, /getAuthHref\('\/saved'\)/);
  assert.match(header, /encodeURIComponent\(href\)/);
});

test('parsed-result handoff is explicit and scrolls final review to the top', () => {
  assert.match(textParseForm, /Gunakan hasil parsing/);
  assert.doesNotMatch(textParseForm, /Edit Manual/);
  assert.match(textParseForm, /className="w-full flex items-center justify-center gap-2 btn-primary"/);
  assert.match(createClient, /window\.scrollTo\(\{ top: 0, behavior: 'smooth' \}\)/);
});

test('guest users still see saved/profile in mobile nav and are redirected to login callback', () => {
  assert.doesNotMatch(mobileNav, /if \(item\.requiresAuth && !session\) return null/);
  assert.match(mobileNav, /getHref\(item\.href, item\.requiresAuth\)/);
  assert.match(mobileNav, /encodeURIComponent\(href\)/);
});

test('mobile nav exposes pasang as a right floating action button above the menubar', () => {
  assert.doesNotMatch(mobileNav, /isPrimary:\s*true/);
  assert.match(mobileNav, /absolute -top-7 right-4/);
  assert.match(mobileNav, /href=\{getHref\('\/listings\/create'\)\}/);
});

test('search CTA copy is consistent and parser CTA points to create listing', () => {
  assert.match(homePage, /Cari Properti/);
  assert.doesNotMatch(homePage, /Jelajahi Properti/);
  assert.doesNotMatch(savedPage, /Jelajahi Properti/);
  assert.match(savedPage, /Cari Properti/);
  assert.doesNotMatch(homePage, /Masuk & coba parser/);
  assert.match(homePage, /Paste listing saya/);
});

test('profile includes account settings form with full-width save button', () => {
  assert.match(profileClient, /Kelola data profil dan preferensi akun Propti kamu\./);
  assert.match(profileClient, /Simpan Pengaturan/);
  assert.match(profileClient, /className="btn-primary w-full inline-flex items-center justify-center gap-2 disabled:opacity-60"/);
});
