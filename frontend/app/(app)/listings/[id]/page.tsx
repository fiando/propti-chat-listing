'use client';

import { use } from 'react';
import { useListing, useSaveListing, useSavedListings } from '@/hooks/useListings';
import { ListingDetail } from '@/components/listings/ListingDetail';
import { useDeleteListing } from '@/hooks/useListings';
import { useRouter } from 'next/navigation';
import { Loader2, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { useSession } from 'next-auth/react';

export default function ListingDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: listing, isLoading, error } = useListing(id);
  const { mutateAsync: deleteListing, isPending: isDeleting } = useDeleteListing();
  const { mutateAsync: toggleSave, isPending: isSaving } = useSaveListing();
  const router = useRouter();
  const { data: session, status } = useSession();
  const { data: savedData } = useSavedListings({ enabled: status === 'authenticated' });

  const handleDelete = async () => {
    if (!confirm('Yakin ingin menghapus iklan ini?')) return;
    await deleteListing(id);
    router.push('/listings');
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
      </div>
    );
  }

  if (error || !listing) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center">
        <p className="text-gray-500 mb-4">Iklan tidak ditemukan atau sudah dihapus.</p>
        <Link href="/search" className="btn-primary inline-flex items-center gap-2">
          <ArrowLeft className="w-4 h-4" />
          Lihat Iklan Lainnya
        </Link>
      </div>
    );
  }

  const isOwner = session?.user && (listing.userId === (session as { user?: { id?: string } }).user?.id);
  const savedIds = savedData?.items.map((item) => item.listingId) ?? [];
  const isSaved = savedIds.includes(id);

  const handleSave = async () => {
    await toggleSave({ id, saved: isSaved });
  };

  return (
    <ListingDetail
      listing={listing}
      isOwner={isOwner || false}
      isSaved={isSaved}
      isSaving={isSaving}
      onSave={status === 'authenticated' ? handleSave : undefined}
      onDelete={isDeleting ? undefined : handleDelete}
    />
  );
}
