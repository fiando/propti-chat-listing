import type { Metadata } from 'next';
import Link from 'next/link';
import { ArrowLeft, Home, MessageCircle } from 'lucide-react';
import { KprCalculator } from '@/components/common/KprCalculator';

export const metadata: Metadata = {
  title: 'Kalkulator KPR — Propti',
  description: 'Hitung estimasi angsuran KPR berdasarkan harga properti, uang muka, suku bunga, dan tenor. Bantu calon pembeli siap sebelum survei.',
};

export default function KprPage() {
  return (
    <div className="bg-[#F8F9FA] min-h-screen">
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="mb-6">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Kembali
          </Link>
        </div>

        <div className="text-center mb-8">
          <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
            Tools Gratis
          </span>
          <h1 className="text-3xl font-black text-gray-900 mb-3">
            Kalkulator KPR
          </h1>
          <p className="text-gray-500 max-w-xl mx-auto">
            Hitung estimasi angsuran bulanan sebelum survei atau presentasi properti ke calon pembeli. Bantu mereka siap secara finansial lebih awal.
          </p>
        </div>

        <KprCalculator />

        <div className="mt-8 grid sm:grid-cols-2 gap-4">
          <div className="card p-5">
            <h3 className="font-bold text-gray-900 mb-2">Tips untuk Agen</h3>
            <ul className="space-y-1.5 text-sm text-gray-600">
              <li>• Tunjukkan kalkulator ini saat presentasi properti ke calon pembeli</li>
              <li>• Bantu mereka hitung DP yang perlu disiapkan sejak awal</li>
              <li>• Bandingkan beberapa skenario tenor untuk temukan yang cocok</li>
              <li>• KPR biasanya butuh DP minimal 10–20% dari harga properti</li>
            </ul>
          </div>
          <div className="card p-5">
            <h3 className="font-bold text-gray-900 mb-2">Tips untuk Pembeli</h3>
            <ul className="space-y-1.5 text-sm text-gray-600">
              <li>• Suku bunga KPR bank Indonesia rata-rata 9–12% per tahun</li>
              <li>• Tenor lebih panjang = angsuran lebih kecil, tapi bunga total lebih besar</li>
              <li>• Angsuran ideal tidak lebih dari 30% dari penghasilan bulanan</li>
              <li>• Siapkan juga biaya notaris, BPHTB, dan KPR (3–5% dari harga)</li>
            </ul>
          </div>
        </div>

        <div className="mt-8 bg-gradient-to-r from-brand-primary to-brand-secondary rounded-3xl p-6 text-white text-center">
          <h3 className="text-lg font-bold mb-2">Sudah ada properti yang ingin dijual?</h3>
          <p className="text-white/80 text-sm mb-4">
            Buat listing rapi dari teks WhatsApp dan sertakan link Propti saat berbagi ke calon pembeli.
          </p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <Link
              href="/listings/create"
              className="inline-flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-6 py-3 rounded-xl hover:bg-brand-light transition-colors text-sm"
            >
              <MessageCircle className="w-4 h-4 text-[#25D366]" />
              Paste listing saya
            </Link>
            <Link
              href="/search"
              className="inline-flex items-center justify-center gap-2 border-2 border-white/40 text-white font-semibold px-6 py-3 rounded-xl hover:bg-white/10 transition-colors text-sm"
            >
              <Home className="w-4 h-4" />
              Cari Properti
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
