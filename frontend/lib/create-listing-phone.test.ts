import test from 'node:test';
import assert from 'node:assert/strict';

const helpersModule = await import('./create-listing-phone.ts').catch(() => ({} as Record<string, unknown>));

const shouldRequirePhone =
  (helpersModule.shouldRequirePhone as
    | ((input: { profilePhone?: string | null; phoneOverride?: string | null }) => boolean)
    | undefined) ??
  (() => true);

const getPhoneModalSubmitLabel =
  (helpersModule.getPhoneModalSubmitLabel as
    | ((input: { isSavingPhone: boolean; isSubmittingListing: boolean }) => string)
    | undefined) ??
  (() => 'Simpan & Posting Iklan');

test('does not require phone again when a fresh phone override was just provided', () => {
  assert.equal(
    shouldRequirePhone({ profilePhone: '', phoneOverride: '08123456789' }),
    false
  );
});

test('shows explicit progress text while saving phone and while posting listing', () => {
  assert.equal(
    getPhoneModalSubmitLabel({ isSavingPhone: true, isSubmittingListing: false }),
    'Menyimpan nomor telepon...'
  );
  assert.equal(
    getPhoneModalSubmitLabel({ isSavingPhone: false, isSubmittingListing: true }),
    'Memasang iklan...'
  );
});
