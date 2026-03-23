import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildListingShareMessage,
  buildWhatsAppShareUrl,
  summarizeOwnerListings,
} from './listing-share.ts';

const baseListing = {
  listingId: 'listing-123',
  userId: 'user-1',
  title: 'Rumah Minimalis Depok',
  description: 'Rumah siap huni dekat stasiun.',
  price: 850000000,
  priceUnit: 'total',
  listingType: 'sell',
  status: 'active',
  propertyDetails: {
    landArea: 120,
    buildingArea: 90,
    bedrooms: 3,
    bathrooms: 2,
    amenities: [],
  },
  location: {
    address: 'Jl. Margonda Raya No. 1',
    city: 'Depok',
    district: 'Beji',
    province: 'Jawa Barat',
  },
  images: [],
  videos: [],
  imageCount: 0,
  premiumFeatures: {
    isPremium: false,
    isFeatured: false,
  },
  views: 12,
  saves: 4,
  moderationStatus: 'approved',
  createdAt: '2026-03-20T00:00:00Z',
  updatedAt: '2026-03-20T00:00:00Z',
};

test('buildListingShareMessage creates a seller-ready summary for WhatsApp distribution', () => {
  assert.equal(
    buildListingShareMessage(baseListing, 'https://propti.id/listings/listing-123'),
    'Rumah Minimalis Depok\nRp 850 Jt\nBeji, Depok\nLihat detail lengkap di Propti:\nhttps://propti.id/listings/listing-123'
  );
});

test('buildWhatsAppShareUrl encodes the seller-ready message into a wa.me deep link', () => {
  const message =
    'Rumah Minimalis Depok\nRp 850 Jt\nBeji, Depok\nLihat detail lengkap di Propti:\nhttps://propti.id/listings/listing-123';

  assert.equal(
    buildWhatsAppShareUrl(message),
    `https://wa.me/?text=${encodeURIComponent(message)}`
  );
});

test('summarizeOwnerListings aggregates owner listing counts and intent signals', () => {
  const summary = summarizeOwnerListings([
    baseListing,
    {
      ...baseListing,
      listingId: 'listing-456',
      status: 'archived',
      views: 3,
      saves: 1,
    },
    {
      ...baseListing,
      listingId: 'listing-789',
      views: 8,
      saves: 2,
    },
  ]);

  assert.deepEqual(summary, {
    totalListings: 3,
    activeListings: 2,
    totalViews: 23,
    totalSaves: 7,
  });
});
