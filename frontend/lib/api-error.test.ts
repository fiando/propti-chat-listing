import test from 'node:test';
import assert from 'node:assert/strict';

const helpersModule = await import('./api-error.ts').catch(() => ({} as Record<string, unknown>));

const getApiErrorMessage =
  (helpersModule.getApiErrorMessage as
    | ((payload: unknown, fallbackMessage?: string) => string)
    | undefined) ??
  ((_payload: unknown, fallbackMessage?: string) =>
    fallbackMessage || 'Terjadi kesalahan. Silakan coba lagi.');

test('extracts backend error field when message is absent', () => {
  assert.equal(
    getApiErrorMessage(
      {
        error: 'free tier allows at most 3 listing(s)',
        code: 403,
      },
      'Request failed with status code 403'
    ),
    'free tier allows at most 3 listing(s)'
  );
});

test('keeps the backend rate-limit message for contact reveal errors', () => {
  assert.equal(
    getApiErrorMessage(
      {
        error: 'Terlalu banyak membuka kontak penjual. Coba lagi dalam 10 menit.',
        code: 429,
      },
      'Request failed with status code 429'
    ),
    'Terlalu banyak membuka kontak penjual. Coba lagi dalam 10 menit.'
  );
});
