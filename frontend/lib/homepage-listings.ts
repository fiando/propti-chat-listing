import type { Listing, ListingsResponse } from '@/types';

const DEFAULT_LIMIT = 6;

type ListingsApiResponse = Partial<ListingsResponse> & {
  items?: Listing[];
  listings?: Listing[];
};

function normalizeListingsResponse(data: ListingsApiResponse): Listing[] {
  return data.items ?? data.listings ?? [];
}

function hasFutureExpiry(listing: Listing, now: number): boolean {
  if (!listing.expiresAt) {
    return true;
  }

  const expiresAt = new Date(listing.expiresAt).getTime();
  return Number.isFinite(expiresAt) && expiresAt > now;
}

function getEligibleHomepageListings(listings: Listing[]) {
  const now = Date.now();
  return listings.filter(
    (listing) =>
      listing.status === 'active' &&
      listing.moderationStatus === 'approved' &&
      hasFutureExpiry(listing, now)
  );
}

function sortNewestFirst(listings: Listing[]) {
  return [...listings].sort((left, right) => right.createdAt.localeCompare(left.createdAt));
}

export type HomepageListingSection =
  | {
      kind: 'featured';
      title: 'Listing Pilihan';
      subtitle: string;
      items: Listing[];
    }
  | {
      kind: 'latest';
      title: 'Listing Terbaru';
      subtitle: string;
      items: Listing[];
    }
  | { kind: 'empty'; title: ''; subtitle: ''; items: [] };

export function selectHomepageListings(listings: Listing[], limit = DEFAULT_LIMIT): Listing[] {
  return buildHomepageListingSection(listings, limit).items;
}

export function buildHomepageListingSection(
  listings: Listing[],
  limit = DEFAULT_LIMIT
): HomepageListingSection {
  const eligible = getEligibleHomepageListings(listings);

  if (eligible.length === 0) {
    return {
      kind: 'empty',
      title: '',
      subtitle: '',
      items: [],
    };
  }

  const featured = sortNewestFirst(
    eligible.filter((listing) => listing.premiumFeatures?.isFeatured)
  );
  const nonFeatured = sortNewestFirst(
    eligible.filter((listing) => !listing.premiumFeatures?.isFeatured)
  );

  if (featured.length > 0) {
    return {
      kind: 'featured',
      title: 'Listing Pilihan',
      subtitle: 'Properti aktif yang sudah lolos moderasi dan siap dihubungi.',
      items: [...featured, ...nonFeatured].slice(0, limit),
    };
  }

  return {
    kind: 'latest',
    title: 'Listing Terbaru',
    subtitle: 'Listing aktif terbaru yang sudah lolos moderasi dan siap dihubungi.',
    items: nonFeatured.slice(0, limit),
  };
}

export function selectEligibleHomepageListings(listings: Listing[], limit = DEFAULT_LIMIT): Listing[] {
  const eligible = listings.filter(
    (listing) =>
      listing.status === 'active' &&
      listing.moderationStatus === 'approved' &&
      hasFutureExpiry(listing, Date.now())
  );
  return sortNewestFirst(eligible).slice(0, limit);
}

export async function getHomepageListings(
  limit = DEFAULT_LIMIT
): Promise<HomepageListingSection> {
  const baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

  try {
    const response = await fetch(`${baseURL}/listings`, {
      next: { revalidate: 60 },
    });
    if (!response.ok) {
      return {
        kind: 'empty',
        title: '',
        subtitle: '',
        items: [],
      };
    }

    const data = (await response.json()) as ListingsApiResponse;
    return buildHomepageListingSection(normalizeListingsResponse(data), limit);
  } catch {
    return {
      kind: 'empty',
      title: '',
      subtitle: '',
      items: [],
    };
  }
}
