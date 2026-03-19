import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import {
  CREATE_LISTING_ACCESS_ERROR_MESSAGE,
  getCreateListingAccessState,
  getCreateListingErrorMessage,
} from './create-listing-errors.ts';

const createPageSource = readFileSync(
  new URL('../components/listings/CreateListingClient.tsx', import.meta.url),
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
  assert.equal(
    getCreateListingErrorMessage(new Error('premium tier allows at most 15 active listing(s)')),
    'Paket Premium maksimal 15 listing aktif. Arsipkan salah satu listing untuk memasang iklan baru.'
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

test('keeps create access unresolved while the quota summary is refetching', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      isListingsFetching: true,
      activeListingsCount: 2,
      listings: [{ status: 'active' }, { status: 'active' }],
      totalListings: 2,
    }),
    {
      status: 'checking',
      activeListingsCount: 2,
    }
  );
});

test('keeps create access unresolved until this mount has a fresh access result', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      hasFreshAccessResult: false,
      activeListingsCount: 2,
    }),
    {
      status: 'checking',
      activeListingsCount: 2,
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
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
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
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
    }
  );
});

test('surfaces a retryable access error when the quota request fails or returns no usable summary', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
      hasListingsError: true,
    }),
    {
      status: 'error',
      activeListingsCount: 0,
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
    }
  );

  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: false,
    }),
    {
      status: 'error',
      activeListingsCount: 0,
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
    }
  );
});

test('blocks premium users after 15 active listings', () => {
  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: true,
      activeListingsCount: 15,
    }),
    {
      status: 'blocked',
      activeListingsCount: 15,
      message: 'Paket Premium maksimal 15 listing aktif. Arsipkan salah satu listing untuk memasang iklan baru.',
    }
  );

  assert.deepEqual(
    getCreateListingAccessState({
      isAuthenticated: true,
      isPremium: true,
      activeListingsCount: 14,
    }),
    {
      status: 'ready',
      activeListingsCount: 14,
    }
  );
});
