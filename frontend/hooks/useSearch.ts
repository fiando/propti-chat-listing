'use client';

import { useCallback, useTransition } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { getListings } from '@/lib/api';
import type { SearchParams } from '@/types';

function parseSearchParams(searchParams: URLSearchParams): SearchParams {
  return {
    q: searchParams.get('q') || undefined,
    city: searchParams.get('city') || undefined,
    listingType: (searchParams.get('listingType') as SearchParams['listingType']) || undefined,
    priceMin: searchParams.get('priceMin')
      ? Number(searchParams.get('priceMin'))
      : undefined,
    priceMax: searchParams.get('priceMax')
      ? Number(searchParams.get('priceMax'))
      : undefined,
    bedrooms: searchParams.get('bedrooms')
      ? Number(searchParams.get('bedrooms'))
      : undefined,
    sortBy: searchParams.get('sortBy') || undefined,
    page: searchParams.get('page') ? Number(searchParams.get('page')) : 1,
  };
}

export function useSearch() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [, startTransition] = useTransition();

  const currentParams = parseSearchParams(searchParams);

  const { data, isLoading, isFetching, error } = useQuery({
    queryKey: ['search', currentParams],
    queryFn: () => getListings(currentParams),
  });

  const updateSearch = useCallback(
    (newParams: SearchParams) => {
      const params = new URLSearchParams();
      Object.entries(newParams).forEach(([key, value]) => {
        if (value !== undefined && value !== '' && value !== null) {
          params.set(key, String(value));
        }
      });

      startTransition(() => {
        router.push(`${pathname}?${params.toString()}`);
      });
    },
    [router, pathname]
  );

  const setPage = useCallback(
    (page: number) => {
      updateSearch({ ...currentParams, page });
    },
    [currentParams, updateSearch]
  );

  return {
    params: currentParams,
    listings: data?.items || [],
    total: data?.total || 0,
    page: data?.page || 1,
    isLoading,
    isFetching,
    error,
    updateSearch,
    setPage,
  };
}
