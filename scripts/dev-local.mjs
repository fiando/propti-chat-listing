#!/usr/bin/env node

import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const DEFAULT_STAGE = 'dev';
const DEFAULT_LOCAL_DYNAMODB_ENDPOINT = 'http://localhost:8000';
const DEFAULT_LOCAL_AWS_REGION = 'ap-southeast-1';
const DEFAULT_LOCAL_AWS_ACCESS_KEY_ID = 'local';
const DEFAULT_LOCAL_AWS_SECRET_ACCESS_KEY = 'local';
const LOCAL_DEV_PORTS = [3000, 3001];

export function parseCliArgs(argv = process.argv.slice(2)) {
  const options = {};

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];

    if (arg === '--backend-env-file') {
      options.backendEnvFile = argv[index + 1];
      index += 1;
    }
  }

  return options;
}

export function buildDevLocalPlan(rootDir = process.cwd(), options = {}) {
  const projectRoot = path.resolve(rootDir);
  const frontendEnvFile = path.join(projectRoot, 'frontend/.env.local');
  const backendEnvFile = path.join(projectRoot, options.backendEnvFile ?? 'backend/.env.local');

  return {
    projectRoot,
    envFiles: [frontendEnvFile, backendEnvFile],
    backend: {
      cwd: path.join(projectRoot, 'backend'),
      envFile: backendEnvFile,
      buildCommand: null, // go run compiles on the fly
      startCommand: {
        command: 'go',
        args: ['run', './cmd/localserver'],
      },
    },
    frontend: {
      cwd: path.join(projectRoot, 'frontend'),
      buildCommand: {
        command: 'npm',
        args: ['run', 'build'],
      },
      startCommand: {
        command: 'npm',
        args: ['run', 'start', '--', '--hostname', '0.0.0.0', '--port', '3000'],
      },
    },
  };
}

export function findMissingEnvFiles(envFiles, existingEnvFiles = null) {
  const existing = existingEnvFiles ?? new Set(envFiles.filter((envFile) => fs.existsSync(envFile)));
  return envFiles.filter((envFile) => !existing.has(envFile));
}

export function parseDotenv(envContent) {
  const parameters = {};

  for (const rawLine of envContent.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) {
      continue;
    }

    const separatorIndex = line.indexOf('=');
    if (separatorIndex === -1) {
      continue;
    }

    const key = line.slice(0, separatorIndex).trim();
    let value = line.slice(separatorIndex + 1).trim();

    if (!key) {
      continue;
    }

    const isWrappedInMatchingQuotes =
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"));

    if (isWrappedInMatchingQuotes) {
      value = value.slice(1, -1);
    }

    parameters[key] = value;
  }

  return parameters;
}

export function canonicalizeLocalLoopbackUrl(rawUrl) {
  if (!rawUrl) {
    return rawUrl;
  }

  try {
    const parsed = new URL(rawUrl);
    if (parsed.hostname !== '127.0.0.1') {
      return rawUrl;
    }
    parsed.hostname = 'localhost';
    return parsed.toString();
  } catch {
    return rawUrl;
  }
}

export function buildFrontendEnvOverrides(frontendEnvContent) {
  const envVars = parseDotenv(frontendEnvContent);
  const nextAuthUrl = canonicalizeLocalLoopbackUrl(envVars.NEXTAUTH_URL || 'http://localhost:3000');
  const nextPublicApiUrl = canonicalizeLocalLoopbackUrl(envVars.NEXT_PUBLIC_API_URL || 'http://localhost:3001');

  return {
    NEXTAUTH_URL: nextAuthUrl,
    NEXTAUTH_URL_INTERNAL: nextAuthUrl,
    NEXT_PUBLIC_API_URL: nextPublicApiUrl,
  };
}

export function parseListeningPidsFromSsOutput(output, targetPorts = []) {
  const ports = new Set(targetPorts.map((port) => Number(port)));
  const seen = new Set();
  const owners = [];

  for (const line of output.split(/\r?\n/)) {
    const portMatch = line.match(/:(\d+)\s+/);
    if (!portMatch) {
      continue;
    }
    const port = Number(portMatch[1]);
    if (!ports.has(port)) {
      continue;
    }

    for (const pidMatch of line.matchAll(/pid=(\d+)/g)) {
      const pid = Number(pidMatch[1]);
      const key = `${port}:${pid}`;
      if (seen.has(key)) {
        continue;
      }
      seen.add(key);
      owners.push({ port, pid });
    }
  }

  return owners;
}

export function shouldStopPortOwner(processCwd, projectRoot) {
  if (!processCwd || !projectRoot) {
    return false;
  }

  const normalizedProcessCwd = path.resolve(processCwd);
  const normalizedProjectRoot = path.resolve(projectRoot);
  return (
    normalizedProcessCwd === normalizedProjectRoot ||
    normalizedProcessCwd.startsWith(`${normalizedProjectRoot}${path.sep}`)
  );
}

export function buildSamEnvOverrides(envContent) {
  const parameters = parseDotenv(envContent);

  return {
    Parameters: parameters,
  };
}

export function shouldBootstrapLocalDynamoDB(endpointUrl) {
  if (!endpointUrl) {
    return false;
  }

  try {
    const url = new URL(endpointUrl);
    return url.hostname === '127.0.0.1' || url.hostname === 'localhost';
  } catch {
    return false;
  }
}

export function resolveContainerRuntimeOptions({
  dockerAvailable,
  podmanAvailable,
  runtimeDir = process.env.XDG_RUNTIME_DIR,
  uid = typeof process.getuid === 'function' ? process.getuid() : null,
}) {
  if (dockerAvailable || !podmanAvailable) {
    return {
      envOverrides: {},
      podmanSocketPath: null,
      shouldUsePodmanSocket: false,
    };
  }

  const effectiveRuntimeDir = runtimeDir || (typeof uid === 'number' ? `/run/user/${uid}` : null);
  if (!effectiveRuntimeDir) {
    return {
      envOverrides: {},
      podmanSocketPath: null,
      shouldUsePodmanSocket: false,
    };
  }

  const podmanSocketPath = path.join(effectiveRuntimeDir, 'podman', 'podman.sock');
  return {
    envOverrides: {
      DOCKER_HOST: `unix://${podmanSocketPath}`,
    },
    podmanSocketPath,
    shouldUsePodmanSocket: true,
  };
}

function resolveStage(envVars) {
  return envVars.STAGE || DEFAULT_STAGE;
}

function resolveTableName(envVars, envKey, defaultBaseName) {
  const configured = envVars[envKey];
  if (configured) {
    return configured;
  }

  return `${defaultBaseName}-${resolveStage(envVars)}`;
}

export function buildLocalDynamoTableDefinitions(envVars) {
  return [
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_LISTINGS_TABLE', 'propti-listings'),
      attributeDefinitions: [
        ['PK', 'S'],
        ['SK', 'S'],
        ['listingId', 'S'],
        ['userId', 'S'],
        ['createdAt', 'S'],
      ],
      keySchema: [
        ['PK', 'HASH'],
        ['SK', 'RANGE'],
      ],
      globalSecondaryIndexes: [
        { IndexName: 'listingId-index', KeySchema: [['listingId', 'HASH']] },
        { IndexName: 'userId-createdAt-index', KeySchema: [['userId', 'HASH'], ['createdAt', 'RANGE']] },
      ],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_USERS_TABLE', 'propti-users'),
      attributeDefinitions: [
        ['PK', 'S'],
        ['SK', 'S'],
        ['googleId', 'S'],
        ['whatsAppLinkedPhone', 'S'],
      ],
      keySchema: [
        ['PK', 'HASH'],
        ['SK', 'RANGE'],
      ],
      globalSecondaryIndexes: [
        { IndexName: 'googleId-index', KeySchema: [['googleId', 'HASH']] },
        { IndexName: 'whatsAppLinkedPhone-index', KeySchema: [['whatsAppLinkedPhone', 'HASH']] },
      ],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_TRANSACTIONS_TABLE', 'propti-transactions'),
      attributeDefinitions: [
        ['PK', 'S'],
        ['SK', 'S'],
        ['userId', 'S'],
        ['orderId', 'S'],
        ['createdAt', 'S'],
      ],
      keySchema: [
        ['PK', 'HASH'],
        ['SK', 'RANGE'],
      ],
      globalSecondaryIndexes: [
        { IndexName: 'userId-createdAt-index', KeySchema: [['userId', 'HASH'], ['createdAt', 'RANGE']] },
        { IndexName: 'orderId-index', KeySchema: [['orderId', 'HASH']] },
      ],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_MODERATIONS_TABLE', 'propti-moderations'),
      attributeDefinitions: [
        ['PK', 'S'],
        ['SK', 'S'],
        ['listingId', 'S'],
      ],
      keySchema: [
        ['PK', 'HASH'],
        ['SK', 'RANGE'],
      ],
      globalSecondaryIndexes: [{ IndexName: 'listingId-index', KeySchema: [['listingId', 'HASH']] }],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_LEADS_TABLE', 'propti-leads'),
      attributeDefinitions: [
        ['PK', 'S'],
        ['SK', 'S'],
        ['ownerUserId', 'S'],
        ['createdAt', 'S'],
      ],
      keySchema: [
        ['PK', 'HASH'],
        ['SK', 'RANGE'],
      ],
      globalSecondaryIndexes: [
        { IndexName: 'ownerUserId-createdAt-index', KeySchema: [['ownerUserId', 'HASH'], ['createdAt', 'RANGE']] },
      ],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_UPLOAD_SESSIONS_TABLE', 'propti-upload-sessions'),
      attributeDefinitions: [['sessionId', 'S']],
      keySchema: [['sessionId', 'HASH']],
      globalSecondaryIndexes: [],
    },
    {
      tableName: resolveTableName(envVars, 'DYNAMODB_WHATSAPP_SESSIONS_TABLE', 'propti-whatsapp-sessions'),
      attributeDefinitions: [['sessionId', 'S']],
      keySchema: [['sessionId', 'HASH']],
      globalSecondaryIndexes: [],
    },
    {
      tableName:
        envVars.DYNAMODB_OTP_CHALLENGES_TABLE ||
        envVars.DYNAMODB_WHATSAPP_OTP_TABLE ||
        `propti-whatsapp-otp-${resolveStage(envVars)}`,
      attributeDefinitions: [['otpId', 'S']],
      keySchema: [['otpId', 'HASH']],
      globalSecondaryIndexes: [],
    },
  ];
}

function createSamEnvOverridesFile(envFile) {
  const overrides = buildSamEnvOverrides(fs.readFileSync(envFile, 'utf8'));
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'propti-sam-'));
  const tempFile = path.join(tempDir, 'env.json');
  fs.writeFileSync(tempFile, JSON.stringify(overrides));

  return {
    tempDir,
    tempFile,
  };
}

function runCommand(cwd, command, args, envOverrides = {}, options = {}) {
  const stdio = options.stdio ?? 'inherit';
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      env: {
        ...process.env,
        ...envOverrides,
      },
      stdio,
    });

    child.on('error', reject);
    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }

      reject(new Error(`${command} ${args.join(' ')} exited with ${signal ?? code}`));
    });
  });
}

async function commandSucceeds(cwd, command, args, envOverrides = {}) {
  try {
    await runCommand(cwd, command, args, envOverrides, { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

function runCommandCapture(cwd, command, args, envOverrides = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      env: {
        ...process.env,
        ...envOverrides,
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString();
    });

    child.on('error', reject);
    child.on('exit', (code) => {
      resolve({
        code: code ?? 1,
        stdout,
        stderr,
      });
    });
  });
}

function flattenAttributeDefinitions(attributeDefinitions) {
  return attributeDefinitions.map(([attributeName, attributeType]) => `AttributeName=${attributeName},AttributeType=${attributeType}`);
}

function flattenKeySchema(keySchema) {
  return keySchema.map(([attributeName, keyType]) => `AttributeName=${attributeName},KeyType=${keyType}`);
}

function buildGlobalSecondaryIndexes(globalSecondaryIndexes) {
  return globalSecondaryIndexes.map((index) => ({
    IndexName: index.IndexName,
    KeySchema: index.KeySchema.map(([attributeName, keyType]) => ({
      AttributeName: attributeName,
      KeyType: keyType,
    })),
    Projection: {
      ProjectionType: 'ALL',
    },
  }));
}

async function bootstrapLocalDynamoDB(cwd, envVars) {
  const endpointUrl = envVars.DYNAMODB_ENDPOINT_URL || DEFAULT_LOCAL_DYNAMODB_ENDPOINT;
  const awsEnv = {
    AWS_EC2_METADATA_DISABLED: 'true',
    AWS_REGION: envVars.AWS_REGION || DEFAULT_LOCAL_AWS_REGION,
    AWS_ACCESS_KEY_ID: envVars.AWS_ACCESS_KEY_ID || DEFAULT_LOCAL_AWS_ACCESS_KEY_ID,
    AWS_SECRET_ACCESS_KEY: envVars.AWS_SECRET_ACCESS_KEY || DEFAULT_LOCAL_AWS_SECRET_ACCESS_KEY,
  };

  console.log(`Ensuring local DynamoDB tables exist at ${endpointUrl}`);

  for (const definition of buildLocalDynamoTableDefinitions(envVars)) {
    const exists = await commandSucceeds(
      cwd,
      'aws',
      ['dynamodb', 'describe-table', '--table-name', definition.tableName, '--endpoint-url', endpointUrl],
      awsEnv,
    );

    if (exists) {
      continue;
    }

    const args = [
      'dynamodb',
      'create-table',
      '--table-name',
      definition.tableName,
      '--billing-mode',
      'PAY_PER_REQUEST',
      '--attribute-definitions',
      ...flattenAttributeDefinitions(definition.attributeDefinitions),
      '--key-schema',
      ...flattenKeySchema(definition.keySchema),
      '--endpoint-url',
      endpointUrl,
    ];

    if (definition.globalSecondaryIndexes.length > 0) {
      args.push('--global-secondary-indexes', JSON.stringify(buildGlobalSecondaryIndexes(definition.globalSecondaryIndexes)));
    }

    await runCommand(cwd, 'aws', args, awsEnv);
  }
}

function startLongRunningCommand(cwd, command, args, envOverrides = {}) {
  return spawn(command, args, {
    cwd,
    env: {
      ...process.env,
      ...envOverrides,
    },
    stdio: 'inherit',
    detached: process.platform !== 'win32',
  });
}

async function waitForFile(filePath, timeoutMs = 5000, intervalMs = 100) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    if (fs.existsSync(filePath)) {
      return true;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  return fs.existsSync(filePath);
}

async function waitForCommandSuccess(cwd, command, args, envOverrides = {}, timeoutMs = 5000, intervalMs = 250) {
  const startedAt = Date.now();
  while (Date.now() - startedAt < timeoutMs) {
    if (await commandSucceeds(cwd, command, args, envOverrides)) {
      return true;
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  return commandSucceeds(cwd, command, args, envOverrides);
}

async function ensureContainerRuntime(cwd) {
  const dockerAvailable = await commandSucceeds(cwd, 'docker', ['info']);
  const podmanAvailable = await commandSucceeds(cwd, 'podman', ['--version']);
  const runtimeOptions = resolveContainerRuntimeOptions({
    dockerAvailable,
    podmanAvailable,
  });

  if (!runtimeOptions.shouldUsePodmanSocket || !runtimeOptions.podmanSocketPath) {
    return {
      envOverrides: {},
      auxiliaryChildren: [],
    };
  }

  const socketDir = path.dirname(runtimeOptions.podmanSocketPath);
  fs.mkdirSync(socketDir, { recursive: true });

  const podmanRuntimeReady = async () =>
    waitForCommandSuccess(cwd, 'docker', ['info'], runtimeOptions.envOverrides, 1500, 250);

  if (await podmanRuntimeReady()) {
    console.log(`Using Podman runtime via ${runtimeOptions.envOverrides.DOCKER_HOST}`);
    return {
      envOverrides: runtimeOptions.envOverrides,
      auxiliaryChildren: [],
    };
  }

  if (!fs.existsSync(runtimeOptions.podmanSocketPath)) {
    await commandSucceeds(cwd, 'systemctl', ['--user', 'start', 'podman.socket']);
    await waitForFile(runtimeOptions.podmanSocketPath, 2000);
  }

  if (await podmanRuntimeReady()) {
    console.log(`Using Podman runtime via ${runtimeOptions.envOverrides.DOCKER_HOST}`);
    return {
      envOverrides: runtimeOptions.envOverrides,
      auxiliaryChildren: [],
    };
  }

  let podmanService = null;
  console.log('Starting Podman API service for AWS SAM...');
  podmanService = startLongRunningCommand(
    cwd,
    'podman',
    ['system', 'service', '--time=0', runtimeOptions.envOverrides.DOCKER_HOST],
    runtimeOptions.envOverrides,
  );

  const socketReady = await waitForFile(runtimeOptions.podmanSocketPath, 5000);
  const runtimeReady = socketReady && (await podmanRuntimeReady());
  if (!runtimeReady) {
    terminateChild(podmanService);
    throw new Error(
      `Failed to initialize Podman runtime at ${runtimeOptions.podmanSocketPath}. Ensure podman and podman-docker are installed, then retry ./scripts/dev-local.mjs.`,
    );
  }

  console.log(`Using Podman runtime via ${runtimeOptions.envOverrides.DOCKER_HOST}`);
  return {
    envOverrides: runtimeOptions.envOverrides,
    auxiliaryChildren: podmanService ? [podmanService] : [],
  };
}

async function getListeningPortOwners(cwd, ports) {
  try {
    const { code, stdout } = await runCommandCapture(cwd, 'ss', ['-ltnp']);
    if (code !== 0) {
      return [];
    }
    return parseListeningPidsFromSsOutput(stdout, ports);
  } catch {
    return [];
  }
}

function processExists(pid) {
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

async function terminatePid(pid) {
  if (!processExists(pid)) {
    return;
  }

  try {
    process.kill(pid, 'SIGTERM');
  } catch {
    return;
  }

  const deadline = Date.now() + 2000;
  while (Date.now() < deadline) {
    if (!processExists(pid)) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }

  if (processExists(pid)) {
    try {
      process.kill(pid, 'SIGKILL');
    } catch {
      // Ignore failure to kill; caller will report if port remains occupied.
    }
  }
}

function readProcessCwd(pid) {
  try {
    return fs.readlinkSync(`/proc/${pid}/cwd`);
  } catch {
    return null;
  }
}

async function cleanupProjectOwnedPortConflicts(cwd, projectRoot, ports) {
  const owners = await getListeningPortOwners(cwd, ports);
  for (const owner of owners) {
    const ownerCwd = readProcessCwd(owner.pid);
    if (!shouldStopPortOwner(ownerCwd, projectRoot)) {
      continue;
    }

    console.log(`Stopping stale local process on port ${owner.port} (pid ${owner.pid})...`);
    await terminatePid(owner.pid);
  }

  const unresolvedOwners = await getListeningPortOwners(cwd, ports);
  if (unresolvedOwners.length === 0) {
    return;
  }

  const details = unresolvedOwners.map((owner) => `${owner.port} (pid ${owner.pid})`).join(', ');
  throw new Error(
    `Local ports already in use: ${details}. Stop them and rerun ./scripts/dev-local.mjs.`,
  );
}

function terminateChild(child) {
  if (!child || child.killed) {
    return;
  }

  if (process.platform !== 'win32' && typeof child.pid === 'number') {
    try {
      process.kill(-child.pid, 'SIGTERM');
      return;
    } catch {
      // Fall through to direct child termination if the process group is unavailable.
    }
  }

  child.kill('SIGTERM');
}

async function main() {
  const cliOptions = parseCliArgs();
  const plan = buildDevLocalPlan(process.cwd(), cliOptions);
  const missingEnvFiles = findMissingEnvFiles(plan.envFiles);

  if (missingEnvFiles.length > 0) {
    console.error('Missing required local env files:');
    for (const envFile of missingEnvFiles) {
      console.error(`- ${path.relative(plan.projectRoot, envFile)}`);
    }
    console.error('Create them from the matching *.example files before starting local development.');
    process.exitCode = 1;
    return;
  }

  const backendEnvContent = fs.readFileSync(plan.backend.envFile, 'utf8');
  const backendEnvVars = parseDotenv(backendEnvContent);
  const frontendEnvContent = fs.readFileSync(plan.envFiles[0], 'utf8');
  const frontendEnvOverrides = buildFrontendEnvOverrides(frontendEnvContent);
  await cleanupProjectOwnedPortConflicts(plan.backend.cwd, plan.projectRoot, LOCAL_DEV_PORTS);
  const containerRuntime = await ensureContainerRuntime(plan.backend.cwd);

  if (shouldBootstrapLocalDynamoDB(backendEnvVars.DYNAMODB_ENDPOINT_URL)) {
    await bootstrapLocalDynamoDB(plan.backend.cwd, backendEnvVars);
  }

  if (plan.backend.buildCommand) {
    console.log('Building backend Lambda binaries...');
    await runCommand(plan.backend.cwd, plan.backend.buildCommand.command, plan.backend.buildCommand.args);
  }

  let backend;
  if (plan.backend.buildCommand === null) {
    // Native go run mode — no Docker/SAM needed. Pass .env vars directly.
    console.log('Starting backend on http://localhost:3001 (native go run, no Docker required)');
    backend = startLongRunningCommand(
      plan.backend.cwd,
      plan.backend.startCommand.command,
      plan.backend.startCommand.args,
      { PORT: '3001', ...backendEnvVars },
    );
  } else {
    // SAM local mode
    const samEnvOverrides = createSamEnvOverridesFile(plan.backend.envFile);
    const backendStartCommand = {
      command: plan.backend.startCommand.command,
      args: [...plan.backend.startCommand.args.slice(0, -1), samEnvOverrides.tempFile],
    };
    console.log('Starting backend on http://localhost:3001');
    backend = startLongRunningCommand(
      plan.backend.cwd,
      backendStartCommand.command,
      backendStartCommand.args,
      containerRuntime.envOverrides,
    );
  }

  if (plan.frontend.buildCommand) {
    console.log('Building frontend production bundle (this may take a minute)...');
    await runCommand(plan.frontend.cwd, plan.frontend.buildCommand.command, plan.frontend.buildCommand.args, frontendEnvOverrides);
  }

  console.log('Starting frontend on http://localhost:3000');
  if (Object.keys(frontendEnvOverrides).length > 0) {
    console.log('Canonicalizing frontend local origin settings:', frontendEnvOverrides);
  }
  const frontend = startLongRunningCommand(
    plan.frontend.cwd,
    plan.frontend.startCommand.command,
    plan.frontend.startCommand.args,
    frontendEnvOverrides,
  );

  const runningChildren = [...containerRuntime.auxiliaryChildren, backend, frontend];
  let shuttingDown = false;

  const shutdown = (exitCode = 0) => {
    if (shuttingDown) {
      return;
    }

    shuttingDown = true;
    for (const child of runningChildren) {
      terminateChild(child);
    }

    if (typeof samEnvOverrides !== 'undefined' && samEnvOverrides?.tempDir) {
      fs.rmSync(samEnvOverrides.tempDir, { recursive: true, force: true });
    }

    setTimeout(() => {
      process.exit(exitCode);
    }, 200);
  };

  process.on('SIGINT', () => shutdown(0));
  process.on('SIGTERM', () => shutdown(0));

  for (const [name, child] of [['backend', backend], ['frontend', frontend]]) {
    child.on('error', (error) => {
      console.error(`${name} failed to start:`, error);
      shutdown(1);
    });

    child.on('exit', (code, signal) => {
      if (shuttingDown) {
        return;
      }

      console.error(`${name} exited unexpectedly with ${signal ?? code}. Stopping local stack.`);
      shutdown(code ?? 1);
    });
  }
}

const entryFile = fileURLToPath(import.meta.url);
if (process.argv[1] === entryFile) {
  main().catch((error) => {
    console.error(error instanceof Error ? error.message : error);
    process.exit(1);
  });
}
