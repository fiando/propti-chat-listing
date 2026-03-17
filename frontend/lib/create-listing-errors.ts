export const FREE_TIER_LISTING_LIMIT = 3;
export const LISTING_ACCESS_CHECK_PAGE_SIZE = 100;
export const FREE_TIER_LISTING_LIMIT_MESSAGE =
  `Paket gratis hanya bisa memiliki ${FREE_TIER_LISTING_LIMIT} listing aktif. Upgrade ke Premium untuk memasang iklan baru.`;
export const CREATE_LISTING_ACCESS_ERROR_MESSAGE =
  'Kami belum bisa memastikan slot listing aktifmu. Coba lagi sebentar lagi.';

type ListingLimitEntry = { status?: string | null } | null | undefined;
type CreateListingAccessState = {
  status: 'ready' | 'checking' | 'blocked' | 'error';
  activeListingsCount: number;
  message?: string;
};

export function getActiveListingCount(
  listings: Array<ListingLimitEntry> | null | undefined
): number {
  return (listings ?? []).filter((listing) => listing?.status === 'active').length;
}

export function isCreateListingLimitReached(input: {
  isPremium?: boolean;
  activeListingsCount?: number;
  listings?: Array<ListingLimitEntry> | null;
  limit?: number;
}): boolean {
  if (input.isPremium) {
    return false;
  }

  const activeListingsCount = input.activeListingsCount ?? getActiveListingCount(input.listings);
  return activeListingsCount >= (input.limit ?? FREE_TIER_LISTING_LIMIT);
}

export function getCreateListingAccessState(input: {
  isAuthenticated?: boolean;
  isPremium?: boolean;
  isAuthLoading?: boolean;
  isListingsLoading?: boolean;
  isListingsFetching?: boolean;
  hasListingsError?: boolean;
  activeListingsCount?: number;
  listings?: Array<ListingLimitEntry> | null;
  totalListings?: number;
  limit?: number;
}): CreateListingAccessState {
  const activeListingsCount = input.activeListingsCount ?? getActiveListingCount(input.listings);

  if (!input.isAuthenticated || input.isPremium) {
    return { status: 'ready', activeListingsCount };
  }

  if (input.isAuthLoading || input.isListingsLoading || input.isListingsFetching || input.listings == null) {
    return { status: 'checking', activeListingsCount };
  }

  const listedCount = input.listings?.length ?? 0;
  const hasIncompleteListingData =
    typeof input.totalListings === 'number' && input.totalListings !== listedCount;

  if (input.hasListingsError || hasIncompleteListingData) {
    return {
      status: 'error',
      activeListingsCount,
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
    };
  }

  if (isCreateListingLimitReached(input)) {
    return {
      status: 'blocked',
      activeListingsCount,
      message: FREE_TIER_LISTING_LIMIT_MESSAGE,
    };
  }

  return { status: 'ready', activeListingsCount };
}

export function getCreateListingErrorMessage(error: unknown): string {
  const message = error instanceof Error ? error.message : '';
  const normalizedMessage = message.trim().toLowerCase();

  if (
    normalizedMessage.includes('free tier allows at most 3 listing') ||
    normalizedMessage.includes('free tier allows at most 3 active listing')
  ) {
    return FREE_TIER_LISTING_LIMIT_MESSAGE;
  }

  return message || 'Terjadi kesalahan saat memasang iklan. Silakan coba lagi.';
}
