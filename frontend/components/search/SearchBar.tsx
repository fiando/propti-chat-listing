'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, SlidersHorizontal, X, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { getCitySuggestions, getProvinceSuggestions } from '@/lib/api';
import type { SearchParams } from '@/types';

interface SearchBarProps {
  initialParams?: SearchParams;
  onSearch: (params: SearchParams) => void;
}

export function SearchBar({ initialParams = {}, onSearch }: SearchBarProps) {
  const [q, setQ] = useState(initialParams.q || '');
  const [province, setProvince] = useState(initialParams.province || '');
  const [city, setCity] = useState(initialParams.city || '');
  const [listingType, setListingType] = useState(initialParams.listingType || '');
  const [showFilters, setShowFilters] = useState(false);
  const [selectedProvinceId, setSelectedProvinceId] = useState('');
  const [provinceQuery, setProvinceQuery] = useState('');
  const [cityQuery, setCityQuery] = useState('');

  const { data: provinces = [], isLoading: loadingProvinces } = useQuery({
    queryKey: ['search-provinces', provinceQuery],
    queryFn: () => getProvinceSuggestions(provinceQuery || undefined),
    staleTime: Infinity,
  });

  const { data: cities = [], isLoading: loadingCities } = useQuery({
    queryKey: ['search-cities', selectedProvinceId, cityQuery],
    queryFn: () => getCitySuggestions(selectedProvinceId, cityQuery || undefined),
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
    setQ(initialParams.q || '');
    setProvince(initialParams.province || '');
    setCity(initialParams.city || '');
    setListingType(initialParams.listingType || '');
    setSelectedProvinceId('');
    setProvinceQuery('');
    setCityQuery('');
  }, [initialParams.city, initialParams.listingType, initialParams.province, initialParams.q]);

  const handleSearch = () => {
    onSearch({
      q: q || undefined,
      province: province || undefined,
      city: city || undefined,
      listingType: (listingType as SearchParams['listingType']) || undefined,
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
    onSearch({});
  };

  const hasFilters = q || province || city || listingType;

  return (
    <div className="bg-white rounded-2xl shadow-card border border-gray-100 p-4">
      {/* Main search row */}
      <div className="flex gap-3">
        <div className="flex-1 relative">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Cari properti, lokasi, atau kata kunci..."
            className="w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand-accent focus:border-transparent bg-gray-50 transition-all"
          />
        </div>

        <button
          onClick={() => setShowFilters(!showFilters)}
          className={cn(
            'flex items-center gap-2 px-4 py-3 rounded-xl border font-medium text-sm transition-all',
            showFilters
              ? 'border-brand-primary bg-brand-light text-brand-primary'
              : 'border-gray-200 text-gray-600 hover:border-gray-300'
          )}
        >
          <SlidersHorizontal className="w-4 h-4" />
          <span className="hidden sm:inline">Filter</span>
          {hasFilters && (
            <span className="w-2 h-2 bg-brand-primary rounded-full" />
          )}
        </button>

        <button
          onClick={handleSearch}
          className="btn-primary text-sm px-6 py-3"
        >
          Cari
        </button>
      </div>

      {/* Expanded filters */}
      {showFilters && (
        <div className="mt-4 pt-4 border-t border-gray-100 grid grid-cols-1 sm:grid-cols-4 gap-3">
          {/* Province */}
          <div className="relative">
            <label className="label text-xs">Provinsi</label>
            <input
              value={provinceQuery}
              onChange={(e) => setProvinceQuery(e.target.value)}
              className="input-field mb-2 text-sm"
              placeholder="Cari provinsi..."
            />
            <div className="relative">
              <select
                value={selectedProvinceId}
                onChange={(e) => {
                  const selectedId = e.target.value;
                  const selectedProvince = provinces.find((item) => item.id === selectedId);
                  setSelectedProvinceId(selectedId);
                  setCityQuery('');
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
            <input
              value={cityQuery}
              onChange={(e) => setCityQuery(e.target.value)}
              disabled={!selectedProvinceId}
              className="input-field mb-2 text-sm disabled:bg-gray-50 disabled:text-gray-400"
              placeholder={!selectedProvinceId ? 'Pilih provinsi dulu' : 'Cari kota...'}
            />
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
