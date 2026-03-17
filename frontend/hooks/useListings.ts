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
  trackListingView,
  createListing,
  updateListing,
  deleteListing,
  getMyListings,
  saveListing,
  unsaveListing,
  getSavedListings,
} from '@/lib/api';
import type { SearchParams, CreateListingRequest } from '@/types';

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

export function useTrackListingView() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => trackListingView(id),
    onSuccess: (listing) => {
      queryClient.setQueryData(['listing', listing.listingId], listing);
    },
  });
}

export function useMyListings(params: SearchParams = {}, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['my-listings', params],
    queryFn: () => getMyListings(params),
    enabled: options?.enabled ?? true,
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
    },
  });
}

export function useUpdateListing(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CreateListingRequest>) => updateListing(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['listing', id] });
      queryClient.invalidateQueries({ queryKey: ['my-listings'] });
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
