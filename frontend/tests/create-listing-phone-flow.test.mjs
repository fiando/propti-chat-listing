import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

test('listing form keeps phone in create-only schema and create-only UI gating', () => {
  const formFile = readFileSync(new URL('../components/listings/ListingForm.tsx', import.meta.url), 'utf8');
  const baseSchemaMatch = formFile.match(/const listingBaseSchema = z\.object\(\{([\s\S]*?)\n\}\);/);

  assert.ok(baseSchemaMatch, 'expected a shared base listing schema');
  assert.doesNotMatch(baseSchemaMatch[1], /phone:/);
  assert.match(
    formFile,
    /const createListingSchema = listingBaseSchema\.extend\(\{\s*phone: z\.string\(\)\.min\(1, 'Nomor telepon harus diisi'\),\s*\}\);/s
  );
  assert.match(formFile, /const editListingSchema = listingBaseSchema;/);
  assert.match(
    formFile,
    /resolver: zodResolver\(mode === 'create' \? createListingSchema : editListingSchema\)/
  );
  assert.match(
    formFile,
    /\{mode === 'create' && \(\s*<div>\s*<label className="label">Nomor Telepon \*<\/label>/s
  );
});

test('create listing flow prefills and saves phone from profile', () => {
  const createPageFile = readFileSync(new URL('../app/(app)/listings/create/page.tsx', import.meta.url), 'utf8');
  const createClientFile = readFileSync(new URL('../components/listings/CreateListingClient.tsx', import.meta.url), 'utf8');

  assert.match(createPageFile, /initialProfilePhone=\{profile\?\.phone \?\? ''\}/);
  assert.match(createClientFile, /function seedDraftPhone\(/);
  assert.match(
    createClientFile,
    /const \[formDraft, setFormDraft\] = useState<Partial<ListingFormValues> \| null>\(\(\) =>\s*seedDraftPhone\(null, normalizedInitialProfilePhone\)\s*\)/
  );
  assert.match(
    createClientFile,
    /setFormDraft\(\s*seedDraftPhone\(\(draft\.formValues as Partial<ListingFormValues> \| undefined\) \|\| null, normalizedInitialProfilePhone\)\s*\)/
  );
  assert.match(createClientFile, /await updateProfile\(\{ phone: data\.phone\.trim\(\) \}\)/);
  assert.match(createClientFile, /initialFormValues=\{formDraft \|\| undefined\}/);
});
