'use client';

import { useSession } from 'next-auth/react';
import { Heart, Loader2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ListingGrid } from '@/components/listings/ListingGrid';
import { useSaveListing, useSavedListings } from '@/hooks/useListings';

export default function SavedPage() {
  const { status } = useSession();
  const router = useRouter();
  const { data, isLoading } = useSavedListings({ enabled: status === 'authenticated' });
  const { mutateAsync: toggleSave, isPending: isSaving } = useSaveListing();

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  const listings = data?.items ?? [];
  const savedIds = listings.map((listing) => listing.listingId);

  const handleSave = async (listingId: string) => {
    await toggleSave({ id: listingId, saved: savedIds.includes(listingId) });
  };

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <Heart className="w-6 h-6 text-red-500 fill-red-500" />
          <h1 className="text-2xl font-black text-brand-primary">Properti Tersimpan</h1>
        </div>
        <p className="text-gray-500">
          {isLoading ? 'Memuat properti tersimpan...' : `${listings.length} properti tersimpan`}
        </p>
      </div>

      {isLoading || status === 'loading' ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
        </div>
      ) : listings.length === 0 ? (
        <div className="text-center py-16">
          <div className="w-20 h-20 bg-red-50 rounded-full flex items-center justify-center mx-auto mb-6">
            <Heart className="w-10 h-10 text-red-200" />
          </div>
          <h3 className="text-lg font-bold text-gray-900 mb-2">Belum Ada Properti Tersimpan</h3>
          <p className="text-gray-500 mb-6">
            Simpan listing yang kamu suka agar lebih mudah dibuka lagi nanti.
          </p>
          <Link href="/search" className="btn-primary inline-flex">
            Cari Properti
          </Link>
        </div>
      ) : (
        <ListingGrid listings={listings} savedIds={savedIds} onSave={isSaving ? undefined : handleSave} />
      )}
    </div>
  );
}
