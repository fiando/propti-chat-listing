'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import { useSession } from 'next-auth/react';
import { useSearch } from '@/hooks/useSearch';
import { useSaveListing, useSavedListings } from '@/hooks/useListings';
import { SearchBar } from '@/components/search/SearchBar';
import { FilterPanel } from '@/components/search/FilterPanel';
import { ListingGrid } from '@/components/listings/ListingGrid';
import { parseSearchIntent } from '@/lib/api';
import { Loader2, ChevronLeft, ChevronRight } from 'lucide-react';
import { serializeSearchParams } from '@/lib/search-params';
import type { SearchParams } from '@/types';

const SORT_OPTIONS = [
  { value: 'newest', label: 'Terbaru' },
  { value: 'price_asc', label: 'Harga Terendah' },
  { value: 'price_desc', label: 'Harga Tertinggi' },
  { value: 'popular', label: 'Paling Banyak Dilihat' },
];

function SearchResults() {
  const { params, listings, total, page, isLoading, isFetching, updateSearch, setPage } = useSearch();
  const { status } = useSession();
  const { data: savedData } = useSavedListings({ enabled: status === 'authenticated' });
  const { mutateAsync: toggleSave, isPending: isSaving } = useSaveListing();
  const [stagedParams, setStagedParams] = useState<SearchParams>(params);
  const [isSmartSearching, setIsSmartSearching] = useState(false);
  const [smartSearchError, setSmartSearchError] = useState<string | null>(null);

  const totalPages = Math.ceil(total / 12);
  const savedIds = savedData?.items.map((listing) => listing.listingId) ?? [];
  const committedFilterKey = useMemo(
    () => serializeSearchParams({ ...params, page: 1 }).toString(),
    [params]
  );
  const stagedFilterKey = useMemo(
    () => serializeSearchParams({ ...stagedParams, page: 1 }).toString(),
    [stagedParams]
  );

  // Sync staged params when URL (committed) params actually change.
  // Using committedFilterKey (serialized string) instead of the `params` object
  // avoids an infinite re-render loop — `params` is a new reference every render
  // since parseSearchParams always creates a fresh object.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => { setStagedParams(params); }, [committedFilterKey]);

  useEffect(() => {
    if (committedFilterKey === stagedFilterKey) {
      return;
    }

    const timeout = window.setTimeout(() => {
      updateSearch({
        ...stagedParams,
        page: 1,
      });
    }, 400);

    return () => window.clearTimeout(timeout);
  }, [committedFilterKey, stagedFilterKey, stagedParams, updateSearch]);

  const handleSave = async (listingId: string) => {
    await toggleSave({ id: listingId, saved: savedIds.includes(listingId) });
  };

  const handleSearch = async (nextParams: SearchParams) => {
    const searchMode = nextParams.searchMode ?? 'manual';
    const smartQuery = nextParams.smartQuery?.trim();
    setSmartSearchError(null);

    if (searchMode !== 'smart') {
      const manualParams: SearchParams = {
        ...nextParams,
        searchMode,
        q: nextParams.q?.trim() || undefined,
        smartQuery: undefined,
      };
      setStagedParams(manualParams);
      updateSearch(manualParams);
      return;
    }

    if (!smartQuery) {
      const emptySmartParams: SearchParams = {
        ...nextParams,
        searchMode,
        smartQuery: undefined,
        q: undefined,
      };
      setStagedParams(emptySmartParams);
      updateSearch(emptySmartParams);
      return;
    }

    setIsSmartSearching(true);
    try {
      const parsed = await parseSearchIntent(smartQuery);
      const mergedParams: SearchParams = {
        ...parsed.searchParams,
        searchMode,
        smartQuery,
        q: undefined,
        province: nextParams.province || parsed.searchParams.province,
        city: nextParams.city || parsed.searchParams.city,
        listingType: nextParams.listingType || parsed.searchParams.listingType,
        page: 1,
      };

      setStagedParams(mergedParams);
      updateSearch(mergedParams);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Cari dengan kalimat gagal diproses.';
      setSmartSearchError(message);
    } finally {
      setIsSmartSearching(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Search bar */}
        <div className="mb-6">
          <SearchBar
            initialParams={params}
            onSearch={handleSearch}
            isSearching={isSmartSearching}
            errorMessage={smartSearchError || undefined}
          />
        </div>

        <div className="flex flex-col lg:flex-row gap-6">
          {/* Filter sidebar */}
          <aside className="lg:w-64 flex-shrink-0">
            <div className="sticky top-24">
              <FilterPanel params={stagedParams} onChange={setStagedParams} />
            </div>
          </aside>

        {/* Results */}
          <div className="flex-1 min-w-0">
            {/* Results header */}
            <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div className="min-w-0">
                <h1 className="break-words text-lg font-bold text-gray-900">
                  {params.smartQuery || params.q ? `Hasil pencarian "${params.smartQuery || params.q}"` : 'Semua Properti'}
                </h1>
                <p className="text-sm text-gray-500">
                  {isLoading ? 'Mencari...' : `${total.toLocaleString()} properti ditemukan`}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-3 sm:justify-end">
                <div className="flex w-full items-center gap-2 sm:w-auto">
                  <label htmlFor="sort-results" className="text-sm font-medium text-gray-600">
                    Urutkan
                  </label>
                  <select
                    id="sort-results"
                    value={stagedParams.sortBy || 'newest'}
                    onChange={(event) =>
                      setStagedParams((current) => ({
                        ...current,
                        sortBy: event.target.value,
                      }))
                    }
                    className="min-w-0 flex-1 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 focus:border-brand-accent focus:outline-none sm:flex-none"
                  >
                    {SORT_OPTIONS.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
                {isFetching && !isLoading && (
                  <Loader2 className="w-5 h-5 text-brand-secondary animate-spin" />
                )}
              </div>
            </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <div className="text-center">
                <Loader2 className="w-10 h-10 text-brand-primary animate-spin mx-auto mb-3" />
                <p className="text-gray-500 text-sm">Mencari properti...</p>
              </div>
            </div>
          ) : (
            <>
              <ListingGrid
                listings={listings}
                savedIds={savedIds}
                onSave={status === 'authenticated' && !isSaving ? handleSave : undefined}
                emptyMessage="Tidak ada properti yang sesuai dengan filter pencarian."
              />

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 mt-8">
                  <button
                    onClick={() => setPage(page - 1)}
                    disabled={page <= 1}
                    className="w-10 h-10 rounded-xl border border-gray-200 flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </button>

                  {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                    const pageNum = i + 1;
                    return (
                      <button
                        key={pageNum}
                        onClick={() => setPage(pageNum)}
                        className={`w-10 h-10 rounded-xl border text-sm font-semibold transition-all ${
                          page === pageNum
                            ? 'bg-brand-primary text-white border-brand-primary'
                            : 'border-gray-200 text-gray-600 hover:bg-gray-50'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}

                  <button
                    onClick={() => setPage(page + 1)}
                    disabled={page >= totalPages}
                    className="w-10 h-10 rounded-xl border border-gray-200 flex items-center justify-center hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
        </div>
      }
    >
      <SearchResults />
    </Suspense>
  );
}
