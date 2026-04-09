export const FREE_TIER_LISTING_LIMIT = 3;
export const BASIC_TIER_LISTING_LIMIT = 8;
export const PREMIUM_TIER_LISTING_LIMIT = 20;
export const PRO_TIER_LISTING_LIMIT = 50;
export const LISTING_ACCESS_CHECK_PAGE_SIZE = 100;
export const FREE_TIER_LISTING_LIMIT_MESSAGE =
  `Paket gratis hanya bisa memiliki ${FREE_TIER_LISTING_LIMIT} listing aktif. Upgrade paket untuk memasang iklan baru.`;
export const BASIC_TIER_LISTING_LIMIT_MESSAGE =
  `Paket Basic maksimal ${BASIC_TIER_LISTING_LIMIT} listing aktif. Arsipkan salah satu listing untuk memasang iklan baru.`;
export const PREMIUM_TIER_LISTING_LIMIT_MESSAGE =
  `Paket Premium maksimal ${PREMIUM_TIER_LISTING_LIMIT} listing aktif. Arsipkan salah satu listing untuk memasang iklan baru.`;
export const PRO_TIER_LISTING_LIMIT_MESSAGE =
  `Paket Pro maksimal ${PRO_TIER_LISTING_LIMIT} listing aktif. Arsipkan salah satu listing untuk memasang iklan baru.`;
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
  tier?: 'free' | 'basic' | 'premium' | 'pro';
  activeListingsCount?: number;
  listings?: Array<ListingLimitEntry> | null;
  limit?: number;
}): boolean {
  const activeListingsCount = input.activeListingsCount ?? getActiveListingCount(input.listings);
  return activeListingsCount >= (input.limit ?? getListingLimit(input.tier));
}

export function getCreateListingAccessState(input: {
  isAuthenticated?: boolean;
  tier?: 'free' | 'basic' | 'premium' | 'pro';
  isAuthLoading?: boolean;
  isListingsLoading?: boolean;
  isListingsFetching?: boolean;
  hasListingsError?: boolean;
  hasFreshAccessResult?: boolean;
  activeListingsCount?: number;
  listings?: Array<ListingLimitEntry> | null;
  totalListings?: number;
  limit?: number;
}): CreateListingAccessState {
  const activeListingsCount = input.activeListingsCount ?? getActiveListingCount(input.listings);
  const hasResolvedActiveListingsCount = typeof input.activeListingsCount === 'number';
  const hasResolvedListings = Array.isArray(input.listings);

  if (!input.isAuthenticated) {
    return { status: 'ready', activeListingsCount };
  }

  if (input.isAuthLoading || input.isListingsLoading || input.isListingsFetching) {
    return { status: 'checking', activeListingsCount };
  }

  if (input.hasFreshAccessResult === false) {
    return { status: 'checking', activeListingsCount };
  }

  if (input.hasListingsError || (!hasResolvedActiveListingsCount && !hasResolvedListings)) {
    return {
      status: 'error',
      activeListingsCount,
      message: CREATE_LISTING_ACCESS_ERROR_MESSAGE,
    };
  }

  const listedCount = input.listings?.length ?? 0;
  const hasIncompleteListingData =
    hasResolvedListings && typeof input.totalListings === 'number' && input.totalListings !== listedCount;

  if (hasIncompleteListingData) {
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
      message: getListingLimitMessage(input.tier),
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

  if (
    normalizedMessage.includes('basic tier allows at most 8 listing') ||
    normalizedMessage.includes('basic tier allows at most 8 active listing')
  ) {
    return BASIC_TIER_LISTING_LIMIT_MESSAGE;
  }

  if (
    normalizedMessage.includes('premium tier allows at most 20 listing') ||
    normalizedMessage.includes('premium tier allows at most 20 active listing')
  ) {
    return PREMIUM_TIER_LISTING_LIMIT_MESSAGE;
  }

  if (
    normalizedMessage.includes('pro tier allows at most 50 listing') ||
    normalizedMessage.includes('pro tier allows at most 50 active listing')
  ) {
    return PRO_TIER_LISTING_LIMIT_MESSAGE;
  }

  return message || 'Terjadi kesalahan saat memasang iklan. Silakan coba lagi.';
}

function getListingLimit(tier: 'free' | 'basic' | 'premium' | 'pro' = 'free') {
  switch (tier) {
    case 'basic':
      return BASIC_TIER_LISTING_LIMIT;
    case 'premium':
      return PREMIUM_TIER_LISTING_LIMIT;
    case 'pro':
      return PRO_TIER_LISTING_LIMIT;
    default:
      return FREE_TIER_LISTING_LIMIT;
  }
}

function getListingLimitMessage(tier: 'free' | 'basic' | 'premium' | 'pro' = 'free') {
  switch (tier) {
    case 'basic':
      return BASIC_TIER_LISTING_LIMIT_MESSAGE;
    case 'premium':
      return PREMIUM_TIER_LISTING_LIMIT_MESSAGE;
    case 'pro':
      return PRO_TIER_LISTING_LIMIT_MESSAGE;
    default:
      return FREE_TIER_LISTING_LIMIT_MESSAGE;
  }
}
