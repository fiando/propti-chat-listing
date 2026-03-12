'use client';

import { useState, useRef } from 'react';
import { MessageCircle, Sparkles, Loader2, CheckCircle, AlertTriangle, ArrowRight, Edit, MapPin, ClipboardPaste } from 'lucide-react';
import { parseListingText } from '@/lib/api';
import type { ParsedListing } from '@/types';
import { formatPrice } from '@/lib/utils';
import { useToast } from '@/app/toaster';

const EXAMPLE_TEXT = `Dijual rumah 2 lantai, 3KT 2KM
LT 120m2 LB 90m2 SHM
Harga 850jt nego
Lok Depok Beji dkt tol Cijago
Carport, taman, ruang tamu, dapur
Hub: 08123456789`;

interface TextParseFormProps {
  onParsed: (result: ParsedListing) => void;
  onManualFill: () => void;
}

export function TextParseForm({ onParsed, onManualFill }: TextParseFormProps) {
  const [text, setText] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ParsedListing | null>(null);
  const [error, setError] = useState<string | null>(null);
  const resultRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();
  const locationSuggestion = result?.locationSuggestion;
  const hasLocationSuggestion =
    !!locationSuggestion?.province ||
    !!locationSuggestion?.city ||
    !!locationSuggestion?.district ||
    !!locationSuggestion?.normalizedAddress;

  const handleParse = async () => {
    if (!text.trim()) return;
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const parsed = await parseListingText(text);
      setResult(parsed);
      toast('AI selesai memproses! Gulir ke bawah untuk melihat hasilnya.', 'success');
      // Small delay so the result card renders before scrolling into view
      setTimeout(() => {
        resultRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 100);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memproses teks. Silakan coba lagi.');
    } finally {
      setLoading(false);
    }
  };

  const handlePaste = async () => {
    try {
      const clipboardText = await navigator.clipboard.readText();
      if (clipboardText) setText(clipboardText);
    } catch {
      toast('Tidak dapat mengakses clipboard. Tempel teks secara manual.', 'warning');
    }
  };

  const handleUseResult = () => {
    if (result) onParsed(result);
  };

  return (
    <div className="space-y-6">
      {/* Input area */}
      <div className="card p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 bg-[#25D366] rounded-xl flex items-center justify-center">
            <MessageCircle className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="font-bold text-gray-900">Paste Teks Iklan dari WhatsApp</h2>
            <p className="text-sm text-gray-500">AI akan otomatis membaca dan merapikan iklanmu</p>
          </div>
        </div>

        <div className="relative">
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder={`Contoh:\n${EXAMPLE_TEXT}`}
            rows={8}
            className="w-full border-2 border-gray-200 rounded-xl px-4 py-3 text-sm font-mono text-gray-700 focus:outline-none focus:border-[#25D366] transition-colors resize-none bg-gray-50 placeholder:text-gray-400"
            disabled={loading}
          />
          {text && (
            <button
              onClick={() => setText('')}
              className="absolute top-3 right-3 text-gray-400 hover:text-gray-600 text-xs"
            >
              Hapus
            </button>
          )}
        </div>

        <div className="flex items-center justify-between mt-2 text-xs text-gray-400">
          <div className="flex items-center gap-3">
            <span>{text.length} karakter</span>
            <button
              type="button"
              onClick={handlePaste}
              disabled={loading}
              className="flex items-center gap-1 text-[#25D366] hover:text-[#1da851] font-medium transition-colors disabled:opacity-50"
            >
              <ClipboardPaste className="w-3.5 h-3.5" />
              Tempel
            </button>
          </div>
          <button
            onClick={() => setText(EXAMPLE_TEXT)}
            className="text-brand-secondary hover:text-brand-primary transition-colors"
          >
            Gunakan contoh teks
          </button>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 mt-4">
          <button
            onClick={handleParse}
            disabled={!text.trim() || loading}
            className="flex-1 flex items-center justify-center gap-2 bg-[#25D366] text-white font-semibold py-3 px-6 rounded-xl hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-md hover:shadow-lg"
          >
            {loading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Sedang merapikan...
              </>
            ) : (
              <>
                <Sparkles className="w-5 h-5" />
                Rapikan Otomatis
              </>
            )}
          </button>
          <button
            onClick={onManualFill}
            className="flex items-center justify-center gap-2 border-2 border-gray-200 text-gray-600 font-medium py-3 px-4 rounded-xl hover:bg-gray-50 transition-colors text-sm"
          >
            <Edit className="w-4 h-4" />
            Isi Manual
          </button>
        </div>
      </div>

      {/* Loading state */}
      {loading && (
        <div className="card p-8 text-center">
          <div className="w-16 h-16 bg-brand-light rounded-full flex items-center justify-center mx-auto mb-4">
            <Sparkles className="w-8 h-8 text-brand-primary animate-pulse" />
          </div>
          <h3 className="font-semibold text-gray-900 mb-2">Sedang merapikan detail dari chat Anda</h3>
          <p className="text-gray-500 text-sm">Harap tunggu sebentar...</p>
          <div className="mt-4 flex justify-center gap-1">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="w-2 h-2 bg-brand-primary rounded-full animate-bounce"
                style={{ animationDelay: `${i * 0.15}s` }}
              />
            ))}
          </div>
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="card p-5 border border-red-100 bg-red-50">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
            <div>
              <p className="font-medium text-red-700">Gagal memproses teks</p>
              <p className="text-sm text-red-600 mt-1">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Result */}
      {result && !loading && (
        <div ref={resultRef} className="card p-6 border-2 border-brand-accent/40">
          {/* Header */}
          <div className="flex items-center gap-3 mb-5">
            <div className="w-10 h-10 bg-brand-light rounded-xl flex items-center justify-center">
              <CheckCircle className="w-5 h-5 text-brand-primary" />
            </div>
            <div>
              <h3 className="font-bold text-gray-900">Hasil yang Terdeteksi</h3>
              <div className="flex items-center gap-2 mt-0.5">
                <div className="flex-1 h-1.5 bg-gray-200 rounded-full overflow-hidden w-24">
                  <div
                    className="h-full bg-brand-primary rounded-full"
                    style={{ width: `${result.confidence * 100}%` }}
                  />
                </div>
                <span className="text-xs text-gray-500">
                  Akurasi {Math.round(result.confidence * 100)}%
                </span>
              </div>
            </div>
          </div>

          {/* Parsed data */}
          <div className="grid grid-cols-2 gap-3 mb-5">
            {[
              { label: 'Judul', value: result.title },
              { label: 'Harga', value: result.price ? formatPrice(result.price) : '-' },
              {
                label: 'Luas Tanah',
                value: result.propertyDetails?.landArea
                  ? `${result.propertyDetails.landArea} m²`
                  : '-',
              },
              {
                label: 'Luas Bangunan',
                value: result.propertyDetails?.buildingArea
                  ? `${result.propertyDetails.buildingArea} m²`
                  : '-',
              },
              {
                label: 'Kamar Tidur',
                value: result.propertyDetails?.bedrooms
                  ? `${result.propertyDetails.bedrooms} KT`
                  : '-',
              },
              {
                label: 'Kamar Mandi',
                value: result.propertyDetails?.bathrooms
                  ? `${result.propertyDetails.bathrooms} KM`
                  : '-',
              },
              {
                label: 'Sertifikat',
                value: result.propertyDetails?.legalStatus || '-',
              },
              { label: 'Alamat', value: result.address || '-' },
            ].map(({ label, value }) => (
              <div key={label} className="bg-gray-50 rounded-xl p-3">
                <div className="text-xs text-gray-500 mb-0.5">{label}</div>
                <div className="font-semibold text-gray-900 text-sm line-clamp-1">{value}</div>
              </div>
            ))}
          </div>

          {/* Location suggestion */}
          {hasLocationSuggestion && (
            <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 mb-4">
              <p className="text-xs font-semibold text-blue-700 mb-2 flex items-center gap-1.5">
                <MapPin className="w-3.5 h-3.5" />
                Saran perbaikan lokasi
              </p>
              <div className="grid grid-cols-3 gap-2 text-xs">
                {locationSuggestion?.province && (
                  <div>
                    <span className="text-blue-500">Provinsi</span>
                    <p className="font-medium text-blue-900 mt-0.5">{locationSuggestion.province}</p>
                  </div>
                )}
                {locationSuggestion?.city && (
                  <div>
                    <span className="text-blue-500">Kota/Kab</span>
                    <p className="font-medium text-blue-900 mt-0.5">{locationSuggestion.city}</p>
                  </div>
                )}
                {locationSuggestion?.district && (
                  <div>
                    <span className="text-blue-500">Kecamatan</span>
                    <p className="font-medium text-blue-900 mt-0.5">{locationSuggestion.district}</p>
                  </div>
                )}
              </div>
              {locationSuggestion?.normalizedAddress && (
                <p className="text-xs text-blue-600 mt-2 border-t border-blue-200 pt-2">
                  {locationSuggestion.normalizedAddress}
                </p>
              )}
            </div>
          )}

          {/* Warnings */}
          {result.warnings?.length > 0 && (
            <div className="bg-amber-50 border border-amber-200 rounded-xl p-3 mb-4">
              <p className="text-xs font-semibold text-amber-700 mb-1.5 flex items-center gap-1">
                <AlertTriangle className="w-3.5 h-3.5" />
                Perlu diverifikasi:
              </p>
              <ul className="space-y-1">
                {result.warnings.map((w, i) => (
                  <li key={i} className="text-xs text-amber-600">
                    • {w}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Action buttons */}
          <div className="flex flex-col sm:flex-row gap-3">
            <button
              onClick={handleUseResult}
              className="flex-1 flex items-center justify-center gap-2 btn-primary"
            >
              <CheckCircle className="w-4 h-4" />
              Gunakan hasil ini
              <ArrowRight className="w-4 h-4" />
            </button>
            <button
              onClick={onManualFill}
              className="flex items-center justify-center gap-2 border-2 border-gray-200 text-gray-600 font-medium py-3 px-4 rounded-xl hover:bg-gray-50 transition-colors text-sm"
            >
              <Edit className="w-4 h-4" />
              Edit Manual
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
