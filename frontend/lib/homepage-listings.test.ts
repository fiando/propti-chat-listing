import test from 'node:test';
import assert from 'node:assert/strict';
import { buildHomepageListingSection } from './homepage-listings.ts';

const baseListing = {
  title: 'Sample Listing',
  status: 'active',
  moderationStatus: 'approved',
  createdAt: '2026-03-10T00:00:00Z',
  location: { city: 'Depok', district: 'Beji' },
  propertyDetails: { landArea: 120, buildingArea: 90, bedrooms: 3, bathrooms: 2 },
  images: [],
  listingType: 'sell',
  price: 850000000,
  premiumFeatures: { isFeatured: false, isPremium: false },
};

type HomepageListingFixture = typeof baseListing & { listingId: string };

function createListing({
  listingId,
  premiumFeatures,
  ...overrides
}: Partial<HomepageListingFixture> & Pick<HomepageListingFixture, 'listingId'>) {
  return {
    ...baseListing,
    ...overrides,
    listingId,
    premiumFeatures: {
      ...baseListing.premiumFeatures,
      ...premiumFeatures,
    },
  };
}

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

test('buildHomepageListingSection returns featured mode when approved featured listings exist', () => {
  const selected = buildHomepageListingSection(listings as never[], 3);

  assert.equal(selected.kind, 'featured');
  assert.equal(selected.title, 'Listing Pilihan');
  assert.equal(selected.subtitle, 'Properti aktif yang sudah lolos moderasi dan siap dihubungi.');
  assert.deepEqual(selected.items.map((listing) => listing.listingId), [
    'new-featured',
    'old-featured',
    'new-standard',
  ]);
});

test('featured homepage section prefers featured listings first and uses credible copy', () => {
  const section = buildHomepageListingSection([
    createListing({
      listingId: 'regular-1',
      createdAt: '2026-03-15T00:00:00Z',
      premiumFeatures: { isFeatured: false, isPremium: false },
    }),
    createListing({
      listingId: 'featured-1',
      createdAt: '2026-03-14T00:00:00Z',
      premiumFeatures: { isFeatured: true, isPremium: true },
    }),
  ] as never[]);

  assert.equal(section.kind, 'featured');
  assert.equal(section.items[0]?.listingId, 'featured-1');
  assert.equal(section.title, 'Listing Pilihan');
  assert.equal(section.subtitle, 'Properti aktif yang sudah lolos moderasi dan siap dihubungi.');
});

test('buildHomepageListingSection falls back to newest approved listings when no featured listings exist', () => {
  const selected = buildHomepageListingSection(
    [
      {
        listingId: 'older-standard',
        title: 'Older Standard',
        status: 'active',
        moderationStatus: 'approved',
        premiumFeatures: { isFeatured: false, isPremium: false },
        createdAt: '2026-03-09T00:00:00Z',
        location: { city: 'Depok', district: 'Beji' },
        propertyDetails: { landArea: 120, buildingArea: 90, bedrooms: 3, bathrooms: 2 },
        images: [],
        listingType: 'sell',
        price: 850000000,
      },
      {
        listingId: 'newer-standard',
        title: 'Newer Standard',
        status: 'active',
        moderationStatus: 'approved',
        premiumFeatures: { isFeatured: false, isPremium: false },
        createdAt: '2026-03-11T00:00:00Z',
        location: { city: 'Bogor', district: 'Cibinong' },
        propertyDetails: { landArea: 90, buildingArea: 65, bedrooms: 2, bathrooms: 1 },
        images: [],
        listingType: 'sell',
        price: 650000000,
      },
    ] as never[],
    2
  );

  assert.equal(selected.kind, 'latest');
  assert.equal(selected.title, 'Listing Terbaru');
  assert.equal(selected.subtitle, 'Listing aktif terbaru yang sudah lolos moderasi dan siap dihubungi.');
  assert.deepEqual(selected.items.map((listing) => listing.listingId), [
    'newer-standard',
    'older-standard',
  ]);
});

test('buildHomepageListingSection returns empty mode when no approved active listings exist', () => {
  const selected = buildHomepageListingSection([], 6);

  assert.equal(selected.kind, 'empty');
  assert.equal(selected.title, '');
  assert.deepEqual(selected.items, []);
});
