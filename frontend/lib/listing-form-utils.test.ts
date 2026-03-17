import test from 'node:test';
import assert from 'node:assert/strict';

import {
  formatListingPriceInput,
  normalizeAmenityIds,
  parseListingPriceInput,
} from './listing-form-utils.ts';

test('formatListingPriceInput adds thousands separators on the fly', () => {
  assert.equal(formatListingPriceInput(850000000), '850.000.000');
  assert.equal(formatListingPriceInput(0), '');
});

test('parseListingPriceInput strips non-digits and returns a numeric value', () => {
  assert.equal(parseListingPriceInput('850.000.000'), 850000000);
  assert.equal(parseListingPriceInput('Rp 1.250.000'), 1250000);
  assert.equal(parseListingPriceInput(''), 0);
});

test('normalizeAmenityIds maps parsed amenity labels into known amenity ids without duplicates', () => {
  assert.deepEqual(
    normalizeAmenityIds(['carport', 'taman', 'ruang tamu', 'dapur', 'Ruang Tamu']),
    ['carport', 'taman', 'ruang_tamu', 'dapur']
  );
});
