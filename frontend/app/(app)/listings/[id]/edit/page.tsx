'use client';

import { use } from 'react';
import { useListing, useUpdateListing } from '@/hooks/useListings';
import { ListingForm } from '@/components/listings/ListingForm';
import { useRouter } from 'next/navigation';
import { Loader2, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import type { CreateListingRequest } from '@/types';
import type { ListingFormValues } from '@/components/listings/ListingForm';

export default function EditListingPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const { data: listing, isLoading } = useListing(resolvedParams.id);
  const { mutateAsync: updateListing, isPending } = useUpdateListing(resolvedParams.id);
  const router = useRouter();

  const handleSubmit = async (data: ListingFormValues) => {
    const payload: Partial<CreateListingRequest> = {
      title: data.title,
      description: data.description,
      price: data.price,
      priceUnit: data.priceUnit,
      listingType: data.listingType,
      propertyDetails: {
        landArea: data.landArea || 0,
        buildingArea: data.buildingArea || 0,
        bedrooms: data.bedrooms,
        bathrooms: data.bathrooms,
        frontWidth: data.frontWidth,
        orientation: data.orientation,
        legalStatus: data.legalStatus,
        powerConsumption: data.powerConsumption,
        amenities: data.amenities,
      },
      location: {
        address: data.address,
        province: data.province,
        city: data.city,
        district: data.district,
      },
    };
    await updateListing(payload);
    router.push(`/listings/${resolvedParams.id}`);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 text-brand-primary animate-spin" />
      </div>
    );
  }

  if (!listing) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center">
        <p className="text-gray-500 mb-4">Iklan tidak ditemukan.</p>
        <Link href="/listings" className="btn-secondary inline-flex items-center gap-2">
          <ArrowLeft className="w-4 h-4" />
          Kembali ke Iklan Saya
        </Link>
      </div>
    );
  }

  const initialData = {
    title: listing.title,
    description: listing.description,
    price: listing.price,
    priceUnit: listing.priceUnit,
    propertyDetails: listing.propertyDetails,
    address: listing.location?.address || '',
  };

  const initialLocation = {
    address: listing.location?.address || '',
    province: listing.location?.province || '',
    city: listing.location?.city || '',
    district: listing.location?.district || '',
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="flex items-center gap-3 mb-8">
        <Link
          href={`/listings/${resolvedParams.id}`}
          className="flex items-center gap-2 text-gray-500 hover:text-gray-700 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          Kembali
        </Link>
        <h1 className="text-2xl font-black text-brand-primary">Edit Iklan</h1>
      </div>

      <ListingForm
        initialData={initialData}
        initialLocation={initialLocation}
        listingId={resolvedParams.id}
        onSubmit={handleSubmit}
        isLoading={isPending}
        mode="edit"
      />
    </div>
  );
}
