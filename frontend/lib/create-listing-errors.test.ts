import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const helpersModule = await import('./create-listing-errors.ts').catch(() => ({} as Record<string, unknown>));

const getCreateListingErrorMessage =
  (helpersModule.getCreateListingErrorMessage as
    | ((error: unknown) => string)
    | undefined) ??
  ((error: unknown) => (error instanceof Error ? error.message : 'Terjadi kesalahan saat memasang iklan.'));
const getCreateListingAccessState =
  (helpersModule.getCreateListingAccessState as
    | ((input: {
        isAuthenticated?: boolean;
        isPremium?: boolean;
        isAuthLoading?: boolean;
        isListingsLoading?: boolean;
        hasListingsError?: boolean;
        listings?: Array<{ status?: string } | null | undefined> | null;
        totalListings?: number;
      }) => { status: string; activeListingsCount: number; message?: string })
    | undefined) ??
  ((input: {
    isAuthenticated?: boolean;
    isPremium?: boolean;
    isAuthLoading?: boolean;
    isListingsLoading?: boolean;
    hasListingsError?: boolean;
    listings?: Array<{ status?: string } | null | undefined> | null;
    totalListings?: number;
  }) => ({
    status: input.isPremium ? 'ready' : 'ready',
    activeListingsCount: input.listings?.filter((listing) => listing?.status === 'active').length ?? 0,
  }));

const createPageSource = readFileSync(
  new URL('../app/(app)/listings/create/page.tsx', import.meta.url),
  'utf8'
);

test('maps free tier listing limit errors to a clear blocking message', () => {
  assert.equal(
    getCreateListingErrorMessage(new Error('free tier allows at most 3 listing(s)')),
    'Paket gratis hanya bisa memiliki 3 listing aktif. Upgrade ke Premium untuk memasang iklan baru.'
  );
  assert.equal(
    getCreateListingErrorMessage(new Error('free tier allows at most 3 active listing(s)')),
    'Paket gratis hanya bisa memiliki 3 listing aktif. Upgrade ke Premium untuk memasang iklan baru.'
  );
  assert.match(
    getCreateListingErrorMessage(new Error('free tier allows at most 3 listing(s)')),
    /3 listing/i
  );
});

test('create listing page catches submit errors and surfaces them with a user-facing helper', () => {
  assert.match(createPageSource, /getCreateListingErrorMessage/);
  assert.match(createPageSource, /toast\(getCreateListingErrorMessage\(error\), 'error'\)/);
});

test('keeps create access unresolved while auth or listing checks are still loading', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      isAuthLoading: true,
      listings: [{ status: 'active' }],
      totalListings: 1,
    }),
    {
      status: 'checking',
      activeListingsCount: 1,
    }
  );

  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      isListingsLoading: true,
      listings: [{ status: 'active' }],
      totalListings: 1,
    }),
    {
      status: 'checking',
      activeListingsCount: 1,
    }
  );
});

test('fails closed when listing access data is incomplete or the quota check errors', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      listings: [{ status: 'active' }, { status: 'inactive' }],
      totalListings: 4,
    }),
    {
      status: 'error',
      activeListingsCount: 1,
      message: 'Kami belum bisa memastikan slot listing aktifmu. Coba lagi sebentar lagi.',
    }
  );

  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      hasListingsError: true,
      listings: [{ status: 'active' }],
      totalListings: 1,
    }),
    {
      status: 'error',
      activeListingsCount: 1,
      message: 'Kami belum bisa memastikan slot listing aktifmu. Coba lagi sebentar lagi.',
    }
  );
});
