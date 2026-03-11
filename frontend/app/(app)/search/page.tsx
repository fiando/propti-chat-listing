'use client';

import { Suspense } from 'react';
import { useSearch } from '@/hooks/useSearch';
import { SearchBar } from '@/components/search/SearchBar';
import { FilterPanel } from '@/components/search/FilterPanel';
import { ListingGrid } from '@/components/listings/ListingGrid';
import { Loader2, ChevronLeft, ChevronRight } from 'lucide-react';
import type { SearchParams } from '@/types';

function SearchResults() {
  const { params, listings, total, page, isLoading, isFetching, updateSearch, setPage } = useSearch();

  const totalPages = Math.ceil(total / 12);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Search bar */}
      <div className="mb-6">
        <SearchBar initialParams={params} onSearch={updateSearch} />
      </div>

      <div className="flex flex-col lg:flex-row gap-6">
        {/* Filter sidebar */}
        <aside className="lg:w-64 flex-shrink-0">
          <div className="sticky top-24">
            <FilterPanel params={params} onChange={updateSearch} />
          </div>
        </aside>

        {/* Results */}
        <div className="flex-1 min-w-0">
          {/* Results header */}
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-lg font-bold text-gray-900">
                {params.q ? `Hasil pencarian "${params.q}"` : 'Semua Properti'}
              </h1>
              <p className="text-sm text-gray-500">
                {isLoading ? 'Mencari...' : `${total.toLocaleString()} properti ditemukan`}
              </p>
            </div>
            {isFetching && !isLoading && (
              <Loader2 className="w-5 h-5 text-brand-secondary animate-spin" />
            )}
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
