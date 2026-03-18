import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');
const premiumModal = readFileSync(new URL('../components/premium/PremiumUpgradeModal.tsx', import.meta.url), 'utf8');

test('free package copy says first 3 listings are free, not 1 per month', () => {
  assert.match(profilePage, /3 iklan pertama gratis/i);
  assert.doesNotMatch(profilePage, /1 iklan aktif per bulan/i);
});

test('premium modal copy reflects 15-photo cap and free tier 3-listing baseline', () => {
  assert.match(premiumModal, /15 foto/i);
  assert.match(premiumModal, /gratis.*3/i);
  assert.doesNotMatch(premiumModal, /foto tidak terbatas/i);
  assert.doesNotMatch(premiumModal, /gratis hanya 1/i);
  assert.doesNotMatch(premiumModal, /30 foto/i);
});

test('premium modal copy only promises shipped benefits and clarifies boost is separate', () => {
  assert.match(premiumModal, /statistik dasar/i);
  assert.match(premiumModal, /iklan unggulan.*terpisah/i);
  assert.match(premiumModal, /berbayar per listing/i);
  assert.doesNotMatch(premiumModal, /statistik detail/i);
  assert.doesNotMatch(premiumModal, /penjual terverifikasi/i);
  assert.doesNotMatch(premiumModal, /prioritas dukungan pelanggan/i);
});

test('premium modal requires phone completion before calling upgrade endpoint', () => {
  assert.match(premiumModal, /profilePhone/);
  assert.match(premiumModal, /Lengkapi nomor telepon/i);
  assert.match(premiumModal, /returnTo=\/profile#premium/);
  assert.match(premiumModal, /if \(!profilePhone\?\.trim\(\)\)/);
  assert.match(premiumModal, /upgradePremium\(\)/);
});

test('premium modal allows upgrade when phone is already completed', () => {
  assert.match(premiumModal, /if \(!profilePhone\?\.trim\(\)\)\s*\{[\s\S]*return;[\s\S]*\}/);
  assert.match(premiumModal, /const result = await upgradePremium\(\)/);
});
