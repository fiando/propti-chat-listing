import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const packageJson = JSON.parse(readFileSync(join(process.cwd(), 'package.json'), 'utf8'));
const rootLayout = readFileSync(join(process.cwd(), 'app', 'layout.tsx'), 'utf8');

test('frontend declares Vercel Analytics and renders it from the App Router root layout', () => {
  assert.ok(packageJson.dependencies['@vercel/analytics']);
  assert.match(rootLayout, /import\s+\{\s*Analytics\s*\}\s+from\s+'@vercel\/analytics\/next';/);
  assert.match(rootLayout, /<Analytics\s*\/>/);
});
