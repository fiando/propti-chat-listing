#!/usr/bin/env node

import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

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

export function buildSamEnvOverrides(envContent) {
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

  return {
    Parameters: parameters,
  };
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

function runCommand(cwd, command, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      env: process.env,
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
