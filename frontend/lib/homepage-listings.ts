import type { Listing, ListingsResponse } from '@/types';

const DEFAULT_LIMIT = 6;

const CURATED_FALLBACK_LISTINGS: Listing[] = [
  createFallbackListing({
    listingId: 'fallback-depok-modern',
    title: 'Rumah Minimalis Modern Depok',
    price: 850_000_000,
    listingType: 'sell',
    city: 'Depok',
    district: 'Beji',
    landArea: 120,
    buildingArea: 90,
    bedrooms: 3,
    bathrooms: 2,
    isFeatured: true,
  }),
  createFallbackListing({
    listingId: 'fallback-bsd-ruko',
    title: 'Ruko 3 Lantai Strategis BSD City',
    price: 2_500_000_000,
    listingType: 'sell',
    city: 'Tangerang Selatan',
    district: 'BSD City',
    landArea: 90,
    buildingArea: 270,
    bedrooms: 0,
    bathrooms: 3,
  }),
  createFallbackListing({
    listingId: 'fallback-kemang-apartment',
    title: 'Apartemen 2BR Kemang Jakarta Selatan',
    price: 8_500_000,
    listingType: 'rent',
    city: 'Jakarta Selatan',
    district: 'Kemang',
    landArea: 0,
    buildingArea: 65,
    bedrooms: 2,
    bathrooms: 1,
    isFeatured: true,
  }),
  createFallbackListing({
    listingId: 'fallback-bandung-cluster',
    title: 'Rumah Cluster Premium Bandung',
    price: 1_200_000_000,
    listingType: 'sell',
    city: 'Bandung',
    district: 'Dago',
    landArea: 180,
    buildingArea: 150,
    bedrooms: 4,
    bathrooms: 3,
  }),
  createFallbackListing({
    listingId: 'fallback-bogor-land',
    title: 'Kavling Siap Bangun Bogor',
    price: 320_000_000,
    listingType: 'sell',
    city: 'Bogor',
    district: 'Cibinong',
    landArea: 200,
    buildingArea: 0,
    bedrooms: 0,
    bathrooms: 0,
  }),
  createFallbackListing({
    listingId: 'fallback-surabaya-kost',
    title: 'Kost Eksklusif AC Surabaya',
    price: 1_800_000,
    listingType: 'rent',
    city: 'Surabaya',
    district: 'Rungkut',
    landArea: 0,
    buildingArea: 20,
    bedrooms: 1,
    bathrooms: 1,
  }),
];

type ListingsApiResponse = Partial<ListingsResponse> & {
  items?: Listing[];
  listings?: Listing[];
};

type FallbackListingInput = {
  listingId: string;
  title: string;
  price: number;
  listingType: Listing['listingType'];
  city: string;
  district: string;
  landArea: number;
  buildingArea: number;
  bedrooms: number;
  bathrooms: number;
  isFeatured?: boolean;
};

function createFallbackListing(input: FallbackListingInput): Listing {
  return {
    listingId: input.listingId,
    userId: 'homepage-fallback',
    title: input.title,
    description: `${input.title} di ${input.district}, ${input.city}`,
    price: input.price,
    priceUnit: input.listingType === 'rent' ? 'monthly' : 'total',
    listingType: input.listingType,
    status: 'active',
    propertyDetails: {
      landArea: input.landArea,
      buildingArea: input.buildingArea,
      bedrooms: input.bedrooms,
      bathrooms: input.bathrooms,
      amenities: [],
    },
    location: {
      address: `${input.district}, ${input.city}`,
      city: input.city,
      district: input.district,
    },
    images: [],
    videos: [],
    imageCount: 0,
    premiumFeatures: {
      isPremium: Boolean(input.isFeatured),
      isFeatured: Boolean(input.isFeatured),
    },
    views: 0,
    saves: 0,
    moderationStatus: 'approved',
    createdAt: '2026-03-01T00:00:00Z',
    updatedAt: '2026-03-01T00:00:00Z',
  };
}

function normalizeListingsResponse(data: ListingsApiResponse): Listing[] {
  return data.items ?? data.listings ?? [];
}

export function selectHomepageListings(listings: Listing[], limit = DEFAULT_LIMIT): Listing[] {
  const eligible = listings.filter(
    (listing) => listing.status === 'active' && listing.moderationStatus === 'approved'
  );

  if (eligible.length === 0) {
    return CURATED_FALLBACK_LISTINGS.slice(0, limit);
  }

  const featured = eligible
    .filter((listing) => listing.premiumFeatures?.isFeatured)
    .sort((left, right) => right.createdAt.localeCompare(left.createdAt));
  const nonFeatured = eligible
    .filter((listing) => !listing.premiumFeatures?.isFeatured)
    .sort((left, right) => right.createdAt.localeCompare(left.createdAt));

  return [...featured, ...nonFeatured].slice(0, limit);
}

export async function getHomepageListings(limit = DEFAULT_LIMIT): Promise<Listing[]> {
  const baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

  try {
    const response = await fetch(`${baseURL}/listings`, {
      next: { revalidate: 300 },
    });
    if (!response.ok) {
      return CURATED_FALLBACK_LISTINGS.slice(0, limit);
    }

    const data = (await response.json()) as ListingsApiResponse;
    const selected = selectHomepageListings(normalizeListingsResponse(data), limit);

    if (selected.length >= limit) {
      return selected;
    }

    const fallbackRemainder = CURATED_FALLBACK_LISTINGS.filter(
      (fallback) => !selected.some((listing) => listing.listingId === fallback.listingId)
    ).slice(0, limit - selected.length);

    return [...selected, ...fallbackRemainder];
  } catch {
    return CURATED_FALLBACK_LISTINGS.slice(0, limit);
  }
}
