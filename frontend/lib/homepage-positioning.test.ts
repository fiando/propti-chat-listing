import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const homePage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');

test('homepage hero sells Propti as a balanced buyer and seller workspace', () => {
  assert.match(homePage, /Workspace properti untuk penjual & pembeli/i);
  assert.match(homePage, /Pasang Listing Lebih Rapi/i);
  assert.match(homePage, /Cari Properti Lebih Yakin/i);
  assert.match(homePage, /Bayar untuk workflow aktif, bukan sekadar upload massal/i);
  assert.doesNotMatch(homePage, /Semudah Chat WhatsApp/i);
});

test('homepage proof section reinforces Propti win angle as workflow and closing support', () => {
  assert.match(homePage, /Cari, simpan, dan shortlist properti yang relevan/i);
  assert.match(homePage, /Sudut menang Propti: bantu closing, bukan tambah portal/i);
  assert.match(homePage, /Portal besar menang di traffic/i);
  assert.match(homePage, /Fitur pembeli seperti cari, simpan, dan alat bantu keputusan tetap bisa dipakai gratis/i);
  assert.match(homePage, /Tetap opsional, bukan langkah wajib untuk semua transaksi/i);
  assert.match(homePage, /Limit dihitung dari listing aktif yang sedang berjalan/i);
  assert.doesNotMatch(homePage, /OLX|Rumah123|Ray White/i);
});
