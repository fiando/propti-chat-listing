'use client';

import { signOut } from 'next-auth/react';
import { useState } from 'react';
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
} from 'lucide-react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { PremiumUpgradeModal } from '@/components/premium/PremiumUpgradeModal';
import { SubscriptionStatusBadge } from '@/components/premium/SubscriptionStatusBadge';
import Link from 'next/link';
import { formatDate } from '@/lib/utils';
import { getSubscriptionStatus } from '@/lib/subscription-status';
import { shouldShowRenewalCTA } from '@/lib/premium-renewal';
import {
  createWhatsAppLinkChallenge,
  disconnectWhatsAppLink,
  getWhatsAppLinkStatus,
  verifyWhatsAppLink,
} from '@/lib/api';
import {
  getWhatsAppWriteEligibilityCopy,
  normalizeWhatsAppLinkPhone,
} from '@/lib/whatsapp-linking';
import type { User as UserProfile } from '@/types';

type ProfilePageClientProps = {
  profile: UserProfile;
  sessionUser?: {
    name?: string | null;
    email?: string | null;
    image?: string | null;
  } | null;
};

export function ProfilePageClient({ profile, sessionUser }: ProfilePageClientProps) {
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const [selectedUpgradeTier, setSelectedUpgradeTier] = useState<'basic' | 'premium' | 'pro'>('premium');
  const [waPhoneInput, setWaPhoneInput] = useState(profile.phone ?? '');
  const [otpCode, setOtpCode] = useState('');
  const [challengeId, setChallengeId] = useState('');
  const [waSuccessMessage, setWaSuccessMessage] = useState<string | null>(null);

  const subscriptionStatus = getSubscriptionStatus({ authStatus: 'authenticated', profile });
  const tier = profile.subscription.tier;
  const isPaidTier = tier !== 'free';
  const isPremium = subscriptionStatus === 'active' || subscriptionStatus === 'expiring_soon';
  const showRenewalCTA = shouldShowRenewalCTA(subscriptionStatus);
  const activeListingsCount = profile.subscription.activeListingsCount ?? 0;
  const currentPriceLabel =
    tier === 'basic' ? 'Rp 59rb/bulan' : tier === 'premium' ? 'Rp 129rb/bulan' : tier === 'pro' ? 'Rp 199rb/bulan' : 'Rp 0';

  const displayName = profile.name || sessionUser?.name || '';
  const displayEmail = profile.email || sessionUser?.email || '';
  const displayAvatar = profile.profilePicture || sessionUser?.image || null;
  const initials = displayName.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase();

  const waStatusQuery = useQuery({
    queryKey: ['whatsapp-link-status'],
    queryFn: () => getWhatsAppLinkStatus(),
  });

  const linkChallengeMutation = useMutation({
    mutationFn: (phone: string) => createWhatsAppLinkChallenge(phone),
    onSuccess: (result) => {
      setChallengeId(result.challengeId);
      setWaPhoneInput(result.phone);
      setWaSuccessMessage(`OTP dikirim. Berlaku sampai ${new Date(result.expiresAt).toLocaleTimeString('id-ID')}.`);
    },
  });

  const verifyLinkMutation = useMutation({
    mutationFn: () => verifyWhatsAppLink(challengeId, otpCode),
    onSuccess: async () => {
      setOtpCode('');
      setChallengeId('');
      setWaSuccessMessage('WhatsApp berhasil terverifikasi.');
      await waStatusQuery.refetch();
    },
  });

  const disconnectLinkMutation = useMutation({
    mutationFn: () => disconnectWhatsAppLink(),
    onSuccess: async () => {
      setChallengeId('');
      setOtpCode('');
      setWaSuccessMessage('Koneksi WhatsApp berhasil diputus.');
      await waStatusQuery.refetch();
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
    verifyLinkMutation.isPending ||
    disconnectLinkMutation.isPending;
  const canVerify = challengeId.trim() !== '' && otpCode.trim().length === 6;

  const handleRequestOtp = async () => {
    setWaSuccessMessage(null);
    const normalizedPhone = normalizeWhatsAppLinkPhone(waPhoneInput);
    await linkChallengeMutation.mutateAsync(normalizedPhone);
  };

  const handleVerifyOtp = async () => {
    setWaSuccessMessage(null);
    await verifyLinkMutation.mutateAsync();
  };

  const handleDisconnect = async () => {
    setWaSuccessMessage(null);
    await disconnectLinkMutation.mutateAsync();
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
          <div>
            <div className="flex items-center gap-4 bg-amber-50 border border-amber-200 rounded-2xl p-4 mb-3">
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

            {showRenewalCTA && (
                <button
                type="button"
                onClick={() => setShowPremiumModal(true)}
                className="w-full btn-gold flex items-center justify-center gap-2"
              >
                <RefreshCw className="w-4 h-4" />
                Perpanjang Paket - {currentPriceLabel}
              </button>
            )}
          </div>
        ) : (
          <div>
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

            <p className="text-sm font-semibold text-gray-700 mb-3">Pilih paket upgrade:</p>
            <div className="grid grid-cols-3 gap-3 mb-4">
              {(
                [
                  { key: 'basic', label: 'Basic', price: 'Rp 59rb/bln', highlight: false },
                  { key: 'premium', label: 'Premium', price: 'Rp 129rb/bln', highlight: true },
                  { key: 'pro', label: 'Pro', price: 'Rp 199rb/bln', highlight: false },
                ] as const
              ).map((plan) => (
                <button
                  key={plan.key}
                  type="button"
                  onClick={() => {
                    setSelectedUpgradeTier(plan.key);
                    setShowPremiumModal(true);
                  }}
                  className={`flex flex-col items-center p-3 rounded-2xl border-2 transition-all text-center ${
                    plan.highlight
                      ? 'border-brand-gold bg-amber-50 text-amber-700'
                      : 'border-gray-200 bg-white text-gray-700 hover:border-brand-gold hover:bg-amber-50'
                  }`}
                >
                  <Crown className={`w-5 h-5 mb-1 ${plan.highlight ? 'text-brand-gold' : 'text-gray-400'}`} />
                  <span className="text-xs font-bold">{plan.label}</span>
                  <span className="text-xs text-gray-500 mt-0.5">{plan.price}</span>
                  {plan.highlight && (
                    <span className="mt-1 text-[10px] font-semibold bg-brand-gold text-white px-1.5 py-0.5 rounded-full">
                      Populer
                    </span>
                  )}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="card p-2 mb-4">
        {[
          { icon: Home, label: 'Iklan Saya', href: '/listings' },
          { icon: Star, label: 'Properti Tersimpan', href: '/saved' },
          { icon: Settings, label: 'Pengaturan', href: '/settings' },
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
        <div className="flex items-center gap-2 mb-4">
          <MessageCircle className="w-5 h-5 text-brand-primary" />
          <h3 className="font-bold text-gray-900">Hubungkan WhatsApp</h3>
        </div>
        <div className={`rounded-2xl border px-4 py-3 mb-4 ${waCopy.tone === 'success' ? 'border-emerald-200 bg-emerald-50' : 'border-amber-200 bg-amber-50'}`}>
          <p className={`text-sm font-semibold ${waCopy.tone === 'success' ? 'text-emerald-700' : 'text-amber-700'}`}>Status WA Write</p>
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
            onClick={() => void handleRequestOtp()}
            disabled={isWaBusy || normalizeWhatsAppLinkPhone(waPhoneInput) === ''}
            className="btn-primary inline-flex items-center justify-center gap-2 disabled:opacity-60"
          >
            {linkChallengeMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <ShieldCheck className="w-4 h-4" />}
            Minta OTP
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

        <div className="mt-4 p-4 rounded-2xl border border-gray-100 bg-gray-50">
          <p className="text-sm font-semibold text-gray-900 mb-2">Verifikasi OTP</p>
          <input
            value={otpCode}
            onChange={(event) => setOtpCode(event.target.value.replace(/\D+/g, '').slice(0, 6))}
            inputMode="numeric"
            placeholder="Masukkan 6 digit OTP"
            className="input-field mb-3"
            disabled={isWaBusy}
          />
          <button
            type="button"
            onClick={() => void handleVerifyOtp()}
            disabled={isWaBusy || !canVerify}
            className="w-full btn-primary inline-flex items-center justify-center gap-2 disabled:opacity-60"
          >
            {verifyLinkMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Check className="w-4 h-4" />}
            Verifikasi OTP
          </button>
        </div>

        {(waStatusQuery.isError || linkChallengeMutation.isError || verifyLinkMutation.isError || disconnectLinkMutation.isError) && (
          <div className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
            {(waStatusQuery.error as Error)?.message ||
              (linkChallengeMutation.error as Error)?.message ||
              (verifyLinkMutation.error as Error)?.message ||
              (disconnectLinkMutation.error as Error)?.message ||
              'Terjadi kendala saat mengelola koneksi WhatsApp.'}
          </div>
        )}

        {waSuccessMessage && (
          <div className="mt-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
            {waSuccessMessage}
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
        mode={showRenewalCTA ? 'renew' : 'upgrade'}
        currentRenewDate={profile.subscription.renewDate}
        selectedTier={showRenewalCTA ? (tier === 'free' ? 'basic' : tier) : selectedUpgradeTier}
      />
    </div>
  );
}
