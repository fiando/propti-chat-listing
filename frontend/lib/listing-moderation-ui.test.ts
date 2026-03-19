import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const listingCardFile = readFileSync(
  new URL('../components/listings/ListingCard.tsx', import.meta.url),
  'utf8'
);
const listingDetailFile = readFileSync(
  new URL('../components/listings/ListingDetail.tsx', import.meta.url),
  'utf8'
);
const editPageFile = readFileSync(
  new URL('../app/(app)/listings/[id]/edit/page.tsx', import.meta.url),
  'utf8'
);

test('owner listing cards switch pending and rejected listings to neutral moderation placeholders', () => {
  assert.match(listingCardFile, /const isModerationHidden = listing\.moderationStatus !== 'approved'/);
  assert.match(listingCardFile, /const listingImage = isModerationHidden \? undefined : getListingCardImage\(listing\)/);
  assert.match(listingCardFile, /Iklan sedang direview/);
  assert.match(listingCardFile, /Iklan ditolak/);
});

test('listing detail hides owner content for pending and rejected states and removes edit action', () => {
  assert.match(listingDetailFile, /const shouldHideOwnerContent = isOwner && listing\.moderationStatus !== 'approved'/);
  assert.match(listingDetailFile, /Konten iklan disembunyikan selama proses moderasi/);
  assert.match(listingDetailFile, /if \(shouldHideOwnerContent\) \{/);
  assert.match(listingDetailFile, /Hapus Iklan/);
});

test('edit page blocks pending and rejected listings from reopening abusive content in the form', () => {
  assert.match(editPageFile, /const isEditLocked = listing\.moderationStatus !== 'approved'/);
  assert.match(editPageFile, /Iklan ini tidak bisa diedit/);
  assert.match(editPageFile, /if \(isEditLocked\) \{/);
});
