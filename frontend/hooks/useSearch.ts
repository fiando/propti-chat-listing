'use client';

import { useCallback, useTransition } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { getListings } from '@/lib/api';
import type { SearchParams } from '@/types';
import { parseSearchParams, serializeSearchParams } from '@/lib/search-params';

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
      const params = serializeSearchParams(newParams);

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
