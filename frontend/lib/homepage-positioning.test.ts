import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('homepage hero sells Propti as a balanced buyer and seller workspace', () => {
  assert.match(homePage, /Workspace properti untuk penjual & pembeli/i);
  assert.match(homePage, /Pasang Listing Lebih Rapi/i);
  assert.match(homePage, /Cari Properti Lebih Yakin/i);
  assert.match(homePage, /Bukan portal listing massal/i);
  assert.doesNotMatch(homePage, /Semudah Chat WhatsApp/i);
});

test('homepage proof section reinforces must-have buyer and seller MVP value', () => {
  assert.match(homePage, /Cari, simpan, dan shortlist properti yang relevan/i);
  assert.match(homePage, /MVP yang seimbang untuk dua sisi transaksi/i);
  assert.match(homePage, /Fitur buyer seperti cari, simpan, dan kalkulator KPR tetap bisa dipakai gratis/i);
  assert.doesNotMatch(homePage, /OLX|Rumah123/i);
});
