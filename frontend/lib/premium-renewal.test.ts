import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldShowRenewalCTA, getRenewalUXCopy } from './premium-renewal.ts';
import type { SubscriptionStatus } from '@/types';

// shouldShowRenewalCTA
test('shouldShowRenewalCTA returns true for expiring_soon', () => {
  assert.equal(shouldShowRenewalCTA('expiring_soon'), true);
});

test('shouldShowRenewalCTA returns true for expired', () => {
  assert.equal(shouldShowRenewalCTA('expired'), true);
});

test('shouldShowRenewalCTA returns false for active', () => {
  assert.equal(shouldShowRenewalCTA('active'), false);
});

test('shouldShowRenewalCTA returns false for free', () => {
  assert.equal(shouldShowRenewalCTA('free'), false);
});

test('shouldShowRenewalCTA returns false for loading', () => {
  assert.equal(shouldShowRenewalCTA('loading'), false);
});

// getRenewalUXCopy
test('getRenewalUXCopy for expired returns Perpanjang Premium CTA', () => {
  const copy = getRenewalUXCopy('expired', undefined);
  assert.equal(copy.ctaText, 'Perpanjang Premium');
  assert.ok(copy.heading.length > 0);
});

test('getRenewalUXCopy for expiring_soon includes expiry date in body', () => {
  const renewDate = '2026-04-10T00:00:00Z';
  const copy = getRenewalUXCopy('expiring_soon', renewDate);
  assert.equal(copy.ctaText, 'Perpanjang Premium');
  assert.ok(copy.body.includes('2026') || copy.body.includes('April'));
});

test('getRenewalUXCopy for active/free returns upgrade copy', () => {
  const statuses: SubscriptionStatus[] = ['active', 'free', 'loading'];
  for (const status of statuses) {
    const copy = getRenewalUXCopy(status, undefined);
    assert.equal(copy.ctaText, 'Upgrade ke Premium');
  }
});
