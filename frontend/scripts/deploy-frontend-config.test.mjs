import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const workflowPath = join(process.cwd(), '..', '.github', 'workflows', 'deploy-frontend.yml');
const workflow = readFileSync(workflowPath, 'utf8');

test('frontend deploy workflow injects required auth runtime secrets into Vercel deploys', () => {
  assert.match(workflow, /NEXTAUTH_URL:\s+\$\{\{\s*secrets\.NEXTAUTH_URL\s*\}\}/);
  assert.match(workflow, /NEXTAUTH_SECRET:\s+\$\{\{\s*secrets\.NEXTAUTH_SECRET\s*\}\}/);
  assert.match(workflow, /GOOGLE_CLIENT_ID:\s+\$\{\{\s*secrets\.GOOGLE_CLIENT_ID\s*\}\}/);
  assert.match(workflow, /GOOGLE_CLIENT_SECRET:\s+\$\{\{\s*secrets\.GOOGLE_CLIENT_SECRET\s*\}\}/);
  assert.match(workflow, /--env NEXTAUTH_URL="\$\{\{\s*secrets\.NEXTAUTH_URL\s*\}\}"/);
  assert.match(workflow, /--env NEXTAUTH_SECRET="\$\{\{\s*secrets\.NEXTAUTH_SECRET\s*\}\}"/);
  assert.match(workflow, /--env GOOGLE_CLIENT_ID="\$\{\{\s*secrets\.GOOGLE_CLIENT_ID\s*\}\}"/);
  assert.match(workflow, /--env GOOGLE_CLIENT_SECRET="\$\{\{\s*secrets\.GOOGLE_CLIENT_SECRET\s*\}\}"/);
});
