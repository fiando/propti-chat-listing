'use client';

import { useState } from 'react';
import { Crown, X, Check, Loader2 } from 'lucide-react';
import { upgradePremium } from '@/lib/api';
import Link from 'next/link';

interface PremiumUpgradeModalProps {
  isOpen: boolean;
  onClose: () => void;
  profilePhone?: string | null;
}

const PREMIUM_FEATURES = [
  'Upload hingga 30 foto per iklan (gratis maksimal 3)',
  'Iklan tampil lebih lama (gratis 30 hari)',
  'Posting lebih dari 3 listing gratis',
  'Statistik dasar listing: dilihat dan disimpan',
];

export function PremiumUpgradeModal({ isOpen, onClose, profilePhone }: PremiumUpgradeModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleUpgrade = async () => {
    if (!profilePhone?.trim()) {
      setError('Lengkapi nomor telepon dulu sebelum upgrade Premium.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const result = await upgradePremium();
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
        aria-label="Tutup modal upgrade"
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
            <Crown className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-black">Propti Premium</h2>
          <p className="text-white/80 mt-1 text-sm">Tambah kapasitas listing dan media tanpa janji fitur yang belum aktif</p>
          <div className="mt-4">
            <span className="text-4xl font-black">Rp 49rb</span>
            <span className="text-white/70 text-sm ml-1">/bulan</span>
          </div>
        </div>

        {/* Features */}
        <div className="p-6">
          <ul className="space-y-3 mb-6">
            {PREMIUM_FEATURES.map((feature) => (
              <li key={feature} className="flex items-start gap-3">
                <div className="w-5 h-5 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                  <Check className="w-3 h-3 text-green-600" />
                </div>
                <span className="text-sm text-gray-700">{feature}</span>
              </li>
            ))}
          </ul>

          {error && (
            <div className="bg-red-50 border border-red-100 rounded-xl p-3 mb-4 text-sm text-red-600">
              {error}
            </div>
          )}

          {!profilePhone?.trim() && (
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-3 mb-4 text-sm text-amber-700">
              Lengkapi nomor telepon dulu di pengaturan akun supaya proses upgrade bisa lanjut.
              <Link href="/settings?returnTo=/profile#premium" className="ml-1 font-semibold underline">
                Buka pengaturan
              </Link>
            </div>
          )}

          <div className="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-3 mb-4 text-sm text-amber-800">
            Iklan unggulan atau promosi berbayar per listing dan dibayar terpisah dari paket
            Premium ini. Boost tetap memakai checkout terpisah saat tersedia.
          </div>

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
                Upgrade Sekarang
              </>
            )}
          </button>

          <p className="text-center text-xs text-gray-400 mt-3">
            Pembayaran aman melalui DOKU. Batalkan kapan saja.
          </p>
        </div>
      </div>
    </div>
  );
}
