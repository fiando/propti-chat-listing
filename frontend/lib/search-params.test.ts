import test from 'node:test';
import assert from 'node:assert/strict';

import { parseSearchParams, serializeSearchParams } from './search-params.ts';

test('parseSearchParams reads comprehensive property filters from the url', () => {
  const params = new URLSearchParams({
    q: 'rumah keluarga',
    province: 'DKI Jakarta',
    city: 'Jakarta Selatan',
    listingType: 'sell',
    priceMin: '500000000',
    priceMax: '2000000000',
    bedrooms: '3',
    bathrooms: '2',
    buildingAreaMin: '90',
    buildingAreaMax: '180',
    landAreaMin: '120',
    landAreaMax: '250',
    legalStatus: 'SHM - Sertifikat Hak Milik',
    amenities: 'carport,kolam_renang',
    sortBy: 'popular',
    page: '2',
  });

  assert.deepEqual(parseSearchParams(params), {
    q: 'rumah keluarga',
    province: 'DKI Jakarta',
    city: 'Jakarta Selatan',
    listingType: 'sell',
    priceMin: 500000000,
    priceMax: 2000000000,
    bedrooms: 3,
    bathrooms: 2,
    buildingAreaMin: 90,
    buildingAreaMax: 180,
    landAreaMin: 120,
    landAreaMax: 250,
    legalStatus: 'SHM - Sertifikat Hak Milik',
    amenities: ['carport', 'kolam_renang'],
    sortBy: 'popular',
    page: 2,
  });
});

test('serializeSearchParams keeps amenities as a compact comma-separated value', () => {
  const params = serializeSearchParams({
    q: 'rumah keluarga',
    bathrooms: 2,
    buildingAreaMin: 90,
    buildingAreaMax: 180,
    landAreaMin: 120,
    landAreaMax: 250,
    legalStatus: 'SHM - Sertifikat Hak Milik',
    amenities: ['carport', 'kolam_renang'],
    page: 1,
  });

  assert.equal(params.get('q'), 'rumah keluarga');
  assert.equal(params.get('bathrooms'), '2');
  assert.equal(params.get('buildingAreaMin'), '90');
  assert.equal(params.get('buildingAreaMax'), '180');
  assert.equal(params.get('landAreaMin'), '120');
  assert.equal(params.get('landAreaMax'), '250');
  assert.equal(params.get('legalStatus'), 'SHM - Sertifikat Hak Milik');
  assert.equal(params.get('amenities'), 'carport,kolam_renang');
  assert.equal(params.get('page'), '1');
});
