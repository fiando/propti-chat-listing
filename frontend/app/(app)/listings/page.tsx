'use client';

import { useSession } from 'next-auth/react';
import { useMyListings } from '@/hooks/useListings';
import { ListingGrid } from '@/components/listings/ListingGrid';
import { Plus, Loader2, Home } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

export default function MyListingsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();
  const { data, isLoading } = useMyListings();

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  const listings = data?.items || [];

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
            <h1 className="text-2xl font-black text-brand-primary">Iklan Saya</h1>
            <p className="text-gray-500 mt-1">
              {isLoading
                ? 'Memuat...'
                : `${listings.length} iklan`}
            </p>
          </div>
        <Link href="/listings/create" className="btn-primary flex items-center gap-2 text-sm">
          <Plus className="w-4 h-4" />
          Pasang Iklan Baru
        </Link>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
        </div>
      ) : listings.length === 0 ? (
        <div className="text-center py-16">
          <div className="w-20 h-20 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <Home className="w-10 h-10 text-gray-300" />
          </div>
          <h3 className="text-lg font-bold text-gray-900 mb-2">
            Belum Ada Iklan
          </h3>
          <p className="text-gray-500 mb-6">
            Pasang iklan pertamamu sekarang dan jangkau ribuan calon pembeli
          </p>
          <Link href="/listings/create" className="btn-primary inline-flex items-center gap-2">
            <Plus className="w-4 h-4" />
            Pasang Iklan Gratis
          </Link>
        </div>
      ) : (
        <ListingGrid
          listings={listings}
          emptyMessage="Belum ada iklan."
        />
      )}
    </div>
  );
}
