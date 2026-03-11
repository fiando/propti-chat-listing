import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatPrice(price: number): string {
  if (price >= 1_000_000_000) {
    return `Rp ${(price / 1_000_000_000).toFixed(1)} Mlr`;
  }
  if (price >= 1_000_000) {
    return `Rp ${(price / 1_000_000).toFixed(0)} Jt`;
  }
  return `Rp ${price.toLocaleString('id-ID')}`;
}

export function formatArea(area: number): string {
  return `${area} m²`;
}

export function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(new Date(dateStr));
}

export function formatRelativeDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays === 0) return 'Hari ini';
  if (diffDays === 1) return 'Kemarin';
  if (diffDays < 7) return `${diffDays} hari lalu`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} minggu lalu`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} bulan lalu`;
  return `${Math.floor(diffDays / 365)} tahun lalu`;
}

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim();
}

export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '...';
}

export const LISTING_TYPE_LABELS: Record<string, string> = {
  sell: 'Dijual',
  rent: 'Disewa',
};

export const PRICE_UNIT_LABELS: Record<string, string> = {
  per_unit: 'Harga Total',
  per_month: 'Per Bulan',
  negotiable: 'Nego',
};

export const LEGAL_STATUS_OPTIONS = [
  'SHM - Sertifikat Hak Milik',
  'HGB - Hak Guna Bangunan',
  'SHSRS - Strata Title',
  'Girik / Letter C',
  'AJB - Akta Jual Beli',
  'Lainnya',
];

export const ORIENTATION_OPTIONS = [
  'Utara',
  'Selatan',
  'Timur',
  'Barat',
  'Timur Laut',
  'Barat Laut',
  'Tenggara',
  'Barat Daya',
];

export const AMENITIES_OPTIONS = [
  { id: 'ruang_tamu', label: 'Ruang Tamu' },
  { id: 'ruang_keluarga', label: 'Ruang Keluarga' },
  { id: 'dapur', label: 'Dapur' },
  { id: 'carport', label: 'Carport' },
  { id: 'garasi', label: 'Garasi' },
  { id: 'taman', label: 'Taman' },
  { id: 'kolam_renang', label: 'Kolam Renang' },
  { id: 'balkon', label: 'Balkon' },
  { id: 'gudang', label: 'Gudang' },
  { id: 'ac', label: 'AC' },
  { id: 'keamanan_24jam', label: 'Keamanan 24 Jam' },
  { id: 'cctv', label: 'CCTV' },
  { id: 'one_gate_system', label: 'One Gate System' },
  { id: 'masjid', label: 'Masjid / Mushola' },
  { id: 'playground', label: 'Playground' },
  { id: 'jogging_track', label: 'Jogging Track' },
];

export const INDONESIAN_CITIES = [
  'Jakarta Selatan',
  'Jakarta Utara',
  'Jakarta Barat',
  'Jakarta Timur',
  'Jakarta Pusat',
  'Depok',
  'Bekasi',
  'Tangerang',
  'Tangerang Selatan',
  'Bogor',
  'Bandung',
  'Surabaya',
  'Yogyakarta',
  'Semarang',
  'Medan',
  'Makassar',
  'Bali / Denpasar',
  'Palembang',
  'Batam',
  'Balikpapan',
];
