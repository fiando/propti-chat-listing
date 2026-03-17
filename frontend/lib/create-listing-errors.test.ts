import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const helpersModule = await import('./create-listing-errors.ts').catch(() => ({} as Record<string, unknown>));

const getCreateListingErrorMessage =
  (helpersModule.getCreateListingErrorMessage as
    | ((error: unknown) => string)
    | undefined) ??
  ((error: unknown) => (error instanceof Error ? error.message : 'Terjadi kesalahan saat memasang iklan.'));

const createPageSource = readFileSync(
  new URL('../app/(app)/listings/create/page.tsx', import.meta.url),
  'utf8'
);

test('maps free tier listing limit errors to a clear Indonesian message', () => {
  assert.equal(
    getCreateListingErrorMessage(new Error('free tier allows at most 3 listing(s)')),
    'Paket gratis hanya bisa memasang 3 listing. Upgrade ke Premium untuk memasang lebih banyak.'
  );
});

test('create listing page catches submit errors and surfaces them with a user-facing helper', () => {
  assert.match(createPageSource, /getCreateListingErrorMessage/);
  assert.match(createPageSource, /toast\(getCreateListingErrorMessage\(error\), 'error'\)/);
});
