'use client';

import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from '@tanstack/react-query';
import {
  getListings,
  getListing,
  getOwnerListing,
  trackListingView,
  revealListingContact,
  createListing,
  updateListing,
  relistListing,
  deleteListing,
  getMyListings,
  saveListing,
  unsaveListing,
  getSavedListings,
} from '@/lib/api';
import { collectActiveListingCount } from '@/lib/my-listing-access';
import type {
  Listing,
  SearchParams,
  CreateListingRequest,
  UpdateListingRequest,
  ContactRevealChannel,
} from '@/types';
import { LISTING_ACCESS_CHECK_PAGE_SIZE } from '@/lib/create-listing-errors';

export function useListings(params: SearchParams = {}) {
  return useQuery({
    queryKey: ['listings', params],
    queryFn: () => getListings(params),
    placeholderData: keepPreviousData,
  });
}

export function useListing(id: string) {
  return useQuery({
    queryKey: ['listing', id],
    queryFn: () => getListing(id),
    enabled: !!id,
  });
}

export function useOwnerListing(id: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['owner-listing', id],
    queryFn: () => getOwnerListing(id),
    enabled: Boolean(id) && (options?.enabled ?? true),
  });
}

export function useTrackListingView() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => trackListingView(id),
    onSuccess: (listing) => {
      queryClient.setQueryData(['listing', listing.listingId], (previous: Listing | undefined) => {
        if (!previous) {
          return listing;
        }

        return {
          ...previous,
          ...listing,
          sellerName: previous?.sellerName ?? listing.sellerName,
          hasSellerPhone: previous?.hasSellerPhone ?? listing.hasSellerPhone,
        };
      });
    },
  });
}

export function useRevealListingContact() {
  return useMutation({
    mutationFn: ({ id, channel }: { id: string; channel: ContactRevealChannel }) =>
      revealListingContact(id, channel),
  });
}

export function useMyListings(
  params: SearchParams = {},
  options?: { enabled?: boolean; userId?: string | null; keepPreviousData?: boolean }
) {
  return useQuery({
    queryKey: ['my-listings', options?.userId ?? null, params],
    queryFn: () => getMyListings(params),
    enabled: options?.enabled ?? true,
    placeholderData: options?.keepPreviousData === false ? undefined : keepPreviousData,
  });
}

export function useMyListingQuotaSummary(options?: {
  enabled?: boolean;
  userId?: string | null;
  activeLimit?: number;
}) {
  return useQuery({
    queryKey: ['my-listing-quota-summary', options?.userId ?? null, options?.activeLimit ?? null],
    queryFn: () =>
      collectActiveListingCount({
        limit: options?.activeLimit ?? 3,
        pageSize: LISTING_ACCESS_CHECK_PAGE_SIZE,
        fetchPage: ({ page, pageSize }) => getMyListings({ page, pageSize }),
      }),
    enabled: options?.enabled ?? true,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    retry: 1,
  });
}

export function useSavedListings(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['saved-listings'],
    queryFn: getSavedListings,
    enabled: options?.enabled ?? true,
  });
}

export function useCreateListing() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateListingRequest) => createListing(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listing-quota-summary'] });
    },
  });
}

export function useUpdateListing(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateListingRequest) => updateListing(id, data),
    onSuccess: (updatedListing) => {
      // Immediately populate owner cache with fresh server response so the
      // detail page shows the real moderationStatus (e.g. "pending") without
      // waiting for the background refetch to complete.
      queryClient.setQueryData(['owner-listing', id], updatedListing);
      queryClient.removeQueries({ queryKey: ['listing', id] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listing-quota-summary'] });
    },
  });
}

export function useDeleteListing() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteListing(id),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: ['listing', id] });
      queryClient.invalidateQueries({ queryKey: ['listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listing-quota-summary'] });
    },
  });
}

export function useRelistListing() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => relistListing(id),
    onSuccess: (listing) => {
      queryClient.setQueryData(['owner-listing', listing.listingId], listing);
      queryClient.removeQueries({ queryKey: ['listing', listing.listingId] });
      queryClient.invalidateQueries({ queryKey: ['listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listing-quota-summary'] });
    },
  });
}

export function useSaveListing() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, saved }: { id: string; saved: boolean }) =>
      saved ? unsaveListing(id) : saveListing(id),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['saved-listings'] });
      queryClient.invalidateQueries({ queryKey: ['listings'] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
      queryClient.invalidateQueries({ queryKey: ['listing', variables.id] });
    },
  });
}
