import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('homepage hero promotes WhatsApp-direct listing creation', () => {
  assert.match(homePage, /Pasang Listing Properti/i);
  assert.match(homePage, /Langsung dari WhatsApp/i);
  assert.match(homePage, /Kirim pesan ke nomor WhatsApp Propti/i);
  assert.match(homePage, /voice note/i);
  assert.doesNotMatch(homePage, /Semudah Chat WhatsApp/i);
});

test('homepage proof section highlights WhatsApp-first USPs', () => {
  assert.match(homePage, /Kirim WA, listing langsung terbuat/i);
  assert.match(homePage, /Voice note diubah jadi listing otomatis/i);
  assert.match(homePage, /ditranskripsi otomatis/i);
  assert.match(homePage, /Kelola listing sepenuhnya dari WhatsApp/i);
});
