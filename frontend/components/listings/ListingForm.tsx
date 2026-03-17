'use client';

import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Loader2, Plus, Minus, ChevronDown, ChevronUp, Trash2 } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import {
  LEGAL_STATUS_OPTIONS,
  ORIENTATION_OPTIONS,
  AMENITIES_OPTIONS,
  cn,
  formatAmenityLabel,
} from '@/lib/utils';
import { getProvinceSuggestions, getCitySuggestions, getDistrictSuggestions } from '@/lib/api';
import type { ParsedListing, Location, ListingType } from '@/types';
import { ImageUpload } from './ImageUpload';
import {
  formatListingPriceInput,
  normalizeAmenityIds,
  parseListingPriceInput,
} from '@/lib/listing-form-utils';

const listingSchema = z.object({
  title: z.string().min(10, 'Judul minimal 10 karakter').max(200, 'Judul maksimal 200 karakter'),
  description: z.string().min(20, 'Deskripsi minimal 20 karakter'),
  price: z.number({ invalid_type_error: 'Harga harus berupa angka' }).positive('Harga harus lebih dari 0'),
  priceUnit: z.enum(['total', 'monthly', 'yearly']),
  listingType: z.enum(['sell', 'rent']),
  landArea: z.number().min(0).optional(),
  buildingArea: z.number().min(0).optional(),
  bedrooms: z.number().min(0).int(),
  bathrooms: z.number().min(0).int(),
  frontWidth: z.preprocess(
    (val) => (typeof val === 'number' && isNaN(val) ? undefined : val),
    z.number().min(0).optional()
  ),
  orientation: z.string().optional(),
  legalStatus: z.string().optional(),
  powerConsumption: z.string().optional(),
  amenities: z.array(z.string()),
  address: z.string().min(5, 'Alamat harus diisi'),
  province: z.string().min(1, 'Provinsi harus dipilih'),
  city: z.string().min(1, 'Kota harus dipilih'),
  district: z.string().optional(),
  images: z.array(z.string()),
});

type ListingFormValues = z.infer<typeof listingSchema>;

export type { ListingFormValues };

interface ListingFormProps {
  initialData?: Partial<ParsedListing> & {
    listingType?: ListingType;
    images?: string[];
  };
  initialLocation?: Partial<Location>;
  initialFormValues?: Partial<ListingFormValues>;
  listingId?: string;
  onSubmit: (data: ListingFormValues) => Promise<void>;
  isLoading?: boolean;
  mode?: 'create' | 'edit';
  isPremium?: boolean;
}

function CounterField({
  value,
  onChange,
  min = 0,
  label,
}: {
  value: number;
  onChange: (v: number) => void;
  min?: number;
  label: string;
}) {
  return (
    <div>
      <label className="label">{label}</label>
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={() => onChange(Math.max(min, value - 1))}
          className="w-10 h-10 rounded-xl border border-gray-200 flex items-center justify-center hover:bg-gray-50 transition-colors"
        >
          <Minus className="w-4 h-4" />
        </button>
        <span className="w-8 text-center font-bold text-lg text-gray-900">{value}</span>
        <button
          type="button"
          onClick={() => onChange(value + 1)}
          className="w-10 h-10 rounded-xl border border-gray-200 flex items-center justify-center hover:bg-gray-50 transition-colors"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

function normalizeOrientation(raw: string | undefined): string {
  if (!raw) return '';
  const lower = raw.trim().toLowerCase();
  return ORIENTATION_OPTIONS.find((o) => o.toLowerCase() === lower) ?? '';
}

function normalizeLegalStatus(raw: string | undefined): string {
  if (!raw) return '';
  const upper = raw.trim().toUpperCase();
  // Match by value (short code) or by label prefix (e.g. "SHM - ...")
  return (
    LEGAL_STATUS_OPTIONS.find((o) => o.value.toUpperCase() === upper)?.value ??
    LEGAL_STATUS_OPTIONS.find((o) => o.label.toUpperCase().startsWith(upper + ' '))?.value ??
    ''
  );
}

export function ListingForm({
  initialData,
  initialLocation,
  initialFormValues,
  listingId,
  onSubmit,
  isLoading = false,
  mode = 'create',
  isPremium = false,
}: ListingFormProps) {
  const [selectedProvinceId, setSelectedProvinceId] = useState('');
  const [selectedCityId, setSelectedCityId] = useState('');
  const [showAmenitiesPicker, setShowAmenitiesPicker] = useState(false);

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ListingFormValues>({
    resolver: zodResolver(listingSchema),
    defaultValues: {
      title: initialFormValues?.title || initialData?.title || '',
      description: initialFormValues?.description || initialData?.description || '',
      price: (initialFormValues?.price ?? initialData?.price) || 0,
      priceUnit:
        initialFormValues?.priceUnit ||
        (initialData?.priceUnit as ListingFormValues['priceUnit']) ||
        'total',
      listingType: initialFormValues?.listingType || initialData?.listingType || ('sell' as ListingType),
      landArea: (initialFormValues?.landArea ?? initialData?.propertyDetails?.landArea) || 0,
      buildingArea: (initialFormValues?.buildingArea ?? initialData?.propertyDetails?.buildingArea) || 0,
      bedrooms: (initialFormValues?.bedrooms ?? initialData?.propertyDetails?.bedrooms) || 0,
      bathrooms: (initialFormValues?.bathrooms ?? initialData?.propertyDetails?.bathrooms) || 0,
      frontWidth: initialFormValues?.frontWidth ?? initialData?.propertyDetails?.frontWidth,
      orientation: normalizeOrientation(initialFormValues?.orientation || initialData?.propertyDetails?.orientation),
      legalStatus: normalizeLegalStatus(initialFormValues?.legalStatus || initialData?.propertyDetails?.legalStatus),
      powerConsumption: initialFormValues?.powerConsumption || initialData?.propertyDetails?.powerConsumption || '',
      amenities: normalizeAmenityIds(
        (initialFormValues?.amenities as string[] | undefined) ||
          initialData?.propertyDetails?.amenities ||
          []
      ),
      address: initialFormValues?.address || initialLocation?.address || initialData?.address || '',
      province: initialFormValues?.province || initialLocation?.province || '',
      city: initialFormValues?.city || initialLocation?.city || '',
      district: initialFormValues?.district || initialLocation?.district || '',
      images: initialFormValues?.images || initialData?.images || [],
    },
  });

  const { data: allProvinces = [], isLoading: loadingProvinces } = useQuery({
    queryKey: ['provinces'],
    queryFn: () => getProvinceSuggestions(),
    staleTime: Infinity,
  });

  const { data: allCities = [], isLoading: loadingCities } = useQuery({
    queryKey: ['cities', selectedProvinceId],
    queryFn: () => getCitySuggestions(selectedProvinceId),
    enabled: !!selectedProvinceId,
    staleTime: Infinity,
  });

  const { data: allDistricts = [], isLoading: loadingDistricts } = useQuery({
    queryKey: ['districts', selectedCityId],
    queryFn: () => getDistrictSuggestions(selectedCityId),
    enabled: !!selectedCityId,
    staleTime: Infinity,
  });

  // Resolve initial province name to ID when provinces load
  const initialProvince = initialLocation?.province || '';
  useEffect(() => {
    if (!initialProvince || !allProvinces.length || selectedProvinceId) return;
    const match = allProvinces.find(
      (p) => p.name.toLowerCase() === initialProvince.toLowerCase()
    );
    if (match) setSelectedProvinceId(match.id);
  }, [allProvinces, initialProvince, selectedProvinceId]);

  // Resolve initial city name to ID when cities load
  const initialCity = initialLocation?.city || '';
  useEffect(() => {
    if (!initialCity || !allCities.length || selectedCityId) return;
    const match = allCities.find(
      (c) => c.name.toLowerCase() === initialCity.toLowerCase()
    );
    if (match) setSelectedCityId(match.id);
  }, [allCities, initialCity, selectedCityId]);

  const watchedAmenities = watch('amenities') || [];

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Basic Info */}
      <div className="card p-6 space-y-5">
        <h2 className="font-bold text-gray-900 text-lg border-b pb-3">Informasi Dasar</h2>

        {/* Listing type */}
        <div>
          <label className="label">Tipe Iklan</label>
          <Controller
            name="listingType"
            control={control}
            render={({ field }) => (
              <div className="grid grid-cols-2 gap-3">
                {[
                  { value: 'sell', label: '🏠 Dijual' },
                  { value: 'rent', label: '🔑 Disewa' },
                ].map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => field.onChange(opt.value)}
                    className={cn(
                      'py-3 px-4 rounded-xl border-2 font-semibold text-sm transition-all',
                      field.value === opt.value
                        ? 'border-brand-primary bg-brand-light text-brand-primary'
                        : 'border-gray-200 text-gray-600 hover:border-gray-300'
                    )}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            )}
          />
        </div>

        {/* Title */}
        <div>
          <label className="label">Judul Iklan *</label>
          <input
            {...register('title')}
            className="input-field"
            placeholder="Cth: Rumah Minimalis 3KT Depok Beji, Dekat Tol"
          />
          {errors.title && <p className="text-red-500 text-xs mt-1">{errors.title.message}</p>}
        </div>

        {/* Description */}
        <div>
          <label className="label">Deskripsi *</label>
          <textarea
            {...register('description')}
            rows={5}
            className="input-field resize-none"
            placeholder="Deskripsikan propertimu secara detail: kondisi, keunggulan, akses jalan, fasilitas sekitar, dll."
          />
          {errors.description && (
            <p className="text-red-500 text-xs mt-1">{errors.description.message}</p>
          )}
        </div>

        {/* Price */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">Harga *</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 text-sm font-medium">
                Rp
              </span>
              <Controller
                name="price"
                control={control}
                render={({ field }) => (
                  <input
                    type="text"
                    inputMode="numeric"
                    value={formatListingPriceInput(field.value || 0)}
                    onChange={(event) => field.onChange(parseListingPriceInput(event.target.value))}
                    className="input-field pl-10"
                    placeholder="850.000.000"
                  />
                )}
              />
            </div>
            {errors.price && <p className="text-red-500 text-xs mt-1">{errors.price.message}</p>}
          </div>
          <div>
            <label className="label">Satuan Harga</label>
            <Controller
              name="priceUnit"
              control={control}
              render={({ field }) => (
                <select {...field} className="input-field">
                  <option value="total">Harga Total</option>
                  <option value="monthly">Per Bulan</option>
                  <option value="yearly">Per Tahun</option>
                </select>
              )}
            />
          </div>
        </div>
      </div>

      {/* Property Details */}
      <div className="card p-6 space-y-5">
        <h2 className="font-bold text-gray-900 text-lg border-b pb-3">Detail Properti</h2>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">Luas Tanah (m²)</label>
            <input
              {...register('landArea', { valueAsNumber: true })}
              type="number"
              className="input-field"
              placeholder="120"
            />
          </div>
          <div>
            <label className="label">Luas Bangunan (m²)</label>
            <input
              {...register('buildingArea', { valueAsNumber: true })}
              type="number"
              className="input-field"
              placeholder="90"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-8">
          <Controller
            name="bedrooms"
            control={control}
            render={({ field }) => (
              <CounterField
                value={field.value}
                onChange={field.onChange}
                label="Kamar Tidur"
              />
            )}
          />
          <Controller
            name="bathrooms"
            control={control}
            render={({ field }) => (
              <CounterField
                value={field.value}
                onChange={field.onChange}
                label="Kamar Mandi"
              />
            )}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">Lebar Muka (m) <span className="text-gray-400 font-normal">(opsional)</span></label>
            <input
              {...register('frontWidth', { valueAsNumber: true })}
              type="number"
              className="input-field"
              placeholder="8"
            />
          </div>
          <div>
            <label className="label">Daya Listrik</label>
            <input
              {...register('powerConsumption')}
              className="input-field"
              placeholder="2200 VA"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">Orientasi Bangunan</label>
            <Controller
              name="orientation"
              control={control}
              render={({ field }) => (
                <select {...field} className="input-field">
                  <option value="">Pilih orientasi</option>
                  {ORIENTATION_OPTIONS.map((o) => (
                    <option key={o} value={o}>
                      {o}
                    </option>
                  ))}
                </select>
              )}
            />
          </div>
          <div>
            <label className="label">Status Sertifikat</label>
            <Controller
              name="legalStatus"
              control={control}
              render={({ field }) => (
                <select {...field} className="input-field">
                  <option value="">Pilih sertifikat</option>
                  {LEGAL_STATUS_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </select>
              )}
            />
          </div>
        </div>

        {/* Amenities */}
        <div>
          <div className="flex items-center justify-between gap-3 mb-3">
            <div>
              <label className="label mb-0">Fasilitas</label>
              <p className="text-xs text-gray-500 mt-1">
                Pilih fasilitas penting, lalu buka daftar lengkap hanya saat perlu.
              </p>
            </div>
            <button
              type="button"
              onClick={() => setShowAmenitiesPicker((current) => !current)}
              className="inline-flex items-center gap-2 rounded-xl border border-gray-200 px-3 py-2 text-xs font-semibold text-gray-700 hover:border-brand-primary/40 hover:text-brand-primary transition-colors"
            >
              {showAmenitiesPicker ? 'Tutup daftar' : 'Pilih fasilitas'}
              {showAmenitiesPicker ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </button>
          </div>

          {watchedAmenities.length > 0 ? (
            <div className="mb-3 flex flex-wrap gap-2">
              {watchedAmenities.map((amenityId) => (
                <span
                  key={amenityId}
                  className="inline-flex items-center gap-1 rounded-full border border-brand-primary/20 bg-brand-light px-3 py-1 text-xs font-medium text-brand-primary"
                >
                  {formatAmenityLabel(amenityId)}
                </span>
              ))}
            </div>
          ) : (
            <p className="mb-3 text-sm text-gray-500">Belum ada fasilitas yang dipilih.</p>
          )}

          <Controller
            name="amenities"
            control={control}
            render={({ field }) => (
              <>
                {showAmenitiesPicker && (
                  <div className="rounded-2xl border border-gray-200 bg-gray-50/70 p-4">
                    <div className="mb-3 flex items-center justify-between gap-3">
                      <p className="text-xs text-gray-500">
                        Daftar fasilitas diperluas untuk rumah, apartemen, ruko, gudang, dan properti komersial.
                      </p>
                      {field.value?.length ? (
                        <button
                          type="button"
                          onClick={() => field.onChange([])}
                          className="inline-flex items-center gap-1 text-xs font-medium text-red-500 hover:text-red-600"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                          Kosongkan
                        </button>
                      ) : null}
                    </div>

                    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                      {AMENITIES_OPTIONS.map((amenity) => {
                        const isChecked = field.value?.includes(amenity.id);
                        return (
                          <button
                            key={amenity.id}
                            type="button"
                            onClick={() => {
                              const current = normalizeAmenityIds(field.value || []);
                              if (isChecked) {
                                field.onChange(current.filter((item) => item !== amenity.id));
                              } else {
                                field.onChange(normalizeAmenityIds([...current, amenity.id]));
                              }
                            }}
                            className={cn(
                              'rounded-lg border px-3 py-2 text-left text-xs font-medium transition-all',
                              isChecked
                                ? 'border-brand-primary bg-brand-light text-brand-primary'
                                : 'border-gray-200 text-gray-600 hover:border-gray-300'
                            )}
                          >
                            {amenity.label}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                )}
              </>
            )}
          />
          {watchedAmenities.length > 0 && (
            <p className="text-xs text-brand-secondary mt-2">
              {watchedAmenities.length} fasilitas dipilih
            </p>
          )}
        </div>
      </div>

      {/* Location */}
      <div className="card p-6 space-y-5">
        <h2 className="font-bold text-gray-900 text-lg border-b pb-3">Lokasi</h2>

        <div>
          <label className="label">Alamat Lengkap *</label>
          <textarea
            {...register('address')}
            rows={2}
            className="input-field resize-none"
            placeholder="Jl. Margonda Raya No. 45, Beji, Depok"
          />
          {errors.address && (
            <p className="text-red-500 text-xs mt-1">{errors.address.message}</p>
          )}
        </div>

        {/* Province */}
        <div>
          <label className="label">Provinsi *</label>
          <Controller
            name="province"
            control={control}
            render={({ field }) => (
              <select
                value={selectedProvinceId}
                onChange={(e) => {
                  const opt = allProvinces.find((p) => p.id === e.target.value);
                  setSelectedProvinceId(e.target.value);
                  setSelectedCityId('');
                  field.onChange(opt?.name || '');
                  setValue('city', '');
                  setValue('district', '');
                }}
                className="input-field"
                disabled={loadingProvinces}
              >
                <option value="">
                  {loadingProvinces ? 'Memuat provinsi...' : 'Pilih provinsi'}
                </option>
                {allProvinces.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            )}
          />
          {errors.province && (
            <p className="text-red-500 text-xs mt-1">{errors.province.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* City */}
          <div>
            <label className="label">Kota / Kabupaten *</label>
            <Controller
              name="city"
              control={control}
              render={({ field }) => (
                <select
                  value={selectedCityId}
                  onChange={(e) => {
                    const opt = allCities.find((c) => c.id === e.target.value);
                    setSelectedCityId(e.target.value);
                    field.onChange(opt?.name || '');
                    setValue('district', '');
                  }}
                  className="input-field"
                  disabled={!selectedProvinceId || loadingCities}
                >
                  <option value="">
                    {!selectedProvinceId
                      ? 'Pilih provinsi dulu'
                      : loadingCities
                      ? 'Memuat kota...'
                      : 'Pilih kota'}
                  </option>
                  {allCities.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              )}
            />
            {errors.city && (
              <p className="text-red-500 text-xs mt-1">{errors.city.message}</p>
            )}
          </div>

          {/* District */}
          <div>
            <label className="label">Kecamatan</label>
            <Controller
              name="district"
              control={control}
              render={({ field }) => (
                <select
                  value={
                    loadingDistricts
                      ? ''
                      : allDistricts.find((d) => d.name === field.value)?.id || ''
                  }
                  onChange={(e) => {
                    const opt = allDistricts.find((d) => d.id === e.target.value);
                    field.onChange(opt?.name || e.target.value);
                  }}
                  className="input-field"
                  disabled={!selectedCityId || loadingDistricts}
                >
                  <option value="">
                    {!selectedCityId
                      ? 'Pilih kota dulu'
                      : loadingDistricts
                      ? 'Memuat kecamatan...'
                      : 'Pilih kecamatan (opsional)'}
                  </option>
                  {allDistricts.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              )}
            />
          </div>
        </div>
      </div>

      {/* Images */}
      <div className="card p-6 space-y-4">
        <h2 className="font-bold text-gray-900 text-lg border-b pb-3">Foto Properti</h2>
        <Controller
          name="images"
          control={control}
          render={({ field }) => (
            <ImageUpload
              listingId={listingId}
              images={field.value}
              onChange={field.onChange}
              maxImages={isPremium ? 30 : 3}
            />
          )}
        />
      </div>

      {/* Submit */}
      <button
        type="submit"
        disabled={isLoading}
        className="w-full btn-primary text-base py-4 flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {isLoading ? (
          <>
            <Loader2 className="w-5 h-5 animate-spin" />
            Menyimpan iklan...
          </>
        ) : (
          <>{mode === 'create' ? 'Pasang Iklan Sekarang' : 'Simpan Perubahan'}</>
        )}
      </button>
    </form>
  );
}
