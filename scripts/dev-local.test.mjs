import test from 'node:test';
import assert from 'node:assert/strict';
import path from 'node:path';

import {
  buildFrontendEnvOverrides,
  buildDevLocalPlan,
  buildLocalDynamoTableDefinitions,
  buildSamEnvOverrides,
  canonicalizeLocalLoopbackUrl,
  findMissingEnvFiles,
  parseListeningPidsFromSsOutput,
  parseCliArgs,
  parseDotenv,
  resolveContainerRuntimeOptions,
  shouldStopPortOwner,
  shouldBootstrapLocalDynamoDB,
} from './dev-local.mjs';

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
    args: ['local', 'start-api', '--host', 'localhost', '--port', '3001', '--env-vars', path.join(rootDir, 'backend/.env.local')],
  });

  assert.equal(plan.frontend.cwd, path.join(rootDir, 'frontend'));
  assert.deepEqual(plan.frontend.startCommand, {
    command: 'npm',
    args: ['run', 'dev', '--', '--hostname', '0.0.0.0', '--port', '3000'],
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
    'localhost',
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

test('parseDotenv reads comments, quoted values, and blank values', () => {
  assert.deepEqual(
    parseDotenv(`
# comment
AWS_REGION=ap-southeast-1
DYNAMODB_ENDPOINT_URL="http://localhost:8000"
EMPTY_VALUE=
`),
    {
      AWS_REGION: 'ap-southeast-1',
      DYNAMODB_ENDPOINT_URL: 'http://localhost:8000',
      EMPTY_VALUE: '',
    },
  );
});

test('shouldBootstrapLocalDynamoDB only enables bootstrap for localhost endpoints', () => {
  assert.equal(shouldBootstrapLocalDynamoDB('http://localhost:8000'), true);
  assert.equal(shouldBootstrapLocalDynamoDB('http://127.0.0.1:8000'), true);
  assert.equal(shouldBootstrapLocalDynamoDB('https://dynamodb.ap-southeast-1.amazonaws.com'), false);
  assert.equal(shouldBootstrapLocalDynamoDB(undefined), false);
});

test('buildLocalDynamoTableDefinitions uses local dev defaults and GSIs', () => {
  const definitions = buildLocalDynamoTableDefinitions({});
  const usersTable = definitions.find((definition) => definition.tableName === 'propti-users-dev');

  assert.ok(usersTable);
  assert.deepEqual(
    usersTable.globalSecondaryIndexes.map((index) => index.IndexName),
    ['googleId-index', 'whatsAppLinkedPhone-index'],
  );

  const listingsTable = definitions.find((definition) => definition.tableName === 'propti-listings-dev');
  assert.ok(listingsTable);
  assert.deepEqual(
    listingsTable.globalSecondaryIndexes.map((index) => index.IndexName),
    ['listingId-index', 'userId-createdAt-index'],
  );
});

test('resolveContainerRuntimeOptions keeps defaults when docker is available', () => {
  const options = resolveContainerRuntimeOptions({
    dockerAvailable: true,
    podmanAvailable: true,
    runtimeDir: '/run/user/1000',
    uid: 1000,
  });

  assert.deepEqual(options, {
    envOverrides: {},
    podmanSocketPath: null,
    shouldUsePodmanSocket: false,
  });
});

test('resolveContainerRuntimeOptions uses podman socket when docker is unavailable', () => {
  const options = resolveContainerRuntimeOptions({
    dockerAvailable: false,
    podmanAvailable: true,
    runtimeDir: '/run/user/1000',
    uid: 1000,
  });

  assert.deepEqual(options, {
    envOverrides: {
      DOCKER_HOST: 'unix:///run/user/1000/podman/podman.sock',
    },
    podmanSocketPath: '/run/user/1000/podman/podman.sock',
    shouldUsePodmanSocket: true,
  });
});

test('resolveContainerRuntimeOptions falls back to uid runtime dir when XDG_RUNTIME_DIR is absent', () => {
  const options = resolveContainerRuntimeOptions({
    dockerAvailable: false,
    podmanAvailable: true,
    runtimeDir: '',
    uid: 1234,
  });

  assert.deepEqual(options, {
    envOverrides: {
      DOCKER_HOST: 'unix:///run/user/1234/podman/podman.sock',
    },
    podmanSocketPath: '/run/user/1234/podman/podman.sock',
    shouldUsePodmanSocket: true,
  });
});

test('canonicalizeLocalLoopbackUrl rewrites 127.0.0.1 hostname to localhost', () => {
  assert.equal(
    canonicalizeLocalLoopbackUrl('http://127.0.0.1:3000/callback?callbackUrl=%2F'),
    'http://localhost:3000/callback?callbackUrl=%2F',
  );
});

test('canonicalizeLocalLoopbackUrl keeps non-loopback or invalid urls unchanged', () => {
  assert.equal(canonicalizeLocalLoopbackUrl('http://localhost:3000'), 'http://localhost:3000');
  assert.equal(canonicalizeLocalLoopbackUrl('not-a-url'), 'not-a-url');
});

test('buildFrontendEnvOverrides canonicalizes local auth and api urls', () => {
  const overrides = buildFrontendEnvOverrides(`
NEXTAUTH_URL=http://127.0.0.1:3000
NEXT_PUBLIC_API_URL=http://127.0.0.1:3001
NEXTAUTH_SECRET=secret
`);

  assert.deepEqual(overrides, {
    NEXTAUTH_URL: 'http://localhost:3000/',
    NEXTAUTH_URL_INTERNAL: 'http://localhost:3000/',
    NEXT_PUBLIC_API_URL: 'http://localhost:3001/',
  });
});

test('buildFrontendEnvOverrides preserves localhost values to override inherited env', () => {
  const overrides = buildFrontendEnvOverrides(`
NEXTAUTH_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3001
`);

  assert.deepEqual(overrides, {
    NEXTAUTH_URL: 'http://localhost:3000',
    NEXTAUTH_URL_INTERNAL: 'http://localhost:3000',
    NEXT_PUBLIC_API_URL: 'http://localhost:3001',
  });
});

test('buildFrontendEnvOverrides falls back to local defaults when vars are missing', () => {
  const overrides = buildFrontendEnvOverrides('NEXTAUTH_SECRET=secret');

  assert.deepEqual(overrides, {
    NEXTAUTH_URL: 'http://localhost:3000',
    NEXTAUTH_URL_INTERNAL: 'http://localhost:3000',
    NEXT_PUBLIC_API_URL: 'http://localhost:3001',
  });
});

test('parseListeningPidsFromSsOutput extracts pids for target ports', () => {
  const ssOutput = `State  Recv-Q Send-Q      Local Address:Port  Peer Address:PortProcess
LISTEN 0      511               0.0.0.0:3000       0.0.0.0:*    users:(("next-server",pid=111,fd=24))
LISTEN 0      128             127.0.0.1:3001       0.0.0.0:*    users:(("sam",pid=222,fd=7))
LISTEN 0      128             127.0.0.1:5432       0.0.0.0:*    users:(("postgres",pid=333,fd=7))
`;

  const owners = parseListeningPidsFromSsOutput(ssOutput, [3000, 3001]);
  assert.deepEqual(owners, [
    { port: 3000, pid: 111 },
    { port: 3001, pid: 222 },
  ]);
});

test('shouldStopPortOwner only stops processes under current project root', () => {
  assert.equal(
    shouldStopPortOwner('/home/bobby/Development/IdeaProjects/saas/propti/frontend', '/home/bobby/Development/IdeaProjects/saas/propti'),
    true,
  );
  assert.equal(
    shouldStopPortOwner('/home/bobby/Development/IdeaProjects/another-app', '/home/bobby/Development/IdeaProjects/saas/propti'),
    false,
  );
});
