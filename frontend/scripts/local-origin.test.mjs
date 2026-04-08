import test from 'node:test';
import assert from 'node:assert/strict';

import { getCanonicalLocalhostUrl, shouldCanonicalizeLocalhostUrl } from '../lib/local-origin.js';

test('getCanonicalLocalhostUrl rewrites non-auth local 127 host to localhost', () => {
  assert.equal(
    getCanonicalLocalhostUrl('http://127.0.0.1:3000/login'),
    'http://localhost:3000/login',
  );
});

test('getCanonicalLocalhostUrl keeps /api/auth paths unchanged to preserve OAuth cookies', () => {
  assert.equal(
    getCanonicalLocalhostUrl('http://127.0.0.1:3000/api/auth/error?error=OAuthCallback'),
    null,
  );
});

test('getCanonicalLocalhostUrl leaves localhost unchanged', () => {
  assert.equal(getCanonicalLocalhostUrl('http://localhost:3000/login'), null);
});

test('shouldCanonicalizeLocalhostUrl returns false for /api/auth routes', () => {
  assert.equal(shouldCanonicalizeLocalhostUrl('http://127.0.0.1:3000/api/auth/callback/google'), false);
});
