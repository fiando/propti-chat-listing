import test from 'node:test';
import assert from 'node:assert/strict';

import * as backendAuth from '../lib/backend-auth.js';

const {
  buildBackendAuthPayload,
  exchangeGoogleIdTokenForBackendSession,
  getBackendAuthHeader,
  getBackendProfilePath,
} = backendAuth;

function createUnsignedJwt(payload) {
  return [
    Buffer.from(JSON.stringify({ alg: 'none', typ: 'JWT' })).toString('base64url'),
    Buffer.from(JSON.stringify(payload)).toString('base64url'),
    '',
  ].join('.');
}

test('buildBackendAuthPayload uses Google id token for backend login exchange', () => {
  const payload = buildBackendAuthPayload({ idToken: 'google-id-token' });

  assert.deepEqual(payload, { idToken: 'google-id-token' });
});

test('exchangeGoogleIdTokenForBackendSession posts Google id token to backend auth endpoint', async () => {
  const calls = [];
  const session = await exchangeGoogleIdTokenForBackendSession({
    apiBaseUrl: 'https://api.propti.id',
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

  assert.equal(calls[0].url, 'https://api.propti.id/auth/google');
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

test('refreshBackendAuthTokenIfNeeded refreshes expired backend sessions with a Google refresh token', async () => {
  assert.equal(
    typeof backendAuth.refreshBackendAuthTokenIfNeeded,
    'function',
    'Expected backend auth helper to expose token refresh logic for expired backend sessions'
  );

  const now = 1_700_000_000_000;
  const refreshedToken = createUnsignedJwt({ exp: Math.floor(now / 1000) + 24 * 60 * 60 });

  const result = await backendAuth.refreshBackendAuthTokenIfNeeded({
    token: {
      backendAccessToken: createUnsignedJwt({ exp: Math.floor(now / 1000) - 60 }),
      refreshToken: 'google-refresh-token',
    },
    now: () => now,
    refreshGoogleTokens: async ({ refreshToken }) => {
      assert.equal(refreshToken, 'google-refresh-token');
      return {
        accessToken: 'google-access-token',
        accessTokenExpiresAt: now + 60 * 60 * 1000,
        refreshToken: 'google-refresh-token-2',
        idToken: 'google-id-token-2',
      };
    },
    exchangeGoogleIdTokenForBackendSession: async ({ idToken }) => {
      assert.equal(idToken, 'google-id-token-2');
      return {
        backendAccessToken: refreshedToken,
        backendUser: { userId: 'user-123', email: 'user@example.com' },
      };
    },
  });

  assert.equal(result.backendAccessToken, refreshedToken);
  assert.equal(result.refreshToken, 'google-refresh-token-2');
  assert.equal(result.googleAccessToken, 'google-access-token');
  assert.equal(result.googleAccessTokenExpiresAt, now + 60 * 60 * 1000);
  assert.equal(result.backendAccessTokenExpiresAt, (Math.floor(now / 1000) + 24 * 60 * 60) * 1000);
  assert.deepEqual(result.backendUser, { userId: 'user-123', email: 'user@example.com' });
  assert.equal(result.error, undefined);
});
