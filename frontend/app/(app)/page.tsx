import type { Metadata } from 'next';
import Link from 'next/link';
import { Fragment } from 'react';
import { redirect } from 'next/navigation';
import {
  MessageCircle,
  Sparkles,
  Radio,
  ArrowRight,
  TrendingUp,
  ChevronRight,
  Mic,
  Search,
  Crown,
  Check,
  BarChart3,
  Users,
  Calculator,
  LayoutDashboard,
} from 'lucide-react';
import { ShieldIcon } from '@/components/icons/ShieldIcon';
import { ListingCard } from '@/components/listings/ListingCard';
import { getHomepageListings } from '@/lib/homepage-listings';
import { getAuthenticatedHomeRedirectPath } from '@/lib/home-redirect';
import { getServerAuthProfile } from '@/lib/server-profile';

export const metadata: Metadata = {
  title: 'Propti — Alat Kerja Properti untuk Agen & Pemilik di Indonesia',
};

const HERO_PROOF_POINTS = [
  {
    title: 'Buat iklan properti rapi dari teks WhatsApp',
    desc: 'Paste teks iklan yang sudah kamu punya — AI langsung baca dan rapikan semua detail, lalu buat halaman iklan siap pakai.',
  },
  {
    title: 'Catat semua calon pembeli di satu tempat',
    desc: 'Lacak setiap orang yang tertarik dari pertama kontak hingga transaksi selesai. Tidak ada lagi yang terlewat hanya karena chat tenggelam.',
  },
  {
    title: 'Bagikan ke semua tempat dari satu link',
    desc: 'Satu link Propti jadi sumber utama iklanmu saat promosi di WhatsApp, Instagram, atau situs jual beli mana pun.',
  },
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
    desc: 'AI kami ekstrak semua detail penting: harga, luas tanah, kamar, dan lokasi dalam hitungan detik.',
  },
  {
    step: 3,
    icon: Radio,
    color: 'bg-brand-primary',
    title: 'Tinjau lalu Tayang',
    desc: 'Cek hasil parse, edit seperlunya, lalu terbitkan listing dengan informasi yang lebih rapi.',
  },
];

const PRODUCT_PROOF = [
  {
    icon: <MessageCircle className="w-6 h-6 text-[#25D366]" />,
    bg: 'bg-[#25D366]/10',
    title: 'Mulai dari teks iklan yang sudah kamu punya',
    desc: 'Paste teks WhatsApp untuk membuat draft lebih cepat, lalu jadikan Propti sebagai pusat listing yang lebih rapi, lebih dipercaya, dan siap kamu bagikan.',
    tag: 'Buat listing',
  },
  {
    icon: <Users className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Catat calon pembeli dari pertama kontak hingga deal',
    desc: 'Setiap orang yang tertarik bisa dicatat dan dipindah ke tahap berikutnya: tertarik, survei, negosiasi, deal selesai — semua terlihat jelas dalam satu tampilan.',
    tag: 'Kelola calon pembeli',
  },
  {
    icon: <ShieldIcon className="w-6 h-6 text-blue-600" />,
    bg: 'bg-blue-50',
    title: 'Halaman iklan yang terasa lebih meyakinkan',
    desc: 'Iklan aktif melewati pemeriksaan dasar dan menampilkan detail penting dalam format yang lebih konsisten untuk calon pembeli.',
    tag: 'Lebih dipercaya',
  },
  {
    icon: <TrendingUp className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Statistik performa iklan & calon pembeli',
    desc: 'Lihat berapa banyak calon pembeli yang akhirnya deal, seberapa cepat kamu merespons, dan mana iklan yang paling banyak diminati.',
    tag: 'Statistik',
  },
];

const WHATSAPP_PERKS = [
  {
    icon: MessageCircle,
    title: 'Buat iklan via teks WhatsApp',
    desc: 'Kirim teks iklan ke nomor WhatsApp Propti — AI otomatis baca dan rapikan detail properti, lalu buat iklan siap tayang tanpa perlu buka website.',
  },
  {
    icon: Mic,
    title: 'Buat iklan via pesan suara',
    desc: 'Lebih cepat ngomong? Kirim pesan suara — Propti dengarkan, ubah jadi teks, dan rapikan jadi iklan properti dalam hitungan detik.',
  },
  {
    icon: Search,
    title: 'Cari properti via WhatsApp',
    desc: 'Ketik atau rekam perintah cari ke Propti — hasilnya langsung muncul di chat tanpa perlu buka halaman web.',
  },
];

const PRICING_PLANS = [
  {
    key: 'free',
    label: 'Gratis',
    price: 'Rp 0',
    period: '',
    blurb: 'Coba dulu, tanpa biaya.',
    highlight: false,
    features: [
      '5 listing aktif',
      'Hingga 5 foto per listing',
      'Buat listing via WhatsApp',
      'Listing tayang 60 hari',
    ],
    cta: 'Mulai Gratis',
    href: '/listings/create',
  },
  {
    key: 'premium',
    label: 'Premium',
    price: 'Rp 99.000',
    period: '/bulan',
    blurb: 'Agen aktif dengan listing rutin.',
    highlight: true,
    features: [
      'Kelola 25 listing aktif sekaligus',
      'Hingga 15 foto per listing',
      'Buat & cari via WhatsApp',
      'Voice note 90 menit/bulan',
      'Listing tayang 90 hari',
    ],
    cta: 'Pilih Premium',
    href: '/profile?upgradeTier=premium#premium',
  },
  {
    key: 'pro',
    label: 'Pro',
    price: 'Rp 179.000',
    period: '/bulan',
    blurb: 'Tim agen profesional.',
    highlight: false,
    features: [
      'Kelola 100 listing aktif sekaligus',
      'Hingga 25 foto per listing',
      'Buat & cari via WhatsApp',
      'Voice note 180 menit/bulan',
      'Listing tayang 90 hari',
    ],
    cta: 'Pilih Pro',
    href: '/profile?upgradeTier=pro#premium',
  },
] as const;

export default async function HomePage() {
  const { isAuthenticated, profile } = await getServerAuthProfile();
  const redirectPath = getAuthenticatedHomeRedirectPath({
    isAuthenticated,
    hasProfile: Boolean(profile),
  });
  if (redirectPath) {
    redirect(redirectPath);
  }

  const homepageSection = await getHomepageListings();

  return (
    <div className="bg-[#F8F9FA]">
      <section className="relative overflow-hidden bg-gradient-hero">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-24 -right-24 w-96 h-96 bg-white/5 rounded-full blur-3xl" />
          <div className="absolute -bottom-32 -left-32 w-[500px] h-[500px] bg-white/5 rounded-full blur-3xl" />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-white/3 rounded-full blur-3xl" />
        </div>

        <div className="relative max-w-6xl mx-auto px-4 pt-16 pb-24 md:pt-24 md:pb-32 text-center">
          <div className="inline-flex items-center gap-2 bg-white/15 backdrop-blur-sm border border-white/20 text-white text-sm font-medium px-4 py-2 rounded-full mb-6">
            <span className="w-2 h-2 bg-[#25D366] rounded-full animate-pulse" />
            Alat kerja properti untuk agen & pemilik
          </div>

          <h1 className="text-4xl md:text-5xl lg:text-6xl font-black text-white leading-tight mb-4 text-balance">
            Pasang Iklan Properti Lebih Rapi.
            <br />
            <span className="text-brand-accent">Catat Semua Calon Pembeli dengan Mudah.</span>
          </h1>

          <p className="text-lg md:text-xl text-white/80 max-w-3xl mx-auto mb-10 text-balance">
            Buat iklan properti rapi dari teks WhatsApp, catat semua calon pembeli di satu tempat, dan bagikan ke mana saja dari satu link. Bukan sekadar situs jual beli — ini alat kerja sehari-hari untuk agen dan pemilik properti.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-12">
            <Link
              href="/listings/create"
              className="group flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all duration-200 shadow-xl hover:shadow-2xl hover:-translate-y-0.5 text-lg"
            >
              <MessageCircle className="w-5 h-5 text-[#25D366]" />
              Paste listing saya
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/agent"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all duration-200 text-lg"
            >
              <LayoutDashboard className="w-5 h-5" />
              Lihat Dasbor Kerja Saya
            </Link>
          </div>

          <div className="grid gap-4 md:grid-cols-3 max-w-5xl mx-auto text-left">
            {HERO_PROOF_POINTS.map((point) => (
              <div
                key={point.title}
                className="rounded-2xl border border-white/15 bg-white/10 backdrop-blur-sm p-5 shadow-lg"
              >
                <div className="text-sm font-semibold text-white">{point.title}</div>
                <p className="mt-2 text-sm leading-relaxed text-white/75">{point.desc}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="absolute bottom-0 left-0 right-0">
          <svg viewBox="0 0 1440 60" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <path d="M0 60L1440 60L1440 30C1440 30 1080 0 720 0C360 0 0 30 0 30L0 60Z" fill="#F8F9FA" />
          </svg>
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
            Cara Kerja
          </span>
          <h2 className="section-title">Pasang Iklan dalam 3 Langkah</h2>
          <p className="section-subtitle max-w-xl mx-auto">
            Tidak perlu mulai dari nol. Cukup paste teks WhatsApp-mu, rapikan hasilnya, lalu
            terbitkan satu link listing yang siap dibagikan.
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto_minmax(0,1fr)] md:items-start">
          {HOW_IT_WORKS.map((step, index) => (
            <Fragment key={step.step}>
              <div className="text-center group">
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

              {index < HOW_IT_WORKS.length - 1 && (
                <div className="hidden md:flex items-center justify-center pt-20">
                  <div className="h-0.5 w-12 bg-gradient-to-r from-brand-accent to-brand-secondary rounded-full" />
                </div>
              )}
            </Fragment>
          ))}
        </div>

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
                <span className="text-sm font-semibold text-gray-600">Hasil Rapian AI Otomatis</span>
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
              Paste listing saya
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-4 py-12 md:py-16">
        <div className="bg-gradient-to-r from-[#25D366]/8 to-[#25D366]/3 border border-[#25D366]/20 rounded-3xl p-6 md:p-8">
          <div className="flex flex-col md:flex-row md:items-center gap-6 md:gap-10">
            <div className="md:w-64 flex-shrink-0">
              <span className="inline-flex items-center gap-1.5 bg-[#25D366]/15 text-[#128C7E] text-xs font-semibold px-3 py-1 rounded-full mb-3">
                <MessageCircle className="w-3 h-3" />
                Fitur WhatsApp Listing
              </span>
              <h2 className="text-xl font-bold text-gray-900 mb-2 leading-snug">
                Atau pasang listing langsung via WhatsApp
              </h2>
              <p className="text-sm text-gray-500 leading-relaxed">
                Kirim pesan ke nomor Propti — tanpa buka website, tanpa isi form.
              </p>
              <Link
                href="/profile"
                className="inline-flex items-center gap-1.5 mt-4 text-sm font-semibold text-[#128C7E] hover:text-[#25D366] transition-colors"
              >
                Hubungkan WhatsApp saya
                <ArrowRight className="w-3.5 h-3.5" />
              </Link>
            </div>

            <div className="flex-1 grid sm:grid-cols-3 gap-4">
              {WHATSAPP_PERKS.map((perk) => (
                <div key={perk.title} className="flex flex-col gap-2 bg-white/70 rounded-2xl p-4 border border-[#25D366]/10">
                  <div className="w-8 h-8 bg-[#25D366]/15 rounded-xl flex items-center justify-center">
                    <perk.icon className="w-4 h-4 text-[#128C7E]" />
                  </div>
                  <p className="text-sm font-semibold text-gray-800 leading-snug">{perk.title}</p>
                  <p className="text-xs text-gray-500 leading-relaxed">{perk.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Agent Workspace Spotlight */}
      <section className="bg-white py-16 md:py-24">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-12">
            <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
              Dasbor Kerja Agen
            </span>
            <h2 className="section-title">Bukan Sekadar Iklan — Ini Alat Kerja Harian Agen</h2>
            <p className="section-subtitle max-w-2xl mx-auto">
              Propti dirancang agar agen properti bisa kerja lebih terstruktur: dari pasang iklan,
              catat calon pembeli, sampai deal selesai — semua dari satu dasbor.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            {[
              {
                icon: <Users className="w-6 h-6 text-brand-primary" />,
                bg: 'bg-brand-light',
                title: 'Kelola Calon Pembeli dengan Mudah',
                desc: 'Geser calon pembeli dari "baru" → "tertarik" → "survei" → "deal selesai" dengan satu klik.',
              },
              {
                icon: <BarChart3 className="w-6 h-6 text-brand-secondary" />,
                bg: 'bg-brand-light',
                title: 'Statistik Penjualan',
                desc: 'Lihat berapa persen calon pembeli yang akhirnya beli, seberapa cepat kamu merespons, dan performa iklan secara keseluruhan.',
              },
              {
                icon: <MessageCircle className="w-6 h-6 text-[#25D366]" />,
                bg: 'bg-[#25D366]/10',
                title: 'Calon Pembeli Masuk via WhatsApp',
                desc: 'Setiap orang yang menghubungi via WhatsApp bisa langsung dicatat sebagai calon pembeli tanpa perlu input manual.',
              },
              {
                icon: <Calculator className="w-6 h-6 text-brand-primary" />,
                bg: 'bg-amber-50',
                title: 'Cek Budget Calon Pembeli',
                desc: 'Bantu calon pembeli memperkirakan kemampuan beli mereka langsung dari halaman iklan properti.',
              },
            ].map((item) => (
              <div key={item.title} className="card p-5">
                <div className={`w-12 h-12 ${item.bg} rounded-xl flex items-center justify-center mb-3`}>
                  {item.icon}
                </div>
                <h3 className="font-bold text-gray-900 mb-1 text-sm">{item.title}</h3>
                <p className="text-xs text-gray-500 leading-relaxed">{item.desc}</p>
              </div>
            ))}
          </div>

          <div className="bg-gradient-to-r from-brand-primary/5 to-brand-secondary/5 border border-brand-primary/10 rounded-3xl p-6 md:p-8 flex flex-col md:flex-row items-center gap-6">
            <div className="flex-1">
              <h3 className="text-xl font-bold text-gray-900 mb-2">
                Sudah punya iklan properti? Coba fitur kerja hariannya gratis
              </h3>
              <p className="text-gray-500 text-sm leading-relaxed">
                Semua fitur pencatatan calon pembeli, statistik, dan pengelolaan iklan tersedia mulai dari paket gratis. Tidak perlu kartu kredit.
              </p>
            </div>
            <div className="flex flex-col sm:flex-row gap-3 flex-shrink-0">
              <Link
                href="/agent"
                className="flex items-center justify-center gap-2 bg-brand-primary text-white font-bold px-6 py-3 rounded-xl hover:bg-brand-secondary transition-colors text-sm shadow-lg"
              >
                <LayoutDashboard className="w-4 h-4" />
                Buka Dasbor Kerja Saya
              </Link>
              <Link
                href="/listings/create"
                className="flex items-center justify-center gap-2 border-2 border-brand-primary text-brand-primary font-bold px-6 py-3 rounded-xl hover:bg-brand-light transition-colors text-sm"
              >
                <MessageCircle className="w-4 h-4" />
                Buat listing baru
              </Link>
            </div>
          </div>
        </div>
      </section>

      {homepageSection.items.length > 0 && (
        <section className="bg-[#F8F9FA] py-16 md:py-24">
          <div className="max-w-6xl mx-auto px-4">
            <div className="flex items-center justify-between mb-10">
              <div>
                <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-3">
                  Sudah Lolos Moderasi
                </span>
                <h2 className="section-title">{homepageSection.title}</h2>
                <p className="section-subtitle">{homepageSection.subtitle}</p>
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
              {homepageSection.items.map((listing) => (
                <ListingCard key={listing.listingId} listing={listing} />
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
      )}

      <section className="max-w-6xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <span className="inline-block bg-brand-light text-brand-primary text-sm font-semibold px-4 py-1.5 rounded-full mb-4">
            Fitur Lengkap
          </span>
          <h2 className="section-title">Semua yang Dibutuhkan Agen & Pemilik Properti</h2>
          <p className="section-subtitle max-w-2xl mx-auto">
            Dari membuat listing, mengelola calon pembeli, hingga mempresentasikan properti — semua tersedia di satu platform.
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {PRODUCT_PROOF.map((item) => (
            <div key={item.title} className="card p-6 flex gap-5">
              <div
                className={`w-14 h-14 ${item.bg} rounded-2xl flex items-center justify-center flex-shrink-0`}
              >
                {item.icon}
              </div>
              <div>
                <span className="text-xs font-semibold text-brand-secondary uppercase tracking-wide">
                  {item.tag}
                </span>
                <h3 className="text-lg font-bold text-gray-900 mt-1 mb-2">{item.title}</h3>
                <p className="text-gray-500 text-sm leading-relaxed">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* KPR Calculator section */}
      <section className="bg-white py-16 md:py-20">
        <div className="max-w-4xl mx-auto px-4">
          <div className="text-center mb-8">
            <span className="inline-block bg-amber-50 text-amber-700 text-sm font-semibold px-4 py-1.5 rounded-full mb-4 border border-amber-200">
              Tools Gratis
            </span>
            <h2 className="section-title">Kalkulator KPR</h2>
            <p className="section-subtitle max-w-xl mx-auto">
              Bantu calon pembeli memperkirakan kemampuan beli mereka sebelum survei. Tunjukkan saat presentasi properti agar mereka bisa siap lebih awal.
            </p>
          </div>
          <div className="card p-6 md:p-8">
            <div className="grid md:grid-cols-2 gap-6">
              <div className="bg-gray-50 rounded-2xl p-6 flex flex-col justify-center items-center text-center">
                <Calculator className="w-10 h-10 text-brand-primary mb-3 opacity-60" />
                <p className="text-gray-600 text-sm font-medium">
                  Kalkulator interaktif tersedia di halaman penuh
                </p>
                <p className="text-gray-400 text-xs mt-1 mb-4">
                  Hitung DP, estimasi cicilan, dan kemampuan beli calon pembeli dengan mudah.
                </p>
                <Link
                  href="/kpr"
                  className="inline-flex items-center gap-2 bg-brand-primary text-white font-semibold px-5 py-2.5 rounded-xl hover:bg-brand-secondary transition-colors text-sm"
                >
                  <Calculator className="w-4 h-4" />
                  Buka Kalkulator KPR
                  <ArrowRight className="w-4 h-4" />
                </Link>
              </div>
              <div className="space-y-3">
                {[
                  ['Harga Properti', 'Rp 750.000.000'],
                  ['Uang Muka (20%)', 'Rp 150.000.000'],
                  ['Jumlah Pinjaman', 'Rp 600.000.000'],
                  ['Bunga 10.5% / 15 thn', '→'],
                  ['Estimasi Angsuran', 'Rp 6,6 Jt / bulan'],
                ].map(([label, value]) => (
                  <div key={label} className={`flex items-center justify-between text-sm rounded-xl px-4 py-2.5 ${label === 'Estimasi Angsuran' ? 'bg-brand-light font-bold text-brand-primary' : 'bg-gray-50 text-gray-600'}`}>
                    <span>{label}</span>
                    <span className={label === 'Estimasi Angsuran' ? 'text-brand-primary' : 'font-medium text-gray-800'}>{value}</span>
                  </div>
                ))}
                <p className="text-xs text-gray-400 text-center">*contoh ilustrasi, gunakan kalkulator untuk angka aktual</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Pricing section */}
      <section className="bg-[#F8F9FA] py-16 md:py-24">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-12">
            <span className="inline-block bg-amber-50 text-amber-700 text-sm font-semibold px-4 py-1.5 rounded-full mb-4 border border-amber-200">
              Harga Paket
            </span>
            <h2 className="section-title">Pilih Paket Sesuai Aktivitas Kerjamu</h2>
            <p className="section-subtitle max-w-2xl mx-auto">
              Mulai gratis, upgrade kapan saja. Paket berbayar membuka lebih banyak listing aktif, lebih banyak foto per iklan, fitur WhatsApp bot, dan pembuatan iklan via pesan suara.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {PRICING_PLANS.map((plan) => (
              <div
                key={plan.key}
                className={`card p-6 flex flex-col relative ${
                  plan.highlight ? 'border-2 border-brand-gold shadow-xl' : ''
                }`}
              >
                {plan.highlight && (
                  <div className="absolute -top-3.5 left-1/2 -translate-x-1/2">
                    <span className="bg-brand-gold text-white text-xs font-bold px-3 py-1 rounded-full whitespace-nowrap flex items-center gap-1">
                      <Crown className="w-3 h-3" />
                      Paling Populer
                    </span>
                  </div>
                )}
                <div className="mb-5">
                  <div className="flex items-center gap-2 mb-2">
                    {plan.key !== 'free' && (
                      <Crown className={`w-4 h-4 ${plan.highlight ? 'text-brand-gold' : 'text-gray-400'}`} />
                    )}
                    <span className="font-bold text-gray-900">{plan.label}</span>
                  </div>
                  <div>
                    <span className="text-2xl font-black text-gray-900">{plan.price}</span>
                    {plan.period && <span className="text-gray-400 text-sm ml-1">{plan.period}</span>}
                  </div>
                  <p className="text-xs text-gray-500 mt-1">{plan.blurb}</p>
                </div>
                <ul className="space-y-2.5 flex-1 mb-6">
                  {plan.features.map((f) => (
                    <li key={f} className="flex items-start gap-2 text-sm text-gray-600">
                      <Check className="w-3.5 h-3.5 text-green-500 flex-shrink-0 mt-0.5" />
                      {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={plan.href}
                  className={`w-full flex items-center justify-center gap-2 py-3 px-4 rounded-xl font-semibold text-sm transition-all ${
                    plan.highlight
                      ? 'bg-brand-gold text-white hover:opacity-90 shadow-lg'
                      : plan.key === 'free'
                        ? 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                        : 'border-2 border-brand-gold text-amber-700 hover:bg-amber-50'
                  }`}
                >
                  {plan.key !== 'free' && <Crown className="w-4 h-4" />}
                  {plan.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="bg-gradient-hero py-16 md:py-20 relative overflow-hidden">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-16 -right-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
          <div className="absolute -bottom-16 -left-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
        </div>
        <div className="relative max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl md:text-4xl font-black text-white mb-4">
            Mulai kerja lebih terstruktur hari ini
          </h2>
          <p className="text-white/80 text-lg mb-8 max-w-2xl mx-auto">
            Buat iklan properti rapi dari teks WhatsApp, catat semua calon pembeli di satu tempat, dan raih lebih banyak transaksi di seluruh Indonesia.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/listings/create"
              className="flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all shadow-xl text-lg"
            >
              <MessageCircle className="w-5 h-5 text-[#25D366]" />
              Paste listing saya
            </Link>
            <Link
              href="/agent"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all text-lg"
            >
              <LayoutDashboard className="w-5 h-5" />
              Lihat Dasbor Kerja Saya
            </Link>
          </div>
          <p className="text-white/50 text-sm mt-6">
            Gratis untuk mulai. Tidak perlu kartu kredit.
          </p>
        </div>
      </section>
    </div>
  );
}
