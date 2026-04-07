import test from 'node:test';
import assert from 'node:assert/strict';

import { getCanonicalLocalhostUrl } from '../lib/local-origin.js';

test('getCanonicalLocalhostUrl rewrites local 127 host to localhost', () => {
  assert.equal(
    getCanonicalLocalhostUrl('http://127.0.0.1:3000/api/auth/error?error=OAuthCallback'),
    'http://localhost:3000/api/auth/error?error=OAuthCallback',
  );
});

test('getCanonicalLocalhostUrl leaves localhost unchanged', () => {
  assert.equal(getCanonicalLocalhostUrl('http://localhost:3000/login'), null);
});
