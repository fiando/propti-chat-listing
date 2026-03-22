'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, SlidersHorizontal, X, ChevronDown, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { getCitySuggestions, getProvinceSuggestions } from '@/lib/api';
import type { SearchParams } from '@/types';

type SearchMode = 'manual' | 'smart';

interface SearchBarProps {
  initialParams?: SearchParams;
  onSearch: (params: SearchParams) => void | Promise<void>;
  isSearching?: boolean;
  errorMessage?: string;
}

export function SearchBar({
  initialParams = {},
  onSearch,
  isSearching = false,
  errorMessage,
}: SearchBarProps) {
  const [q, setQ] = useState(initialParams.smartQuery || initialParams.q || '');
  const [searchMode, setSearchMode] = useState<SearchMode>(
    initialParams.searchMode || (initialParams.smartQuery ? 'smart' : 'manual')
  );
  const [province, setProvince] = useState(initialParams.province || '');
  const [city, setCity] = useState(initialParams.city || '');
  const [listingType, setListingType] = useState(initialParams.listingType || '');
  const [showFilters, setShowFilters] = useState(false);
  const [selectedProvinceId, setSelectedProvinceId] = useState('');

  const { data: provinces = [], isLoading: loadingProvinces } = useQuery({
    queryKey: ['search-provinces'],
    queryFn: () => getProvinceSuggestions(),
    staleTime: Infinity,
  });

  const { data: cities = [], isLoading: loadingCities } = useQuery({
    queryKey: ['search-cities', selectedProvinceId],
    queryFn: () => getCitySuggestions(selectedProvinceId),
    enabled: !!selectedProvinceId,
    staleTime: Infinity,
  });

  useEffect(() => {
    if (!province || !provinces.length || selectedProvinceId) return;
    const match = provinces.find((item) => item.name.toLowerCase() === province.toLowerCase());
    if (match) {
      setSelectedProvinceId(match.id);
    }
  }, [province, provinces, selectedProvinceId]);

  useEffect(() => {
    setQ(initialParams.smartQuery || initialParams.q || '');
    setSearchMode(initialParams.searchMode || (initialParams.smartQuery ? 'smart' : 'manual'));
    setProvince(initialParams.province || '');
    setCity(initialParams.city || '');
    setListingType(initialParams.listingType || '');
    setSelectedProvinceId('');
  }, [
    initialParams.city,
    initialParams.listingType,
    initialParams.province,
    initialParams.q,
    initialParams.searchMode,
    initialParams.smartQuery,
  ]);

  const handleSearch = () => {
    const trimmedQuery = q.trim();
    void onSearch({
      ...initialParams,
      searchMode,
      smartQuery: searchMode === 'smart' ? trimmedQuery || undefined : undefined,
      q: searchMode === 'manual' ? trimmedQuery || undefined : undefined,
      province: province || undefined,
      city: city || undefined,
      listingType: (listingType as SearchParams['listingType']) || undefined,
      page: 1,
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSearch();
  };

  const clearFilters = () => {
    setQ('');
    setProvince('');
    setCity('');
    setListingType('');
    setSelectedProvinceId('');
    void onSearch({
      ...initialParams,
      searchMode,
      smartQuery: undefined,
      q: undefined,
      province: undefined,
      city: undefined,
      listingType: undefined,
      page: 1,
    });
  };

  const hasFilters = q || province || city || listingType;

  return (
    <div className="bg-white rounded-2xl shadow-card border border-gray-100 p-4">
      <div className="flex flex-col gap-3">
        <div className="grid grid-cols-2 gap-2">
          {([
            { value: 'manual', label: 'Search biasa' },
            { value: 'smart', label: 'Cari dengan kalimat' },
          ] as const).map((option) => (
            <label
              key={option.value}
              className={cn(
                'flex cursor-pointer items-center justify-center gap-2 rounded-xl border px-3 py-2.5 text-sm font-medium transition-all',
                searchMode === option.value
                  ? 'border-brand-primary bg-brand-light text-brand-primary'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300'
              )}
            >
              <input
                type="radio"
                name="search-mode"
                value={option.value}
                checked={searchMode === option.value}
                onChange={() => setSearchMode(option.value)}
                className="sr-only"
              />
              <span>{option.label}</span>
            </label>
          ))}
        </div>

        <div className="relative sm:hidden">
          <Search className="absolute left-3.5 top-4 w-5 h-5 text-gray-400" />
          <textarea
            value={q}
            onChange={(e) => setQ(e.target.value.slice(0, 200))}
            placeholder={
              searchMode === 'smart'
                ? 'Cari dengan kalimat: rumah dijual di Jogja 500 juta - 1 M, SHM, AC, CCTV...'
                : 'Cari lokasi, judul, atau kata kunci properti...'
            }
            rows={searchMode === 'smart' ? 3 : 2}
            maxLength={200}
            className="min-h-[104px] w-full resize-none rounded-xl border border-gray-200 bg-gray-50 pl-11 pr-4 py-3.5 text-sm leading-6 transition-all focus:border-transparent focus:outline-none focus:ring-2 focus:ring-brand-accent"
          />
        </div>

        <div className="relative hidden sm:block">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            value={q}
            onChange={(e) => setQ(e.target.value.slice(0, 200))}
            onKeyDown={handleKeyDown}
            placeholder={
              searchMode === 'smart'
                ? 'Cari dengan kalimat: rumah dijual di Jogja 500 juta - 1 M, SHM, AC, CCTV...'
                : 'Cari lokasi, judul, atau kata kunci properti...'
            }
            maxLength={200}
            className="w-full pl-11 pr-4 py-3.5 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand-accent focus:border-transparent bg-gray-50 transition-all"
          />
        </div>

        <div className="grid grid-cols-2 gap-3 sm:flex sm:items-center">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              'flex items-center justify-center gap-2 px-4 py-3 rounded-xl border font-medium text-sm transition-all sm:w-auto',
              showFilters
                ? 'border-brand-primary bg-brand-light text-brand-primary'
                : 'border-gray-200 text-gray-600 hover:border-gray-300'
            )}
          >
            <SlidersHorizontal className="w-4 h-4" />
            <span>Filter</span>
            {hasFilters && <span className="w-2 h-2 bg-brand-primary rounded-full" />}
          </button>

          <button
            onClick={handleSearch}
            disabled={isSearching}
            className="btn-primary text-sm px-6 py-3 w-full sm:w-auto"
          >
            {isSearching ? (
              <span className="flex items-center justify-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Memproses...
              </span>
            ) : (
              'Cari'
            )}
          </button>
        </div>
      </div>

      <div className="mt-3 flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-xs text-gray-500">
          {searchMode === 'smart' ? (
            <>
              <span className="font-semibold text-brand-primary">Cari dengan kalimat.</span>{' '}
              Semua filter dan urutan bisa terisi otomatis dari satu pencarian.
            </>
          ) : (
            'Search biasa lebih cepat dan langsung pakai filter yang kamu pilih sendiri.'
          )}
        </p>
        <div className="flex items-center gap-3">
          {searchMode === 'smart' && (
            <p className="text-xs text-gray-400 sm:hidden">Tulis beberapa baris jika query-mu panjang.</p>
          )}
          <p className="text-xs text-gray-400">{q.length}/200</p>
        </div>
        {errorMessage && <p className="text-xs font-medium text-red-500">{errorMessage}</p>}
      </div>

      {/* Expanded filters */}
      {showFilters && (
        <div className="mt-4 pt-4 border-t border-gray-100 grid grid-cols-1 sm:grid-cols-4 gap-3">
          {/* Province */}
          <div className="relative">
            <label className="label text-xs">Provinsi</label>
            <div className="relative">
              <select
                value={selectedProvinceId}
                onChange={(e) => {
                  const selectedId = e.target.value;
                  const selectedProvince = provinces.find((item) => item.id === selectedId);
                  setSelectedProvinceId(selectedId);
                  setProvince(selectedProvince?.name || '');
                  setCity('');
                }}
                className="input-field appearance-none pr-8 text-sm"
              >
                <option value="">{loadingProvinces ? 'Memuat provinsi...' : 'Semua Provinsi'}</option>
                {provinces.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </select>
              <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
            </div>
          </div>

          {/* City */}
          <div className="relative">
            <label className="label text-xs">Kota</label>
            <div className="relative">
              <select
                value={city}
                onChange={(e) => setCity(e.target.value)}
                disabled={!selectedProvinceId}
                className="input-field appearance-none pr-8 text-sm disabled:bg-gray-50 disabled:text-gray-400"
              >
                <option value="">
                  {!selectedProvinceId
                    ? 'Pilih provinsi dulu'
                    : loadingCities
                    ? 'Memuat kota...'
                    : 'Semua Kota'}
                </option>
                {cities.map((item) => (
                  <option key={item.id} value={item.name}>
                    {item.name}
                  </option>
                ))}
              </select>
              <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
            </div>
          </div>

          {/* Type */}
          <div>
            <label className="label text-xs">Tipe Iklan</label>
            <div className="relative">
              <select
                value={listingType}
                onChange={(e) => setListingType(e.target.value)}
                className="input-field appearance-none pr-8 text-sm"
              >
                <option value="">Jual &amp; Sewa</option>
                <option value="sell">Dijual</option>
                <option value="rent">Disewa</option>
              </select>
              <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
            </div>
          </div>

          {/* Clear */}
          {hasFilters && (
            <div className="flex items-end">
              <button
                onClick={clearFilters}
                className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-red-500 transition-colors py-3"
              >
                <X className="w-4 h-4" />
                Hapus Filter
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
