import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');
const premiumModal = readFileSync(new URL('../components/premium/PremiumUpgradeModal.tsx', import.meta.url), 'utf8');
const imageUpload = readFileSync(new URL('../components/listings/ImageUpload.tsx', import.meta.url), 'utf8');
const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('free package copy says first 3 listings are free, not 1 per month', () => {
  assert.match(profilePage, /3 iklan pertama gratis/i);
  assert.doesNotMatch(profilePage, /1 iklan aktif per bulan/i);
  assert.doesNotMatch(profilePage, /statistik dasar|insight/i);
});

test('premium modal copy reflects tiered caps and durations', () => {
  assert.match(premiumModal, /Premium: maksimal 15 foto/i);
  assert.match(premiumModal, /Premium: maksimal 20 listing aktif/i);
  assert.match(premiumModal, /Basic: maksimal 8 foto/i);
  assert.match(premiumModal, /Basic: maksimal 6 listing aktif/i);
  assert.match(premiumModal, /Pro: maksimal 20 foto/i);
  assert.match(premiumModal, /Pro: maksimal 50 listing aktif/i);
  assert.match(premiumModal, /Premium: tayang sampai 90 hari/i);
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
  assert.match(profilePage, /Upgrade ke Basic\/Premium\/Pro/i);
  assert.doesNotMatch(profilePage, /Rp 49rb/i);
});

test('image upload upsell copy also reflects the 15-photo premium cap', () => {
  assert.match(imageUpload, /ImageLimits\.premium/);
  assert.doesNotMatch(imageUpload, /30 foto/i);
});

test('premium and homepage copy avoid statistics or insight claims', () => {
  assert.match(premiumModal, /WA full read\/write/i);
  assert.match(premiumModal, /Voice hingga 60 menit/i);
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
  assert.match(premiumModal, /const result = await upgradePremium\(selectedTier\)/);
});
