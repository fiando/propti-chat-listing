import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const listingDetailFile = readFileSync(
  new URL('../components/listings/ListingDetail.tsx', import.meta.url),
  'utf8'
);
const settingsPageFile = readFileSync(new URL('../app/(app)/settings/page.tsx', import.meta.url), 'utf8');

test('listing detail uses a zoom-capable lightbox for the gallery popup', () => {
  assert.match(listingDetailFile, /yet-another-react-lightbox/);
  assert.match(listingDetailFile, /yet-another-react-lightbox\/plugins\/zoom/);
  assert.match(listingDetailFile, /plugins=\{\[Zoom\]\}/);
  assert.match(listingDetailFile, /const \[isLightboxOpen, setIsLightboxOpen\] = useState\(false\)/);
  assert.match(listingDetailFile, /onClick=\{\(\) => setIsLightboxOpen\(true\)\}/);
});

test('settings page only shows the optional role placeholder when no role is selected', () => {
  assert.match(settingsPageFile, /\{role === '' && <option value="">Pilih peran akun \(opsional\)<\/option>\}/);
});
