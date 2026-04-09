import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const profilePage = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');
const premiumModal = readFileSync(new URL('../components/premium/PremiumUpgradeModal.tsx', import.meta.url), 'utf8');
const imageUpload = readFileSync(new URL('../components/listings/ImageUpload.tsx', import.meta.url), 'utf8');
const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('free package copy says 5 active listings are free, not 1 per month', () => {
  assert.match(profilePage, /5 listing aktif gratis/i);
  assert.doesNotMatch(profilePage, /1 iklan aktif per bulan/i);
  assert.doesNotMatch(profilePage, /statistik dasar|insight/i);
});

test('premium modal copy reflects current paid caps and free starter limits', () => {
  assert.match(premiumModal, /Maksimal 15 foto per iklan/i);
  assert.match(premiumModal, /Maksimal 25 listing aktif/i);
  assert.match(premiumModal, /Maksimal 25 foto per iklan/i);
  assert.match(premiumModal, /Maksimal 100 listing aktif/i);
  assert.match(premiumModal, /Iklan tayang hingga 90 hari/i);
  assert.match(premiumModal, /Paket gratis: maksimal 5 foto/i);
  assert.match(premiumModal, /5 listing aktif/i);
  assert.match(premiumModal, /tayang 60 hari/i);
  assert.match(premiumModal, /Limit dihitung dari listing aktif yang sedang berjalan/i);
  assert.doesNotMatch(premiumModal, /foto tidak terbatas/i);
  assert.doesNotMatch(premiumModal, /gratis hanya 1/i);
  assert.doesNotMatch(premiumModal, /30 foto/i);
  assert.doesNotMatch(premiumModal, /lebih dari 5 listing/i);
  assert.doesNotMatch(premiumModal, /gratis maksimal 3/i);
});

test('profile page shows tier selection for all users including paid', () => {
  assert.match(profilePage, /Pilih paket upgrade:/i);
  assert.match(profilePage, /Ganti atau perpanjang paket:/i);
  assert.match(profilePage, /Rp 99rb\/bln/i);
  assert.match(profilePage, /Rp 179rb\/bln/i);
  assert.doesNotMatch(profilePage, /Rp 59rb|Rp 129rb|Rp 199rb/i);
});

test('image upload upsell copy also reflects the 15-photo premium cap', () => {
  assert.match(imageUpload, /ImageLimits\.premium/);
  assert.doesNotMatch(imageUpload, /30 foto/i);
});

test('premium modal WA copy says create and search only, no edit or delete', () => {
  assert.match(premiumModal, /Buat & cari via WhatsApp/i);
  assert.doesNotMatch(premiumModal, /edit & hapus/i);
  assert.doesNotMatch(premiumModal, /tanpa edit\/hapus/i);
  assert.match(premiumModal, /Voice note hingga 90 menit/i);
  assert.doesNotMatch(premiumModal, /penjual terverifikasi/i);
  assert.doesNotMatch(premiumModal, /prioritas dukungan pelanggan/i);
});

test('homepage copy avoids internal MVP wording and speaks to end users', () => {
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
  assert.match(premiumModal, /grid gap-3 md:grid-cols-2/);
  assert.doesNotMatch(premiumModal, /label:\s*'Basic'/);
});

test('premium modal clearly differentiates upgrade and downgrade actions', () => {
  assert.match(premiumModal, /Downgrade ke/);
  assert.match(premiumModal, /Upgrade ke/);
  assert.match(premiumModal, /Perpanjang Paket/);
});

test('homepage pricing links point to profile page, not separate upgrade page', () => {
  assert.match(homePage, /\/profile\?upgradeTier=premium#premium/);
  assert.match(homePage, /\/profile\?upgradeTier=pro#premium/);
  assert.doesNotMatch(homePage, /\/profile\?upgradeTier=basic#premium/);
  assert.doesNotMatch(homePage, /\/upgrade\?tier=/);
});

test('homepage free tier advertises WhatsApp creation only, while paid tiers add search', () => {
  assert.match(homePage, /key: 'free'[\s\S]*?Buat listing via WhatsApp[\s\S]*?cta: 'Mulai Gratis'/);
  assert.doesNotMatch(homePage, /key: 'free'[\s\S]*?Buat & cari via WhatsApp[\s\S]*?cta: 'Mulai Gratis'/);
  assert.match(homePage, /Buat & cari via WhatsApp/);
  assert.match(homePage, /fitur WhatsApp, dan voice note/i);
  assert.match(homePage, /listing aktif yang sedang berjalan/i);
  assert.doesNotMatch(homePage, /Semakin banyak listing yang perlu kamu kelola aktif sekaligus/i);
});

test('profile page auto-opens modal from upgradeTier query param', () => {
  assert.match(profilePage, /upgradeTierParam/);
  assert.match(profilePage, /setShowPremiumModal\(true\)/);
  assert.doesNotMatch(profilePage, /\['basic', 'premium', 'pro'\]/);
  assert.match(profilePage, /\['premium', 'pro'\]/);
});
