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
  assert.match(homePage, /workspace transaksi properti/i);
  assert.match(homePage, /Fitur pembeli seperti cari, simpan, dan alat bantu keputusan tetap bisa dipakai gratis/i);
  assert.doesNotMatch(homePage, /Kalkulator KPR|\/kpr/i);
  assert.doesNotMatch(homePage, /\/agent|Agent Tool|Dasbor Penjual/i);
  assert.doesNotMatch(homePage, /OLX|Rumah123/i);
});
