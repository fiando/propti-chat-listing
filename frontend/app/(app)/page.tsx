import type { Metadata } from 'next';
import Link from 'next/link';
import {
  MessageCircle,
  Sparkles,
  Radio,
  ArrowRight,
  Home,
  TrendingUp,
  Users,
  Star,
  CheckCircle,
  ChevronRight,
  MapPin,
  Bed,
  Bath,
  Maximize2,
} from 'lucide-react';
import { formatPrice } from '@/lib/utils';
import { ShieldIcon } from '@/components/icons/ShieldIcon';

export const metadata: Metadata = {
  title: 'Propti — Jual Properti Semudah Chat WhatsApp',
};

const MOCK_LISTINGS = [
  {
    id: '1',
    title: 'Rumah Minimalis Modern Depok',
    price: 850_000_000,
    city: 'Depok',
    district: 'Beji',
    bedrooms: 3,
    bathrooms: 2,
    landArea: 120,
    buildingArea: 90,
    image: null,
    isFeatured: true,
    listingType: 'sell' as const,
  },
  {
    id: '2',
    title: 'Ruko 3 Lantai Strategis Tangerang',
    price: 2_500_000_000,
    city: 'Tangerang Selatan',
    district: 'BSD City',
    bedrooms: 0,
    bathrooms: 3,
    landArea: 90,
    buildingArea: 270,
    image: null,
    isFeatured: false,
    listingType: 'sell' as const,
  },
  {
    id: '3',
    title: 'Apartemen 2BR Kemang Jakarta Selatan',
    price: 8_500_000,
    city: 'Jakarta Selatan',
    district: 'Kemang',
    bedrooms: 2,
    bathrooms: 1,
    landArea: 0,
    buildingArea: 65,
    image: null,
    isFeatured: true,
    listingType: 'rent' as const,
  },
  {
    id: '4',
    title: 'Rumah Cluster Premium Bandung',
    price: 1_200_000_000,
    city: 'Bandung',
    district: 'Dago',
    bedrooms: 4,
    bathrooms: 3,
    landArea: 180,
    buildingArea: 150,
    image: null,
    isFeatured: false,
    listingType: 'sell' as const,
  },
  {
    id: '5',
    title: 'Kavling Siap Bangun Bogor',
    price: 320_000_000,
    city: 'Bogor',
    district: 'Cibinong',
    bedrooms: 0,
    bathrooms: 0,
    landArea: 200,
    buildingArea: 0,
    image: null,
    isFeatured: false,
    listingType: 'sell' as const,
  },
  {
    id: '6',
    title: 'Kost Eksklusif AC Surabaya',
    price: 1_800_000,
    city: 'Surabaya',
    district: 'Rungkut',
    bedrooms: 1,
    bathrooms: 1,
    landArea: 0,
    buildingArea: 20,
    image: null,
    isFeatured: false,
    listingType: 'rent' as const,
  },
];

const STATS = [
  { value: '10.000+', label: 'Properti', icon: Home },
  { value: '5.000+', label: 'Pengguna', icon: Users },
  { value: '99%', label: 'Kepuasan', icon: Star },
];

const HOW_IT_WORKS = [
  {
    step: 1,
    icon: MessageCircle,
    color: 'bg-[#25D366]',
    title: 'Paste Teks Iklan',
    desc: 'Copy paste iklan propertimu dari WhatsApp atau ketik langsung. Tidak perlu format khusus.',
  },
  {
    step: 2,
    icon: Sparkles,
    color: 'bg-brand-gold',
    title: 'AI Rapikan Otomatis',
    desc: 'AI kami ekstrak semua detail: harga, luas tanah, kamar, lokasi — dalam hitungan detik.',
  },
  {
    step: 3,
    icon: Radio,
    color: 'bg-brand-primary',
    title: 'Iklan Tayang',
    desc: 'Iklan kamu langsung live di Propti dan bisa ditemukan ribuan pencari properti.',
  },
];

const TESTIMONIALS = [
  {
    name: 'Budi Santoso',
    role: 'Agen Properti, Jakarta',
    text: 'Dulu butuh 30 menit buat upload satu listing. Sekarang 60 detik sudah tayang. Propti benar-benar revolusi!',
    avatar: 'BS',
  },
  {
    name: 'Siti Rahayu',
    role: 'Pemilik Kost, Bandung',
    text: 'Saya tinggal copy paste dari grup WA, langsung jadi iklan yang rapih. Mudah banget!',
    avatar: 'SR',
  },
  {
    name: 'Ahmad Fauzi',
    role: 'Developer Perumahan, Surabaya',
    text: 'Fitur AI parsing-nya luar biasa. Tidak perlu isi form panjang lagi. Highly recommended!',
    avatar: 'AF',
  },
];

function MockListingCard({ listing }: { listing: (typeof MOCK_LISTINGS)[0] }) {
  const typeLabel = listing.listingType === 'sell' ? 'Dijual' : 'Disewa';
  const typeBg = listing.listingType === 'sell' ? 'bg-brand-primary' : 'bg-blue-600';
  const priceLabel =
    listing.listingType === 'rent'
      ? `${formatPrice(listing.price)}/bln`
      : formatPrice(listing.price);

  const gradients = [
    'from-brand-primary to-brand-secondary',
    'from-blue-600 to-blue-400',
    'from-purple-600 to-purple-400',
    'from-amber-600 to-amber-400',
    'from-teal-600 to-teal-400',
    'from-rose-600 to-rose-400',
  ];
  const gradient = gradients[parseInt(listing.id) % gradients.length];

  return (
    <Link href={`/listings/${listing.id}`} className="card block group cursor-pointer">
      {/* Image area */}
      <div className={`relative h-48 bg-gradient-to-br ${gradient} rounded-t-2xl overflow-hidden`}>
        <div className="absolute inset-0 flex items-center justify-center opacity-20">
          <Home className="w-20 h-20 text-white" />
        </div>
        <div className="absolute top-3 left-3 flex gap-2">
          <span className={`${typeBg} text-white text-xs font-bold px-2.5 py-1 rounded-full`}>
            {typeLabel}
          </span>
          {listing.isFeatured && (
            <span className="bg-brand-gold text-white text-xs font-bold px-2.5 py-1 rounded-full flex items-center gap-1">
              <Star className="w-3 h-3" /> Unggulan
            </span>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="p-4">
        <h3 className="font-semibold text-gray-900 group-hover:text-brand-primary transition-colors line-clamp-1 mb-1">
          {listing.title}
        </h3>
        <p className="text-brand-primary font-bold text-lg mb-2">{priceLabel}</p>

        <div className="flex items-center gap-1 text-gray-500 text-xs mb-3">
          <MapPin className="w-3 h-3 flex-shrink-0" />
          <span className="truncate">
            {listing.district}, {listing.city}
          </span>
        </div>

        <div className="flex items-center gap-3 text-xs text-gray-500 border-t pt-3">
          {listing.landArea > 0 && (
            <span className="flex items-center gap-1">
              <Maximize2 className="w-3 h-3" />
              {listing.landArea} m²
            </span>
          )}
          {listing.buildingArea > 0 && (
            <span className="flex items-center gap-1">
              <Home className="w-3 h-3" />
              {listing.buildingArea} m²
            </span>
          )}
          {listing.bedrooms > 0 && (
            <span className="flex items-center gap-1">
              <Bed className="w-3 h-3" />
              {listing.bedrooms} KT
            </span>
          )}
          {listing.bathrooms > 0 && (
            <span className="flex items-center gap-1">
              <Bath className="w-3 h-3" />
              {listing.bathrooms} KM
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}

export default function HomePage() {
  return (
    <div className="bg-[#F8F9FA]">
      {/* ── HERO ── */}
      <section className="relative overflow-hidden bg-gradient-hero">
        {/* Decorative background shapes */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-24 -right-24 w-96 h-96 bg-white/5 rounded-full blur-3xl" />
          <div className="absolute -bottom-32 -left-32 w-[500px] h-[500px] bg-white/5 rounded-full blur-3xl" />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-white/3 rounded-full blur-3xl" />
        </div>

        <div className="relative max-w-6xl mx-auto px-4 pt-16 pb-24 md:pt-24 md:pb-32 text-center">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 bg-white/15 backdrop-blur-sm border border-white/20 text-white text-sm font-medium px-4 py-2 rounded-full mb-6">
            <span className="w-2 h-2 bg-[#25D366] rounded-full animate-pulse" />
            Cara tercepat owner & agen pasang listing properti
          </div>

          {/* Headline */}
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-black text-white leading-tight mb-4 text-balance">
            Jual properti
            <br />
            <span className="text-brand-accent">semudah chat.</span>
          </h1>

          {/* Sub-headline */}
          <p className="text-lg md:text-xl text-white/80 max-w-2xl mx-auto mb-10 text-balance">
            Paste detail properti dari WhatsApp, AI kami ubah jadi listing rapi, lengkap, dan siap
            tayang — dalam hitungan detik. Gratis untuk 3 iklan pertama.
          </p>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
            <Link
              href="/listings/create"
              className="group flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all duration-200 shadow-xl hover:shadow-2xl hover:-translate-y-0.5 text-lg"
            >
              <MessageCircle className="w-5 h-5 text-[#25D366]" />
              Paste listing saya
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/search"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all duration-200 text-lg"
            >
              <Home className="w-5 h-5" />
              Cari Properti
            </Link>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-6 max-w-md mx-auto">
            {STATS.map((s) => (
              <div key={s.label} className="text-center">
                <div className="text-2xl md:text-3xl font-black text-white">{s.value}</div>
                <div className="text-white/60 text-xs md:text-sm mt-1">{s.label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Wave divider */}
        <div className="absolute bottom-0 left-0 right-0">
          <svg viewBox="0 0 1440 60" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M0 60L1440 60L1440 30C1440 30 1080 0 720 0C360 0 0 30 0 30L0 60Z" fill="#F8F9FA" />
          </svg>
        </div>
      </section>

      {/* ── HOW IT WORKS ── */}
      <section className="max-w-6xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
            Cara Kerja
          </span>
          <h2 className="section-title">Pasang Iklan dalam 3 Langkah</h2>
          <p className="section-subtitle max-w-xl mx-auto">
            Tidak perlu isi formulir panjang. Cukup paste teks WhatsApp-mu dan biarkan AI yang bekerja.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-6 md:gap-8 relative">
          {/* Connector line for desktop */}
          <div className="hidden md:block absolute top-14 left-1/3 right-1/3 h-0.5 bg-gradient-to-r from-brand-accent to-brand-secondary z-0" />

          {HOW_IT_WORKS.map((step) => (
            <div key={step.step} className="relative z-10 text-center group">
              <div className="card p-8 hover:-translate-y-1 transition-transform duration-300">
                <div
                  className={`w-16 h-16 ${step.color} rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg group-hover:scale-110 transition-transform duration-300`}
                >
                  <step.icon className="w-8 h-8 text-white" />
                </div>
                <div className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">
                  Langkah {step.step}
                </div>
                <h3 className="text-lg font-bold text-gray-900 mb-3">{step.title}</h3>
                <p className="text-gray-500 text-sm leading-relaxed">{step.desc}</p>
              </div>
            </div>
          ))}
        </div>

        {/* AI Parse Demo Preview */}
        <div className="mt-12 bg-gradient-to-r from-[#25D366]/10 to-brand-light/50 border border-[#25D366]/20 rounded-3xl p-6 md:p-8">
          <div className="grid md:grid-cols-2 gap-8 items-center">
            <div>
              <div className="flex items-center gap-2 mb-3">
                <div className="w-6 h-6 bg-[#25D366] rounded-full flex items-center justify-center">
                  <MessageCircle className="w-3 h-3 text-white" />
                </div>
                <span className="text-sm font-semibold text-gray-600">Contoh Teks WhatsApp</span>
              </div>
              <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100 font-mono text-sm text-gray-600 leading-relaxed">
                Dijual rumah 2 lantai, 3KT 2KM, LT 120m2 LB 90m2 SHM, harga 850jt nego, lok Depok
                Beji dkt tol, hubungi 08123456789
              </div>
            </div>
            <div>
              <div className="flex items-center gap-2 mb-3">
                <div className="w-6 h-6 bg-brand-gold rounded-full flex items-center justify-center">
                  <Sparkles className="w-3 h-3 text-white" />
                </div>
                <span className="text-sm font-semibold text-gray-600">Hasil AI Parse Otomatis</span>
              </div>
              <div className="bg-white rounded-2xl p-4 shadow-sm border border-brand-accent/30 space-y-2">
                {[
                  ['Judul', 'Rumah 2 Lantai Depok Beji'],
                  ['Harga', 'Rp 850 Jt (Nego)'],
                  ['Luas Tanah', '120 m²'],
                  ['Luas Bangunan', '90 m²'],
                  ['Kamar Tidur', '3 KT'],
                  ['Sertifikat', 'SHM'],
                ].map(([key, val]) => (
                  <div key={key} className="flex items-center justify-between text-sm">
                    <span className="text-gray-500">{key}</span>
                    <span className="font-medium text-gray-800">{val}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="text-center mt-6">
            <Link
              href="/listings/create"
              className="inline-flex items-center gap-2 bg-[#25D366] text-white font-semibold px-6 py-3 rounded-xl hover:opacity-90 transition-opacity shadow-lg"
            >
              <MessageCircle className="w-4 h-4" />
              Coba Sekarang — Gratis!
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* ── FEATURED LISTINGS ── */}
      <section className="bg-white py-16 md:py-24">
        <div className="max-w-6xl mx-auto px-4">
          <div className="flex items-center justify-between mb-10">
            <div>
              <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-3">
                Properti Terbaru
              </span>
              <h2 className="section-title">Properti Pilihan</h2>
              <p className="section-subtitle">Temukan properti impianmu dari ribuan listing terpercaya</p>
            </div>
            <Link
              href="/search"
              className="hidden md:flex items-center gap-2 text-brand-primary font-semibold hover:text-brand-secondary transition-colors group"
            >
              Lihat Semua
              <ChevronRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {MOCK_LISTINGS.map((listing) => (
              <MockListingCard key={listing.id} listing={listing} />
            ))}
          </div>

          <div className="text-center mt-8 md:hidden">
            <Link href="/search" className="btn-secondary inline-flex items-center gap-2">
              Lihat Semua Properti
              <ChevronRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* ── FEATURES ── */}
      <section className="max-w-6xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
            Fitur Unggulan
          </span>
          <h2 className="section-title">Kenapa Pilih Propti?</h2>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {[
            {
              icon: <MessageCircle className="w-6 h-6 text-[#25D366]" />,
              bg: 'bg-[#25D366]/10',
              title: 'AI WhatsApp Parser',
              desc: 'Paste teks iklanmu dari WhatsApp — AI kami otomatis ekstrak judul, harga, luas, kamar, dan lokasi dengan akurasi tinggi.',
              tag: 'Teknologi AI',
            },
            {
              icon: <Sparkles className="w-6 h-6 text-brand-gold" />,
              bg: 'bg-amber-50',
              title: 'Listing Premium',
              desc: 'Jadikan iklanmu tampil di posisi teratas dengan fitur Featured Listing. Dapatkan 10x lebih banyak calon pembeli.',
              tag: 'Fitur Premium',
            },
            {
              icon: <TrendingUp className="w-6 h-6 text-brand-primary" />,
              bg: 'bg-brand-light',
              title: 'Analitik Gratis',
              desc: 'Pantau berapa kali iklanmu dilihat, disimpan, dan dihubungi. Data real-time untuk semua pengguna.',
              tag: 'Analytics',
            },
            {
              icon: <ShieldIcon className="w-6 h-6 text-blue-600" />,
              bg: 'bg-blue-50',
              title: 'Moderasi Otomatis',
              desc: 'Setiap iklan melalui proses moderasi cepat untuk memastikan kualitas dan keamanan transaksi.',
              tag: 'Aman & Terpercaya',
            },
          ].map((f, i) => (
            <div key={i} className="card p-6 flex gap-5">
              <div className={`w-14 h-14 ${f.bg} rounded-2xl flex items-center justify-center flex-shrink-0`}>
                {f.icon}
              </div>
              <div>
                <span className="text-xs font-semibold text-brand-secondary uppercase tracking-wide">
                  {f.tag}
                </span>
                <h3 className="text-lg font-bold text-gray-900 mt-1 mb-2">{f.title}</h3>
                <p className="text-gray-500 text-sm leading-relaxed">{f.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* ── TESTIMONIALS ── */}
      <section className="bg-brand-primary/5 py-16 md:py-24">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-12">
            <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
              Testimoni
            </span>
            <h2 className="section-title">Kata Mereka Tentang Propti</h2>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            {TESTIMONIALS.map((t, i) => (
              <div key={i} className="card p-6">
                <div className="flex items-center gap-1 mb-4">
                  {Array.from({ length: 5 }).map((_, j) => (
                    <Star key={j} className="w-4 h-4 fill-brand-gold text-brand-gold" />
                  ))}
                </div>
                <p className="text-gray-600 text-sm leading-relaxed mb-5 italic">
                  &ldquo;{t.text}&rdquo;
                </p>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-gradient-hero rounded-full flex items-center justify-center text-white font-bold text-sm">
                    {t.avatar}
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900 text-sm">{t.name}</div>
                    <div className="text-gray-500 text-xs">{t.role}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA BANNER ── */}
      <section className="bg-gradient-hero py-16 md:py-20 relative overflow-hidden">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-16 -right-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
          <div className="absolute -bottom-16 -left-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
        </div>
        <div className="relative max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl md:text-4xl font-black text-white mb-4">
            Mulai Jual Propertimu Sekarang
          </h2>
          <p className="text-white/80 text-lg mb-8 max-w-xl mx-auto">
            Daftar gratis, pasang 3 listing pertama tanpa biaya. Tidak perlu kartu kredit.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/login"
              className="flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all shadow-xl text-lg"
            >
              <CheckCircle className="w-5 h-5 text-brand-primary" />
              Daftar Gratis Sekarang
            </Link>
            <Link
              href="/search"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all text-lg"
            >
              Jelajahi Properti
            </Link>
          </div>
          <p className="text-white/50 text-sm mt-6">
            Sudah bergabung 5.000+ pengguna. Gratis selamanya untuk listing dasar.
          </p>
        </div>
      </section>
    </div>
  );
}
