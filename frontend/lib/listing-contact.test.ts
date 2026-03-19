import test from 'node:test';
import assert from 'node:assert/strict';
import { buildListingContactLinks, normalizeContactPhone } from './listing-contact.ts';

test('normalizeContactPhone converts Indonesian local numbers to international format', () => {
  assert.equal(normalizeContactPhone('0812 3456 7890'), '6281234567890');
  assert.equal(normalizeContactPhone('+62 812-3456-7890'), '6281234567890');
});

test('buildListingContactLinks returns WhatsApp and phone links when phone exists', () => {
  assert.deepEqual(buildListingContactLinks('081234567890'), {
    whatsappUrl:
      'https://wa.me/6281234567890?text=Halo%2C%20saya%20tertarik%20dengan%20properti%20ini%20di%20Propti.%20Apakah%20masih%20tersedia%3F',
    phoneUrl: 'tel:+6281234567890',
  });
});

test('buildListingContactLinks returns null links when phone is empty', () => {
  assert.deepEqual(buildListingContactLinks(''), {
    whatsappUrl: null,
    phoneUrl: null,
  });
});
