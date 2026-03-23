import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildPostLoginCallbackUrl,
  getPostLoginDecision,
  getSafeAuthCallbackUrl,
} from './auth-callback.ts';

test('wraps safe callback urls through the callback sync page', () => {
  assert.equal(
    buildPostLoginCallbackUrl('/saved?tab=recent'),
    '/callback?callbackUrl=%2Fsaved%3Ftab%3Drecent'
  );
});

test('falls back to the home page for auth callback urls that point back into auth routes', () => {
  assert.equal(buildPostLoginCallbackUrl('/login?callbackUrl=%2Fprofile'), '/callback?callbackUrl=%2F');
  assert.equal(getSafeAuthCallbackUrl('/api/auth/signin/google'), '/');
});

test('keeps waiting while an unauthenticated callback is still syncing the session', () => {
  assert.equal(
    getPostLoginDecision({
      status: 'unauthenticated',
      syncAttempt: 1,
      maxSyncAttempts: 3,
    }),
    'wait'
  );
});

test('forces a target redirect once the session is authenticated', () => {
  assert.equal(
    getPostLoginDecision({
      status: 'authenticated',
      syncAttempt: 0,
      maxSyncAttempts: 3,
    }),
    'redirect-target'
  );
});

test('returns users to login after the callback exhausts its sync retries', () => {
  assert.equal(
    getPostLoginDecision({
      status: 'unauthenticated',
      syncAttempt: 3,
      maxSyncAttempts: 3,
    }),
    'redirect-login'
  );
});
