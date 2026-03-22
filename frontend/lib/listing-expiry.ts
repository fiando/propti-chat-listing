import type { Listing } from '@/types';

type ListingExpiryInfo = {
  kind: 'active' | 'expired';
  label: string;
  detail: string;
  tone: string;
};

function formatExpiryDate(expiresAt: string): string {
  return new Date(expiresAt).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function getTime(value: string | undefined): number | null {
  if (!value) {
    return null;
  }

  const parsed = new Date(value).getTime();
  return Number.isFinite(parsed) ? parsed : null;
}

export function isListingExpired(listing: Pick<Listing, 'expiresAt'>, now = Date.now()): boolean {
  const expiresAt = getTime(listing.expiresAt);
  if (expiresAt === null) {
    return false;
  }

  return expiresAt <= now;
}

export function shouldShowRelistAction(
  listing: Pick<Listing, 'status' | 'expiresAt'>,
  now = Date.now()
): boolean {
  return listing.status === 'archived' && isListingExpired(listing, now);
}

export function getListingExpiryInfo(
  listing: Pick<Listing, 'status' | 'expiresAt'>,
  now = Date.now()
): ListingExpiryInfo | null {
  if (!listing.expiresAt) {
    return null;
  }

  const expiresAt = getTime(listing.expiresAt);
  if (expiresAt === null) {
    return null;
  }

  const formattedDate = formatExpiryDate(listing.expiresAt);
  if (listing.status === 'archived' || expiresAt <= now) {
    return {
      kind: 'expired',
      label: 'Arsip',
      detail: `Listing berakhir pada ${formattedDate}`,
      tone: 'border-gray-200 bg-gray-50 text-gray-700',
    };
  }

  return {
    kind: 'active',
    label: `Aktif sampai ${formattedDate}`,
    detail: 'Listing masih tampil ke publik sampai tanggal ini.',
    tone: 'border-emerald-200 bg-emerald-50 text-emerald-700',
  };
}

