import test from 'node:test';
import assert from 'node:assert/strict';
import { collectActiveListingCount } from './my-listing-access.ts';

test('walks additional pages until the create-limit decision can be made', async () => {
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
