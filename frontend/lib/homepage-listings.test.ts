import test from 'node:test';
import assert from 'node:assert/strict';
import { selectHomepageListings } from './homepage-listings.ts';

const listings = [
  {
    listingId: 'old-featured',
    title: 'Old Featured',
    status: 'active',
    moderationStatus: 'approved',
    premiumFeatures: { isFeatured: true, isPremium: true },
    createdAt: '2026-03-10T00:00:00Z',
    location: { city: 'Depok', district: 'Beji' },
    propertyDetails: { landArea: 120, buildingArea: 90, bedrooms: 3, bathrooms: 2 },
    images: [],
    listingType: 'sell',
    price: 850000000,
  },
  {
    listingId: 'new-standard',
    title: 'New Standard',
    status: 'active',
    moderationStatus: 'approved',
    premiumFeatures: { isFeatured: false, isPremium: false },
    createdAt: '2026-03-12T00:00:00Z',
    location: { city: 'Jakarta Selatan', district: 'Kemang' },
    propertyDetails: { landArea: 0, buildingArea: 65, bedrooms: 2, bathrooms: 1 },
    images: [],
    listingType: 'rent',
    price: 8500000,
  },
  {
    listingId: 'new-featured',
    title: 'New Featured',
    status: 'active',
    moderationStatus: 'approved',
    premiumFeatures: { isFeatured: true, isPremium: true },
    createdAt: '2026-03-13T00:00:00Z',
    location: { city: 'Bandung', district: 'Dago' },
    propertyDetails: { landArea: 180, buildingArea: 150, bedrooms: 4, bathrooms: 3 },
    images: [],
    listingType: 'sell',
    price: 1200000000,
  },
  {
    listingId: 'pending-listing',
    title: 'Pending Listing',
    status: 'active',
    moderationStatus: 'pending',
    premiumFeatures: { isFeatured: true, isPremium: true },
    createdAt: '2026-03-14T00:00:00Z',
    location: { city: 'Bogor', district: 'Cibinong' },
    propertyDetails: { landArea: 200, buildingArea: 0, bedrooms: 0, bathrooms: 0 },
    images: [],
    listingType: 'sell',
    price: 320000000,
  },
];

test('selectHomepageListings prefers approved active featured listings, then newest listings', () => {
  const selected = selectHomepageListings(listings as never[], 3);
  assert.deepEqual(selected.map((listing) => listing.listingId), ['new-featured', 'old-featured', 'new-standard']);
});

test('selectHomepageListings falls back to curated listings when no real listings qualify', () => {
  const selected = selectHomepageListings([], 6);
  assert.equal(selected.length, 6);
  assert.ok(selected.every((listing) => listing.title));
});
