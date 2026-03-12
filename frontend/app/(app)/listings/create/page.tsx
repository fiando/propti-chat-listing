'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { MessageCircle, PenLine, ArrowLeft } from 'lucide-react';
import { TextParseForm } from '@/components/listings/TextParseForm';
import { ListingForm } from '@/components/listings/ListingForm';
import { useCreateListing } from '@/hooks/useListings';
import type { ParsedListing, CreateListingRequest, Location } from '@/types';
import Link from 'next/link';
import type { ListingFormValues } from '@/components/listings/ListingForm';

type Step = 'choose' | 'parse' | 'form';

export default function CreateListingPage() {
  const router = useRouter();
  const [step, setStep] = useState<Step>('choose');
  const [parsedData, setParsedData] = useState<ParsedListing | null>(null);
  const [parsedLocation, setParsedLocation] = useState<Partial<Location> | null>(null);
  const { mutateAsync: createListing, isPending } = useCreateListing();

  const handleParsed = (result: ParsedListing) => {
    setParsedData(result);
    setParsedLocation({
      address: result.locationSuggestion?.normalizedAddress || result.address || '',
      province: result.locationSuggestion?.province || '',
      city: result.locationSuggestion?.city || '',
      district: result.locationSuggestion?.district || '',
    });
    setStep('form');
  };

  const handleSubmit = async (data: ListingFormValues) => {
    const payload: CreateListingRequest = {
      title: data.title,
      description: data.description,
      price: data.price,
      priceUnit: data.priceUnit,
      listingType: data.listingType,
      propertyDetails: {
        landArea: data.landArea || 0,
        buildingArea: data.buildingArea || 0,
        bedrooms: data.bedrooms,
        bathrooms: data.bathrooms,
        frontWidth: data.frontWidth,
        orientation: data.orientation,
        legalStatus: data.legalStatus,
        powerConsumption: data.powerConsumption,
        amenities: data.amenities,
      },
      location: {
        address: data.address,
        province: data.province,
        city: data.city,
        district: data.district,
      },
      images: data.images,
    };

    const listing = await createListing(payload);
    router.push(`/listings/${listing.listingId}`);
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      {/* Back button */}
      {step !== 'choose' && (
        <button
          onClick={() => setStep(step === 'form' && !parsedData ? 'choose' : step === 'form' ? 'parse' : 'choose')}
          className="flex items-center gap-2 text-gray-500 hover:text-gray-700 mb-6 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          Kembali
        </button>
      )}

      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-black text-brand-primary">
          {step === 'choose' && 'Pasang Iklan Properti'}
          {step === 'parse' && 'Paste Teks Iklanmu'}
          {step === 'form' && 'Detail Iklan'}
        </h1>
        <p className="text-gray-500 mt-1">
          {step === 'choose' && 'Pilih cara memasang iklan yang paling mudah untukmu'}
          {step === 'parse' && 'Copy paste dari WhatsApp atau chat lain, AI kami yang rapikan'}
          {step === 'form' && 'Lengkapi informasi propertimu'}
        </p>

        {/* Step indicator */}
        <div className="flex items-center gap-2 mt-4">
          {(['choose', 'parse', 'form'] as Step[]).map((s, i) => (
            <div key={s} className="flex items-center gap-2">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold transition-all ${
                  step === s
                    ? 'bg-brand-primary text-white'
                    : i < ['choose', 'parse', 'form'].indexOf(step)
                    ? 'bg-brand-accent text-white'
                    : 'bg-gray-200 text-gray-400'
                }`}
              >
                {i + 1}
              </div>
              {i < 2 && <div className="w-8 h-0.5 bg-gray-200" />}
            </div>
          ))}
        </div>
      </div>

      {/* Step: Choose method */}
      {step === 'choose' && (
        <div className="grid sm:grid-cols-2 gap-4">
          <button
            onClick={() => setStep('parse')}
            className="card p-6 text-left hover:-translate-y-1 transition-all duration-300 group border-2 border-transparent hover:border-[#25D366]/30"
          >
            <div className="w-14 h-14 bg-[#25D366] rounded-2xl flex items-center justify-center mb-4 shadow-lg group-hover:scale-110 transition-transform">
              <MessageCircle className="w-7 h-7 text-white" />
            </div>
            <h3 className="font-bold text-gray-900 text-lg mb-2">Paste Teks Iklan</h3>
            <p className="text-gray-500 text-sm leading-relaxed">
              Copy paste teks iklan dari WhatsApp, grup, atau chat. AI kami ekstrak semua detail otomatis.
            </p>
            <div className="mt-4 inline-flex items-center gap-1 text-[#25D366] text-sm font-semibold">
              ⚡ Paling Cepat — 60 detik
            </div>
          </button>

          <button
            onClick={() => setStep('form')}
            className="card p-6 text-left hover:-translate-y-1 transition-all duration-300 group border-2 border-transparent hover:border-brand-accent/30"
          >
            <div className="w-14 h-14 bg-brand-primary rounded-2xl flex items-center justify-center mb-4 shadow-lg group-hover:scale-110 transition-transform">
              <PenLine className="w-7 h-7 text-white" />
            </div>
            <h3 className="font-bold text-gray-900 text-lg mb-2">Isi Manual</h3>
            <p className="text-gray-500 text-sm leading-relaxed">
              Isi formulir detail properti secara manual. Cocok untuk properti dengan spesifikasi khusus.
            </p>
            <div className="mt-4 inline-flex items-center gap-1 text-brand-secondary text-sm font-semibold">
              📋 Kontrol Penuh
            </div>
          </button>
        </div>
      )}

      {/* Step: AI parse */}
      {step === 'parse' && (
        <TextParseForm
          onParsed={handleParsed}
          onManualFill={() => setStep('form')}
        />
      )}

      {/* Step: Form */}
      {step === 'form' && (
        <ListingForm
          initialData={parsedData || undefined}
          initialLocation={parsedLocation || undefined}
          onSubmit={handleSubmit}
          isLoading={isPending}
          mode="create"
        />
      )}
    </div>
  );
}
