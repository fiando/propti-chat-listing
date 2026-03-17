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

export function formatPriceFull(price: number): string {
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
  total: 'Harga Total',
  monthly: 'Per Bulan',
  yearly: 'Per Tahun',
};

export const LEGAL_STATUS_OPTIONS = [
  { value: 'SHM', label: 'SHM - Sertifikat Hak Milik' },
  { value: 'HGB', label: 'HGB - Hak Guna Bangunan' },
  { value: 'SHSRS', label: 'SHSRS - Strata Title' },
  { value: 'Girik', label: 'Girik / Letter C' },
  { value: 'AJB', label: 'AJB - Akta Jual Beli' },
  { value: 'Lainnya', label: 'Lainnya' },
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
  { id: 'teras', label: 'Teras' },
  { id: 'kanopi', label: 'Kanopi' },
  { id: 'kolam_renang', label: 'Kolam Renang' },
  { id: 'balkon', label: 'Balkon' },
  { id: 'gudang', label: 'Gudang' },
  { id: 'ruang_makan', label: 'Ruang Makan' },
  { id: 'ruang_kerja', label: 'Ruang Kerja' },
  { id: 'ruang_cuci', label: 'Ruang Cuci' },
  { id: 'kamar_pembantu', label: 'Kamar Pembantu' },
  { id: 'kamar_mandi_pembantu', label: 'KM Pembantu' },
  { id: 'tempat_jemuran', label: 'Area Jemur' },
  { id: 'pantry', label: 'Pantry' },
  { id: 'ac', label: 'AC' },
  { id: 'water_heater', label: 'Water Heater' },
  { id: 'kitchen_set', label: 'Kitchen Set' },
  { id: 'furnished', label: 'Fully Furnished' },
  { id: 'semi_furnished', label: 'Semi Furnished' },
  { id: 'internet_wifi', label: 'Internet / WiFi' },
  { id: 'tv_kabel', label: 'TV Kabel' },
  { id: 'pompa_air', label: 'Pompa Air' },
  { id: 'sumur_bor', label: 'Sumur Bor' },
  { id: 'pdams', label: 'Air PDAM' },
  { id: 'listrik_3_phase', label: 'Listrik 3 Phase' },
  { id: 'keamanan_24jam', label: 'Keamanan 24 Jam' },
  { id: 'cctv', label: 'CCTV' },
  { id: 'one_gate_system', label: 'One Gate System' },
  { id: 'akses_kartu', label: 'Access Card' },
  { id: 'lift', label: 'Lift' },
  { id: 'lobi', label: 'Lobi / Reception' },
  { id: 'gym', label: 'Gym / Fitness' },
  { id: 'clubhouse', label: 'Clubhouse' },
  { id: 'masjid', label: 'Masjid / Mushola' },
  { id: 'playground', label: 'Playground' },
  { id: 'jogging_track', label: 'Jogging Track' },
  { id: 'lapangan_olahraga', label: 'Lapangan Olahraga' },
  { id: 'function_room', label: 'Function Room' },
  { id: 'loading_dock', label: 'Loading Dock' },
  { id: 'akses_container', label: 'Akses Kontainer' },
  { id: 'akses_truk', label: 'Akses Truk' },
  { id: 'fire_safety', label: 'Fire Safety' },
  { id: 'dekat_tol', label: 'Dekat Tol' },
  { id: 'jalan_lebar', label: 'Akses Jalan Lebar' },
];

const AMENITY_LABELS = new Map(AMENITIES_OPTIONS.map((amenity) => [amenity.id, amenity.label]));

export function formatAmenityLabel(value: string): string {
  const normalizedValue = value.trim();

  if (!normalizedValue) {
    return '';
  }

  return (
    AMENITY_LABELS.get(normalizedValue) ||
    normalizedValue
      .replace(/[_-]+/g, ' ')
      .split(/\s+/)
      .filter(Boolean)
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ')
  );
}

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
