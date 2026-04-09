'use client';

import { signOut } from 'next-auth/react';
import { useEffect, useState } from 'react';
import {
  User,
  Mail,
  Crown,
  Home,
  TrendingUp,
  Settings,
  LogOut,
  Edit2,
  Check,
  Star,
  RefreshCw,
  Loader2,
  MessageCircle,
  ShieldCheck,
  Unplug,
  Plus,
  Search,
} from 'lucide-react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { PremiumUpgradeModal } from '@/components/premium/PremiumUpgradeModal';
import { SubscriptionStatusBadge } from '@/components/premium/SubscriptionStatusBadge';
import Link from 'next/link';
import { formatDate } from '@/lib/utils';
import { getSubscriptionStatus } from '@/lib/subscription-status';
import { shouldShowRenewalCTA } from '@/lib/premium-renewal';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  updateProfile,
  createWhatsAppLinkChallenge,
  disconnectWhatsAppLink,
  getWhatsAppLinkStatus,
} from '@/lib/api';
import {
  getWhatsAppWriteEligibilityCopy,
  normalizeWhatsAppLinkPhone,
} from '@/lib/whatsapp-linking';
import { buildProfileUpdatePayload } from '@/lib/profile-update-payload';
import type { UpdateProfileRequest, User as UserProfile, UserPreferences, SubscriptionTier } from '@/types';

type ProfilePageClientProps = {
  profile: UserProfile;
  sessionUser?: {
    name?: string | null;
    email?: string | null;
    image?: string | null;
  } | null;
};

export function ProfilePageClient({ profile, sessionUser }: ProfilePageClientProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const [selectedUpgradeTier, setSelectedUpgradeTier] = useState<'basic' | 'premium' | 'pro'>('premium');
  const [role, setRole] = useState<'buyer' | 'seller' | 'both' | ''>(profile.role ?? '');
  const [notifications, setNotifications] = useState(profile.preferences?.notifications ?? true);
  const [saved, setSaved] = useState(false);
  const [profilePhoneInput, setProfilePhoneInput] = useState(profile.phone ?? '');
  const [savedProfilePhone, setSavedProfilePhone] = useState(profile.phone ?? '');
  const [waPhoneInput, setWaPhoneInput] = useState(profile.phone ?? '');
  const [challengeId, setChallengeId] = useState('');
  const [waChallengeMessage, setWaChallengeMessage] = useState('');
  const [waChallengeLink, setWaChallengeLink] = useState('');
  const [isAutoCheckingVerification, setIsAutoCheckingVerification] = useState(false);
  const [waSuccessMessage, setWaSuccessMessage] = useState<string | null>(null);
  const returnTo = searchParams.get('returnTo');
  const upgradeTierParam = searchParams.get('upgradeTier');
  const DEFAULT_PREFERENCES: UserPreferences = {
    favoriteLocations: [],
    searchHistory: [],
    notifications: true,
  };

  const subscriptionStatus = getSubscriptionStatus({ authStatus: 'authenticated', profile });
  const tier = profile.subscription.tier;
  const isPaidTier = tier !== 'free';
  const isPremium = subscriptionStatus === 'active' || subscriptionStatus === 'expiring_soon';
  const showRenewalCTA = shouldShowRenewalCTA(subscriptionStatus);
  const activeListingsCount = profile.subscription.activeListingsCount ?? 0;
  const currentPriceLabel =
    tier === 'basic' ? 'Rp 59rb/bulan' : tier === 'premium' ? 'Rp 129rb/bulan' : tier === 'pro' ? 'Rp 199rb/bulan' : 'Rp 0';

  const TIER_ORDER: Record<SubscriptionTier, number> = { free: 0, basic: 1, premium: 2, pro: 3 };

  useEffect(() => {
    if (upgradeTierParam && ['basic', 'premium', 'pro'].includes(upgradeTierParam)) {
      setSelectedUpgradeTier(upgradeTierParam as 'basic' | 'premium' | 'pro');
      setShowPremiumModal(true);
    }
  }, [upgradeTierParam]);

  const displayName = profile.name || sessionUser?.name || '';
  const displayEmail = profile.email || sessionUser?.email || '';
  const displayAvatar = profile.profilePicture || sessionUser?.image || null;
  const initials = displayName.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase();

  const waStatusQuery = useQuery({
    queryKey: ['whatsapp-link-status'],
    queryFn: () => getWhatsAppLinkStatus(),
  });

  const runAutoVerificationCheck = async (expectedPhone: string) => {
    setIsAutoCheckingVerification(true);
    for (let attempt = 0; attempt < 30; attempt++) {
      try {
        const status = await getWhatsAppLinkStatus();
        if (status.eligible && status.linkedPhone === expectedPhone) {
          await waStatusQuery.refetch();
          setChallengeId('');
          setWaChallengeMessage('');
          setWaChallengeLink('');
          setWaSuccessMessage('WhatsApp listing berhasil aktif.');
          setIsAutoCheckingVerification(false);
          return;
        }
      } catch {
        // keep polling; manual check remains available.
      }
      await new Promise((resolve) => setTimeout(resolve, 2000));
    }
    setIsAutoCheckingVerification(false);
    setWaSuccessMessage('Belum terverifikasi otomatis. Kirim pesan challenge persis seperti yang ditampilkan, lalu tekan cek status.');
  };

  const linkChallengeMutation = useMutation({
    mutationFn: (phone: string) => createWhatsAppLinkChallenge(phone),
    onSuccess: (result) => {
      setChallengeId(result.challengeId);
      setWaPhoneInput(result.phone);
      setWaChallengeMessage(result.messageText ?? '');
      setWaChallengeLink(result.messageLink ?? '');
      setWaSuccessMessage(`Challenge siap dikirim. Berlaku sampai ${new Date(result.expiresAt).toLocaleTimeString('id-ID')}.`);
      if (result.messageLink) {
        window.open(result.messageLink, '_blank', 'noopener,noreferrer');
      }
      void runAutoVerificationCheck(result.phone);
    },
  });

  const refreshWaStatusMutation = useMutation({
    mutationFn: () => getWhatsAppLinkStatus(),
    onSuccess: async () => {
      const latest = await waStatusQuery.refetch();
      if (latest.data?.eligible) {
        setChallengeId('');
        setWaChallengeMessage('');
        setWaChallengeLink('');
        setWaSuccessMessage('WhatsApp listing berhasil aktif.');
      } else {
        setWaSuccessMessage('Belum terverifikasi. Kirim pesan challenge persis seperti yang ditampilkan, lalu cek status lagi.');
      }
    },
  });

  const disconnectLinkMutation = useMutation({
    mutationFn: () => disconnectWhatsAppLink(),
    onSuccess: async () => {
      setChallengeId('');
      setWaChallengeMessage('');
      setWaChallengeLink('');
      setWaSuccessMessage('Koneksi WhatsApp berhasil diputus.');
      await waStatusQuery.refetch();
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: UpdateProfileRequest) => updateProfile(payload),
    onSuccess: (updatedProfile) => {
      const nextPhone = updatedProfile.phone ?? '';
      const previousSavedPhone = savedProfilePhone;

      setSavedProfilePhone(nextPhone);
      setProfilePhoneInput(nextPhone);
      if (waPhoneInput === previousSavedPhone) {
        setWaPhoneInput(nextPhone);
      }
      setSaved(true);
      if (returnTo) {
        router.push(returnTo);
      }
    },
  });

  const waStatus = waStatusQuery.data ?? {
    eligible: false,
    isLinked: false,
    linkedPhone: '',
    reason: '',
  };
  const waCopy = getWhatsAppWriteEligibilityCopy(waStatus);
  const isWaBusy =
    waStatusQuery.isLoading ||
    linkChallengeMutation.isPending ||
    refreshWaStatusMutation.isPending ||
    disconnectLinkMutation.isPending ||
    isAutoCheckingVerification;

  const handleActivateWhatsAppListing = async () => {
    setWaSuccessMessage(null);
    const normalizedPhone = normalizeWhatsAppLinkPhone(waPhoneInput);
    await linkChallengeMutation.mutateAsync(normalizedPhone);
  };

  const handleRefreshVerification = async () => {
    setWaSuccessMessage(null);
    await refreshWaStatusMutation.mutateAsync();
  };

  const handleDisconnect = async () => {
    setWaSuccessMessage(null);
    await disconnectLinkMutation.mutateAsync();
  };

  const handleSubmitSettings = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaved(false);
    const payload: UpdateProfileRequest = buildProfileUpdatePayload({
      phone: profilePhoneInput,
      role,
      notifications,
      preferences: profile.preferences ?? DEFAULT_PREFERENCES,
    });
    await updateMutation.mutateAsync(payload);
  };

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-black text-brand-primary mb-8">Profil Saya</h1>

      <div className="card p-6 mb-4">
        <div className="flex items-center gap-5">
          {displayAvatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={displayAvatar}
              alt={displayName}
              className="w-20 h-20 rounded-2xl object-cover shadow-md"
            />
          ) : (
            <div className="w-20 h-20 bg-gradient-hero rounded-2xl flex items-center justify-center text-white text-2xl font-black shadow-md">
              {initials}
            </div>
          )}
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h2 className="text-xl font-bold text-gray-900">{displayName}</h2>
              {isPaidTier && (
                <span className="flex items-center gap-1 bg-amber-50 text-amber-600 text-xs font-semibold px-2 py-0.5 rounded-full border border-amber-200">
                  <Crown className="w-3 h-3" />
                  {tier === 'basic' ? 'Basic' : tier === 'premium' ? 'Premium' : 'Pro'}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1.5 text-gray-500 text-sm">
              <Mail className="w-3.5 h-3.5" />
              {displayEmail}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4 mt-6 pt-6 border-t">
          {[
            {
              icon: Home,
              label: 'Iklan Aktif',
              value: activeListingsCount,
            },
            {
              icon: User,
              label: 'Tipe Akun',
              value: tier === 'free' ? 'Gratis' : tier === 'basic' ? 'Basic' : tier === 'premium' ? 'Premium' : 'Pro',
            },
            {
              icon: TrendingUp,
              label: 'Bergabung',
              value: formatDate(profile.createdAt).split(' ').slice(-1)[0],
            },
          ].map((stat) => (
            <div key={stat.label} className="text-center">
              <div className="w-10 h-10 bg-brand-light rounded-xl flex items-center justify-center mx-auto mb-2">
                <stat.icon className="w-5 h-5 text-brand-primary" />
              </div>
              <div className="font-bold text-gray-900">{stat.value}</div>
              <div className="text-xs text-gray-500">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>

      <div className="card p-6 mb-4" id="premium">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-bold text-gray-900">Paket Berlangganan</h3>
        </div>

        {isPaidTier || subscriptionStatus === 'expired' ? (
          <div className="flex items-center gap-4 bg-amber-50 border border-amber-200 rounded-2xl p-4 mb-4">
            <div className="w-12 h-12 bg-brand-gold rounded-xl flex items-center justify-center">
              <Crown className="w-6 h-6 text-white" />
            </div>
            <div className="flex-1">
              <p className="font-bold text-amber-700">
                Propti {tier === 'basic' ? 'Basic' : tier === 'premium' ? 'Premium' : tier === 'pro' ? 'Pro' : 'Premium'}
              </p>
              <SubscriptionStatusBadge
                status={subscriptionStatus}
                renewDate={profile.subscription.renewDate}
                tier={tier}
                className="mt-1 text-xs"
              />
            </div>
            {isPremium && <Check className="w-5 h-5 text-amber-500" />}
          </div>
        ) : (
          <div className="bg-gray-50 rounded-2xl p-4 mb-4">
            <div className="flex items-center justify-between mb-3">
              <span className="font-semibold text-gray-700">Paket Gratis</span>
              <span className="badge bg-gray-100 text-gray-500">Aktif</span>
            </div>
            <ul className="space-y-2 text-sm text-gray-600">
             {['3 iklan pertama gratis', 'Maksimal 3 foto per iklan', 'Iklan tayang 30 hari'].map((f) => (
                <li key={f} className="flex items-center gap-2">
                  <Check className="w-3.5 h-3.5 text-gray-400" />
                  {f}
                </li>
              ))}
            </ul>
          </div>
        )}

        <p className="text-sm font-semibold text-gray-700 mb-3">
          {isPaidTier ? 'Ganti atau perpanjang paket:' : 'Pilih paket upgrade:'}
        </p>
        <div className="grid grid-cols-3 gap-3 mb-4">
          {(
            [
              { key: 'basic', label: 'Basic', price: 'Rp 59rb/bln', highlight: false },
              { key: 'premium', label: 'Premium', price: 'Rp 129rb/bln', highlight: true },
              { key: 'pro', label: 'Pro', price: 'Rp 199rb/bln', highlight: false },
            ] as const
          ).map((plan) => {
            const isCurrent = tier === plan.key;
            const isUpgrade = TIER_ORDER[plan.key] > TIER_ORDER[tier];
            const actionLabel = isCurrent
              ? (showRenewalCTA ? 'Perpanjang' : 'Paket Aktif')
              : isUpgrade
                ? 'Upgrade'
                : 'Ganti';
            const isDisabled = isCurrent && !showRenewalCTA;
            return (
              <button
                key={plan.key}
                type="button"
                disabled={isDisabled}
                onClick={() => {
                  setSelectedUpgradeTier(plan.key);
                  setShowPremiumModal(true);
                }}
                className={`flex flex-col items-center p-3 rounded-2xl border-2 transition-all text-center ${
                  isCurrent
                    ? 'border-amber-300 bg-amber-50 text-amber-700'
                    : plan.highlight
                      ? 'border-brand-gold bg-amber-50 text-amber-700'
                      : 'border-gray-200 bg-white text-gray-700 hover:border-brand-gold hover:bg-amber-50'
                } ${isDisabled ? 'opacity-60 cursor-default' : ''}`}
              >
                <Crown className={`w-5 h-5 mb-1 ${isCurrent || plan.highlight ? 'text-brand-gold' : 'text-gray-400'}`} />
                <span className="text-xs font-bold">{plan.label}</span>
                <span className="text-xs text-gray-500 mt-0.5">{plan.price}</span>
                <span className={`mt-1 text-[10px] font-semibold px-1.5 py-0.5 rounded-full ${
                  isCurrent
                    ? 'bg-amber-200 text-amber-800'
                    : isUpgrade
                      ? 'bg-green-100 text-green-700'
                      : 'bg-gray-100 text-gray-600'
                }`}>
                  {actionLabel}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      <div className="card p-2 mb-4">
        {[
          { icon: Home, label: 'Iklan Saya', href: '/listings' },
          { icon: Star, label: 'Properti Tersimpan', href: '/saved' },
        ].map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 px-4 py-3.5 rounded-xl hover:bg-gray-50 transition-colors group"
          >
            <div className="w-8 h-8 bg-brand-light rounded-lg flex items-center justify-center">
              <item.icon className="w-4 h-4 text-brand-primary" />
            </div>
            <span className="font-medium text-gray-700 group-hover:text-gray-900 flex-1">
              {item.label}
            </span>
            <Edit2 className="w-4 h-4 text-gray-300 group-hover:text-gray-400" />
          </Link>
        ))}
      </div>

      <div className="card p-6 mb-4">
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Settings className="w-5 h-5 text-brand-primary" />
            <h2 className="font-bold text-gray-900">Pengaturan Akun</h2>
          </div>
          <p className="text-gray-500 text-sm">Kelola data profil dan preferensi akun Propti kamu.</p>
        </div>

        {returnTo && (
          <div className="mb-6 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
            Kamu bisa lanjut pasang iklan web seperti biasa. Aktivasi WhatsApp listing bersifat opsional
            dan butuh challenge verifikasi nomor.
          </div>
        )}

        <form onSubmit={handleSubmitSettings} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="block">
              <span className="text-sm font-medium text-gray-700">Nama</span>
              <input value={displayName} disabled className="input-field mt-1 bg-gray-50 text-gray-500" />
            </label>

            <label className="block">
              <span className="text-sm font-medium text-gray-700">Email</span>
              <input value={displayEmail} disabled className="input-field mt-1 bg-gray-50 text-gray-500" />
            </label>

            <label className="block md:col-span-2">
              <span className="text-sm font-medium text-gray-700">Nomor Telepon</span>
              <input
                value={profilePhoneInput}
                onChange={(event) => setProfilePhoneInput(event.target.value)}
                placeholder="Contoh: 081234567890"
                className="input-field mt-1"
              />
              <span className="mt-1 block text-xs text-gray-500">
                Simpan nomor ini untuk kontak listing web. Aktivasi WhatsApp listing tetap opsional di bagian bawah.
              </span>
            </label>

            <label className="block md:col-span-2">
              <span className="text-sm font-medium text-gray-700">Peran Akun</span>
              <select
                value={role}
                onChange={(event) => setRole(event.target.value as 'buyer' | 'seller' | 'both' | '')}
                className="input-field mt-1"
              >
                {role === '' && <option value="">Pilih peran akun (opsional)</option>}
                <option value="buyer">Pencari Properti</option>
                <option value="seller">Penjual / Agen</option>
                <option value="both">Keduanya</option>
              </select>
            </label>
          </div>

          <label className="flex items-start gap-3 rounded-2xl border border-gray-100 p-4">
            <input
              type="checkbox"
              checked={notifications}
              onChange={(event) => setNotifications(event.target.checked)}
              className="mt-1 h-4 w-4 rounded border-gray-300 text-brand-primary focus:ring-brand-primary"
            />
            <span>
              <span className="block font-semibold text-gray-900">Aktifkan notifikasi</span>
              <span className="text-sm text-gray-500">
                Dapatkan pembaruan penting untuk aktivitas akun dan listing.
              </span>
            </span>
          </label>

          {updateMutation.isError && (
            <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
              {(updateMutation.error as Error).message || 'Gagal menyimpan pengaturan.'}
            </div>
          )}

          {saved && !updateMutation.isPending && (
            <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              Pengaturan berhasil diperbarui.
            </div>
          )}

          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="btn-primary w-full inline-flex items-center justify-center gap-2 disabled:opacity-60"
          >
            {updateMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Check className="w-4 h-4" />}
            Simpan Pengaturan
          </button>
        </form>
      </div>

      <div className="card p-6 mb-4">
        <div className="flex items-center gap-2 mb-4">
          <MessageCircle className="w-5 h-5 text-brand-primary" />
          <h3 className="font-bold text-gray-900">Hubungkan WhatsApp</h3>
        </div>
        <div className={`rounded-2xl border px-4 py-3 mb-4 ${waCopy.tone === 'success' ? 'border-emerald-200 bg-emerald-50' : 'border-amber-200 bg-amber-50'}`}>
          <p className={`text-sm font-semibold ${waCopy.tone === 'success' ? 'text-emerald-700' : 'text-amber-700'}`}>Status WhatsApp</p>
          <p className={`mt-1 text-sm ${waCopy.tone === 'success' ? 'text-emerald-700' : 'text-amber-700'}`}>{waCopy.title}</p>
          <p className={`mt-1 text-xs ${waCopy.tone === 'success' ? 'text-emerald-700' : 'text-amber-700'}`}>{waCopy.description}</p>
          {waStatus.linkedPhone && (
            <p className="mt-2 text-xs text-gray-600">Nomor terhubung: +{waStatus.linkedPhone}</p>
          )}
        </div>

        <label className="block mb-3">
          <span className="text-sm font-medium text-gray-700">Nomor WhatsApp</span>
          <input
            value={waPhoneInput}
            onChange={(event) => setWaPhoneInput(event.target.value)}
            placeholder="Contoh: 081234567890"
            className="input-field mt-1"
            disabled={isWaBusy}
          />
        </label>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <button
            type="button"
            onClick={() => void handleActivateWhatsAppListing()}
            disabled={isWaBusy || normalizeWhatsAppLinkPhone(waPhoneInput) === ''}
            className="btn-primary inline-flex items-center justify-center gap-2 disabled:opacity-60"
          >
            {linkChallengeMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <ShieldCheck className="w-4 h-4" />}
            Aktifkan WhatsApp Listing
          </button>

          <button
            type="button"
            onClick={() => void handleDisconnect()}
            disabled={isWaBusy || !waStatus.isLinked}
            className="inline-flex items-center justify-center gap-2 rounded-xl border border-red-200 bg-red-50 py-3 px-4 text-sm font-semibold text-red-600 hover:bg-red-100 disabled:opacity-60"
          >
            {disconnectLinkMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Unplug className="w-4 h-4" />}
            Putuskan WhatsApp
          </button>
        </div>

        {challengeId && (
          <div className="mt-4 p-4 rounded-2xl border border-gray-100 bg-gray-50">
            <p className="text-sm font-semibold text-gray-900 mb-2">Verifikasi lewat WhatsApp</p>
            <p className="text-xs text-gray-600 mb-2">Kirim pesan ini tanpa diubah dari nomor WhatsApp yang kamu daftarkan:</p>
            <div className="rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm font-semibold text-gray-900 mb-3">
              {waChallengeMessage}
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <a
                href={waChallengeLink || '#'}
                target="_blank"
                rel="noreferrer"
                className={`btn-primary inline-flex items-center justify-center gap-2 ${waChallengeLink ? '' : 'pointer-events-none opacity-60'}`}
              >
                Kirim Pesan di WhatsApp
              </a>
              <button
                type="button"
                onClick={() => void handleRefreshVerification()}
                disabled={isWaBusy}
                className="w-full btn-primary inline-flex items-center justify-center gap-2 disabled:opacity-60"
              >
                {refreshWaStatusMutation.isPending || isAutoCheckingVerification ? <Loader2 className="w-4 h-4 animate-spin" /> : <Check className="w-4 h-4" />}
                Cek Status Verifikasi
              </button>
            </div>
          </div>
        )}

        {(waStatusQuery.isError || linkChallengeMutation.isError || refreshWaStatusMutation.isError || disconnectLinkMutation.isError) && (
          <div className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
            {(waStatusQuery.error as Error)?.message ||
              (linkChallengeMutation.error as Error)?.message ||
              (refreshWaStatusMutation.error as Error)?.message ||
              (disconnectLinkMutation.error as Error)?.message ||
              'Terjadi kendala saat mengelola koneksi WhatsApp.'}
          </div>
        )}

        {waSuccessMessage && (
          <div className="mt-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
            {waSuccessMessage}
          </div>
        )}

        {waStatus.eligible && (
          <div className="mt-4 rounded-2xl border border-blue-100 bg-blue-50 px-4 py-4">
            <p className="text-sm font-bold text-blue-800 mb-3">Cara pakai fitur WhatsApp</p>
            <div className="space-y-3">
              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 w-7 h-7 rounded-full bg-blue-100 flex items-center justify-center">
                  <Plus className="w-3.5 h-3.5 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-blue-800">Pasang Iklan</p>
                  <p className="text-xs text-blue-600 mt-0.5">Ketik atau kirim voice note deskripsi propertimu langsung ke WhatsApp Propti. AI kami otomatis buatkan draftnya.</p>
                  <p className="text-xs font-mono bg-blue-100 text-blue-700 rounded-lg px-2 py-1 mt-1 inline-block">Contoh: &quot;jual rumah 2 lantai di Ciputat harga 750jt&quot;</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 w-7 h-7 rounded-full bg-blue-100 flex items-center justify-center">
                  <Search className="w-3.5 h-3.5 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-blue-800">Cari Properti</p>
                  <p className="text-xs text-blue-600 mt-0.5">Awali pesan dengan <span className="font-semibold">cari</span> lalu tuliskan kebutuhanmu. Bisa pakai teks atau voice note.</p>
                  <p className="text-xs font-mono bg-blue-100 text-blue-700 rounded-lg px-2 py-1 mt-1 inline-block">Contoh: &quot;cari rumah 3 kamar dekat sekolah di Depok&quot;</p>
                </div>
              </div>
            </div>
            {process.env.NEXT_PUBLIC_WHATSAPP_NUMBER && (
              <a
                href={`https://wa.me/${process.env.NEXT_PUBLIC_WHATSAPP_NUMBER}`}
                target="_blank"
                rel="noreferrer"
                className="mt-4 inline-flex items-center gap-2 text-xs font-semibold text-blue-700 hover:text-blue-900"
              >
                <MessageCircle className="w-3.5 h-3.5" />
                Buka WhatsApp Propti
              </a>
            )}
          </div>
        )}
      </div>

      <button
        type="button"
        onClick={() => signOut({ callbackUrl: '/' })}
        className="w-full flex items-center justify-center gap-2 border-2 border-red-100 text-red-500 font-semibold py-3 px-4 rounded-2xl hover:bg-red-50 transition-colors"
      >
        <LogOut className="w-4 h-4" />
        Keluar dari Propti
      </button>

      <PremiumUpgradeModal
        isOpen={showPremiumModal}
        onClose={() => setShowPremiumModal(false)}
        mode={showRenewalCTA && selectedUpgradeTier === tier ? 'renew' : 'upgrade'}
        currentRenewDate={profile.subscription.renewDate}
        currentTier={tier}
        selectedTier={selectedUpgradeTier}
      />
    </div>
  );
}
