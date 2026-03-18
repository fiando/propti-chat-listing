import test from 'node:test';
import assert from 'node:assert/strict';

import {
  clearRevealedContactCache,
  getCachedRevealedContact,
  getOrLoadRevealedContact,
} from './revealed-contact-cache.ts';

test('getOrLoadRevealedContact caches the resolved contact for later clicks', async () => {
  clearRevealedContactCache();

  let calls = 0;
  const first = await getOrLoadRevealedContact('listing-1', async () => {
    calls += 1;
    return {
      sellerName: 'Budi',
      sellerPhone: '08123456789',
    };
  });

  const second = await getOrLoadRevealedContact('listing-1', async () => {
    calls += 1;
    return {
      sellerName: 'Siti',
      sellerPhone: '08999999999',
    };
  });

  assert.equal(calls, 1);
  assert.deepEqual(first, second);
  assert.deepEqual(getCachedRevealedContact('listing-1'), first);
});

test('getOrLoadRevealedContact reuses the in-flight request for repeated clicks', async () => {
  clearRevealedContactCache();

  let calls = 0;
  const loader = async () => {
    calls += 1;
    await new Promise((resolve) => setTimeout(resolve, 5));
    return {
      sellerName: 'Budi',
      sellerPhone: '08123456789',
    };
  };

  const [first, second] = await Promise.all([
    getOrLoadRevealedContact('listing-2', loader),
    getOrLoadRevealedContact('listing-2', loader),
  ]);

  assert.equal(calls, 1);
  assert.deepEqual(first, second);
});
