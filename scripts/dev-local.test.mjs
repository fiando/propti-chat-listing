import test from 'node:test';
import assert from 'node:assert/strict';
import path from 'node:path';

import { buildDevLocalPlan, buildSamEnvOverrides, findMissingEnvFiles, parseCliArgs } from './dev-local.mjs';

test('buildDevLocalPlan uses app-level .env.local files and localhost ports', () => {
  const rootDir = '/workspace/propti';
  const plan = buildDevLocalPlan(rootDir);

  assert.deepEqual(plan.envFiles, [
    path.join(rootDir, 'frontend/.env.local'),
    path.join(rootDir, 'backend/.env.local'),
  ]);

  assert.equal(plan.backend.cwd, path.join(rootDir, 'backend'));
  assert.deepEqual(plan.backend.buildCommand, {
    command: 'make',
    args: ['build'],
  });
  assert.deepEqual(plan.backend.startCommand, {
    command: 'sam',
    args: ['local', 'start-api', '--host', '127.0.0.1', '--port', '3001', '--env-vars', path.join(rootDir, 'backend/.env.local')],
  });

  assert.equal(plan.frontend.cwd, path.join(rootDir, 'frontend'));
  assert.deepEqual(plan.frontend.startCommand, {
    command: 'npm',
    args: ['run', 'dev', '--', '--hostname', '127.0.0.1', '--port', '3000'],
  });
});

test('buildDevLocalPlan can use a custom backend env file', () => {
  const rootDir = '/workspace/propti';
  const plan = buildDevLocalPlan(rootDir, {
    backendEnvFile: 'backend/.env.development',
  });

  assert.deepEqual(plan.envFiles, [
    path.join(rootDir, 'frontend/.env.local'),
    path.join(rootDir, 'backend/.env.development'),
  ]);
  assert.equal(plan.backend.envFile, path.join(rootDir, 'backend/.env.development'));
  assert.deepEqual(plan.backend.startCommand.args, [
    'local',
    'start-api',
    '--host',
    '127.0.0.1',
    '--port',
    '3001',
    '--env-vars',
    path.join(rootDir, 'backend/.env.development'),
  ]);
});

test('findMissingEnvFiles reports only missing local env files', () => {
  const missing = findMissingEnvFiles(['/tmp/frontend/.env.local', '/tmp/backend/.env.local'], new Set(['/tmp/backend/.env.local']));

  assert.deepEqual(missing, ['/tmp/frontend/.env.local']);
});

test('buildSamEnvOverrides converts dotenv content into SAM Parameters JSON', () => {
  const overrides = buildSamEnvOverrides(`
# Comment line
JWT_SECRET=replace-with-jwt-secret
PUBLIC_API_BASE_URL="http://localhost:3001"
EMPTY_VALUE=
`);

  assert.deepEqual(overrides, {
    Parameters: {
      JWT_SECRET: 'replace-with-jwt-secret',
      PUBLIC_API_BASE_URL: 'http://localhost:3001',
      EMPTY_VALUE: '',
    },
  });
});

test('parseCliArgs accepts backend env file override', () => {
  assert.deepEqual(
    parseCliArgs(['--backend-env-file', 'backend/.env.development']),
    { backendEnvFile: 'backend/.env.development' },
  );
});
