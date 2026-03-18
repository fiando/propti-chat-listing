'use client';

import Link from 'next/link';
import { Heart, MapPin, Bed, Bath, Maximize2, Home, Eye, Star } from 'lucide-react';
import { cn, formatPrice, LISTING_TYPE_LABELS } from '@/lib/utils';
import type { Listing } from '@/types';
import { getListingCardImage } from '@/lib/listing-images';

interface ListingCardProps {
  listing: Listing;
  onSave?: (id: string) => void;
  isSaved?: boolean;
}

export function ListingCard({ listing, onSave, isSaved = false }: ListingCardProps) {
  const handleSave = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onSave?.(listing.listingId);
  };

  const priceLabel =
    listing.listingType === 'rent'
      ? `${formatPrice(listing.price)}/bln`
      : formatPrice(listing.price);

  const typeLabel = LISTING_TYPE_LABELS[listing.listingType];
  const typeBg = listing.listingType === 'sell' ? 'bg-brand-primary' : 'bg-blue-600';
  const listingImage = getListingCardImage(listing);

  return (
    <Link href={`/listings/${listing.listingId}`} className="card block group cursor-pointer">
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
          <span className={`${typeBg} text-white text-xs font-bold px-2.5 py-1 rounded-full`}>
            {typeLabel}
          </span>
          {listing.premiumFeatures?.isFeatured && (
            <span className="bg-brand-gold text-white text-xs font-bold px-2.5 py-1 rounded-full flex items-center gap-1">
              <Star className="w-3 h-3" />
              Unggulan
            </span>
          )}
        </div>

        {/* Save button */}
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

        {/* Views */}
        {listing.views > 0 && (
          <div className="absolute bottom-3 right-3 flex items-center gap-1 bg-black/50 text-white text-xs px-2 py-1 rounded-full">
            <Eye className="w-3 h-3" />
            {listing.views.toLocaleString()}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        <h3 className="font-semibold text-gray-900 group-hover:text-brand-primary transition-colors line-clamp-1 mb-1">
          {listing.title}
        </h3>
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
      </div>
    </Link>
  );
}
