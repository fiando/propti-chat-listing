'use client';

import { useState } from 'react';
import { Crown, X, Check, Loader2, RefreshCw } from 'lucide-react';
import { upgradePremium } from '@/lib/api';
import type { SubscriptionTier } from '@/types';

type PremiumModalMode = 'upgrade' | 'renew';

interface PremiumUpgradeModalProps {
  isOpen: boolean;
  onClose: () => void;
  mode?: PremiumModalMode;
  currentRenewDate?: string;
  selectedTier?: Exclude<SubscriptionTier, 'free'>;
}

const TIER_CONFIG = {
  basic: {
    label: 'Basic',
    price: 'Rp 59.000',
    blurb: 'Untuk seller yang mulai serius beriklan.',
    features: [
      'Basic: maksimal 8 foto per iklan',
      'Basic: maksimal 6 listing aktif',
      'WA text read + create (tanpa edit/hapus)',
      'Voice hingga 20 menit per bulan',
      'Masa tayang sampai 90 hari',
    ],
  },
  premium: {
    label: 'Premium',
    price: 'Rp 129.000',
    blurb: 'Untuk performa listing yang lebih agresif.',
    features: [
      'Premium: maksimal 15 foto per iklan',
      'Premium: maksimal 20 listing aktif',
      'WA full read/write',
      'Voice hingga 60 menit per bulan',
      'Premium: tayang sampai 90 hari',
    ],
  },
  pro: {
    label: 'Pro',
    price: 'Rp 199.000',
    blurb: 'Untuk tim agen dengan volume listing tinggi.',
    features: [
      'Pro: maksimal 20 foto per iklan',
      'Pro: maksimal 50 listing aktif',
      'WA full read/write',
      'Voice hingga 120 menit per bulan',
      'Masa tayang sampai 90 hari',
    ],
  },
} as const;

export function PremiumUpgradeModal({
  isOpen,
  onClose,
  mode = 'upgrade',
  currentRenewDate,
  selectedTier = 'premium',
}: PremiumUpgradeModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const isRenewal = mode === 'renew';
  const config = TIER_CONFIG[selectedTier];

  const expiryDateStr = currentRenewDate
    ? new Date(currentRenewDate).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
    : null;

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await upgradePremium(selectedTier);
      window.location.href = result.paymentUrl;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memproses pembayaran.');
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        aria-label="Tutup modal"
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-md bg-white rounded-3xl shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-brand-gold to-amber-500 p-6 text-white text-center">
          <button
            type="button"
            onClick={onClose}
            className="absolute top-4 right-4 w-8 h-8 bg-white/20 rounded-full flex items-center justify-center hover:bg-white/30 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
          <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center mx-auto mb-3">
            {isRenewal ? <RefreshCw className="w-8 h-8 text-white" /> : <Crown className="w-8 h-8 text-white" />}
          </div>
          <h2 className="text-2xl font-black">
            {isRenewal ? `Perpanjang ${config.label}` : `Propti ${config.label}`}
          </h2>
          {isRenewal && expiryDateStr ? (
            <p className="text-white/80 mt-1 text-sm">Berlaku sampai {expiryDateStr}. Perpanjang untuk +1 bulan.</p>
          ) : (
            <p className="text-white/80 mt-1 text-sm">{config.blurb}</p>
          )}
          <div className="mt-4">
            <span className="text-4xl font-black">{config.price}</span>
            <span className="text-white/70 text-sm ml-1">/bulan</span>
          </div>
        </div>

        {/* Features */}
        <div className="p-6">
          <ul className="space-y-3 mb-6">
            {config.features.map((feature) => (
              <li key={feature} className="flex items-start gap-3">
                <div className="w-5 h-5 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                  <Check className="w-3 h-3 text-green-600" />
                </div>
                <span className="text-sm text-gray-700">{feature}</span>
              </li>
            ))}
          </ul>

          {!isRenewal && (
            <div className="mb-6 rounded-2xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
              Paket gratis: maksimal 3 foto per iklan, 3 listing aktif, tayang 30 hari.
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-100 rounded-xl p-3 mb-4 text-sm text-red-600">
              {error}
            </div>
          )}

          <button
            type="button"
            onClick={handleSubmit}
            disabled={loading}
            className="w-full btn-gold flex items-center justify-center gap-2 text-base py-4"
          >
            {loading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Memproses...
              </>
            ) : isRenewal ? (
              <>
                <RefreshCw className="w-5 h-5" />
                Perpanjang Paket
              </>
            ) : (
              <>
                <Crown className="w-5 h-5" />
                Upgrade Paket
              </>
            )}
          </button>

          <p className="text-center text-xs text-gray-400 mt-1">
            Pembayaran aman melalui DOKU. Batalkan kapan saja.
          </p>
        </div>
      </div>
    </div>
  );
}
