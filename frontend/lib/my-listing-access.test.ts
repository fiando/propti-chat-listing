import test from 'node:test';
import assert from 'node:assert/strict';

const accessModule = await import('./my-listing-access.ts').catch(() => ({} as Record<string, unknown>));

const collectActiveListingCount =
  (accessModule.collectActiveListingCount as
    | ((input: {
        limit: number;
        pageSize?: number;
        fetchPage: (params: { page: number; pageSize: number }) => Promise<{
          items: Array<{ status?: string | null }>;
          total: number;
          page: number;
        }>;
      }) => Promise<{ activeListingsCount: number; totalListings: number; fetchedPages: number }>)
    | undefined) ??
  (async (input: {
    limit: number;
    pageSize?: number;
    fetchPage: (params: { page: number; pageSize: number }) => Promise<{
      items: Array<{ status?: string | null }>;
      total: number;
      page: number;
    }>;
  }) => {
    const response = await input.fetchPage({ page: 1, pageSize: input.pageSize ?? 100 });
    return {
      activeListingsCount: response.items.filter((listing) => listing.status === 'active').length,
      totalListings: response.total,
      fetchedPages: 1,
    };
  });

test('walks additional pages until the active listing limit is reached', async () => {
  const calls: Array<{ page: number; pageSize: number }> = [];

  const result = await collectActiveListingCount({
    limit: 1,
    pageSize: 100,
    fetchPage: async ({ page, pageSize }) => {
      calls.push({ page, pageSize });

      if (page === 1) {
        return {
          items: Array.from({ length: 100 }, () => ({ status: 'archived' })),
          total: 101,
          page,
        };
      }

      return {
        items: [{ status: 'active' }],
        total: 101,
        page,
      };
    },
  });

  assert.deepEqual(calls, [
    { page: 1, pageSize: 100 },
    { page: 2, pageSize: 100 },
  ]);
  assert.deepEqual(result, {
    activeListingsCount: 1,
    totalListings: 101,
    fetchedPages: 2,
  });
});

test('stops fetching once the limit is already reached', async () => {
  const calls: number[] = [];

  const result = await collectActiveListingCount({
    limit: 3,
    pageSize: 100,
    fetchPage: async ({ page }) => {
      calls.push(page);

      return {
        items: [{ status: 'active' }, { status: 'active' }, { status: 'active' }],
        total: 300,
        page,
      };
    },
  });

  assert.deepEqual(calls, [1]);
  assert.deepEqual(result, {
    activeListingsCount: 3,
    totalListings: 300,
    fetchedPages: 1,
  });
});
