import test from 'node:test';
import assert from 'node:assert/strict';

import { getAuthenticatedHomeRedirectPath } from './home-redirect.ts';

test('redirects authenticated visitors with profile to my listings dashboard', () => {
  assert.equal(getAuthenticatedHomeRedirectPath({ isAuthenticated: true, hasProfile: true }), '/listings');
});

test('keeps landing page for logged-out visitors', () => {
  assert.equal(getAuthenticatedHomeRedirectPath({ isAuthenticated: false, hasProfile: false }), null);
});

test('keeps landing page when auth exists but profile is unavailable', () => {
  assert.equal(getAuthenticatedHomeRedirectPath({ isAuthenticated: true, hasProfile: false }), null);
});
