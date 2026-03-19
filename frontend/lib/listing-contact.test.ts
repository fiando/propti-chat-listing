import test from 'node:test';
import assert from 'node:assert/strict';
import { buildListingContactLinks, buildWhatsAppMessage, normalizeContactPhone } from './listing-contact.ts';

test('normalizeContactPhone converts Indonesian local numbers to international format', () => {
  assert.equal(normalizeContactPhone('0812 3456 7890'), '6281234567890');
  assert.equal(normalizeContactPhone('+62 812-3456-7890'), '6281234567890');
});

test('buildWhatsAppMessage returns default message without title or url', () => {
  assert.equal(
    buildWhatsAppMessage(),
    'Halo, saya tertarik dengan properti di Propti. Apakah masih tersedia?'
  );
});

test('buildWhatsAppMessage includes title when provided', () => {
  assert.equal(
    buildWhatsAppMessage('Rumah Minimalis Bekasi'),
    'Halo, saya tertarik dengan properti *Rumah Minimalis Bekasi* di Propti. Apakah masih tersedia?'
  );
});

test('buildWhatsAppMessage includes title and url when both provided', () => {
  assert.equal(
    buildWhatsAppMessage('Rumah Minimalis Bekasi', 'https://propti.id/listings/abc123'),
    'Halo, saya tertarik dengan properti *Rumah Minimalis Bekasi* di Propti. Apakah masih tersedia?\nhttps://propti.id/listings/abc123'
  );
});

test('buildListingContactLinks returns WhatsApp and phone links with dynamic message when phone exists', () => {
  const result = buildListingContactLinks(
    '081234567890',
    'Rumah Minimalis Bekasi',
    'https://propti.id/listings/abc123'
  );
  const expectedMessage =
    'Halo, saya tertarik dengan properti *Rumah Minimalis Bekasi* di Propti. Apakah masih tersedia?\nhttps://propti.id/listings/abc123';
  assert.equal(result.whatsappUrl, `https://wa.me/6281234567890?text=${encodeURIComponent(expectedMessage)}`);
  assert.equal(result.phoneUrl, 'tel:+6281234567890');
});

test('buildListingContactLinks without title/url falls back to default message', () => {
  const result = buildListingContactLinks('081234567890');
  const expectedMessage = 'Halo, saya tertarik dengan properti di Propti. Apakah masih tersedia?';
  assert.equal(result.whatsappUrl, `https://wa.me/6281234567890?text=${encodeURIComponent(expectedMessage)}`);
});

test('buildListingContactLinks returns null links when phone is empty', () => {
  assert.deepEqual(buildListingContactLinks(''), {
    whatsappUrl: null,
    phoneUrl: null,
  });
});
