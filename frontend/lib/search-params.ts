import type { SearchParams } from '@/types';

function parseNumber(value: string | null): number | undefined {
  if (!value) {
    return undefined;
  }

  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function parseAmenities(value: string | null): string[] | undefined {
  if (!value) {
    return undefined;
  }

  const amenities = value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);

  return amenities.length > 0 ? amenities : undefined;
}

export function parseSearchParams(searchParams: URLSearchParams): SearchParams {
  return {
    searchMode: (searchParams.get('searchMode') as SearchParams['searchMode']) || undefined,
    smartQuery: searchParams.get('smartQuery') || undefined,
    q: searchParams.get('q') || undefined,
    province: searchParams.get('province') || undefined,
    city: searchParams.get('city') || undefined,
    listingType: (searchParams.get('listingType') as SearchParams['listingType']) || undefined,
    priceMin: parseNumber(searchParams.get('priceMin')),
    priceMax: parseNumber(searchParams.get('priceMax')),
    bedrooms: parseNumber(searchParams.get('bedrooms')),
    bathrooms: parseNumber(searchParams.get('bathrooms')),
    buildingAreaMin: parseNumber(searchParams.get('buildingAreaMin')),
    buildingAreaMax: parseNumber(searchParams.get('buildingAreaMax')),
    landAreaMin: parseNumber(searchParams.get('landAreaMin')),
    landAreaMax: parseNumber(searchParams.get('landAreaMax')),
    legalStatus: searchParams.get('legalStatus') || undefined,
    amenities: parseAmenities(searchParams.get('amenities')),
    sortBy: searchParams.get('sortBy') || undefined,
    page: parseNumber(searchParams.get('page')) || 1,
  };
}

export function serializeSearchParams(params: SearchParams): URLSearchParams {
  const nextParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === '' || value === null) {
      return;
    }

    if (Array.isArray(value)) {
      if (value.length > 0) {
        nextParams.set(key, value.join(','));
      }
      return;
    }

    nextParams.set(key, String(value));
  });

  return nextParams;
}
