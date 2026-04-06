import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

test('listing form schema includes phone and create flow references it', () => {
  const formFile = readFileSync(new URL('../components/listings/ListingForm.tsx', import.meta.url), 'utf8');

  assert.match(formFile, /phone:/);
  assert.match(formFile, /Nomor Telepon/);
  assert.match(formFile, /mode === 'create'/);
  assert.match(formFile, /editListingSchema/);
});
