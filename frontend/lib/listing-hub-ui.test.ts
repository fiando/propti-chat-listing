import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';

const createListingClientFile = readFileSync(
  new URL('../components/listings/CreateListingClient.tsx', import.meta.url),
  'utf8'
);
const myListingsPageFile = readFileSync(
  new URL('../app/(app)/listings/page.tsx', import.meta.url),
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
const listingLayoutPath = new URL('../app/(app)/listings/[id]/layout.tsx', import.meta.url);
const listingLayoutFile = existsSync(listingLayoutPath) ? readFileSync(listingLayoutPath, 'utf8') : '';
const robotsPath = new URL('../app/robots.ts', import.meta.url);
const robotsFile = existsSync(robotsPath) ? readFileSync(robotsPath, 'utf8') : '';
const sitemapPath = new URL('../app/sitemap.ts', import.meta.url);
const sitemapFile = existsSync(sitemapPath) ? readFileSync(sitemapPath, 'utf8') : '';

test('listing detail route exports dynamic metadata for canonical sharing and previews', () => {
  assert.match(listingLayoutFile, /export async function generateMetadata/);
  assert.match(listingLayoutFile, /alternates:\s*\{\s*canonical:/);
  assert.match(listingLayoutFile, /buildListingMetadata\(listing\)/);
});

test('owner listing dashboard surfaces summary metrics and share helpers', () => {
  assert.match(myListingsPageFile, /totalViews/);
  assert.match(myListingsPageFile, /totalSaves/);
  assert.match(myListingsPageFile, /totalContactReveals/);
  assert.match(myListingsPageFile, /onShareToWhatsApp/);
  assert.match(myListingsPageFile, /onCopyShareLink/);
});

test('owner listing cards expose one-tap WhatsApp and copy-link actions', () => {
  assert.match(listingCardFile, /Bagikan WA/);
  assert.match(listingCardFile, /Salin Link/);
  assert.match(listingCardFile, /onShareToWhatsApp\?: \(listing: Listing\) => void/);
  assert.match(listingCardFile, /onCopyShareLink\?: \(listing: Listing\) => void/);
});

test('create listing flow redirects owners to a share-prompted detail page after publish', () => {
  assert.match(createListingClientFile, /router\.push\(`\/listings\/\$\{listing\.listingId\}\?sharePrompt=1`\)/);
});

test('owner detail page reads a share prompt flag and renders explicit distribution copy', () => {
  assert.match(listingDetailPageFile, /useSearchParams/);
  assert.match(listingDetailPageFile, /sharePrompt/);
  assert.match(listingDetailFile, /Bagikan link ini ke WhatsApp/);
  assert.match(listingDetailFile, /Salin link iklan/);
  assert.match(listingDetailFile, /contactReveals/);
});

test('app exposes robots and sitemap metadata routes for listing discovery', () => {
  assert.match(robotsFile, /export default function robots/);
  assert.match(robotsFile, /sitemap:\s*['"`]https:\/\/propti\.id\/sitemap\.xml['"`]/);
  assert.match(sitemapFile, /export default async function sitemap/);
  assert.match(sitemapFile, /\/listings\/\$\{listing\.listingId\}/);
});
