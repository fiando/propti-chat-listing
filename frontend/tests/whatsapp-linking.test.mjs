import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  getWhatsAppWriteEligibilityCopy,
  normalizeWhatsAppLinkPhone,
} from '../lib/whatsapp-linking.ts';

test('normalizeWhatsAppLinkPhone normalizes Indonesian phone inputs for OTP challenge', () => {
  assert.equal(normalizeWhatsAppLinkPhone('0812 3456 7890'), '6281234567890');
  assert.equal(normalizeWhatsAppLinkPhone('+62 812-3456-7890'), '6281234567890');
  assert.equal(normalizeWhatsAppLinkPhone('6281234567890'), '6281234567890');
});

test('getWhatsAppWriteEligibilityCopy gives actionable connect/verify/write messaging', () => {
  assert.deepEqual(
    getWhatsAppWriteEligibilityCopy({ eligible: false, isLinked: false, reason: 'whatsapp phone is not linked' }),
    {
      tone: 'warning',
      title: 'WhatsApp belum terhubung',
      description: 'Hubungkan nomor WhatsApp kamu dulu, lalu minta OTP untuk verifikasi.',
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
      title: 'Nomor sudah terhubung, tinggal verifikasi OTP',
      description: 'Masukkan kode OTP agar akun kamu bisa pakai fitur WhatsApp write.',
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

test('profile page exposes connect, otp verify, disconnect, and eligibility copy', () => {
  const profilePageFile = readFileSync(new URL('../components/profile/ProfilePageClient.tsx', import.meta.url), 'utf8');

  assert.match(profilePageFile, /Hubungkan WhatsApp/);
  assert.match(profilePageFile, /Minta OTP/);
  assert.match(profilePageFile, /Verifikasi OTP/);
  assert.match(profilePageFile, /Putuskan WhatsApp/);
  assert.match(profilePageFile, /Status WA Write/);
});
