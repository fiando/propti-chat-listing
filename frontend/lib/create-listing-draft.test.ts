import test from 'node:test';
import assert from 'node:assert/strict';
import {
  clearCreateListingDraft,
  loadCreateListingDraft,
  saveCreateListingDraft,
  type CreateListingDraft,
} from './create-listing-draft.ts';

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

test('saveCreateListingDraft and loadCreateListingDraft preserve parsed data and form draft', () => {
  const storage = new MemoryStorage();
  const draft: CreateListingDraft = {
    step: 'form',
    parseText: 'Dijual rumah 3KT 2KM Depok',
    parsedData: {
      title: 'Rumah Depok 3KT',
      description: 'Rumah siap huni',
      price: 850000000,
      priceUnit: 'total',
      propertyDetails: {
        landArea: 120,
        buildingArea: 90,
        bedrooms: 3,
        bathrooms: 2,
        amenities: [],
      },
      address: 'Jl. Margonda Raya No. 1',
      confidence: 0.88,
      requiresManualReview: false,
      warnings: [],
    },
    parsedLocation: {
      address: 'Jl. Margonda Raya No. 1',
      province: 'Jawa Barat',
      city: 'Depok',
      district: 'Beji',
    },
    formValues: {
      title: 'Rumah Depok 3KT',
      description: 'Rumah siap huni dekat tol',
      price: 850000000,
      priceUnit: 'total',
      listingType: 'sell',
      landArea: 120,
      buildingArea: 90,
      bedrooms: 3,
      bathrooms: 2,
      amenities: [],
      address: 'Jl. Margonda Raya No. 1',
      province: 'Jawa Barat',
      city: 'Depok',
      district: 'Beji',
      images: ['https://example.com/image-1.jpg'],
    },
  };

  saveCreateListingDraft(storage, draft);

  assert.deepEqual(loadCreateListingDraft(storage), draft);
});

test('loadCreateListingDraft returns null for malformed draft payloads', () => {
  const storage = new MemoryStorage();
  storage.setItem('propti:create-listing-draft', '{invalid');

  assert.equal(loadCreateListingDraft(storage), null);
});

test('clearCreateListingDraft removes stored draft state', () => {
  const storage = new MemoryStorage();
  saveCreateListingDraft(storage, { step: 'parse', parseText: 'text' });

  clearCreateListingDraft(storage);

  assert.equal(loadCreateListingDraft(storage), null);
});
