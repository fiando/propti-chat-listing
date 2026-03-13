'use client';

import { useSession } from 'next-auth/react';
import { Heart } from 'lucide-react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function SavedPage() {
  const { status } = useSession();
  const router = useRouter();

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <Heart className="w-6 h-6 text-red-500 fill-red-500" />
          <h1 className="text-2xl font-black text-brand-primary">Properti Tersimpan</h1>
        </div>
        <p className="text-gray-500">Fitur simpan listing sedang disiapkan.</p>
      </div>

      <div className="text-center py-16">
        <div className="w-20 h-20 bg-red-50 rounded-full flex items-center justify-center mx-auto mb-6">
          <Heart className="w-10 h-10 text-red-200" />
        </div>
        <h3 className="text-lg font-bold text-gray-900 mb-2">Segera Hadir</h3>
        <p className="text-gray-500 mb-6">
          Untuk saat ini, halaman ini tidak lagi memanggil endpoint yang belum tersedia di backend.
        </p>
        <Link href="/search" className="btn-primary inline-flex">
          Jelajahi Properti
        </Link>
      </div>
    </div>
  );
}
