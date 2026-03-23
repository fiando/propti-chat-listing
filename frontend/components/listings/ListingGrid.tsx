import { ListingCard } from './ListingCard';
import type { Listing } from '@/types';
import { Home } from 'lucide-react';

interface ListingGridProps {
  listings: Listing[];
  savedIds?: string[];
  onSave?: (id: string) => void;
  onDelete?: (id: string) => void;
  onRelist?: (id: string) => void;
  onShareToWhatsApp?: (listing: Listing) => void;
  onCopyShareLink?: (listing: Listing) => void;
  deletingId?: string;
  relistingId?: string;
  emptyMessage?: string;
}

export function ListingGrid({
  listings,
  savedIds = [],
  onSave,
  onDelete,
  onRelist,
  onShareToWhatsApp,
  onCopyShareLink,
  deletingId,
  relistingId,
  emptyMessage = 'Belum ada properti yang ditemukan.',
}: ListingGridProps) {
  if (listings.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="w-20 h-20 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <Home className="w-10 h-10 text-gray-300" />
        </div>
        <p className="text-gray-500 font-medium">{emptyMessage}</p>
        <p className="text-gray-400 text-sm mt-1">Coba ubah filter pencarianmu</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      {listings.map((listing) => (
        <ListingCard
          key={listing.listingId}
          listing={listing}
          isSaved={savedIds.includes(listing.listingId)}
          onSave={onSave}
          onDelete={onDelete}
          onRelist={onRelist}
          onShareToWhatsApp={onShareToWhatsApp}
          onCopyShareLink={onCopyShareLink}
          isDeleting={deletingId === listing.listingId}
          isRelisting={relistingId === listing.listingId}
        />
      ))}
    </div>
  );
}
