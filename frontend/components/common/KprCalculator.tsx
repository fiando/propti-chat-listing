'use client';

import { useState, useMemo } from 'react';
import { Calculator, Info } from 'lucide-react';

function formatRupiah(value: number): string {
  if (value >= 1_000_000_000) {
    const b = value / 1_000_000_000;
    return `Rp ${b % 1 === 0 ? b.toFixed(0) : b.toFixed(2)} Miliar`;
  }
  if (value >= 1_000_000) {
    const jt = value / 1_000_000;
    return `Rp ${jt % 1 === 0 ? jt.toFixed(0) : jt.toFixed(1)} Jt`;
  }
  return `Rp ${value.toLocaleString('id-ID')}`;
}

interface KprResult {
  loanAmount: number;
  monthlyPayment: number;
  totalPayment: number;
  totalInterest: number;
  dpAmount: number;
}

function calculateKpr(
  propertyPrice: number,
  dpPercent: number,
  annualRate: number,
  tenorYears: number,
): KprResult | null {
  if (propertyPrice <= 0 || annualRate <= 0 || tenorYears <= 0) return null;
  const dpAmount = (propertyPrice * dpPercent) / 100;
  const loanAmount = propertyPrice - dpAmount;
  if (loanAmount <= 0) return null;

  const monthlyRate = annualRate / 100 / 12;
  const n = tenorYears * 12;
  const monthlyPayment = loanAmount * (monthlyRate * Math.pow(1 + monthlyRate, n)) / (Math.pow(1 + monthlyRate, n) - 1);
  const totalPayment = monthlyPayment * n;
  const totalInterest = totalPayment - loanAmount;

  return { loanAmount, monthlyPayment, totalPayment, totalInterest, dpAmount };
}

export function KprCalculator() {
  const [priceRaw, setPriceRaw] = useState('750000000');
  const [dpPercent, setDpPercent] = useState(20);
  const [annualRate, setAnnualRate] = useState(10.5);
  const [tenorYears, setTenorYears] = useState(15);

  const propertyPrice = useMemo(() => (priceRaw ? parseInt(priceRaw, 10) : 0), [priceRaw]);

  const result = useMemo(
    () => calculateKpr(propertyPrice, dpPercent, annualRate, tenorYears),
    [propertyPrice, dpPercent, annualRate, tenorYears],
  );

  const handlePriceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const digits = e.target.value.replace(/\D/g, '');
    setPriceRaw(digits);
  };

  const priceDisplayValue = priceRaw ? parseInt(priceRaw, 10).toLocaleString('id-ID') : '';

  return (
    <div className="card p-6 md:p-8">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-brand-light rounded-xl flex items-center justify-center flex-shrink-0">
          <Calculator className="w-5 h-5 text-brand-primary" />
        </div>
        <div>
          <h2 className="text-lg font-bold text-gray-900">Kalkulator KPR</h2>
          <p className="text-sm text-gray-500">Estimasi angsuran bulanan berdasarkan harga properti</p>
        </div>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Inputs */}
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-1.5">
              Harga Properti
            </label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400 font-medium">Rp</span>
              <input
                type="text"
                inputMode="numeric"
                value={priceDisplayValue}
                onChange={handlePriceChange}
                placeholder="750.000.000"
                className="input-field pl-9 text-sm"
              />
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-sm font-semibold text-gray-700">
                Uang Muka (DP)
              </label>
              <span className="text-sm font-bold text-brand-primary">{dpPercent}%</span>
            </div>
            <input
              type="range"
              min={10}
              max={50}
              step={5}
              value={dpPercent}
              onChange={(e) => setDpPercent(Number(e.target.value))}
              className="w-full accent-brand-primary"
            />
            <div className="flex justify-between text-xs text-gray-400 mt-1">
              <span>10%</span>
              <span>50%</span>
            </div>
            {result && (
              <p className="text-xs text-gray-500 mt-1">
                DP: {formatRupiah(result.dpAmount)}
              </p>
            )}
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-sm font-semibold text-gray-700">
                Suku Bunga / Tahun
              </label>
              <span className="text-sm font-bold text-brand-primary">{annualRate}%</span>
            </div>
            <input
              type="range"
              min={5}
              max={18}
              step={0.5}
              value={annualRate}
              onChange={(e) => setAnnualRate(Number(e.target.value))}
              className="w-full accent-brand-primary"
            />
            <div className="flex justify-between text-xs text-gray-400 mt-1">
              <span>5%</span>
              <span>18%</span>
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-sm font-semibold text-gray-700">
                Tenor KPR
              </label>
              <span className="text-sm font-bold text-brand-primary">{tenorYears} tahun</span>
            </div>
            <input
              type="range"
              min={5}
              max={30}
              step={1}
              value={tenorYears}
              onChange={(e) => setTenorYears(Number(e.target.value))}
              className="w-full accent-brand-primary"
            />
            <div className="flex justify-between text-xs text-gray-400 mt-1">
              <span>5 thn</span>
              <span>30 thn</span>
            </div>
          </div>
        </div>

        {/* Results */}
        <div className="flex flex-col justify-center">
          {result ? (
            <div className="space-y-3">
              <div className="bg-brand-light rounded-2xl p-5 text-center">
                <p className="text-xs font-semibold text-brand-secondary uppercase tracking-wide mb-1">
                  Estimasi Angsuran / Bulan
                </p>
                <p className="text-3xl font-black text-brand-primary">
                  {formatRupiah(result.monthlyPayment)}
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  selama {tenorYears} tahun ({tenorYears * 12} kali angsuran)
                </p>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="bg-gray-50 rounded-xl p-3 text-center">
                  <p className="text-xs text-gray-500 mb-0.5">Jumlah Pinjaman</p>
                  <p className="text-base font-bold text-gray-900">{formatRupiah(result.loanAmount)}</p>
                </div>
                <div className="bg-gray-50 rounded-xl p-3 text-center">
                  <p className="text-xs text-gray-500 mb-0.5">Total Bunga</p>
                  <p className="text-base font-bold text-gray-900">{formatRupiah(result.totalInterest)}</p>
                </div>
                <div className="bg-gray-50 rounded-xl p-3 text-center col-span-2">
                  <p className="text-xs text-gray-500 mb-0.5">Total Bayar Keseluruhan</p>
                  <p className="text-base font-bold text-gray-900">{formatRupiah(result.totalPayment)}</p>
                </div>
              </div>

              <div className="flex items-start gap-2 bg-amber-50 rounded-xl p-3">
                <Info className="w-3.5 h-3.5 text-amber-500 flex-shrink-0 mt-0.5" />
                <p className="text-xs text-amber-700 leading-relaxed">
                  Perhitungan menggunakan metode anuitas. Angsuran aktual dapat berbeda tergantung kebijakan bank dan suku bunga berlaku.
                </p>
              </div>
            </div>
          ) : (
            <div className="text-center text-gray-400 py-8">
              <Calculator className="w-10 h-10 mx-auto mb-2 opacity-30" />
              <p className="text-sm">Isi harga properti untuk melihat estimasi</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
