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
  title: 'Propti — Workspace Properti untuk Penjual & Pembeli di Indonesia',
  description:
    'Propti membantu penjual merapikan listing dan menindaklanjuti pembeli, sambil membantu pembeli mencari, menyimpan, dan menilai kecocokan properti dari satu tempat.',
};

const HERO_PROOF_POINTS = [
  {
    title: 'Penjual bisa mulai dari teks WhatsApp yang sudah ada',
    desc: 'Paste teks iklan yang sudah kamu punya — AI langsung baca dan rapikan detail penting supaya listing lebih cepat tayang.',
  },
  {
    title: 'Pembeli bisa cari, simpan, lalu cek kesiapan beli',
    desc: 'Cari properti dengan filter atau kalimat, simpan shortlist favorit, lalu cek kecocokan budget atau kesiapan beli sebelum lanjut survei.',
  },
  {
    title: 'Bukan portal listing massal — ini workspace transaksi',
    desc: 'Propti membantu setelah ada minat: listing tetap rapi, pembeli tetap bisa kembali, dan penjual tetap bisa follow-up sampai deal.',
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
    tag: 'Untuk penjual',
  },
  {
    icon: <Search className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Cari, simpan, dan shortlist properti yang relevan',
    desc: 'Pembeli bisa pakai search biasa atau cari dengan kalimat, lalu simpan properti yang cocok supaya mudah dibandingkan lagi nanti.',
    tag: 'Untuk pembeli',
  },
  {
    icon: <ShieldIcon className="w-6 h-6 text-blue-600" />,
    bg: 'bg-blue-50',
    title: 'Listing tampil lebih konsisten dan lebih mudah dipercaya',
    desc: 'Iklan aktif melewati pemeriksaan dasar dan menampilkan detail penting dalam format yang lebih konsisten untuk calon pembeli.',
    tag: 'Kepercayaan',
  },
  {
    icon: <Users className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Pembeli dan penjual lanjut dari minat ke keputusan',
    desc: 'Pembeli bisa cek kecocokan budget dan detail penting lebih awal, penjual bisa catat lead dan status follow-up supaya percakapan tidak hilang setelah klik pertama.',
    tag: 'Untuk closing',
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

const TRANSACTION_WORKSPACE = [
  {
    icon: <Search className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Cari dengan filter atau kalimat natural',
    desc: 'Pembeli tidak harus paham format filter. Cukup tulis kebutuhan, lalu rapikan hasilnya dengan filter lanjutan jika perlu.',
  },
  {
    icon: <TrendingUp className="w-6 h-6 text-brand-secondary" />,
    bg: 'bg-brand-light',
    title: 'Shortlist yang bisa dibuka lagi kapan saja',
    desc: 'Listing tersimpan membantu pembeli kembali ke properti yang menarik tanpa perlu mengulang pencarian dari awal.',
  },
  {
    icon: <Calculator className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-amber-50',
    title: 'Alat bantu hitung budget bila diperlukan',
    desc: 'Untuk pembeli yang memakai pembiayaan, simulasi KPR bisa membantu memperkirakan anggaran sebelum lanjut survei.',
  },
  {
    icon: <Users className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Follow-up calon pembeli tetap rapi',
    desc: 'Penjual dan agen bisa lihat siapa yang baru masuk, siapa yang siap survei, sampai siapa yang sudah deal.',
  },
] as const;

const BALANCED_MVP_LANES = [
  {
    title: 'Untuk penjual / agen',
    points: [
      'Buat listing cepat dari teks atau voice note WhatsApp',
      'Bagikan satu link listing yang lebih rapi dan lebih meyakinkan',
      'Catat lead dan pantau follow-up sampai deal selesai',
    ],
  },
  {
    title: 'Untuk pembeli',
    points: [
      'Cari properti dengan filter atau kalimat natural',
      'Simpan shortlist properti yang menarik untuk dibandingkan lagi',
      'Cek kecocokan budget lalu hubungi penjual dari halaman listing',
    ],
  },
] as const;

const PRICING_PLANS = [
  {
    key: 'free',
    label: 'Gratis',
    price: 'Rp 0',
    period: '',
    blurb: 'Coba dulu, tanpa biaya.',
    highlight: false,
    features: [
      '3 listing aktif',
      'Hingga 3 foto per listing',
      'Buat listing via WhatsApp',
      'Listing tayang 30 hari',
    ],
    cta: 'Mulai Gratis',
    href: '/listings/create',
  },
  {
    key: 'basic',
    label: 'Basic',
    price: 'Rp 59.000',
    period: '/bulan',
    blurb: 'Agen pemula aktif di lapangan.',
    highlight: false,
    features: [
      'Kelola 8 listing aktif sekaligus',
      'Hingga 8 foto per listing',
      'Buat & cari via WhatsApp',
      'Voice note 20 menit/bulan',
      'Listing tayang 90 hari',
    ],
    cta: 'Pilih Basic',
    href: '/profile?upgradeTier=basic#premium',
  },
  {
    key: 'premium',
    label: 'Premium',
    price: 'Rp 129.000',
    period: '/bulan',
    blurb: 'Agen aktif dengan listing rutin.',
    highlight: true,
    features: [
      'Kelola 20 listing aktif sekaligus',
      'Hingga 15 foto per listing',
      'Buat & cari via WhatsApp',
      'Voice note 60 menit/bulan',
      'Listing tayang 90 hari',
    ],
    cta: 'Pilih Premium',
    href: '/profile?upgradeTier=premium#premium',
  },
  {
    key: 'pro',
    label: 'Pro',
    price: 'Rp 199.000',
    period: '/bulan',
    blurb: 'Tim agen profesional.',
    highlight: false,
    features: [
      'Kelola 50 listing aktif sekaligus',
      'Hingga 25 foto per listing',
      'Buat & cari via WhatsApp',
      'Voice note 120 menit/bulan',
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
            Workspace properti untuk penjual & pembeli
          </div>

          <h1 className="text-4xl md:text-5xl lg:text-6xl font-black text-white leading-tight mb-4 text-balance">
            Pasang Listing Lebih Rapi.
            <br />
            <span className="text-brand-accent">Cari Properti Lebih Yakin.</span>
          </h1>

          <p className="text-lg md:text-xl text-white/80 max-w-3xl mx-auto mb-10 text-balance">
            Untuk penjual, Propti membantu merapikan listing, membagikannya ke mana saja, dan menindaklanjuti pembeli. Untuk pembeli, Propti membantu mencari, menyimpan, dan mengecek kecocokan properti dari satu tempat. Bukan portal listing massal — ini workspace transaksi properti.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-12">
            <Link
              href="/search"
              className="group flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all duration-200 shadow-xl hover:shadow-2xl hover:-translate-y-0.5 text-lg"
            >
              <Search className="w-5 h-5" />
              Cari properti
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/listings/create"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all duration-200 text-lg"
            >
              <MessageCircle className="w-5 h-5 text-[#25D366]" />
              Buat listing saya
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
          <h2 className="section-title">Penjual bisa tayang cepat, pembeli bisa cek detail lebih cepat</h2>
          <p className="section-subtitle max-w-xl mx-auto">
            Penjual bisa mulai dari teks WhatsApp yang sudah ada. Setelah tayang, pembeli tinggal buka
            listing yang lebih rapi, simpan properti yang cocok, lalu lanjut hubungi penjual atau cek budget bila perlu.
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
                Fitur WhatsApp
              </span>
              <h2 className="text-xl font-bold text-gray-900 mb-2 leading-snug">
                WhatsApp tetap jadi jalur cepat untuk penjual & pembeli
              </h2>
              <p className="text-sm text-gray-500 leading-relaxed">
                Buat listing, kirim voice note, atau cari properti langsung dari chat saat kamu belum
                sempat buka website.
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
              Workspace Transaksi
            </span>
            <h2 className="section-title">Kenapa Propti bukan sekadar portal listing</h2>
            <p className="section-subtitle max-w-2xl mx-auto">
              Portal listing berhenti di impresi. Propti fokus pada langkah setelah orang tertarik:
              pembeli bisa shortlist dan cek kesiapan beli, penjual bisa follow-up sampai deal.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            {TRANSACTION_WORKSPACE.map((item) => (
              <div key={item.title} className="card p-5">
                <div className={`w-12 h-12 ${item.bg} rounded-xl flex items-center justify-center mb-3`}>
                  {item.icon}
                </div>
                <h3 className="font-bold text-gray-900 mb-1 text-sm">{item.title}</h3>
                <p className="text-xs text-gray-500 leading-relaxed">{item.desc}</p>
              </div>
            ))}
          </div>

          <div className="bg-gradient-to-r from-brand-primary/5 to-brand-secondary/5 border border-brand-primary/10 rounded-3xl p-6 md:p-8">
            <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-6 mb-6">
              <div className="flex-1">
                <h3 className="text-xl font-bold text-gray-900 mb-2">
                  MVP yang seimbang untuk dua sisi transaksi
                </h3>
                <p className="text-gray-500 text-sm leading-relaxed">
                  Penjual butuh listing cepat dan follow-up rapi. Pembeli butuh pencarian, shortlist,
                  dan alat bantu keputusan. Itu sebabnya Propti tidak berhenti di halaman iklan.
                </p>
              </div>
              <div className="flex flex-col sm:flex-row gap-3 flex-shrink-0">
                <Link
                  href="/search"
                  className="flex items-center justify-center gap-2 bg-brand-primary text-white font-bold px-6 py-3 rounded-xl hover:bg-brand-secondary transition-colors text-sm shadow-lg"
                >
                  <Search className="w-4 h-4" />
                  Jelajahi Properti
                </Link>
                <Link
                  href="/agent"
                  className="flex items-center justify-center gap-2 border-2 border-brand-primary text-brand-primary font-bold px-6 py-3 rounded-xl hover:bg-brand-light transition-colors text-sm"
                >
                  <LayoutDashboard className="w-4 h-4" />
                  Lihat Dasbor Penjual
                </Link>
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              {BALANCED_MVP_LANES.map((lane) => (
                <div key={lane.title} className="rounded-2xl border border-white/70 bg-white/80 p-5 shadow-sm">
                  <h4 className="font-bold text-gray-900 mb-3">{lane.title}</h4>
                  <ul className="space-y-2.5">
                    {lane.points.map((point) => (
                      <li key={point} className="flex items-start gap-2 text-sm text-gray-600">
                        <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                        {point}
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
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
            Fitur MVP
          </span>
          <h2 className="section-title">Fitur yang harus ada untuk pembeli & penjual sejak launch</h2>
          <p className="section-subtitle max-w-2xl mx-auto">
            Penjual butuh listing yang rapi dan follow-up yang jelas. Pembeli butuh pencarian, shortlist,
            dan alat bantu keputusan tanpa diasumsikan selalu memakai KPR. Semua ini jadi inti produk Propti.
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
              Untuk pembeli yang membutuhkannya, kalkulator ini membantu memperkirakan angsuran bulanan sebelum survei. Tetap opsional, bukan langkah wajib untuk semua transaksi.
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
                  Hitung DP, angsuran bulanan, dan total bunga KPR dengan mudah.
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
              Fitur pembeli seperti cari, simpan, dan alat bantu keputusan tetap bisa dipakai gratis. Paket
              berbayar dipakai saat kamu aktif pasang dan kelola listing dalam jumlah lebih besar.
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
            Mulai dari sisi yang kamu butuhkan hari ini
          </h2>
          <p className="text-white/80 text-lg mb-8 max-w-2xl mx-auto">
            Cari properti, simpan shortlist, cek budget bila perlu, atau mulai dari listing yang sudah kamu punya.
            Propti membantu pembeli dan penjual bertemu di workflow yang lebih rapi.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/search"
              className="flex items-center justify-center gap-2 bg-white text-brand-primary font-bold px-8 py-4 rounded-2xl hover:bg-brand-light transition-all shadow-xl text-lg"
            >
              <Search className="w-5 h-5" />
              Cari properti
            </Link>
            <Link
              href="/listings/create"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all text-lg"
            >
              <MessageCircle className="w-5 h-5 text-[#25D366]" />
              Buat listing saya
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
