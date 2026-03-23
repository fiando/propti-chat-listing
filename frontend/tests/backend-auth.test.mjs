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

test('exchangeGoogleAccessTokenForBackendSession posts Google access token to backend auth endpoint', async () => {
  assert.equal(
    typeof backendAuth.exchangeGoogleAccessTokenForBackendSession,
    'function',
    'Expected backend auth helper to expose access-token exchange for refreshed Google sessions'
  );

  const calls = [];
  const session = await backendAuth.exchangeGoogleAccessTokenForBackendSession({
    apiBaseUrl: 'https://api.propti.id',
    accessToken: 'google-access-token',
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
  assert.equal(calls[0].init.body, JSON.stringify({ accessToken: 'google-access-token' }));
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

test('refreshBackendAuthTokenIfNeeded falls back to Google access token exchange when refresh returns no id token', async () => {
  const now = 1_700_000_000_000;
  const refreshedToken = createUnsignedJwt({ exp: Math.floor(now / 1000) + 24 * 60 * 60 });

  const result = await backendAuth.refreshBackendAuthTokenIfNeeded({
    token: {
      backendAccessToken: createUnsignedJwt({ exp: Math.floor(now / 1000) - 60 }),
      refreshToken: 'google-refresh-token',
      googleId: 'google-user-123',
    },
    now: () => now,
    refreshGoogleTokens: async () => ({
      accessToken: 'google-access-token',
      accessTokenExpiresAt: now + 60 * 60 * 1000,
      refreshToken: 'google-refresh-token-2',
    }),
    exchangeGoogleIdTokenForBackendSession: async () => {
      assert.fail('Expected access-token exchange fallback when Google refresh does not return an id_token');
    },
    exchangeGoogleAccessTokenForBackendSession: async ({ accessToken, googleId }) => {
      assert.equal(accessToken, 'google-access-token');
      assert.equal(googleId, 'google-user-123');
      return {
        backendAccessToken: refreshedToken,
        backendUser: { userId: 'user-123', email: 'user@example.com' },
      };
    },
  });

  assert.equal(result.backendAccessToken, refreshedToken);
  assert.equal(result.refreshToken, 'google-refresh-token-2');
  assert.equal(result.googleAccessToken, 'google-access-token');
  assert.equal(result.backendAccessTokenExpiresAt, (Math.floor(now / 1000) + 24 * 60 * 60) * 1000);
  assert.deepEqual(result.backendUser, { userId: 'user-123', email: 'user@example.com' });
  assert.equal(result.error, undefined);
});

test('refreshBackendAuthTokenIfNeeded refreshes token that expires within 5 minutes', async () => {
  const now = 1_700_000_000_000;
  // Token expires in 4 minutes — within the 5-minute refresh skew window
  const almostExpiredToken = createUnsignedJwt({ exp: Math.floor(now / 1000) + 4 * 60 });
  const refreshedToken = createUnsignedJwt({ exp: Math.floor(now / 1000) + 7 * 24 * 60 * 60 });

  const result = await backendAuth.refreshBackendAuthTokenIfNeeded({
    token: {
      backendAccessToken: almostExpiredToken,
      refreshToken: 'google-refresh-token',
    },
    now: () => now,
    refreshGoogleTokens: async () => ({
      accessToken: 'google-access-token',
      accessTokenExpiresAt: now + 60 * 60 * 1000,
      refreshToken: 'google-refresh-token-2',
      idToken: 'google-id-token-2',
    }),
    exchangeGoogleIdTokenForBackendSession: async () => ({
      backendAccessToken: refreshedToken,
      backendUser: { userId: 'user-123', email: 'user@example.com' },
    }),
  });

  assert.equal(result.backendAccessToken, refreshedToken, 'Token expiring within 5 minutes should be refreshed');
  assert.equal(result.error, undefined);
});

test('refreshBackendAuthTokenIfNeeded skips refresh when token has more than 5 minutes remaining', async () => {
  const now = 1_700_000_000_000;
  // Token expires in 10 minutes — outside the 5-minute refresh skew window
  const validToken = createUnsignedJwt({ exp: Math.floor(now / 1000) + 10 * 60 });

  const result = await backendAuth.refreshBackendAuthTokenIfNeeded({
    token: {
      backendAccessToken: validToken,
      refreshToken: 'google-refresh-token',
    },
    now: () => now,
    refreshGoogleTokens: async () => {
      assert.fail('Should not refresh when token has more than 5 minutes remaining');
    },
  });

  assert.equal(result.backendAccessToken, validToken, 'Token with >5 minutes remaining should not be refreshed');
  assert.equal(result.error, undefined);
});
