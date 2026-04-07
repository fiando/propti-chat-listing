#!/usr/bin/env node

import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const DEFAULT_STAGE = 'dev';
const DEFAULT_LOCAL_DYNAMODB_ENDPOINT = 'http://127.0.0.1:8000';
const DEFAULT_LOCAL_AWS_REGION = 'ap-southeast-1';
const DEFAULT_LOCAL_AWS_ACCESS_KEY_ID = 'local';
const DEFAULT_LOCAL_AWS_SECRET_ACCESS_KEY = 'local';

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
      buildCommand: {
        command: 'make',
        args: ['build'],
      },
      startCommand: {
        command: 'sam',
        args: ['local', 'start-api', '--host', '127.0.0.1', '--port', '3001', '--env-vars', backendEnvFile],
      },
    },
    frontend: {
      cwd: path.join(projectRoot, 'frontend'),
      startCommand: {
        command: 'npm',
        args: ['run', 'dev', '--', '--hostname', '127.0.0.1', '--port', '3000'],
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

function runCommand(cwd, command, args, envOverrides = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      env: {
        ...process.env,
        ...envOverrides,
      },
      stdio: 'inherit',
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
    await runCommand(cwd, command, args, envOverrides);
    return true;
  } catch {
    return false;
  }
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

function startLongRunningCommand(cwd, command, args) {
  return spawn(command, args, {
    cwd,
    env: process.env,
    stdio: 'inherit',
    detached: process.platform !== 'win32',
  });
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

  if (shouldBootstrapLocalDynamoDB(backendEnvVars.DYNAMODB_ENDPOINT_URL)) {
    await bootstrapLocalDynamoDB(plan.backend.cwd, backendEnvVars);
  }

  console.log('Building backend Lambda binaries...');
  await runCommand(plan.backend.cwd, plan.backend.buildCommand.command, plan.backend.buildCommand.args);

  const samEnvOverrides = createSamEnvOverridesFile(plan.backend.envFile);
  const backendStartCommand = {
    command: plan.backend.startCommand.command,
    args: [...plan.backend.startCommand.args.slice(0, -1), samEnvOverrides.tempFile],
  };

  console.log('Starting backend on http://127.0.0.1:3001');
  const backend = startLongRunningCommand(
    plan.backend.cwd,
    backendStartCommand.command,
    backendStartCommand.args,
  );

  console.log('Starting frontend on http://127.0.0.1:3000');
  const frontend = startLongRunningCommand(
    plan.frontend.cwd,
    plan.frontend.startCommand.command,
    plan.frontend.startCommand.args,
  );

  const runningChildren = [backend, frontend];
  let shuttingDown = false;

  const shutdown = (exitCode = 0) => {
    if (shuttingDown) {
      return;
    }

    shuttingDown = true;
    for (const child of runningChildren) {
      terminateChild(child);
    }

    fs.rmSync(samEnvOverrides.tempDir, { recursive: true, force: true });

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
