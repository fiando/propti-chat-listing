'use client';

import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronDown, ChevronUp, RotateCcw } from 'lucide-react';
import { cn, AMENITIES_OPTIONS, LEGAL_STATUS_OPTIONS } from '@/lib/utils';
import { getCitySuggestions, getProvinceSuggestions } from '@/lib/api';
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

const ROOM_COUNT_OPTIONS = [1, 2, 3, 4, 5];

function hasAdvancedFilters(params: SearchParams) {
  return Boolean(
    params.bedrooms ||
      params.bathrooms ||
      params.buildingAreaMin ||
      params.buildingAreaMax ||
      params.landAreaMin ||
      params.landAreaMax ||
      params.legalStatus ||
      params.amenities?.length
  );
}

function countActiveAdvancedFilters(params: SearchParams) {
  let count = 0;

  if (params.bedrooms) count += 1;
  if (params.bathrooms) count += 1;
  if (params.buildingAreaMin || params.buildingAreaMax) count += 1;
  if (params.landAreaMin || params.landAreaMax) count += 1;
  if (params.legalStatus) count += 1;
  if (params.amenities?.length) count += params.amenities.length;

  return count;
}

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
    <div className="mb-4 border-b border-gray-100 pb-4 last:mb-0 last:border-0 last:pb-0">
      <button
        onClick={() => setOpen(!open)}
        className="mb-3 flex w-full items-center justify-between text-left"
      >
        <span className="text-sm font-semibold text-gray-900">{title}</span>
        {open ? (
          <ChevronUp className="h-4 w-4 text-gray-400" />
        ) : (
          <ChevronDown className="h-4 w-4 text-gray-400" />
        )}
      </button>
      {open && <div>{children}</div>}
    </div>
  );
}

function CountFilterGroup({
  label,
  value,
  onChange,
}: {
  label: string;
  value?: number;
  onChange: (nextValue?: number) => void;
}) {
  return (
    <FilterSection title={label} defaultOpen>
      <div className="flex flex-wrap gap-2">
        {ROOM_COUNT_OPTIONS.map((count) => (
          <button
            key={count}
            type="button"
            onClick={() => onChange(value === count ? undefined : count)}
            className={cn(
              'rounded-xl border px-3 py-2 text-sm font-semibold transition-all',
              value === count
                ? 'border-brand-primary bg-brand-light text-brand-primary'
                : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'
            )}
          >
            {count}+
          </button>
        ))}
      </div>
    </FilterSection>
  );
}

function AreaRangeInputs({
  title,
  minValue,
  maxValue,
  minKey,
  maxKey,
  params,
  onChange,
}: {
  title: string;
  minValue?: number;
  maxValue?: number;
  minKey: 'buildingAreaMin' | 'landAreaMin';
  maxKey: 'buildingAreaMax' | 'landAreaMax';
  params: SearchParams;
  onChange: (params: SearchParams) => void;
}) {
  return (
    <FilterSection title={title} defaultOpen>
      <div className="grid grid-cols-2 gap-2">
        <input
          type="number"
          min={0}
          value={minValue ?? ''}
          onChange={(event) =>
            onChange({
              ...params,
              [minKey]: event.target.value ? Number(event.target.value) : undefined,
            })
          }
          placeholder="Min m²"
          className="input-field text-sm"
        />
        <input
          type="number"
          min={0}
          value={maxValue ?? ''}
          onChange={(event) =>
            onChange({
              ...params,
              [maxKey]: event.target.value ? Number(event.target.value) : undefined,
            })
          }
          placeholder="Maks m²"
          className="input-field text-sm"
        />
      </div>
    </FilterSection>
  );
}

export function FilterPanel({ params, onChange }: FilterPanelProps) {
  const [selectedProvinceId, setSelectedProvinceId] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(() => hasAdvancedFilters(params));

  const { data: provinces = [], isLoading: loadingProvinces } = useQuery({
    queryKey: ['filter-provinces'],
    queryFn: () => getProvinceSuggestions(),
    staleTime: Infinity,
  });

  const { data: cities = [], isLoading: loadingCities } = useQuery({
    queryKey: ['filter-cities', selectedProvinceId],
    queryFn: () => getCitySuggestions(selectedProvinceId),
    enabled: !!selectedProvinceId,
    staleTime: Infinity,
  });

  useEffect(() => {
    if (!params.province || !provinces.length) {
      setSelectedProvinceId('');
      return;
    }

    const match = provinces.find((item) => item.name.toLowerCase() === params.province?.toLowerCase());
    setSelectedProvinceId(match?.id || '');
  }, [params.province, provinces]);

  useEffect(() => {
    if (hasAdvancedFilters(params)) {
      setShowAdvanced(true);
    }
  }, [params]);

  const activeAdvancedFilterCount = useMemo(() => countActiveAdvancedFilters(params), [params]);

  const handleReset = () => {
    onChange({ searchMode: params.searchMode, smartQuery: params.smartQuery, q: params.q });
  };

  const hasActiveFilters =
    params.province ||
    params.city ||
    params.priceMin ||
    params.priceMax ||
    params.listingType ||
    hasAdvancedFilters(params);

  return (
    <div className="card p-5">
      <div className="mb-5 flex items-center justify-between">
        <h3 className="font-bold text-gray-900">Filter</h3>
        {hasActiveFilters && (
          <button
            onClick={handleReset}
            className="flex items-center gap-1 text-xs text-red-500 transition-colors hover:text-red-600"
          >
            <RotateCcw className="h-3 w-3" />
            Reset
          </button>
        )}
      </div>

      <FilterSection title="Tipe Iklan">
        <div className="space-y-2">
          {[
            { value: '', label: 'Semua' },
            { value: 'sell', label: '🏠 Dijual' },
            { value: 'rent', label: '🔑 Disewa' },
          ].map((opt) => (
            <label key={opt.value} className="group flex cursor-pointer items-center gap-2.5">
              <input
                type="radio"
                name="listingType"
                checked={opt.value === '' ? !params.listingType : params.listingType === opt.value}
                onChange={() =>
                  onChange({
                    ...params,
                    listingType: opt.value ? (opt.value as 'sell' | 'rent') : undefined,
                  })
                }
                className="h-4 w-4 accent-brand-primary"
              />
              <span className="text-sm text-gray-700 group-hover:text-gray-900">{opt.label}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      <FilterSection title="Provinsi">
        <div className="relative">
          <select
            value={selectedProvinceId}
            onChange={(e) => {
              const nextProvinceId = e.target.value;
              const province = provinces.find((item) => item.id === nextProvinceId);
              setSelectedProvinceId(nextProvinceId);
              onChange({
                ...params,
                province: province?.name || undefined,
                city: undefined,
              });
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
          <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
        </div>
      </FilterSection>

      <FilterSection title="Kota">
        <div className="relative">
          <select
            value={params.city || ''}
            onChange={(e) =>
              onChange({
                ...params,
                city: e.target.value || undefined,
              })
            }
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
          <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
        </div>
      </FilterSection>

      <FilterSection title="Rentang Harga">
        <div className="space-y-2">
          <label className="flex items-center gap-2.5 cursor-pointer">
            <input
              type="radio"
              name="priceRange"
              checked={!params.priceMin && !params.priceMax}
              onChange={() => onChange({ ...params, priceMin: undefined, priceMax: undefined })}
              className="h-4 w-4 accent-brand-primary"
            />
            <span className="text-sm text-gray-700">Semua Harga</span>
          </label>
          {PRICE_RANGES.map((range) => (
            <label key={range.label} className="flex cursor-pointer items-center gap-2.5">
              <input
                type="radio"
                name="priceRange"
                checked={params.priceMin === range.min && params.priceMax === range.max}
                onChange={() =>
                  onChange({ ...params, priceMin: range.min, priceMax: range.max })
                }
                className="h-4 w-4 accent-brand-primary"
              />
              <span className="text-sm text-gray-700">{range.label}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      <div className="pt-1">
        <button
          type="button"
          onClick={() => setShowAdvanced((current) => !current)}
          className="flex w-full items-center justify-between rounded-xl border border-dashed border-gray-200 px-3 py-2.5 text-sm font-semibold text-gray-700 transition-colors hover:border-brand-primary/40 hover:text-brand-primary"
        >
          <span className="flex items-center gap-2">
            Filter Lanjutan
            {activeAdvancedFilterCount > 0 && (
              <span className="rounded-full bg-brand-light px-2 py-0.5 text-xs text-brand-primary">
                {activeAdvancedFilterCount}
              </span>
            )}
          </span>
          {showAdvanced ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </button>
      </div>

      {showAdvanced && (
        <div className="mt-4 rounded-2xl border border-gray-100 bg-gray-50/70 p-4">
          <p className="mb-4 text-xs text-gray-500">
            Fokuskan pencarian dengan spesifikasi properti yang benar-benar penting untuk pembeli atau penyewa.
          </p>

          <CountFilterGroup
            label="Kamar Tidur Minimal"
            value={params.bedrooms}
            onChange={(value) => onChange({ ...params, bedrooms: value })}
          />

          <CountFilterGroup
            label="Kamar Mandi Minimal"
            value={params.bathrooms}
            onChange={(value) => onChange({ ...params, bathrooms: value })}
          />

          <AreaRangeInputs
            title="Luas Bangunan"
            minValue={params.buildingAreaMin}
            maxValue={params.buildingAreaMax}
            minKey="buildingAreaMin"
            maxKey="buildingAreaMax"
            params={params}
            onChange={onChange}
          />

          <AreaRangeInputs
            title="Luas Tanah"
            minValue={params.landAreaMin}
            maxValue={params.landAreaMax}
            minKey="landAreaMin"
            maxKey="landAreaMax"
            params={params}
            onChange={onChange}
          />

          <FilterSection title="Sertifikat" defaultOpen>
            <div className="relative">
              <select
                value={params.legalStatus || ''}
                onChange={(event) =>
                  onChange({
                    ...params,
                    legalStatus: event.target.value || undefined,
                  })
                }
                className="input-field appearance-none pr-8 text-sm"
              >
                <option value="">Semua Sertifikat</option>
                {LEGAL_STATUS_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
              <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            </div>
          </FilterSection>

          <FilterSection title="Fasilitas" defaultOpen>
            <div className="grid grid-cols-2 gap-2">
              {AMENITIES_OPTIONS.map((amenity) => {
                const isActive = params.amenities?.includes(amenity.id);
                return (
                  <button
                    key={amenity.id}
                    type="button"
                    onClick={() => {
                      const nextAmenities = isActive
                        ? (params.amenities || []).filter((item) => item !== amenity.id)
                        : [...(params.amenities || []), amenity.id];

                      onChange({
                        ...params,
                        amenities: nextAmenities.length > 0 ? nextAmenities : undefined,
                      });
                    }}
                    className={cn(
                      'rounded-lg border px-3 py-2 text-left text-xs font-medium transition-all',
                      isActive
                        ? 'border-brand-primary bg-brand-light text-brand-primary'
                        : 'border-gray-200 text-gray-600 hover:border-gray-300'
                    )}
                  >
                    {amenity.label}
                  </button>
                );
              })}
            </div>
          </FilterSection>

        </div>
      )}
    </div>
  );
}
