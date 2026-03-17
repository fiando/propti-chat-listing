const LISTING_ACCESS_CHECK_PAGE_SIZE = 100;

function countActiveListings(listings: Array<{ status?: string | null }>): number {
  return listings.filter((listing) => listing.status === 'active').length;
}

type ListingPage = {
  items: Array<{ status?: string | null }>;
  total: number;
  page: number;
};

export async function collectActiveListingCount(input: {
  limit: number;
  pageSize?: number;
  fetchPage: (params: { page: number; pageSize: number }) => Promise<ListingPage>;
}): Promise<{ activeListingsCount: number; totalListings: number; fetchedPages: number }> {
  let page = 1;
  let fetchedListings = 0;
  let activeListingsCount = 0;
  let totalListings = 0;

  while (true) {
    const response = await input.fetchPage({
      page,
      pageSize: input.pageSize ?? LISTING_ACCESS_CHECK_PAGE_SIZE,
    });
    const items = response.items ?? [];

    totalListings = response.total ?? items.length;
    fetchedListings += items.length;
    activeListingsCount += countActiveListings(items);

    if (activeListingsCount >= input.limit || items.length === 0 || fetchedListings >= totalListings) {
      return {
        activeListingsCount,
        totalListings,
        fetchedPages: page,
      };
    }

    page += 1;
  }
}
