import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(new URL('../app/(app)/profile/page.tsx', import.meta.url), 'utf8');
const premiumModal = readFileSync(new URL('../components/premium/PremiumUpgradeModal.tsx', import.meta.url), 'utf8');

test('free package copy says first 3 listings are free, not 1 per month', () => {
  assert.match(profilePage, /3 iklan pertama gratis/i);
  assert.doesNotMatch(profilePage, /1 iklan aktif per bulan/i);
});

test('premium modal copy reflects 30-photo cap and free tier 3-listing baseline', () => {
  assert.match(premiumModal, /30 foto/i);
  assert.match(premiumModal, /gratis.*3/i);
  assert.doesNotMatch(premiumModal, /foto tidak terbatas/i);
  assert.doesNotMatch(premiumModal, /gratis hanya 1/i);
});
