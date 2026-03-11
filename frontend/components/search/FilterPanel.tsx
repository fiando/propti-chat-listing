'use client';

import { useState } from 'react';
import { ChevronDown, ChevronUp, RotateCcw } from 'lucide-react';
import { INDONESIAN_CITIES, cn, formatPrice } from '@/lib/utils';
import type { SearchParams } from '@/types';

interface FilterPanelProps {
  params: SearchParams;
  onChange: (params: SearchParams) => void;
}

const PRICE_RANGES = [
  { label: 'Di bawah Rp 500 Jt', min: 0, max: 500_000_000 },
  { label: 'Rp 500 Jt - 1 M', min: 500_000_000, max: 1_000_000_000 },
  { label: 'Rp 1 M - 2 M', min: 1_000_000_000, max: 2_000_000_000 },
  { label: 'Rp 2 M - 5 M', min: 2_000_000_000, max: 5_000_000_000 },
  { label: 'Di atas Rp 5 M', min: 5_000_000_000, max: undefined },
];

const BEDROOM_OPTIONS = [0, 1, 2, 3, 4, 5];

function FilterSection({
  title,
  children,
  defaultOpen = true,
}: {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="border-b border-gray-100 pb-4 mb-4 last:border-0 last:mb-0 last:pb-0">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center justify-between w-full text-left mb-3"
      >
        <span className="font-semibold text-gray-900 text-sm">{title}</span>
        {open ? (
          <ChevronUp className="w-4 h-4 text-gray-400" />
        ) : (
          <ChevronDown className="w-4 h-4 text-gray-400" />
        )}
      </button>
      {open && <div>{children}</div>}
    </div>
  );
}

export function FilterPanel({ params, onChange }: FilterPanelProps) {
  const handleReset = () => {
    onChange({ q: params.q });
  };

  const hasActiveFilters =
    params.city || params.priceMin || params.priceMax || params.bedrooms || params.listingType;

  return (
    <div className="card p-5">
      <div className="flex items-center justify-between mb-5">
        <h3 className="font-bold text-gray-900">Filter</h3>
        {hasActiveFilters && (
          <button
            onClick={handleReset}
            className="flex items-center gap-1 text-xs text-red-500 hover:text-red-600 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            Reset
          </button>
        )}
      </div>

      {/* Listing type */}
      <FilterSection title="Tipe Iklan">
        <div className="space-y-2">
          {[
            { value: '', label: 'Semua' },
            { value: 'sell', label: '🏠 Dijual' },
            { value: 'rent', label: '🔑 Disewa' },
          ].map((opt) => (
            <label key={opt.value} className="flex items-center gap-2.5 cursor-pointer group">
              <input
                type="radio"
                name="listingType"
                checked={
                  opt.value === ''
                    ? !params.listingType
                    : params.listingType === opt.value
                }
                onChange={() =>
                  onChange({
                    ...params,
                    listingType: opt.value ? (opt.value as 'sell' | 'rent') : undefined,
                  })
                }
                className="accent-brand-primary w-4 h-4"
              />
              <span className="text-sm text-gray-700 group-hover:text-gray-900">{opt.label}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      {/* City */}
      <FilterSection title="Kota">
        <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
          <label className="flex items-center gap-2.5 cursor-pointer group">
            <input
              type="radio"
              name="city"
              checked={!params.city}
              onChange={() => onChange({ ...params, city: undefined })}
              className="accent-brand-primary w-4 h-4"
            />
            <span className="text-sm text-gray-700">Semua Kota</span>
          </label>
          {INDONESIAN_CITIES.map((city) => (
            <label key={city} className="flex items-center gap-2.5 cursor-pointer group">
              <input
                type="radio"
                name="city"
                checked={params.city === city}
                onChange={() => onChange({ ...params, city })}
                className="accent-brand-primary w-4 h-4"
              />
              <span className="text-sm text-gray-700 group-hover:text-gray-900">{city}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      {/* Price range */}
      <FilterSection title="Rentang Harga">
        <div className="space-y-2">
          <label className="flex items-center gap-2.5 cursor-pointer">
            <input
              type="radio"
              name="priceRange"
              checked={!params.priceMin && !params.priceMax}
              onChange={() => onChange({ ...params, priceMin: undefined, priceMax: undefined })}
              className="accent-brand-primary w-4 h-4"
            />
            <span className="text-sm text-gray-700">Semua Harga</span>
          </label>
          {PRICE_RANGES.map((range) => (
            <label key={range.label} className="flex items-center gap-2.5 cursor-pointer">
              <input
                type="radio"
                name="priceRange"
                checked={params.priceMin === range.min && params.priceMax === range.max}
                onChange={() =>
                  onChange({ ...params, priceMin: range.min, priceMax: range.max })
                }
                className="accent-brand-primary w-4 h-4"
              />
              <span className="text-sm text-gray-700">{range.label}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      {/* Bedrooms */}
      <FilterSection title="Kamar Tidur" defaultOpen={false}>
        <div className="flex flex-wrap gap-2">
          {BEDROOM_OPTIONS.map((n) => (
            <button
              key={n}
              onClick={() =>
                onChange({ ...params, bedrooms: params.bedrooms === n ? undefined : n })
              }
              className={cn(
                'w-10 h-10 rounded-xl border text-sm font-semibold transition-all',
                params.bedrooms === n
                  ? 'border-brand-primary bg-brand-light text-brand-primary'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300'
              )}
            >
              {n === 0 ? '-' : n === 5 ? '5+' : n}
            </button>
          ))}
        </div>
      </FilterSection>

      {/* Sort */}
      <FilterSection title="Urutkan" defaultOpen={false}>
        <div className="space-y-2">
          {[
            { value: 'newest', label: 'Terbaru' },
            { value: 'price_asc', label: 'Harga Terendah' },
            { value: 'price_desc', label: 'Harga Tertinggi' },
            { value: 'popular', label: 'Paling Banyak Dilihat' },
          ].map((opt) => (
            <label key={opt.value} className="flex items-center gap-2.5 cursor-pointer">
              <input
                type="radio"
                name="sortBy"
                checked={params.sortBy === opt.value || (!params.sortBy && opt.value === 'newest')}
                onChange={() => onChange({ ...params, sortBy: opt.value })}
                className="accent-brand-primary w-4 h-4"
              />
              <span className="text-sm text-gray-700">{opt.label}</span>
            </label>
          ))}
        </div>
      </FilterSection>
    </div>
  );
}
