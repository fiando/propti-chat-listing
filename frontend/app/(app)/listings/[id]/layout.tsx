import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import type { Listing } from '@/types';
import { buildListingMetadata } from '@/lib/listing-share';

const LISTING_API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id';

async function getPublicListing(id: string): Promise<Listing | null> {
  try {
    const response = await fetch(`${LISTING_API_BASE}/listings/${id}`, {
      next: { revalidate: 300 },
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as Listing;
  } catch {
    return null;
  }
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const { id } = await params;
  const listing = await getPublicListing(id);

  if (!listing) {
    return {
      title: 'Detail Properti',
      alternates: {
        canonical: `/listings/${id}`,
      },
    };
  }

  return buildListingMetadata(listing);
}

export default function ListingDetailLayout({ children }: { children: ReactNode }) {
  return children;
}
