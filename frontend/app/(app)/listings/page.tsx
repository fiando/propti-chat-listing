'use client';

import { useSession } from 'next-auth/react';
import { useMyListings, useDeleteListing, useRelistListing } from '@/hooks/useListings';
import { ListingGrid } from '@/components/listings/ListingGrid';
import { Plus, Loader2, Home, Eye, Heart } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { useToast } from '@/app/toaster';
import type { Listing } from '@/types';
import {
  buildListingShareMessage,
  buildListingShareUrl,
  buildWhatsAppShareUrl,
  summarizeOwnerListings,
} from '@/lib/listing-share';

export default function MyListingsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();
  const { data, isLoading } = useMyListings();
  const { mutateAsync: deleteListing } = useDeleteListing();
  const { mutateAsync: relistListing } = useRelistListing();
  const [deletingId, setDeletingId] = useState<string | undefined>();
  const [relistingId, setRelistingId] = useState<string | undefined>();
  const { toast } = useToast();

  if (status === 'unauthenticated') {
    router.replace('/login');
    return null;
  }

  const handleDelete = async (id: string) => {
    setDeletingId(id);
    try {
      await deleteListing(id);
    } finally {
      setDeletingId(undefined);
    }
  };

  const handleRelist = async (id: string) => {
    setRelistingId(id);
    try {
      await relistListing(id);
    } finally {
      setRelistingId(undefined);
    }
  };

  const listings = data?.items || [];
  const { totalViews, totalSaves, activeListings } = summarizeOwnerListings(listings);

  const handleShareToWhatsApp = (listing: Listing) => {
    if (typeof window === 'undefined') {
      return;
    }

    const listingUrl = buildListingShareUrl(listing.listingId, window.location.origin);
    const whatsappUrl = buildWhatsAppShareUrl(buildListingShareMessage(listing, listingUrl));
    window.open(whatsappUrl, '_blank', 'noopener,noreferrer');
  };

  const handleCopyShareLink = async (listing: Listing) => {
    if (typeof window === 'undefined') {
      return;
    }

    try {
      await navigator.clipboard.writeText(buildListingShareUrl(listing.listingId, window.location.origin));
      toast('Link iklan berhasil disalin.', 'success');
    } catch {
      toast('Gagal menyalin link iklan.', 'error');
    }
  };

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
          Pasang Iklan
        </Link>
      </div>

      {!isLoading && listings.length > 0 && (
        <div className="mb-8 grid gap-4 sm:grid-cols-3">
          {[
            { icon: Home, label: 'Iklan Aktif', value: activeListings },
            { icon: Eye, label: 'Total Dilihat', value: totalViews },
            { icon: Heart, label: 'Total Disimpan', value: totalSaves },
          ].map((item) => (
            <div key={item.label} className="card p-5">
              <div className="flex items-center gap-3">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-brand-light">
                  <item.icon className="h-5 w-5 text-brand-primary" />
                </div>
                <div>
                  <p className="text-sm text-gray-500">{item.label}</p>
                  <p className="text-2xl font-black text-gray-900">{item.value}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

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
            Pasang Iklan
          </Link>
        </div>
      ) : (
        <ListingGrid
          listings={listings}
          emptyMessage="Belum ada iklan."
          onDelete={handleDelete}
          onRelist={handleRelist}
          onShareToWhatsApp={handleShareToWhatsApp}
          onCopyShareLink={handleCopyShareLink}
          deletingId={deletingId}
          relistingId={relistingId}
        />
      )}
    </div>
  );
}
