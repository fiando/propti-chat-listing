import test from 'node:test';
import assert from 'node:assert/strict';

import {
  getListingExpiryInfo,
  isListingExpired,
  shouldShowRelistAction,
} from './listing-expiry.ts';
import type { Listing } from '@/types';

function createListing(overrides: Partial<Listing> = {}): Listing {
  return {
    listingId: 'listing-1',
    userId: 'user-1',
    title: 'Rumah Depok',
    description: 'Deskripsi',
    price: 850000000,
    priceUnit: 'total',
    listingType: 'sell',
    status: 'active',
    propertyDetails: { landArea: 120, buildingArea: 90, bedrooms: 3, bathrooms: 2, amenities: [] },
    location: { city: 'Depok', district: 'Beji', address: 'Jl. Margonda' },
    images: [],
    videos: [],
    imageCount: 0,
    premiumFeatures: { isPremium: false, isFeatured: false },
    views: 0,
    saves: 0,
    moderationStatus: 'approved',
    createdAt: '2026-03-01T00:00:00Z',
    updatedAt: '2026-03-01T00:00:00Z',
    ...overrides,
  };
}

test('isListingExpired returns true once expiresAt has passed', () => {
  assert.equal(
    isListingExpired(createListing({ expiresAt: '2026-03-10T00:00:00Z' }), Date.parse('2026-03-11T00:00:00Z')),
    true
  );
});

test('shouldShowRelistAction only returns true for archived expired listings', () => {
  assert.equal(
    shouldShowRelistAction(
      createListing({ status: 'archived', expiresAt: '2026-03-10T00:00:00Z' }),
      Date.parse('2026-03-11T00:00:00Z')
    ),
    true
  );
  assert.equal(
    shouldShowRelistAction(
      createListing({ status: 'active', expiresAt: '2026-03-12T00:00:00Z' }),
      Date.parse('2026-03-11T00:00:00Z')
    ),
    false
  );
});

test('getListingExpiryInfo shows active-until feedback for active listings', () => {
  const info = getListingExpiryInfo(
    createListing({ expiresAt: '2026-03-20T00:00:00Z' }),
    Date.parse('2026-03-11T00:00:00Z')
  );

  assert.equal(info?.kind, 'active');
  assert.match(info?.label ?? '', /Aktif sampai/i);
  assert.match(info?.detail ?? '', /masih tayang/i);
});

test('getListingExpiryInfo shows expired archive feedback for archived listings', () => {
  const info = getListingExpiryInfo(
    createListing({ status: 'archived', expiresAt: '2026-03-10T00:00:00Z' }),
    Date.parse('2026-03-11T00:00:00Z')
  );

  assert.equal(info?.kind, 'expired');
  assert.match(info?.label ?? '', /Tayangan berakhir/i);
  assert.match(info?.detail ?? '', /tidak lagi tayang/i);
});
