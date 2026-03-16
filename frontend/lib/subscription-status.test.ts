import test from 'node:test';
import assert from 'node:assert/strict';
import type { User } from '@/types';
import { getSubscriptionStatus } from './subscription-status.ts';

function createUser(overrides: Partial<User['subscription']> = {}): User {
  return {
    userId: 'user-1',
    googleId: 'google-1',
    email: 'user@example.com',
    name: 'User Example',
    role: 'buyer',
    createdAt: '2026-03-16T00:00:00Z',
    lastLoginAt: '2026-03-16T00:00:00Z',
    subscription: {
      tier: 'free',
      monthlyListingsUsed: 0,
      ...overrides,
    },
  };
}

test('returns loading while session is still loading', () => {
  assert.equal(getSubscriptionStatus({ authStatus: 'loading', profile: undefined, now: 0 }), 'loading');
});

test('returns loading for authenticated users before profile loads', () => {
  assert.equal(getSubscriptionStatus({ authStatus: 'authenticated', profile: undefined, now: 0 }), 'loading');
});

test('returns premium when premium membership has a future renew date', () => {
  const profile = createUser({
    tier: 'premium',
    renewDate: '2026-03-20T00:00:00Z',
  });

  assert.equal(
    getSubscriptionStatus({
      authStatus: 'authenticated',
      profile,
      now: Date.parse('2026-03-16T00:00:00Z'),
    }),
    'premium'
  );
});

test('returns free when premium membership is expired', () => {
  const profile = createUser({
    tier: 'premium',
    renewDate: '2026-03-10T00:00:00Z',
  });

  assert.equal(
    getSubscriptionStatus({
      authStatus: 'authenticated',
      profile,
      now: Date.parse('2026-03-16T00:00:00Z'),
    }),
    'free'
  );
});

test('returns free for free accounts after profile loads', () => {
  assert.equal(
    getSubscriptionStatus({
      authStatus: 'authenticated',
      profile: createUser(),
      now: 0,
    }),
    'free'
  );
});
