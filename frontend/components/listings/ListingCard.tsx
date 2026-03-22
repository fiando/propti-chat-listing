'use client';

import Link from 'next/link';
import { Heart, MapPin, Bed, Bath, Maximize2, Home, Eye, Star, Trash2, Loader2, RefreshCcw } from 'lucide-react';
import { cn, formatPrice, LISTING_TYPE_LABELS } from '@/lib/utils';
import type { Listing } from '@/types';
import { getListingCardImage } from '@/lib/listing-images';
import { getListingExpiryInfo, shouldShowRelistAction } from '@/lib/listing-expiry';

interface ListingCardProps {
  listing: Listing;
  onSave?: (id: string) => void;
  isSaved?: boolean;
  onDelete?: (id: string) => void;
  onRelist?: (id: string) => void;
  isDeleting?: boolean;
  isRelisting?: boolean;
}

export function ListingCard({
  listing,
  onSave,
  isSaved = false,
  onDelete,
  onRelist,
  isDeleting = false,
  isRelisting = false,
}: ListingCardProps) {
  const handleSave = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onSave?.(listing.listingId);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onDelete?.(listing.listingId);
  };

  const handleRelist = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onRelist?.(listing.listingId);
  };

  const isRejected = listing.moderationStatus === 'rejected';
  const isPending = listing.moderationStatus === 'pending';
  const isModerationHidden = isRejected;
  const priceLabel =
    listing.listingType === 'rent'
      ? `${formatPrice(listing.price)}/bln`
      : formatPrice(listing.price);

  const typeLabel = LISTING_TYPE_LABELS[listing.listingType];
  const typeBg = listing.listingType === 'sell' ? 'bg-brand-primary' : 'bg-blue-600';
  const listingImage = isModerationHidden ? undefined : getListingCardImage(listing);
  const moderationCopy =
    listing.moderationStatus === 'rejected'
      ? {
          badge: 'Ditolak',
          title: 'Iklan Ditarik Otomatis',
          message: 'Konten ini tidak lolos pemeriksaan otomatis kami. Hapus dan buat iklan baru.',
          tone: 'bg-red-50 text-red-700 border-red-200',
        }
      : {
          badge: 'Sedang direview',
          title: 'Sedang direview',
          message: 'Konten belum tampil di Propti sampai review selesai.',
          tone: 'bg-amber-50 text-amber-700 border-amber-200',
        };
  const expiryInfo = getListingExpiryInfo(listing);
  const isExpiredArchivedListing = shouldShowRelistAction(listing);
  const isOwnerListingCard = Boolean(onDelete || onRelist);
  const href = isExpiredArchivedListing ? `/listings/${listing.listingId}/edit` : `/listings/${listing.listingId}`;

  return (
    <Link href={href} className="card block group cursor-pointer">
      {/* Image */}
      <div className="relative h-48 bg-gradient-to-br from-brand-primary/20 to-brand-secondary/30 rounded-t-2xl overflow-hidden">
        {listingImage ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={listingImage}
            alt={listing.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <Home className="w-16 h-16 text-brand-secondary/30" />
          </div>
        )}

        {/* Badges */}
        <div className="absolute top-3 left-3 flex gap-2">
          {isModerationHidden ? (
            <span className={`border text-xs font-bold px-2.5 py-1 rounded-full ${moderationCopy.tone}`}>
              {moderationCopy.badge}
            </span>
          ) : (
            <>
              <span className={`${typeBg} text-white text-xs font-bold px-2.5 py-1 rounded-full`}>
                {typeLabel}
              </span>
              {isPending && (
                <span className="border text-xs font-bold px-2.5 py-1 rounded-full bg-amber-50 text-amber-700 border-amber-200">
                  Sedang diproses
                </span>
              )}
              {isOwnerListingCard && expiryInfo && (
                <span className={`border text-xs font-bold px-2.5 py-1 rounded-full ${expiryInfo.tone}`}>
                  {expiryInfo.label}
                </span>
              )}
              {!isPending && listing.premiumFeatures?.isFeatured && (
                <span className="bg-brand-gold text-white text-xs font-bold px-2.5 py-1 rounded-full flex items-center gap-1">
                  <Star className="w-3 h-3" />
                  Unggulan
                </span>
              )}
            </>
          )}
        </div>

        {/* Save button */}
        {!isModerationHidden && (
          <button
            onClick={handleSave}
            className={cn(
              'absolute top-3 right-3 w-8 h-8 rounded-full flex items-center justify-center transition-all duration-200',
              isSaved
                ? 'bg-red-500 text-white shadow-md'
                : 'bg-white/90 text-gray-500 hover:bg-white hover:text-red-400 shadow-sm'
            )}
            disabled={!onSave}
            aria-label={isSaved ? 'Hapus dari tersimpan' : 'Simpan properti'}
          >
            <Heart className={cn('w-4 h-4', isSaved && 'fill-white')} />
          </button>
        )}

        {/* Views */}
        {!isModerationHidden && listing.views > 0 && (
          <div className="absolute bottom-3 right-3 flex items-center gap-1 bg-black/50 text-white text-xs px-2 py-1 rounded-full">
            <Eye className="w-3 h-3" />
            {listing.views.toLocaleString()}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        {isModerationHidden ? (
          <div className={`rounded-2xl border px-4 py-4 ${moderationCopy.tone}`}>
            <p className="font-semibold">{moderationCopy.title}</p>
            <p className="mt-1 text-sm leading-6">{moderationCopy.message}</p>
            {onDelete && listing.moderationStatus === 'rejected' && (
              <button
                onClick={handleDelete}
                disabled={isDeleting}
                className="mt-3 flex items-center gap-1.5 text-sm font-medium text-red-600 hover:text-red-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isDeleting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Trash2 className="w-3.5 h-3.5" />}
                Hapus iklan
              </button>
            )}
          </div>
        ) : (
          <>
            <h3 className="font-semibold text-gray-900 group-hover:text-brand-primary transition-colors line-clamp-1 mb-1">
              {listing.title}
            </h3>
            {isPending && (
              <p className="text-xs text-amber-600 mb-1">Belum tampil ke publik · sedang diproses</p>
            )}
            <p className="text-brand-primary font-bold text-lg mb-2">{priceLabel}</p>

             {listing.location?.city && (
               <div className="flex items-center gap-1 text-gray-500 text-xs mb-3">
                <MapPin className="w-3 h-3 flex-shrink-0" />
                <span className="truncate">
                  {listing.location.district
                    ? `${listing.location.district}, ${listing.location.city}`
                    : listing.location.city}
                </span>
                </div>
              )}
            {isOwnerListingCard && expiryInfo && (
              <div className={`mb-3 rounded-xl border px-3 py-2 text-xs ${expiryInfo.tone}`}>
                <p className="font-semibold">{expiryInfo.label}</p>
                <p className="mt-1">{expiryInfo.detail}</p>
              </div>
            )}

             <div className="flex items-center gap-3 text-xs text-gray-500 border-t border-gray-50 pt-3 flex-wrap">
              {listing.propertyDetails?.landArea > 0 && (
                <span className="flex items-center gap-1">
                  <Maximize2 className="w-3 h-3" />
                  LT {listing.propertyDetails.landArea} m²
                </span>
              )}
              {listing.propertyDetails?.buildingArea > 0 && (
                <span className="flex items-center gap-1">
                  <Home className="w-3 h-3" />
                  LB {listing.propertyDetails.buildingArea} m²
                </span>
              )}
              {listing.propertyDetails?.bedrooms > 0 && (
                <span className="flex items-center gap-1">
                  <Bed className="w-3 h-3" />
                  {listing.propertyDetails.bedrooms} KT
                </span>
              )}
              {listing.propertyDetails?.bathrooms > 0 && (
                <span className="flex items-center gap-1">
                  <Bath className="w-3 h-3" />
                  {listing.propertyDetails.bathrooms} KM
                </span>
              )}
            </div>
            {onRelist && isExpiredArchivedListing && (
              <button
                onClick={handleRelist}
                disabled={isRelisting}
                className="mt-3 inline-flex items-center gap-2 rounded-full border border-brand-primary/20 bg-brand-primary/5 px-3 py-2 text-sm font-semibold text-brand-primary transition-colors hover:bg-brand-primary/10 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isRelisting ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCcw className="w-4 h-4" />}
                Relist Iklan
              </button>
            )}
          </>
        )}
      </div>
    </Link>
  );
}
