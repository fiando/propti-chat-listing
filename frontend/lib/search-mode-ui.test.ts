import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const searchBarFile = readFileSync(new URL('../components/search/SearchBar.tsx', import.meta.url), 'utf8');
const searchPageFile = readFileSync(new URL('../app/(app)/search/page.tsx', import.meta.url), 'utf8');

test('search bar exposes an explicit manual vs smart search mode and caps smart input to 200 characters', () => {
  assert.match(searchBarFile, /type SearchMode = 'manual' \| 'smart'/);
  assert.match(searchBarFile, /name="search-mode"/);
  assert.match(searchBarFile, /value: 'manual'/);
  assert.match(searchBarFile, /value: 'smart'/);
  assert.match(searchBarFile, /maxLength=\{200\}/);
  assert.match(searchBarFile, /slice\(0, 200\)/);
});

test('search page only calls parseSearchIntent when smart mode is selected', () => {
  assert.match(searchPageFile, /const searchMode = nextParams\.searchMode \?\? 'manual'/);
  assert.match(searchPageFile, /if \(searchMode !== 'smart'\)/);
  assert.match(searchPageFile, /const smartQuery = nextParams\.smartQuery\?\.trim\(\)/);
});
