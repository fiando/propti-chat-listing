import test from 'node:test';
import assert from 'node:assert/strict';
import {
  getTrackedListingViewKey,
  shouldTrackListingView,
  markListingViewTracked,
} from './listing-view-tracking.ts';

class MemoryStorage implements Storage {
  private readonly store = new Map<string, string>();

  get length() {
    return this.store.size;
  }

  clear() {
    this.store.clear();
  }

  getItem(key: string) {
    return this.store.get(key) ?? null;
  }

  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null;
  }

  removeItem(key: string) {
    this.store.delete(key);
  }

  setItem(key: string, value: string) {
    this.store.set(key, value);
  }
}

test('shouldTrackListingView returns true when listing has never been tracked', () => {
  const storage = new MemoryStorage();

  assert.equal(shouldTrackListingView(storage, 'listing-1', 1_000), true);
});

test('shouldTrackListingView returns false within the cooldown window', () => {
  const storage = new MemoryStorage();
  markListingViewTracked(storage, 'listing-1', 1_000);

  assert.equal(shouldTrackListingView(storage, 'listing-1', 1_000 + 60_000), false);
});

test('shouldTrackListingView returns true after the cooldown window expires', () => {
  const storage = new MemoryStorage();
  markListingViewTracked(storage, 'listing-1', 1_000);

  assert.equal(
    shouldTrackListingView(storage, 'listing-1', 1_000 + 1000 * 60 * 60 * 24 + 1),
    true
  );
});

test('getTrackedListingViewKey namespaces listing IDs predictably', () => {
  assert.equal(
    getTrackedListingViewKey('listing-123'),
    'propti:listing-view:listing-123'
  );
});
