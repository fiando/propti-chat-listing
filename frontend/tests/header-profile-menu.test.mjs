import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

test('Header avoids a fixed full-screen dismiss layer for the profile menu', async () => {
  const headerPath = path.join(__dirname, '..', 'components', 'common', 'Header.tsx');
  const content = await readFile(headerPath, 'utf8');

  assert.equal(
    content.includes('fixed inset-0 z-10'),
    false,
    'Header should not use a fixed full-screen overlay that can block page clicks while the profile menu is open'
  );
});
