import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

test('listing form no longer owns phone input for website create flow', () => {
  const formFile = readFileSync(new URL('../components/listings/ListingForm.tsx', import.meta.url), 'utf8');
  assert.doesNotMatch(formFile, /phone: z\.string\(\)\.min\(1, 'Nomor telepon harus diisi'\)/);
  assert.doesNotMatch(formFile, /const createListingSchema = listingBaseSchema\.extend/);
  assert.doesNotMatch(formFile, /resolver: zodResolver\(mode === 'create' \? createListingSchema : editListingSchema\)/);
  assert.doesNotMatch(formFile, /<label className="label">Nomor Telepon \*<\/label>/);
});

test('create listing flow gates on missing profile phone instead of prefilling form phone', () => {
  const createPageFile = readFileSync(new URL('../app/(app)/listings/create/page.tsx', import.meta.url), 'utf8');
  const createClientFile = readFileSync(new URL('../components/listings/CreateListingClient.tsx', import.meta.url), 'utf8');

  assert.match(createPageFile, /initialProfilePhone=\{profile\?\.phone \?\? ''\}/);
  assert.match(createClientFile, /const \[showPhoneSetupModal, setShowPhoneSetupModal\] = useState\(false\)/);
  assert.match(createClientFile, /if \(isAuthenticated && !normalizedInitialProfilePhone\) \{\s*setShowPhoneSetupModal\(true\);\s*\}/);
  assert.match(createClientFile, /Lengkapi nomor telepon profil dulu/);
  assert.match(createClientFile, /Simpan Nomor di Profil/);
  assert.match(createClientFile, /router\.push\('\/profile\?returnTo=%2Flistings%2Fcreate'\)/);
  assert.doesNotMatch(createClientFile, /function seedDraftPhone\(/);
  assert.doesNotMatch(createClientFile, /await updateProfile\(\{ phone: data\.phone\.trim\(\) \}\)/);
  assert.match(createClientFile, /initialFormValues=\{formDraft \|\| undefined\}/);
});
