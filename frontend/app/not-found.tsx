import Link from 'next/link';
import { ArrowLeft, Home, Search, TriangleAlert } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-[calc(100vh-120px)] bg-gradient-to-b from-brand-light/40 via-white to-white">
      <div className="max-w-4xl mx-auto px-4 py-16 md:py-24">
        <div className="card overflow-hidden border border-brand-secondary/10">
          <div className="bg-gradient-hero px-6 py-12 text-white md:px-10">
            <div className="inline-flex items-center gap-2 rounded-full bg-white/15 px-4 py-2 text-sm font-semibold">
              <TriangleAlert className="w-4 h-4" />
              Halaman tidak ditemukan
            </div>
            <h1 className="mt-6 text-4xl font-black tracking-tight md:text-5xl">404</h1>
            <p className="mt-4 max-w-2xl text-sm text-white/85 md:text-base">
              Maaf, halaman yang kamu cari tidak tersedia, sudah dipindahkan, atau tautannya salah.
            </p>
          </div>

          <div className="px-6 py-8 md:px-10">
            <div className="grid gap-4 md:grid-cols-3">
              {[
                {
                  icon: Home,
                  title: 'Kembali ke Beranda',
                  description: 'Lihat properti unggulan dan update terbaru dari Propti.',
                  href: '/',
                  label: 'Buka Beranda',
                },
                {
                  icon: Search,
                  title: 'Cari Properti',
                  description: 'Lanjutkan pencarian rumah, apartemen, tanah, atau gudang.',
                  href: '/search',
                  label: 'Mulai Cari',
                },
                {
                  icon: ArrowLeft,
                  title: 'Kembali ke Profil',
                  description: 'Balik ke area akun untuk melihat listing, simpanan, dan pengaturan.',
                  href: '/profile',
                  label: 'Buka Profil',
                },
              ].map((item) => (
                <div key={item.href} className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
                  <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-brand-light text-brand-primary">
                    <item.icon className="w-5 h-5" />
                  </div>
                  <h2 className="text-base font-bold text-gray-900">{item.title}</h2>
                  <p className="mt-2 text-sm text-gray-500">{item.description}</p>
                  <Link href={item.href} className="mt-5 inline-flex text-sm font-semibold text-brand-primary hover:text-brand-secondary">
                    {item.label}
                  </Link>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
