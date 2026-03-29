import type { Metadata } from 'next';
import Link from 'next/link';
import { Fragment } from 'react';
import { redirect } from 'next/navigation';
import {
  MessageCircle,
  Sparkles,
  Radio,
  ArrowRight,
  Home,
  TrendingUp,
  ChevronRight,
  Mic,
  Zap,
} from 'lucide-react';
import { ShieldIcon } from '@/components/icons/ShieldIcon';
import { ListingCard } from '@/components/listings/ListingCard';
import { getHomepageListings } from '@/lib/homepage-listings';
import { getAuthenticatedHomeRedirectPath } from '@/lib/home-redirect';
import { getServerAuthProfile } from '@/lib/server-profile';

export const metadata: Metadata = {
  title: 'Propti — Satu Listing Properti yang Lebih Rapi dan Siap Dibagikan',
};

const HERO_PROOF_POINTS = [
  {
    title: 'Satu listing properti yang lebih rapi',
    desc: 'Susun satu halaman listing yang lebih enak dilihat, lebih mudah dicek, dan lebih siap dikirim ke calon pembeli.',
  },
  {
    title: 'Lebih dipercaya sebelum dibagikan',
    desc: 'Detail penting, media, dan alur kontak ditata lebih jelas supaya listing terasa lebih serius dibanding posting acak di chat atau sosial media.',
  },
  {
    title: 'Siap dibagikan ke semua channel',
    desc: 'Setelah rapi di Propti, listing yang sama bisa kamu pakai sebagai halaman utama saat membagikan ke WhatsApp dan channel lain.',
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
    tag: 'Pusat listing',
  },
  {
    icon: <ShieldIcon className="w-6 h-6 text-blue-600" />,
    bg: 'bg-blue-50',
    title: 'Halaman listing yang terasa lebih meyakinkan',
    desc: 'Listing aktif melewati moderasi dasar dan menampilkan detail penting dalam format yang lebih konsisten untuk calon pembeli.',
    tag: 'Lebih dipercaya',
  },
  {
    icon: <Sparkles className="w-6 h-6 text-brand-gold" />,
    bg: 'bg-amber-50',
    title: 'Bukan isi ulang, cukup rapikan lalu terbitkan',
    desc: 'AI mengisi detail utama lebih dulu supaya owner dan agen tinggal merapikan bagian yang perlu dibenahi sebelum listing dibagikan.',
    tag: 'Cepat dipakai',
  },
  {
    icon: <TrendingUp className="w-6 h-6 text-brand-primary" />,
    bg: 'bg-brand-light',
    title: 'Satu tempat untuk siap tayang dan siap share',
    desc: 'Mulai dari pasang listing, media, moderasi, sampai alur kontak calon pembeli dibuat ringkas agar satu link Propti bisa jadi sumber utama listingmu.',
    tag: 'Siap dibagikan',
  },
];

const WHATSAPP_PERKS = [
  {
    icon: MessageCircle,
    title: 'Kirim teks, listing langsung terbuat',
    desc: 'Chat ke nomor WhatsApp Propti dengan teks iklan biasa — AI parse dan simpan listing otomatis tanpa perlu buka website.',
  },
  {
    icon: Mic,
    title: 'Voice note juga bisa',
    desc: 'Lebih cepat ngomong? Kirim pesan suara ke Propti, audio ditranskripsi lalu diparse jadi listing rapi dalam hitungan detik.',
  },
  {
    icon: Zap,
    title: 'Kelola dari WhatsApp',
    desc: 'Cek listing aktif, edit, hapus, atau cari properti — semua lewat perintah chat tanpa perlu login ke website.',
  },
];

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
            Pusat listing untuk owner & agen
          </div>

          <h1 className="text-4xl md:text-5xl lg:text-6xl font-black text-white leading-tight mb-4 text-balance">
            Satu Listing Properti
            <br />
            <span className="text-brand-accent">yang Lebih Rapi</span>
          </h1>

          <p className="text-lg md:text-xl text-white/80 max-w-3xl mx-auto mb-10 text-balance">
            Buat satu halaman listing yang lebih rapi, lebih dipercaya, dan siap dibagikan ke semua
            channel. Mulai dari teks WhatsApp, rapikan detailnya, lalu pakai link Propti sebagai
            pusat listingmu.
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
              href="/search"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all duration-200 text-lg"
            >
              <Home className="w-5 h-5" />
              Cari Properti
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
                <Zap className="w-3 h-3" />
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

      {homepageSection.items.length > 0 && (
        <section className="bg-white py-16 md:py-24">
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
            Bukti Produk
          </span>
          <h2 className="section-title">Yang Bisa kamu Cek Langsung di Propti</h2>
          <p className="section-subtitle max-w-2xl mx-auto">
            Fokus kami bukan klaim besar, tapi alur produk yang membantu kamu menyiapkan satu
            listing utama sebelum dibagikan ke channel lain.
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

      <section className="bg-gradient-hero py-16 md:py-20 relative overflow-hidden">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-16 -right-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
          <div className="absolute -bottom-16 -left-16 w-64 h-64 bg-white/5 rounded-full blur-2xl" />
        </div>
        <div className="relative max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl md:text-4xl font-black text-white mb-4">
            Siapkan satu listing, lalu bagikan ke mana saja
          </h2>
          <p className="text-white/80 text-lg mb-8 max-w-2xl mx-auto">
            Mulai dari satu teks iklan, rapikan jadi halaman listing yang lebih meyakinkan, lalu
            gunakan link Propti itu saat kamu promosi di WhatsApp dan channel lain.
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
              href="/search"
              className="flex items-center justify-center gap-2 border-2 border-white/40 text-white font-bold px-8 py-4 rounded-2xl hover:bg-white/10 transition-all text-lg"
            >
              <Home className="w-5 h-5" />
              Cari Properti
            </Link>
          </div>
          <p className="text-white/50 text-sm mt-6">
            Cocok untuk owner dan agen yang ingin punya satu link listing utama sebelum menyebarkan
            iklannya.
          </p>
        </div>
      </section>
    </div>
  );
}
