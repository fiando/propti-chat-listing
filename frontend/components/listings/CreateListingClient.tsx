'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { MessageCircle, PenLine, ArrowLeft, AlertTriangle, Crown } from 'lucide-react';
import { TextParseForm } from '@/components/listings/TextParseForm';
import { ListingForm } from '@/components/listings/ListingForm';
import { useCreateListing } from '@/hooks/useListings';
import type { ParsedListing, CreateListingRequest, Location } from '@/types';
import type { ListingFormValues } from '@/components/listings/ListingForm';
import {
  clearCreateListingDraft,
  loadCreateListingDraft,
  saveCreateListingDraft,
} from '@/lib/create-listing-draft';
import { useToast } from '@/app/toaster';
import {
  CREATE_LISTING_ACCESS_ERROR_MESSAGE,
  BASIC_TIER_LISTING_LIMIT,
  FREE_TIER_LISTING_LIMIT,
  PREMIUM_TIER_LISTING_LIMIT,
  PRO_TIER_LISTING_LIMIT,
  getCreateListingErrorMessage,
} from '@/lib/create-listing-errors';
import { normalizeAmenityIds } from '@/lib/listing-form-utils';
import { updateProfile, uploadListingImage, prepareListingUpload } from '@/lib/api';
import { uploadPendingListingImages } from '@/lib/listing-images';
import { ImageLimits } from '@/types';

type Step = 'choose' | 'parse' | 'form';

type CreateAccessState = {
  status: 'ready' | 'checking' | 'blocked' | 'error';
  activeListingsCount: number;
  message?: string;
};

type CreateListingClientProps = {
  initialIsAuthenticated: boolean;
  initialTier: 'free' | 'basic' | 'premium' | 'pro';
  initialProfilePhone: string;
  initialCreateAccessState: CreateAccessState;
};

function seedDraftPhone(
  draft: Partial<ListingFormValues> | null,
  profilePhone: string
): Partial<ListingFormValues> | null {
  if (!profilePhone || draft?.phone?.trim()) {
    return draft;
  }

  return {
    ...(draft ?? {}),
    phone: profilePhone,
  };
}

export function CreateListingClient({
  initialIsAuthenticated,
  initialTier,
  initialProfilePhone,
  initialCreateAccessState,
}: CreateListingClientProps) {
  const router = useRouter();
  const [step, setStep] = useState<Step>('choose');
  const [parsedData, setParsedData] = useState<ParsedListing | null>(null);
  const [parsedLocation, setParsedLocation] = useState<Partial<Location> | null>(null);
  const [parseTextDraft, setParseTextDraft] = useState('');
  const normalizedInitialProfilePhone = initialProfilePhone.trim();
  const [formDraft, setFormDraft] = useState<Partial<ListingFormValues> | null>(() =>
    seedDraftPhone(null, normalizedInitialProfilePhone)
  );
  const [draftRestored, setDraftRestored] = useState(false);
  const { mutateAsync: createListing, isPending } = useCreateListing();
  const { toast } = useToast();
  const savedProfilePhoneRef = useRef(normalizedInitialProfilePhone);

  const isAuthenticated = initialIsAuthenticated;
  const isPremium = initialTier !== 'free';
  const createAccessState = initialCreateAccessState;
  const activeListingsCount = createAccessState.activeListingsCount;
  const listingLimit =
    initialTier === 'free'
      ? FREE_TIER_LISTING_LIMIT
      : initialTier === 'basic'
      ? BASIC_TIER_LISTING_LIMIT
      : initialTier === 'premium'
      ? PREMIUM_TIER_LISTING_LIMIT
      : PRO_TIER_LISTING_LIMIT;
  const hasCreateAccessError = createAccessState.status === 'error';
  const isCreateBlocked = createAccessState.status === 'blocked';
  const visibleStep: Step = hasCreateAccessError || isCreateBlocked ? 'choose' : step;

  useEffect(() => {
    if (!isCreateBlocked && !hasCreateAccessError) {
      return;
    }

    setStep('choose');
  }, [hasCreateAccessError, isCreateBlocked]);

  const applyDraft = useCallback((draft: NonNullable<ReturnType<typeof loadCreateListingDraft>>) => {
    setStep(draft.step);
    setParseTextDraft(draft.parseText || '');
    setParsedData(draft.parsedData || null);
    setParsedLocation(draft.parsedLocation || null);
    setFormDraft(
      seedDraftPhone((draft.formValues as Partial<ListingFormValues> | undefined) || null, normalizedInitialProfilePhone)
    );
    setDraftRestored(true);
  }, [normalizedInitialProfilePhone]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    const draft = loadCreateListingDraft(window.localStorage);
    if (!draft) {
      return;
    }

    applyDraft(draft);
  }, [applyDraft]);

  const discardDraft = () => {
    if (typeof window !== 'undefined') {
      clearCreateListingDraft(window.localStorage);
    }

    setDraftRestored(false);
    setParseTextDraft('');
    setParsedData(null);
    setParsedLocation(null);
    setFormDraft(seedDraftPhone(null, normalizedInitialProfilePhone));
    setStep('choose');
    toast('Draft dibuang. Kamu bisa mulai dari awal.', 'success');
  };

  const persistDraft = (draft: {
    step: Step;
    parseText?: string;
    formValues?: Partial<ListingFormValues>;
  }) => {
    if (typeof window === 'undefined') {
      return;
    }

    const sanitizedFormValues = draft.formValues
      ? {
          ...draft.formValues,
          images: [],
        }
      : formDraft
      ? {
          ...formDraft,
          images: [],
        }
      : undefined;

    saveCreateListingDraft(window.localStorage, {
      step: draft.step,
      parseText: draft.parseText ?? parseTextDraft,
      parsedData: parsedData ?? undefined,
      parsedLocation: parsedLocation ?? undefined,
      formValues: sanitizedFormValues,
    });
  };

  const goToLoginWithDraft = (draft: { step: Step; parseText?: string; formValues?: Partial<ListingFormValues> }) => {
    persistDraft(draft);
    router.push(`/login?callbackUrl=${encodeURIComponent('/listings/create')}`);
  };

  const handleUseParsedResult = (result: ParsedListing) => {
    setParsedData(result);
    setFormDraft(seedDraftPhone(null, normalizedInitialProfilePhone));
    setParsedLocation({
      address: result.locationSuggestion?.normalizedAddress || result.address || '',
      province: result.locationSuggestion?.province || '',
      city: result.locationSuggestion?.city || '',
      district: result.locationSuggestion?.district || '',
    });
    setStep('form');
    setDraftRestored(false);

    if (typeof window !== 'undefined') {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const handleRequireAuthForParse = (text: string) => {
    setParseTextDraft(text);
    goToLoginWithDraft({ step: 'parse', parseText: text });
  };

  const submitListing = async (data: ListingFormValues) => {
    const normalizedPhone = data.phone.trim();
    if (normalizedPhone !== savedProfilePhoneRef.current) {
      await updateProfile({ phone: data.phone.trim() });
      savedProfilePhoneRef.current = normalizedPhone;
    }

    const uploadPayload = await uploadPendingListingImages(
      data.images,
      {
        prepareUpload: prepareListingUpload,
        uploadObject: uploadListingImage,
      }
    );

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
        amenities: normalizeAmenityIds(data.amenities),
      },
      location: {
        address: data.address,
        province: data.province,
        city: data.city,
        district: data.district,
      },
      newImageUploadSessionIds: uploadPayload.newImageUploadSessionIds,
      featuredUploadSessionId: uploadPayload.featuredUploadSessionId,
    };

    const listing = await createListing(payload);
    if (typeof window !== 'undefined') {
      clearCreateListingDraft(window.localStorage);
    }
    setDraftRestored(false);
    router.push(`/listings/${listing.listingId}?sharePrompt=1`);
  };

  const handleSubmit = async (data: ListingFormValues) => {
    setFormDraft(data);

    if (!isAuthenticated) {
      goToLoginWithDraft({ step: 'form', formValues: data });
      return;
    }

    if (isCreateBlocked || hasCreateAccessError) {
      toast(createAccessState.message || CREATE_LISTING_ACCESS_ERROR_MESSAGE, 'error');
      return;
    }


    try {
      await submitListing(data);
    } catch (error) {
      toast(getCreateListingErrorMessage(error), 'error');
    }
  };

  const handleStartNew = (nextStep: Step) => {
    if (isCreateBlocked || hasCreateAccessError) {
      return;
    }

    if (draftRestored) {
      if (typeof window !== 'undefined') {
        clearCreateListingDraft(window.localStorage);
      }
      setDraftRestored(false);
      setFormDraft(seedDraftPhone(null, savedProfilePhoneRef.current));
      setParseTextDraft('');
      setParsedData(null);
      setParsedLocation(null);
    }

    setStep(nextStep);
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      {(visibleStep !== 'choose' || (draftRestored && !isCreateBlocked)) && (
        <div className="mb-6 flex items-center justify-between gap-3">
          {visibleStep !== 'choose' ? (
            <button
              type="button"
              onClick={() => setStep(visibleStep === 'form' && !parsedData ? 'choose' : visibleStep === 'form' ? 'parse' : 'choose')}
              className="flex items-center gap-2 text-gray-500 hover:text-gray-700 text-sm font-medium"
            >
              <ArrowLeft className="w-4 h-4" />
              Kembali
            </button>
          ) : <div />}

          {draftRestored ? (
            <button
              type="button"
              onClick={discardDraft}
              className="inline-flex items-center gap-2 rounded-xl border border-red-200 px-3 py-2 text-xs font-semibold text-red-500 hover:border-red-300 hover:text-red-600"
            >
              Buang draft
            </button>
          ) : null}
        </div>
      )}

      <div className="mb-8">
        <h1 className="text-2xl font-black text-brand-primary">
          {visibleStep === 'choose' && 'Pasang Iklan Properti'}
          {visibleStep === 'parse' && 'Paste Teks Iklanmu'}
          {visibleStep === 'form' && 'Detail Iklan'}
        </h1>
        <p className="text-gray-500 mt-1">
          {visibleStep === 'choose' && 'Pilih cara memasang iklan yang paling mudah untukmu'}
          {visibleStep === 'parse' && 'Copy paste dari WhatsApp atau chat lain, AI kami yang rapikan'}
          {visibleStep === 'form' && 'Lengkapi informasi propertimu'}
        </p>

        <div className="flex items-center gap-2 mt-4">
          {(['choose', 'parse', 'form'] as Step[]).map((s, i) => (
            <div key={s} className="flex items-center gap-2">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold transition-all ${
                  visibleStep === s
                    ? 'bg-brand-primary text-white'
                    : i < ['choose', 'parse', 'form'].indexOf(visibleStep)
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

      {visibleStep === 'choose' && (
        <div className="space-y-4">
          {isCreateBlocked && (
            <div className="rounded-3xl border border-amber-200 bg-amber-50 p-6">
              <div className="flex items-start gap-4">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-100">
                  <AlertTriangle className="h-5 w-5 text-amber-600" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-semibold text-amber-700">
                    Batas {listingLimit} listing paket {initialTier === 'free' ? 'gratis' : initialTier} sudah terpakai
                  </p>
                  <h2 className="mt-1 text-xl font-bold text-gray-900">
                    {isPremium
                      ? 'Arsipkan salah satu listing untuk memasang iklan baru'
                      : 'Upgrade dulu untuk memasang iklan baru'}
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                     {createAccessState.message}
                  </p>
                  <p className="mt-2 text-sm text-gray-500">
                    Saat ini kamu sudah punya {activeListingsCount} listing aktif di akun {initialTier === 'free' ? 'gratis' : initialTier}.
                  </p>

                  <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                    {!isPremium && (
                      <Link
                        href="/profile#premium"
                        className="inline-flex items-center justify-center gap-2 rounded-2xl bg-brand-gold px-4 py-3 text-sm font-semibold text-white shadow-md transition hover:brightness-105"
                      >
                        <Crown className="h-4 w-4" />
                        Upgrade ke Premium
                      </Link>
                    )}
                    <Link
                      href="/listings"
                      className="inline-flex items-center justify-center rounded-2xl border-2 border-white px-4 py-3 text-sm font-semibold text-gray-700 transition hover:bg-white"
                    >
                      Kembali ke iklan saya
                    </Link>
                  </div>
                </div>
              </div>
            </div>
          )}

          {hasCreateAccessError && (
            <div className="rounded-3xl border border-red-200 bg-red-50 p-6">
              <div className="flex items-start gap-4">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-red-100">
                  <AlertTriangle className="h-5 w-5 text-red-600" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-semibold text-red-700">
                    Slot listing aktif belum bisa dipastikan
                  </p>
                  <h2 className="mt-1 text-xl font-bold text-gray-900">
                    Coba cek lagi sebentar lagi
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                    {createAccessState.message || CREATE_LISTING_ACCESS_ERROR_MESSAGE}
                  </p>

                  <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                    <button
                      type="button"
                      onClick={() => window.location.reload()}
                      className="inline-flex items-center justify-center rounded-2xl bg-white px-4 py-3 text-sm font-semibold text-gray-700 shadow-sm transition hover:bg-gray-100"
                    >
                      Coba lagi
                    </button>
                    <Link
                      href="/listings"
                      className="inline-flex items-center justify-center rounded-2xl border-2 border-white px-4 py-3 text-sm font-semibold text-gray-700 transition hover:bg-white"
                    >
                      Kembali ke iklan saya
                    </Link>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="grid sm:grid-cols-2 gap-4">
            <button
              type="button"
              onClick={() => handleStartNew('parse')}
              disabled={isCreateBlocked || hasCreateAccessError}
              className="card p-6 text-left transition-all duration-300 group border-2 border-transparent enabled:hover:-translate-y-1 enabled:hover:border-[#25D366]/30 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <div className="w-14 h-14 bg-[#25D366] rounded-2xl flex items-center justify-center mb-4 shadow-lg group-hover:scale-110 transition-transform">
                <MessageCircle className="w-7 h-7 text-white" />
              </div>
              <h3 className="font-bold text-gray-900 text-lg mb-2">Paste Teks Iklan</h3>
              <p className="text-gray-500 text-sm leading-relaxed">
                Copy paste teks iklan dari WhatsApp, grup, atau chat. AI kami ekstrak semua detail otomatis.
              </p>
              <div className="mt-4 inline-flex items-center gap-1 text-[#25D366] text-sm font-semibold">
                ⚡ Paling Cepat - 60 detik
              </div>
            </button>

            <button
              type="button"
              onClick={() => handleStartNew('form')}
              disabled={isCreateBlocked || hasCreateAccessError}
              className="card p-6 text-left transition-all duration-300 group border-2 border-transparent enabled:hover:-translate-y-1 enabled:hover:border-brand-accent/30 disabled:cursor-not-allowed disabled:opacity-50"
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
        </div>
      )}

      {visibleStep === 'parse' && (
        <TextParseForm
          onUseParsedResult={handleUseParsedResult}
          onManualFill={() => setStep('form')}
          initialText={parseTextDraft}
          isAuthenticated={isAuthenticated}
          onRequireAuth={handleRequireAuthForParse}
        />
      )}

      {visibleStep === 'form' && (
        <ListingForm
          initialData={parsedData || undefined}
          initialLocation={parsedLocation || undefined}
          initialFormValues={formDraft || undefined}
          onSubmit={handleSubmit}
          isLoading={isPending}
          mode="create"
          isPremium={isPremium}
          maxImages={ImageLimits[initialTier]}
        />
      )}

    </div>
  );
}
