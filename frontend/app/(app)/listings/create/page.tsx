'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { MessageCircle, PenLine, ArrowLeft, Phone, X } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { TextParseForm } from '@/components/listings/TextParseForm';
import { ListingForm } from '@/components/listings/ListingForm';
import { useCreateListing } from '@/hooks/useListings';
import { useAuth } from '@/hooks/useAuth';
import type { ParsedListing, CreateListingRequest, Location, User } from '@/types';
import type { ListingFormValues } from '@/components/listings/ListingForm';
import {
  clearCreateListingDraft,
  loadCreateListingDraft,
  saveCreateListingDraft,
} from '@/lib/create-listing-draft';
import { useToast } from '@/app/toaster';
import { getCreateListingErrorMessage } from '@/lib/create-listing-errors';
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
  const { isPremium, isAuthenticated, profile } = useAuth();
  const { toast } = useToast();

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

  const handleParsed = (result: ParsedListing) => {
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
      {(step !== 'choose' || draftRestored) && (
        <div className="mb-6 flex items-center justify-between gap-3">
          {step !== 'choose' ? (
            <button
              onClick={() => setStep(step === 'form' && !parsedData ? 'choose' : step === 'form' ? 'parse' : 'choose')}
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
            onClick={() => handleStartNew('parse')}
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
            onClick={() => handleStartNew('form')}
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
          initialText={parseTextDraft}
          isAuthenticated={isAuthenticated}
          onRequireAuth={handleRequireAuthForParse}
        />
      )}

      {/* Step: Form */}
      {step === 'form' && (
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
