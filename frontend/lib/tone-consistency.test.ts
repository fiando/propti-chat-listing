import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const landingPage = readFileSync(new URL('../app/(app)/page.tsx', import.meta.url), 'utf8');
const parseForm = readFileSync(new URL('../components/listings/TextParseForm.tsx', import.meta.url), 'utf8');
const privacyPage = readFileSync(new URL('../app/(app)/privacy/page.tsx', import.meta.url), 'utf8');
const termsPage = readFileSync(new URL('../app/(app)/terms/page.tsx', import.meta.url), 'utf8');

test('product UI copy uses kamu instead of Anda', () => {
  assert.match(landingPage, /Yang Bisa kamu Cek Langsung di Propti/i);
  assert.match(parseForm, /Sedang merapikan detail dari chat kamu/i);
  assert.doesNotMatch(landingPage, /Yang Bisa Anda Cek Langsung di Propti/i);
  assert.doesNotMatch(parseForm, /chat Anda/i);
});

test('legal pages keep formal Anda wording', () => {
  assert.match(privacyPage, /data yang Anda berikan/i);
  assert.match(termsPage, /Dengan menggunakan Propti, Anda setuju/i);
  assert.doesNotMatch(privacyPage, /\bkamu\b/i);
  assert.doesNotMatch(termsPage, /\bKamu\b|\bkamu\b/i);
});
