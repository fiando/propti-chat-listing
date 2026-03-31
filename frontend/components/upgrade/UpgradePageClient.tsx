'use client';

import { useState } from 'react';
import { Crown, Check, Loader2, ArrowLeft } from 'lucide-react';
import { upgradePremium } from '@/lib/api';
import Link from 'next/link';
import type { SubscriptionTier } from '@/types';

type PaidTier = Exclude<SubscriptionTier, 'free'>;

interface UpgradePageClientProps {
  isAuthenticated: boolean;
  currentTier: SubscriptionTier;
  initialTier: PaidTier;
}

const TIER_CONFIG: Record<
  PaidTier,
  {
    label: string;
    price: string;
    priceNum: string;
    blurb: string;
    features: string[];
    highlight: boolean;
  }
> = {
  basic: {
    label: 'Basic',
    price: 'Rp 59.000',
    priceNum: '59rb',
    blurb: 'Untuk seller yang mulai serius beriklan.',
    highlight: false,
    features: [
      'Maksimal 8 foto per iklan',
      'Maksimal 6 listing aktif',
      'WA baca + buat listing (tanpa edit/hapus)',
      'Voice note hingga 20 menit per bulan',
      'Iklan tayang hingga 90 hari',
    ],
  },
  premium: {
    label: 'Premium',
    price: 'Rp 129.000',
    priceNum: '129rb',
    blurb: 'Untuk performa listing yang lebih agresif.',
    highlight: true,
    features: [
      'Maksimal 15 foto per iklan',
      'Maksimal 20 listing aktif',
      'WA baca, buat, edit & hapus listing',
      'Voice note hingga 60 menit per bulan',
      'Iklan tayang hingga 90 hari',
    ],
  },
  pro: {
    label: 'Pro',
    price: 'Rp 199.000',
    priceNum: '199rb',
    blurb: 'Untuk tim agen dengan volume listing tinggi.',
    highlight: false,
    features: [
      'Maksimal 20 foto per iklan',
      'Maksimal 50 listing aktif',
      'WA baca, buat, edit & hapus listing',
      'Voice note hingga 120 menit per bulan',
      'Iklan tayang hingga 90 hari',
    ],
  },
};

const FREE_FEATURES = [
  'Maksimal 3 foto per iklan',
  'Maksimal 3 listing aktif',
  'Iklan tayang 30 hari',
  'Buat listing via WhatsApp',
];

const PAID_TIERS: PaidTier[] = ['basic', 'premium', 'pro'];

export function UpgradePageClient({ currentTier, initialTier }: UpgradePageClientProps) {
  const [activeTier, setActiveTier] = useState<PaidTier>(initialTier);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const config = TIER_CONFIG[activeTier];
  const isCurrentPlan = currentTier === activeTier;

  const handleUpgrade = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await upgradePremium(activeTier);
      window.location.href = result.paymentUrl;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memproses pembayaran.');
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="mb-6">
        <Link href="/profile#premium" className="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700">
          <ArrowLeft className="w-4 h-4" />
          Kembali ke Profil
        </Link>
      </div>

      <div className="text-center mb-10">
        <div className="inline-flex items-center gap-2 bg-amber-50 text-amber-700 text-sm font-semibold px-4 py-2 rounded-full mb-4 border border-amber-200">
          <Crown className="w-4 h-4" />
          Pilih Paket
        </div>
        <h1 className="text-3xl font-black text-gray-900 mb-3">Upgrade Akun Propti Anda</h1>
        <p className="text-gray-500 max-w-xl mx-auto">
          Pilih paket yang sesuai kebutuhan. Semua paket berbayar mendukung listing via WhatsApp dan voice note.
        </p>
      </div>

      {/* Plan comparison grid */}
      <div className="grid md:grid-cols-4 gap-4 mb-8">
        {/* Free plan */}
        <div className="card p-5 border-2 border-gray-100">
          <div className="mb-4">
            <span className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Gratis</span>
            <div className="mt-1">
              <span className="text-2xl font-black text-gray-900">Rp 0</span>
              <span className="text-gray-400 text-sm ml-1">/bulan</span>
            </div>
            <p className="text-xs text-gray-500 mt-1">Mulai tanpa biaya.</p>
          </div>
          <ul className="space-y-2 mb-5">
            {FREE_FEATURES.map((f) => (
              <li key={f} className="flex items-start gap-2 text-sm text-gray-600">
                <Check className="w-3.5 h-3.5 text-gray-400 flex-shrink-0 mt-0.5" />
                {f}
              </li>
            ))}
          </ul>
          {currentTier === 'free' && (
            <div className="w-full text-center py-2 rounded-xl bg-gray-100 text-gray-500 text-sm font-semibold">
              Paket Aktif
            </div>
          )}
        </div>

        {/* Paid plans */}
        {PAID_TIERS.map((tier) => {
          const plan = TIER_CONFIG[tier];
          const isActive = activeTier === tier;
          const isCurrent = currentTier === tier;
          return (
            <button
              key={tier}
              type="button"
              onClick={() => setActiveTier(tier)}
              className={`card p-5 text-left transition-all ${
                isActive
                  ? 'border-2 border-brand-gold shadow-lg'
                  : 'border-2 border-transparent hover:border-amber-200'
              } ${plan.highlight ? 'relative' : ''}`}
            >
              {plan.highlight && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="bg-brand-gold text-white text-xs font-bold px-3 py-1 rounded-full whitespace-nowrap">
                    Paling Populer
                  </span>
                </div>
              )}
              <div className="mb-4">
                <div className="flex items-center gap-2 mb-1">
                  <Crown className={`w-4 h-4 ${isActive ? 'text-brand-gold' : 'text-gray-400'}`} />
                  <span className="text-sm font-semibold text-gray-700 uppercase tracking-wide">{plan.label}</span>
                </div>
                <div className="mt-1">
                  <span className="text-2xl font-black text-gray-900">{plan.price}</span>
                  <span className="text-gray-400 text-sm ml-1">/bulan</span>
                </div>
                <p className="text-xs text-gray-500 mt-1">{plan.blurb}</p>
              </div>
              <ul className="space-y-2 mb-5">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-start gap-2 text-sm text-gray-600">
                    <Check className="w-3.5 h-3.5 text-green-500 flex-shrink-0 mt-0.5" />
                    {f}
                  </li>
                ))}
              </ul>
              {isCurrent ? (
                <div className="w-full text-center py-2 rounded-xl bg-amber-100 text-amber-700 text-sm font-semibold">
                  Paket Aktif
                </div>
              ) : (
                <div
                  className={`w-full text-center py-2 rounded-xl text-sm font-semibold ${
                    isActive
                      ? 'bg-brand-gold text-white'
                      : 'bg-gray-100 text-gray-500'
                  }`}
                >
                  {isActive ? 'Dipilih' : 'Pilih'}
                </div>
              )}
            </button>
          );
        })}
      </div>

      {/* CTA section */}
      <div className="card p-6 text-center max-w-md mx-auto">
        <h2 className="font-bold text-gray-900 mb-1">
          {isCurrentPlan ? 'Ini paket aktif Anda' : `Upgrade ke Propti ${config.label}`}
        </h2>
        <p className="text-sm text-gray-500 mb-4">
          {isCurrentPlan
            ? 'Perpanjang paket sebelum masa berlaku habis.'
            : `${config.price}/bulan · ${config.blurb}`}
        </p>

        {error && (
          <div className="bg-red-50 border border-red-100 rounded-xl p-3 mb-4 text-sm text-red-600">
            {error}
          </div>
        )}

        <button
          type="button"
          onClick={handleUpgrade}
          disabled={loading}
          className="w-full btn-gold flex items-center justify-center gap-2 text-base py-4"
        >
          {loading ? (
            <>
              <Loader2 className="w-5 h-5 animate-spin" />
              Memproses...
            </>
          ) : (
            <>
              <Crown className="w-5 h-5" />
              {isCurrentPlan ? `Perpanjang ${config.label}` : `Upgrade ke ${config.label} — ${config.price}`}
            </>
          )}
        </button>
        <p className="text-xs text-gray-400 mt-2">
          Pembayaran aman melalui DOKU. Batalkan kapan saja.
        </p>
      </div>
    </div>
  );
}
