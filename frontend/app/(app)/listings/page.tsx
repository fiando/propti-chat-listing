'use client';

import { useSession } from 'next-auth/react';
import { useMyListings, useRenewListing } from '@/hooks/useListings';
import { Plus, Loader2, Home, RefreshCw, Clock, AlertCircle } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ListingCard } from '@/components/listings/ListingCard';
import { formatDate } from '@/lib/utils';
import type { Listing } from '@/types';
import { useToast } from '@/app/toaster';

function listingExpiryStatus(listing: Listing): 'expired' | 'expiring-soon' | 'active' | 'no-expiry' {
  if (!listing.expiresAt) return 'no-expiry';
  const expiresAt = new Date(listing.expiresAt);
  const now = new Date();
  if (expiresAt <= now) return 'expired';
  const daysLeft = (expiresAt.getTime() - now.getTime()) / (1000 * 60 * 60 * 24);
  if (daysLeft <= 7) return 'expiring-soon';
  return 'active';
}

function ExpiryBadge({ listing }: { listing: Listing }) {
  const status = listingExpiryStatus(listing);
  if (status === 'no-expiry') return null;

  if (status === 'expired') {
    return (
      <div className="flex items-center gap-1.5 text-red-600 bg-red-50 border border-red-200 rounded-lg px-3 py-1.5 text-xs font-semibold">
        <AlertCircle className="w-3.5 h-3.5" />
        Iklan kadaluarsa
      </div>
    );
  }

  const expiresAt = new Date(listing.expiresAt!);
  if (status === 'expiring-soon') {
    return (
      <div className="flex items-center gap-1.5 text-orange-600 bg-orange-50 border border-orange-200 rounded-lg px-3 py-1.5 text-xs font-semibold">
        <Clock className="w-3.5 h-3.5" />
        Berakhir {formatDate(listing.expiresAt!)}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1.5 text-gray-400 text-xs">
      <Clock className="w-3 h-3" />
      Aktif hingga {formatDate(expiresAt.toISOString())}
    </div>
  );
}

function MyListingItem({ listing }: { listing: Listing }) {
  const { mutate: renew, isPending } = useRenewListing();
  const { toast } = useToast();
  const expiryStatus = listingExpiryStatus(listing);
  const isExpired = expiryStatus === 'expired';

  const handleRenew = (e: React.MouseEvent) => {
    e.preventDefault();
    renew(listing.listingId, {
      onSuccess: () => toast('Iklan berhasil diperpanjang', 'success'),
      onError: () => toast('Gagal memperpanjang iklan', 'error'),
    });
  };

  return (
    <div className="relative">
      {isExpired && (
        <div className="absolute inset-0 bg-white/60 rounded-2xl z-10 pointer-events-none" />
      )}
      <ListingCard listing={listing} />
      <div className="flex items-center justify-between mt-2 px-1">
        <ExpiryBadge listing={listing} />
        {isExpired && (
          <button
            type="button"
            onClick={handleRenew}
            disabled={isPending}
            className="flex items-center gap-1.5 bg-brand-primary text-white text-xs font-semibold px-3 py-1.5 rounded-lg hover:bg-brand-primary/90 transition-colors disabled:opacity-60 z-20 relative"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isPending ? 'animate-spin' : ''}`} />
            {isPending ? 'Memproses...' : 'Perpanjang'}
          </button>
        )}
      </div>
    </div>
  );
}

export default function MyListingsPage() {
  const { status } = useSession();
  const router = useRouter();
  const { data, isLoading } = useMyListings();

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  const listings = data?.items || [];
  const activeCount = listings.filter((l) => {
    const s = listingExpiryStatus(l);
    return s === 'active' || s === 'no-expiry';
  }).length;

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-black text-brand-primary">Iklan Saya</h1>
          <p className="text-gray-500 mt-1">
            {isLoading
              ? 'Memuat...'
              : `${activeCount} iklan aktif${listings.length > activeCount ? `, ${listings.length - activeCount} kadaluarsa` : ''}`}
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
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {listings.map((listing) => (
            <MyListingItem key={listing.listingId} listing={listing} />
          ))}
        </div>
      )}
    </div>
  );
}
