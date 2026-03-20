import test from 'node:test';
import assert from 'node:assert/strict';
import type { User } from '@/types';
import { getSubscriptionStatus, getDaysUntilExpiry, getExpiryMessage } from './subscription-status.ts';

function createUser(overrides: Partial<User['subscription']> = {}, rootOverrides: Partial<Pick<User, 'subscriptionStatus'>> = {}): User {
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
    ...rootOverrides,
  };
}

test('returns loading while session is still loading', () => {
  assert.equal(getSubscriptionStatus({ authStatus: 'loading', profile: undefined, now: 0 }), 'loading');
});

test('returns loading for authenticated users before profile loads', () => {
  assert.equal(getSubscriptionStatus({ authStatus: 'authenticated', profile: undefined, now: 0 }), 'loading');
});

test('returns active when premium membership has 8+ days remaining', () => {
  const profile = createUser({ tier: 'premium', renewDate: '2026-03-20T00:00:00Z' });
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-12T00:00:00Z') }),
    'active'
  );
});

test('returns expiring_soon when 0-7 days remaining', () => {
  const profile = createUser({ tier: 'premium', renewDate: '2026-03-20T00:00:00Z' });
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-16T00:00:00Z') }),
    'expiring_soon'
  );
});

test('returns expiring_soon exactly at 7-day boundary', () => {
  const profile = createUser({ tier: 'premium', renewDate: '2026-03-20T00:00:00Z' });
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-13T00:00:00Z') }),
    'expiring_soon'
  );
});

test('returns expired when renewDate is in the past', () => {
  const profile = createUser({ tier: 'premium', renewDate: '2026-03-10T00:00:00Z' });
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-16T00:00:00Z') }),
    'expired'
  );
});

test('returns expired when premium has no renewDate', () => {
  const profile = createUser({ tier: 'premium', renewDate: undefined });
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-16T00:00:00Z') }),
    'expired'
  );
});

test('returns free for free accounts', () => {
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile: createUser(), now: 0 }),
    'free'
  );
});

test('trusts API-provided subscriptionStatus if present', () => {
  const profile = createUser(
    { tier: 'premium', renewDate: '2026-03-20T00:00:00Z' },
    { subscriptionStatus: 'expiring_soon' }
  );
  assert.equal(
    getSubscriptionStatus({ authStatus: 'authenticated', profile, now: Date.parse('2026-03-12T00:00:00Z') }),
    'expiring_soon'
  );
});

// getDaysUntilExpiry
test('getDaysUntilExpiry returns null for missing renewDate', () => {
  assert.equal(getDaysUntilExpiry(undefined), null);
});

test('getDaysUntilExpiry returns positive days for future date', () => {
  const renewDate = new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString();
  const days = getDaysUntilExpiry(renewDate);
  assert.ok(days !== null && days >= 4 && days <= 6);
});

// getExpiryMessage
test('getExpiryMessage returns expired message for expired status', () => {
  assert.equal(getExpiryMessage('expired', undefined), 'Premium telah berakhir');
});

test('getExpiryMessage returns free message for free status', () => {
  assert.equal(getExpiryMessage('free', undefined), 'Paket gratis');
});
