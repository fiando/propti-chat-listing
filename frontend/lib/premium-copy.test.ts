import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');
const premiumModal = readFileSync(new URL('../components/premium/PremiumUpgradeModal.tsx', import.meta.url), 'utf8');
const upgradePage = readFileSync(new URL('../components/upgrade/UpgradePageClient.tsx', import.meta.url), 'utf8');
const imageUpload = readFileSync(new URL('../components/listings/ImageUpload.tsx', import.meta.url), 'utf8');
const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('free package copy says first 3 listings are free, not 1 per month', () => {
  assert.match(profilePage, /3 iklan pertama gratis/i);
  assert.doesNotMatch(profilePage, /1 iklan aktif per bulan/i);
  assert.doesNotMatch(profilePage, /statistik dasar|insight/i);
});

test('premium modal copy reflects tiered caps and durations', () => {
  assert.match(premiumModal, /Maksimal 15 foto per iklan/i);
  assert.match(premiumModal, /Maksimal 20 listing aktif/i);
  assert.match(premiumModal, /Maksimal 8 foto per iklan/i);
  assert.match(premiumModal, /Maksimal 6 listing aktif/i);
  assert.match(premiumModal, /Maksimal 20 foto per iklan/i);
  assert.match(premiumModal, /Maksimal 50 listing aktif/i);
  assert.match(premiumModal, /Iklan tayang hingga 90 hari/i);
  assert.match(premiumModal, /Paket gratis: maksimal 3 foto/i);
  assert.match(premiumModal, /3 listing aktif/i);
  assert.match(premiumModal, /tayang 30 hari/i);
  assert.doesNotMatch(premiumModal, /foto tidak terbatas/i);
  assert.doesNotMatch(premiumModal, /gratis hanya 1/i);
  assert.doesNotMatch(premiumModal, /30 foto/i);
  assert.doesNotMatch(premiumModal, /lebih dari 3 listing/i);
  assert.doesNotMatch(premiumModal, /gratis maksimal 3/i);
});

test('profile package copy shows tiered package pricing instead of legacy 49rb', () => {
  assert.match(profilePage, /Pilih paket upgrade:/i);
  assert.doesNotMatch(profilePage, /Rp 49rb/i);
});

test('image upload upsell copy also reflects the 15-photo premium cap', () => {
  assert.match(imageUpload, /ImageLimits\.premium/);
  assert.doesNotMatch(imageUpload, /30 foto/i);
});

test('premium and homepage copy avoid statistics or insight claims', () => {
  assert.match(premiumModal, /WA baca, buat, edit & hapus listing/i);
  assert.match(premiumModal, /Voice note hingga 60 menit/i);
  assert.doesNotMatch(premiumModal, /statistik|insight/i);
  assert.doesNotMatch(homePage, /statistik|insight/i);
  assert.doesNotMatch(premiumModal, /statistik detail/i);
  assert.doesNotMatch(premiumModal, /penjual terverifikasi/i);
  assert.doesNotMatch(premiumModal, /prioritas dukungan pelanggan/i);
});

test('homepage copy avoids internal MVP wording and speaks to end users', () => {
  assert.match(homePage, /Satu Listing Properti/i);
  assert.match(homePage, /lebih rapi/i);
  assert.doesNotMatch(homePage, /MVP fokus ke hal yang paling penting/i);
  assert.doesNotMatch(homePage, /MVP focus/i);
});

test('premium modal upgrades without requiring profile phone', () => {
  assert.doesNotMatch(premiumModal, /profilePhone/);
  assert.doesNotMatch(premiumModal, /Lengkapi nomor telepon/i);
  assert.doesNotMatch(premiumModal, /returnTo=\/profile#premium/);
  assert.doesNotMatch(premiumModal, /if \(!profilePhone\?\.trim\(\)\)/);
  assert.match(premiumModal, /const result = await upgradePremium\(activeTier\)/);
});

test('premium modal provides wider desktop layout and plan selector', () => {
  assert.match(premiumModal, /max-w-4xl/);
  assert.match(premiumModal, /grid gap-3 md:grid-cols-3/);
});

test('premium modal clearly differentiates upgrade and downgrade actions', () => {
  assert.match(premiumModal, /Downgrade ke/);
  assert.match(premiumModal, /Upgrade ke/);
  assert.match(premiumModal, /Perpanjang Paket/);
});

test('upgrade page uses wider desktop layout to reduce wrapped plan text', () => {
  assert.match(upgradePage, /max-w-7xl/);
});

test('upgrade page computes per-plan upgrade\/downgrade labels from current tier', () => {
  assert.match(upgradePage, /const TIER_ORDER: Record<SubscriptionTier, number>/);
  assert.match(upgradePage, /const isDowngrade = selectedTierRank < currentTierRank/);
  assert.match(upgradePage, /Downgrade ke/);
  assert.match(upgradePage, /Upgrade ke/);
});
