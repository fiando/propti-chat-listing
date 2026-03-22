import test from 'node:test';
import assert from 'node:assert/strict';

import { shouldRedirectAuthenticatedLoginVisitor } from './login-redirect.ts';

test('redirects login visitors only when the account profile is available', () => {
  assert.equal(
    shouldRedirectAuthenticatedLoginVisitor({ isAuthenticated: true, hasProfile: true }),
    true
  );
});

test('keeps the login page accessible when session exists but profile could not be loaded', () => {
  assert.equal(
    shouldRedirectAuthenticatedLoginVisitor({ isAuthenticated: true, hasProfile: false }),
    false
  );
});

test('does not redirect unauthenticated visitors away from login', () => {
  assert.equal(
    shouldRedirectAuthenticatedLoginVisitor({ isAuthenticated: false, hasProfile: false }),
    false
  );
});
