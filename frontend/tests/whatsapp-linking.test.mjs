import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  getWhatsAppWriteEligibilityCopy,
  normalizeWhatsAppLinkPhone,
} from '../lib/whatsapp-linking.ts';

test('normalizeWhatsAppLinkPhone normalizes Indonesian phone inputs for whatsapp challenge', () => {
  assert.equal(normalizeWhatsAppLinkPhone('0812 3456 7890'), '6281234567890');
  assert.equal(normalizeWhatsAppLinkPhone('+62 812-3456-7890'), '6281234567890');
  assert.equal(normalizeWhatsAppLinkPhone('6281234567890'), '6281234567890');
});

test('getWhatsAppWriteEligibilityCopy gives actionable connect/challenge/write messaging', () => {
  assert.deepEqual(
    getWhatsAppWriteEligibilityCopy({ eligible: false, isLinked: false, reason: 'whatsapp phone is not linked' }),
    {
      tone: 'warning',
      title: 'WhatsApp belum terhubung',
      description: 'Hubungkan nomor WhatsApp kamu lalu kirim pesan challenge untuk verifikasi.',
    }
  );

  assert.deepEqual(
    getWhatsAppWriteEligibilityCopy({
      eligible: false,
      isLinked: true,
      linkedPhone: '6281234567890',
      reason: 'whatsapp identity is not verified',
    }),
    {
      tone: 'warning',
      title: 'Nomor sudah terhubung, tinggal kirim pesan verifikasi WhatsApp',
      description: 'Kirim pesan challenge dari WhatsApp yang sama agar akun kamu bisa pakai fitur WhatsApp write.',
    }
  );

  assert.deepEqual(
    getWhatsAppWriteEligibilityCopy({
      eligible: true,
      isLinked: true,
      linkedPhone: '6281234567890',
    }),
    {
      tone: 'success',
      title: 'WhatsApp terverifikasi',
      description: 'Akun kamu sudah eligible untuk fitur WhatsApp write.',
    }
  );
});

test('api helpers target whatsapp identity endpoints', () => {
  const apiFile = readFileSync(new URL('../lib/api.ts', import.meta.url), 'utf8');

  assert.match(apiFile, /\/auth\/whatsapp\/link-status/);
  assert.match(apiFile, /\/auth\/whatsapp\/link-challenge/);
  assert.match(apiFile, /\/auth\/whatsapp\/link-verify/);
  assert.match(apiFile, /\/auth\/whatsapp\/link/);
});

test('profile page exposes connect, whatsapp challenge verify, disconnect, and eligibility copy', () => {
  const profilePageFile = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');

  assert.match(profilePageFile, /Hubungkan WhatsApp/);
  assert.match(profilePageFile, /Aktifkan WhatsApp Listing/);
  assert.match(profilePageFile, /Kirim Pesan di WhatsApp/);
  assert.match(profilePageFile, /Cek Status Verifikasi/);
  assert.match(profilePageFile, /Putuskan WhatsApp/);
  assert.match(profilePageFile, /Status WA Write/);
});

test('create listing page does not gate website flow with whatsapp otp modal', () => {
  const createListingFile = readFileSync(new URL('../components/listings/CreateListingClient.tsx', import.meta.url), 'utf8');

  assert.doesNotMatch(createListingFile, /Minta OTP WhatsApp/);
  assert.doesNotMatch(createListingFile, /Verifikasi & Pasang Iklan/);
  assert.doesNotMatch(createListingFile, /Hubungkan WhatsApp/);
});
