import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildBackendAuthPayload,
  exchangeGoogleIdTokenForBackendSession,
  getBackendAuthHeader,
  getBackendProfilePath,
} from '../lib/backend-auth.js';

test('buildBackendAuthPayload uses Google id token for backend login exchange', () => {
  const payload = buildBackendAuthPayload({ idToken: 'google-id-token' });

  assert.deepEqual(payload, { idToken: 'google-id-token' });
});

test('exchangeGoogleIdTokenForBackendSession posts Google id token to backend auth endpoint', async () => {
  const calls = [];
  const session = await exchangeGoogleIdTokenForBackendSession({
    apiBaseUrl: 'https://api.propti.id/v1',
    idToken: 'google-id-token',
    fetchImpl: async (url, init) => {
      calls.push({ url, init });
      return {
        ok: true,
        json: async () => ({
          accessToken: 'backend-jwt',
          user: { userId: 'user-123', email: 'user@example.com' },
        }),
      };
    },
  });

  assert.equal(calls[0].url, 'https://api.propti.id/v1/auth/google');
  assert.equal(calls[0].init.method, 'POST');
  assert.equal(calls[0].init.headers['Content-Type'], 'application/json');
  assert.equal(calls[0].init.body, JSON.stringify({ idToken: 'google-id-token' }));
  assert.deepEqual(session, {
    backendAccessToken: 'backend-jwt',
    backendUser: { userId: 'user-123', email: 'user@example.com' },
  });
});

test('getBackendAuthHeader uses backend-issued token instead of raw Google token', () => {
  const header = getBackendAuthHeader({ backendAccessToken: 'backend-jwt', googleAccessToken: 'google-access-token' });

  assert.equal(header, 'Bearer backend-jwt');
});

test('getBackendProfilePath targets backend auth user endpoint', () => {
  assert.equal(getBackendProfilePath(), '/auth/user');
});
