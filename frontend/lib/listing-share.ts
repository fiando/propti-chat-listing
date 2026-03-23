type ShareListingLocation = {
  district?: string;
  city?: string;
  province?: string;
};

type ShareListingRecord = {
  listingId: string;
  title: string;
  description: string;
  price: number;
  status: string;
  views: number;
  saves: number;
  contactReveals?: number;
  location?: ShareListingLocation;
  featuredThumbnailUrl?: string;
};

function formatSharePrice(price: number): string {
  if (price >= 1_000_000_000) {
    return `Rp ${(price / 1_000_000_000).toFixed(1)} Mlr`;
  }
  if (price >= 1_000_000) {
    return `Rp ${(price / 1_000_000).toFixed(0)} Jt`;
  }
  return `Rp ${price.toLocaleString('id-ID')}`;
}

function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength).trimEnd()}...`;
}

export function buildListingShareUrl(listingId: string, origin = 'https://propti.id') {
  return new URL(`/listings/${listingId}`, origin).toString();
}

export function buildListingShareLocation(location?: ShareListingLocation) {
  if (!location) {
    return '';
  }

  return [location.district, location.city].filter(Boolean).join(', ') || location.province || '';
}

export function buildListingShareMessage(listing: ShareListingRecord, listingUrl: string) {
  const lines = [
    listing.title,
    formatSharePrice(listing.price),
    buildListingShareLocation(listing.location),
    'Lihat detail lengkap di Propti:',
    listingUrl,
  ].filter(Boolean);

  return lines.join('\n');
}

export function buildWhatsAppShareUrl(message: string) {
  return `https://wa.me/?text=${encodeURIComponent(message)}`;
}

export function summarizeOwnerListings(
  listings: Array<Pick<ShareListingRecord, 'status' | 'views' | 'saves' | 'contactReveals'>>
) {
  return listings.reduce(
    (summary, listing) => ({
      totalListings: summary.totalListings + 1,
      activeListings: summary.activeListings + (listing.status === 'active' ? 1 : 0),
      totalViews: summary.totalViews + listing.views,
      totalSaves: summary.totalSaves + listing.saves,
      totalContactReveals: summary.totalContactReveals + (listing.contactReveals ?? 0),
    }),
    {
      totalListings: 0,
      activeListings: 0,
      totalViews: 0,
      totalSaves: 0,
      totalContactReveals: 0,
    }
  );
}

export function buildListingMetadata(listing: ShareListingRecord) {
  const canonical = `/listings/${listing.listingId}`;
  const shareDescription = truncateText(
    `${listing.title} — ${formatSharePrice(listing.price)}${buildListingShareLocation(listing.location) ? ` • ${buildListingShareLocation(listing.location)}` : ''}. ${listing.description}`,
    160
  );
  const images = listing.featuredThumbnailUrl
    ? [
        {
          url: listing.featuredThumbnailUrl,
          alt: listing.title,
        },
      ]
    : undefined;

  return {
    title: listing.title,
    description: shareDescription,
    alternates: {
      canonical,
    },
    openGraph: {
      title: listing.title,
      description: shareDescription,
      url: canonical,
      images,
    },
    twitter: {
      card: 'summary_large_image' as const,
      title: listing.title,
      description: shareDescription,
      images: images?.map((image) => image.url),
    },
  };
}
