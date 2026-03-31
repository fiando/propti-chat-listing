'use client';

import { useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Bell, Crown, Loader2, RefreshCw, Save, Settings, ShieldCheck, UserRound } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { PremiumUpgradeModal } from '@/components/premium/PremiumUpgradeModal';
import { SubscriptionStatusBadge } from '@/components/premium/SubscriptionStatusBadge';
import { shouldShowRenewalCTA } from '@/lib/premium-renewal';
import { getSubscriptionStatus } from '@/lib/subscription-status';
import { updateProfile } from '@/lib/api';
import { buildProfileUpdatePayload } from '@/lib/profile-update-payload';
import type { UpdateProfileRequest, UserPreferences } from '@/types';

const DEFAULT_PREFERENCES: UserPreferences = {
  favoriteLocations: [],
  searchHistory: [],
  notifications: true,
};

export default function SettingsPage() {
  const { status } = useSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const { profile, isLoading } = useAuth();
  const [showPremiumModal, setShowPremiumModal] = useState(false);
  const subscriptionStatus = getSubscriptionStatus({ authStatus: status, profile: profile ?? undefined });
  const showRenewalCTA = shouldShowRenewalCTA(subscriptionStatus);
  const [role, setRole] = useState<'buyer' | 'seller' | 'both' | ''>('');
  const [notifications, setNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const returnTo = searchParams.get('returnTo');

  useEffect(() => {
    if (!profile) {
      return;
    }
    setRole(profile.role ?? '');
    setNotifications(profile.preferences?.notifications ?? true);
  }, [profile]);

  const updateMutation = useMutation({
    mutationFn: (payload: UpdateProfileRequest) => updateProfile(payload),
    onSuccess: async () => {
      setSaved(true);
      await queryClient.invalidateQueries({ queryKey: ['profile'] });
      if (returnTo) {
        router.push(returnTo);
      }
    },
  });

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaved(false);

    const payload: UpdateProfileRequest = buildProfileUpdatePayload({
      role,
      notifications,
      preferences: profile?.preferences ?? DEFAULT_PREFERENCES,
    });

    await updateMutation.mutateAsync(payload);
  };

  if (status === 'loading' || isLoading || !profile) {
    return (
      <div className="flex items-center justify-center min-h-[320px]">
        <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-11 h-11 rounded-2xl bg-brand-light flex items-center justify-center">
            <Settings className="w-5 h-5 text-brand-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-black text-brand-primary">Pengaturan Akun</h1>
            <p className="text-gray-500">Kelola data profil dan preferensi akun Propti kamu.</p>
          </div>
        </div>
      </div>

      {returnTo && (
        <div className="mb-6 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
          Hubungkan WhatsApp dulu supaya kamu bisa pasang iklan. Setelah tersambung dan terverifikasi,
          kamu akan dikembalikan ke proses pasang iklan.
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="card p-6">
          <div className="flex items-center gap-2 mb-4">
            <UserRound className="w-5 h-5 text-brand-primary" />
            <h2 className="font-bold text-gray-900">Profil Dasar</h2>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <label className="block">
              <span className="text-sm font-medium text-gray-700">Nama</span>
              <input value={profile.name} disabled className="input-field mt-1 bg-gray-50 text-gray-500" />
            </label>

            <label className="block">
              <span className="text-sm font-medium text-gray-700">Email</span>
              <input value={profile.email} disabled className="input-field mt-1 bg-gray-50 text-gray-500" />
            </label>

            <label className="block">
              <span className="text-sm font-medium text-gray-700">Peran Akun</span>
              <select value={role} onChange={(event) => setRole(event.target.value as 'buyer' | 'seller' | 'both' | '')} className="input-field mt-1">
                {role === '' && <option value="">Pilih peran akun (opsional)</option>}
                <option value="buyer">Pencari Properti</option>
                <option value="seller">Penjual / Agen</option>
                <option value="both">Keduanya</option>
              </select>
            </label>
          </div>
        </div>

        <div className="card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Bell className="w-5 h-5 text-brand-primary" />
            <h2 className="font-bold text-gray-900">Preferensi</h2>
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
        </div>

        <div className="card p-6">
          <div className="flex items-center gap-2 mb-3">
            <ShieldCheck className="w-5 h-5 text-brand-primary" />
            <h2 className="font-bold text-gray-900">Status Akun</h2>
          </div>
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-600">
              Paket aktif saat ini: <span className="font-semibold text-gray-900">{subscriptionStatus === 'active' || subscriptionStatus === 'expiring_soon' ? 'Premium' : 'Gratis'}</span>
            </p>
            <SubscriptionStatusBadge
              status={subscriptionStatus}
              renewDate={profile?.subscription?.renewDate}
            />
          </div>
          {showRenewalCTA && (
            <button
              type="button"
              onClick={() => setShowPremiumModal(true)}
              className="mt-4 w-full btn-gold flex items-center justify-center gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              Perpanjang Premium
            </button>
          )}
          {subscriptionStatus === 'free' && (
            <button
              type="button"
              onClick={() => setShowPremiumModal(true)}
              className="mt-4 w-full btn-gold flex items-center justify-center gap-2"
            >
              <Crown className="w-4 h-4" />
              Upgrade ke Premium
            </button>
          )}
        </div>

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
          {updateMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          Simpan Pengaturan
        </button>
      </form>

      <PremiumUpgradeModal
        isOpen={showPremiumModal}
        onClose={() => setShowPremiumModal(false)}
        mode={showRenewalCTA ? 'renew' : 'upgrade'}
        currentRenewDate={profile?.subscription?.renewDate}
        currentTier={profile?.subscription?.tier ?? 'free'}
        selectedTier={profile?.subscription?.tier === 'free' ? 'basic' : profile?.subscription?.tier ?? 'premium'}
      />
    </div>
  );
}
