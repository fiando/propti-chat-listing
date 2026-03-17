'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { MessageCircle, PenLine, ArrowLeft, Phone, X, AlertTriangle, Crown, Loader2 } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { TextParseForm } from '@/components/listings/TextParseForm';
import { ListingForm } from '@/components/listings/ListingForm';
import { useCreateListing, useMyListingQuotaSummary } from '@/hooks/useListings';
import { useAuth } from '@/hooks/useAuth';
import type { ParsedListing, CreateListingRequest, Location, User } from '@/types';
import type { ListingFormValues } from '@/components/listings/ListingForm';
import {
  clearCreateListingDraft,
  loadCreateListingDraft,
  saveCreateListingDraft,
} from '@/lib/create-listing-draft';
import { useToast } from '@/app/toaster';
import {
  CREATE_LISTING_ACCESS_ERROR_MESSAGE,
  FREE_TIER_LISTING_LIMIT,
  getCreateListingAccessState,
  getCreateListingErrorMessage,
} from '@/lib/create-listing-errors';
import { normalizeAmenityIds } from '@/lib/listing-form-utils';
import { updateProfile } from '@/lib/api';
import { getPhoneModalSubmitLabel, shouldRequirePhone } from '@/lib/create-listing-phone';

type Step = 'choose' | 'parse' | 'form';

export default function CreateListingPage() {
  const router = useRouter();
  const [step, setStep] = useState<Step>('choose');
  const [parsedData, setParsedData] = useState<ParsedListing | null>(null);
  const [parsedLocation, setParsedLocation] = useState<Partial<Location> | null>(null);
  const [parseTextDraft, setParseTextDraft] = useState('');
  const [formDraft, setFormDraft] = useState<Partial<ListingFormValues> | null>(null);
  const [draftRestored, setDraftRestored] = useState(false);
  const [showPhoneModal, setShowPhoneModal] = useState(false);
  const [pendingSubmitData, setPendingSubmitData] = useState<ListingFormValues | null>(null);
  const [phoneInput, setPhoneInput] = useState('');
  const [isSavingPhone, setIsSavingPhone] = useState(false);
  const [isSubmittingListingFromPhoneModal, setIsSubmittingListingFromPhoneModal] = useState(false);
  const queryClient = useQueryClient();
  const { mutateAsync: createListing, isPending } = useCreateListing();
  const { isPremium, isAuthenticated, isLoading: isAuthLoading, profile } = useAuth();
  const {
    data: listingQuotaSummary,
    isLoading: isListingQuotaLoading,
    isError: hasListingQuotaError,
    refetch: refetchListingQuotaSummary,
  } = useMyListingQuotaSummary(
    {
      enabled: isAuthenticated && !isPremium,
      userId: profile?.userId ?? null,
      activeLimit: FREE_TIER_LISTING_LIMIT,
    }
  );
  const { toast } = useToast();
  const createAccessState = getCreateListingAccessState({
    isAuthenticated,
    isPremium,
    isAuthLoading,
    isListingsLoading: isListingQuotaLoading,
    hasListingsError: hasListingQuotaError,
    activeListingsCount: listingQuotaSummary?.activeListingsCount,
  });
  const activeListingsCount = createAccessState.activeListingsCount;
  const isCheckingCreateAccess = createAccessState.status === 'checking';
  const hasCreateAccessError = createAccessState.status === 'error';
  const isCreateBlocked = createAccessState.status === 'blocked';
  const visibleStep: Step =
    isCheckingCreateAccess || hasCreateAccessError || isCreateBlocked ? 'choose' : step;

  useEffect(() => {
    if (!isCreateBlocked && !hasCreateAccessError) {
      return;
    }

    setStep('choose');
    setShowPhoneModal(false);
    setPendingSubmitData(null);
  }, [hasCreateAccessError, isCreateBlocked]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    const draft = loadCreateListingDraft(window.localStorage);
    if (!draft) {
      return;
    }

    // Auto-apply draft immediately — user will see "Buang draft" button on the form
    applyDraft(draft);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const applyDraft = (draft: NonNullable<ReturnType<typeof loadCreateListingDraft>>) => {
    setStep(draft.step);
    setParseTextDraft(draft.parseText || '');
    setParsedData(draft.parsedData || null);
    setParsedLocation(draft.parsedLocation || null);
    setFormDraft((draft.formValues as Partial<ListingFormValues> | undefined) || null);
    setDraftRestored(true);
  };

  const discardDraft = () => {
    if (typeof window !== 'undefined') {
      clearCreateListingDraft(window.localStorage);
    }

    setDraftRestored(false);
    setParseTextDraft('');
    setParsedData(null);
    setParsedLocation(null);
    setFormDraft(null);
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

    saveCreateListingDraft(window.localStorage, {
      step: draft.step,
      parseText: draft.parseText ?? parseTextDraft,
      parsedData: parsedData ?? undefined,
      parsedLocation: parsedLocation ?? undefined,
      formValues: draft.formValues ?? formDraft ?? undefined,
    });
  };

  const goToLoginWithDraft = (draft: { step: Step; parseText?: string; formValues?: Partial<ListingFormValues> }) => {
    persistDraft(draft);
    router.push(`/login?callbackUrl=${encodeURIComponent('/listings/create')}`);
  };

  const handleUseParsedResult = (result: ParsedListing) => {
    setParsedData(result);
    setFormDraft(null);
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
      images: data.images,
    };

    const listing = await createListing(payload);
    if (typeof window !== 'undefined') {
      clearCreateListingDraft(window.localStorage);
    }
    setDraftRestored(false);
    router.push(`/listings/${listing.listingId}`);
  };

  const handleSubmit = async (
    data: ListingFormValues,
    options?: { phoneOverride?: string }
  ) => {
    setFormDraft(data);

    if (!isAuthenticated) {
      goToLoginWithDraft({ step: 'form', formValues: data });
      return;
    }

    if (isCheckingCreateAccess || isCreateBlocked || hasCreateAccessError) {
      toast(createAccessState.message || CREATE_LISTING_ACCESS_ERROR_MESSAGE, 'error');
      return;
    }

    if (
      shouldRequirePhone({
        profilePhone: profile?.phone,
        phoneOverride: options?.phoneOverride,
      })
    ) {
      setPendingSubmitData(data);
      setShowPhoneModal(true);
      return;
    }

    try {
      await submitListing(data);
    } catch (error) {
      toast(getCreateListingErrorMessage(error), 'error');
    }
  };

  const handleStartNew = (nextStep: Step) => {
    if (isCreateBlocked || isCheckingCreateAccess || hasCreateAccessError) {
      return;
    }

    if (draftRestored) {
      if (typeof window !== 'undefined') {
        clearCreateListingDraft(window.localStorage);
      }
      setDraftRestored(false);
      setParseTextDraft('');
      setParsedData(null);
      setParsedLocation(null);
      setFormDraft(null);
    }

    setStep(nextStep);
  };

  const handlePhoneSubmit = async () => {
    const trimmedPhone = phoneInput.trim();

    if (!trimmedPhone || !pendingSubmitData) return;

    if (isCheckingCreateAccess || isCreateBlocked || hasCreateAccessError) {
      setShowPhoneModal(false);
      setPendingSubmitData(null);
      toast(createAccessState.message || CREATE_LISTING_ACCESS_ERROR_MESSAGE, 'error');
      return;
    }

    setIsSavingPhone(true);
    try {
      await updateProfile({ phone: trimmedPhone });
      queryClient.setQueryData<User | undefined>(['profile'], (current) =>
        current ? { ...current, phone: trimmedPhone } : current
      );
    } catch {
      toast('Gagal menyimpan nomor telepon. Coba lagi.', 'error');
      setIsSavingPhone(false);
      return;
    }

    setIsSavingPhone(false);
    setIsSubmittingListingFromPhoneModal(true);

    try {
      await submitListing(pendingSubmitData);
      setPendingSubmitData(null);
      void queryClient.invalidateQueries({ queryKey: ['profile'] });
    } catch (error) {
      toast(getCreateListingErrorMessage(error), 'error');
    } finally {
      setIsSubmittingListingFromPhoneModal(false);
    }
  };

  const isPhoneModalBusy = isSavingPhone || isSubmittingListingFromPhoneModal;

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      {/* Back button */}
      {(visibleStep !== 'choose' || (draftRestored && !isCreateBlocked)) && (
        <div className="mb-6 flex items-center justify-between gap-3">
          {visibleStep !== 'choose' ? (
            <button
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

      {/* Page header */}
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

        {/* Step indicator */}
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


      {/* Step: Choose method */}
      {visibleStep === 'choose' && (
        <div className="space-y-4">
          {isCheckingCreateAccess && (
            <div className="card flex items-center gap-3 p-4 text-sm text-gray-600">
              <Loader2 className="h-4 w-4 animate-spin text-brand-primary" />
              Memeriksa slot listing gratis yang masih tersedia...
            </div>
          )}

          {isCreateBlocked && (
            <div className="rounded-3xl border border-amber-200 bg-amber-50 p-6">
              <div className="flex items-start gap-4">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-100">
                  <AlertTriangle className="h-5 w-5 text-amber-600" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-semibold text-amber-700">
                    Batas {FREE_TIER_LISTING_LIMIT} listing gratis sudah terpakai
                  </p>
                  <h2 className="mt-1 text-xl font-bold text-gray-900">
                    Upgrade dulu untuk memasang iklan baru
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                     {createAccessState.message}
                  </p>
                  <p className="mt-2 text-sm text-gray-500">
                    Saat ini kamu sudah punya {activeListingsCount} listing aktif di akun gratis.
                  </p>

                  <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                    <Link
                      href="/profile#premium"
                      className="inline-flex items-center justify-center gap-2 rounded-2xl bg-brand-gold px-4 py-3 text-sm font-semibold text-white shadow-md transition hover:brightness-105"
                    >
                      <Crown className="h-4 w-4" />
                      Upgrade ke Premium
                    </Link>
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
                       onClick={() => void refetchListingQuotaSummary()}
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
            onClick={() => handleStartNew('parse')}
            disabled={isCreateBlocked || isCheckingCreateAccess || hasCreateAccessError}
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
              ⚡ Paling Cepat — 60 detik
            </div>
          </button>

          <button
            onClick={() => handleStartNew('form')}
            disabled={isCreateBlocked || isCheckingCreateAccess || hasCreateAccessError}
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

      {/* Step: AI parse */}
      {visibleStep === 'parse' && (
        <TextParseForm
          onUseParsedResult={handleUseParsedResult}
          onManualFill={() => setStep('form')}
          initialText={parseTextDraft}
          isAuthenticated={isAuthenticated}
          onRequireAuth={handleRequireAuthForParse}
        />
      )}

      {/* Step: Form */}
      {visibleStep === 'form' && (
        <ListingForm
          initialData={parsedData || undefined}
          initialLocation={parsedLocation || undefined}
          initialFormValues={formDraft || undefined}
          onSubmit={handleSubmit}
          isLoading={isPending}
          mode="create"
          isPremium={isPremium}
        />
      )}

      {/* Phone number modal */}
      {showPhoneModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl">
            <div className="mb-4 flex items-start justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-light">
                  <Phone className="h-5 w-5 text-brand-primary" />
                </div>
                <div>
                  <h2 className="font-bold text-gray-900">Tambah Nomor Telepon</h2>
                  <p className="text-xs text-gray-500">Agar pembeli bisa menghubungimu</p>
                </div>
              </div>
              <button
                onClick={() => setShowPhoneModal(false)}
                disabled={isPhoneModalBusy}
                className="text-gray-400 hover:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <input
              type="tel"
              value={phoneInput}
              onChange={(e) => setPhoneInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !isPhoneModalBusy && void handlePhoneSubmit()}
              placeholder="08xxxxxxxxxx"
              className="input-field mb-4"
              autoFocus
              disabled={isPhoneModalBusy}
            />

            <button
              onClick={handlePhoneSubmit}
              disabled={!phoneInput.trim() || isPhoneModalBusy}
              className="btn-primary w-full disabled:opacity-50"
            >
              {getPhoneModalSubmitLabel({
                isSavingPhone,
                isSubmittingListing: isSubmittingListingFromPhoneModal,
              })}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
