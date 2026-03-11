'use client';

import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Loader2, Plus, Minus } from 'lucide-react';
import {
  LEGAL_STATUS_OPTIONS,
  ORIENTATION_OPTIONS,
  AMENITIES_OPTIONS,
  INDONESIAN_CITIES,
  cn,
} from '@/lib/utils';
import type { ParsedListing, ListingType } from '@/types';
import { ImageUpload } from './ImageUpload';

const listingSchema = z.object({
  title: z.string().min(10, 'Judul minimal 10 karakter').max(200, 'Judul maksimal 200 karakter'),
  description: z.string().min(20, 'Deskripsi minimal 20 karakter'),
  price: z.number({ invalid_type_error: 'Harga harus berupa angka' }).positive('Harga harus lebih dari 0'),
  priceUnit: z.enum(['per_unit', 'per_month', 'negotiable']),
  listingType: z.enum(['sell', 'rent']),
  landArea: z.number().min(0).optional(),
  buildingArea: z.number().min(0).optional(),
  bedrooms: z.number().min(0).int(),
  bathrooms: z.number().min(0).int(),
  frontWidth: z.number().min(0).optional(),
  orientation: z.string().optional(),
  legalStatus: z.string().optional(),
  powerConsumption: z.string().optional(),
  amenities: z.array(z.string()),
  address: z.string().min(5, 'Alamat harus diisi'),
  city: z.string().min(1, 'Kota harus dipilih'),
  district: z.string().optional(),
  images: z.array(z.string()),
});

type ListingFormValues = z.infer<typeof listingSchema>;

export type { ListingFormValues };

interface ListingFormProps {
  initialData?: Partial<ParsedListing>;
  listingId?: string;
  onSubmit: (data: ListingFormValues) => Promise<void>;
  isLoading?: boolean;
  mode?: 'create' | 'edit';
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

export function ListingForm({
  initialData,
  listingId,
  onSubmit,
  isLoading = false,
  mode = 'create',
}: ListingFormProps) {
  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ListingFormValues>({
    resolver: zodResolver(listingSchema),
    defaultValues: {
      title: initialData?.title || '',
      description: initialData?.description || '',
      price: initialData?.price || 0,
      priceUnit:
        (initialData?.priceUnit as ListingFormValues['priceUnit']) || 'per_unit',
      listingType: 'sell' as ListingType,
      landArea: initialData?.propertyDetails?.landArea || 0,
      buildingArea: initialData?.propertyDetails?.buildingArea || 0,
      bedrooms: initialData?.propertyDetails?.bedrooms || 0,
      bathrooms: initialData?.propertyDetails?.bathrooms || 0,
      orientation: initialData?.propertyDetails?.orientation || '',
      legalStatus: initialData?.propertyDetails?.legalStatus || '',
      amenities: initialData?.propertyDetails?.amenities || [],
      address: initialData?.address || '',
      city: '',
      district: '',
      images: [],
    },
  });

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
              <input
                {...register('price', { valueAsNumber: true })}
                type="number"
                className="input-field pl-10"
                placeholder="850000000"
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
                  <option value="per_unit">Harga Total</option>
                  <option value="per_month">Per Bulan</option>
                  <option value="negotiable">Nego</option>
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
            <label className="label">Lebar Muka (m)</label>
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
                    <option key={o} value={o}>
                      {o}
                    </option>
                  ))}
                </select>
              )}
            />
          </div>
        </div>

        {/* Amenities */}
        <div>
          <label className="label">Fasilitas</label>
          <Controller
            name="amenities"
            control={control}
            render={({ field }) => (
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                {AMENITIES_OPTIONS.map((amenity) => {
                  const isChecked = field.value?.includes(amenity.id);
                  return (
                    <button
                      key={amenity.id}
                      type="button"
                      onClick={() => {
                        const current = field.value || [];
                        if (isChecked) {
                          field.onChange(current.filter((a) => a !== amenity.id));
                        } else {
                          field.onChange([...current, amenity.id]);
                        }
                      }}
                      className={cn(
                        'py-2 px-3 rounded-lg border text-xs font-medium text-left transition-all',
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

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">Kota *</label>
            <Controller
              name="city"
              control={control}
              render={({ field }) => (
                <select {...field} className="input-field">
                  <option value="">Pilih kota</option>
                  {INDONESIAN_CITIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
              )}
            />
            {errors.city && <p className="text-red-500 text-xs mt-1">{errors.city.message}</p>}
          </div>
          <div>
            <label className="label">Kecamatan / Kelurahan</label>
            <input
              {...register('district')}
              className="input-field"
              placeholder="Beji"
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
              maxImages={3}
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
