import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('homepage hero sells Propti as the seller listing HQ', () => {
  assert.match(homePage, /Satu Listing Properti/i);
  assert.match(homePage, /yang Lebih Rapi/i);
  assert.match(homePage, /siap dibagikan ke semua\s+channel/i);
  assert.match(homePage, /pusat listingmu/i);
  assert.doesNotMatch(homePage, /Semudah Chat WhatsApp/i);
});

test('homepage proof section reinforces a share-ready listing position', () => {
  assert.match(homePage, /pusat listing yang lebih rapi, lebih dipercaya, dan siap kamu bagikan/i);
  assert.match(homePage, /Satu tempat untuk siap tayang dan siap share/i);
  assert.match(homePage, /sumber utama listingmu/i);
});
