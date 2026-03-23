import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const rootLayoutFile = readFileSync(new URL('../app/layout.tsx', import.meta.url), 'utf8');
const manifestFile = readFileSync(new URL('../public/site.webmanifest', import.meta.url), 'utf8');

test('root layout exports explicit mobile viewport and standalone apple web app metadata', () => {
  assert.match(rootLayoutFile, /export const viewport:\s*Viewport\s*=\s*\{/);
  assert.match(rootLayoutFile, /width:\s*['"]device-width['"]/);
  assert.match(rootLayoutFile, /initialScale:\s*1/);
  assert.match(rootLayoutFile, /viewportFit:\s*['"]cover['"]/);
  assert.match(rootLayoutFile, /appleWebApp:\s*\{/);
  assert.match(rootLayoutFile, /capable:\s*true/);
  assert.match(rootLayoutFile, /title:\s*['"]Propti['"]/);
});

test('manifest keeps the install title short for the PWA shell', () => {
  assert.match(manifestFile, /"name"\s*:\s*"Propti"/);
  assert.match(manifestFile, /"short_name"\s*:\s*"Propti"/);
  assert.match(manifestFile, /"display"\s*:\s*"standalone"/);
});
