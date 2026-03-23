import type { MetadataRoute } from 'next';
import type { Listing } from '@/types';

const SITE_URL = 'https://propti.id';
const LISTING_API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

type ListingsApiResponse = {
  items?: Listing[];
  listings?: Listing[];
};

async function getActiveListings(): Promise<Listing[]> {
  try {
    const response = await fetch(`${LISTING_API_BASE}/listings?pageSize=500`, {
      next: { revalidate: 1800 },
    });

    if (!response.ok) {
      return [];
    }

    const data = (await response.json()) as ListingsApiResponse;
    return data.items ?? data.listings ?? [];
  } catch {
    return [];
  }
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const listings = await getActiveListings();

  const listingEntries = listings.map((listing) => ({
    url: `${SITE_URL}/listings/${listing.listingId}`,
    lastModified: listing.updatedAt || listing.createdAt,
    changeFrequency: 'daily' as const,
    priority: 0.8,
  }));

  return [
    {
      url: SITE_URL,
      lastModified: new Date().toISOString(),
      changeFrequency: 'daily',
      priority: 1,
    },
    {
      url: `${SITE_URL}/listings`,
      lastModified: new Date().toISOString(),
      changeFrequency: 'hourly',
      priority: 0.9,
    },
    ...listingEntries,
  ];
}
