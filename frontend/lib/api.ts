import axios, { AxiosError } from 'axios';
import { getSession } from 'next-auth/react';
import type {
  Listing,
  ListingsResponse,
  ParsedListing,
  SearchParams,
  User,
  CreateListingRequest,
  FeatureListingRequest,
  PaymentResponse,
  LocationSuggestion,
  LocationOption,
} from '@/types';
import { getBackendAuthHeader, getBackendProfilePath } from '@/lib/backend-auth';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
});

apiClient.interceptors.request.use(async (config) => {
  const session = await getSession();
  if (session?.user) {
    const token = (session as { accessToken?: string }).accessToken;
    const authHeader = getBackendAuthHeader({ backendAccessToken: token });
    if (authHeader) {
      config.headers.Authorization = authHeader;
    }
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
    const message =
      (error.response?.data as { message?: string })?.message ||
      error.message ||
      'Terjadi kesalahan. Silakan coba lagi.';
    return Promise.reject(new Error(message));
  }
);

type BackendListingsResponse = {
  items?: Listing[];
  listings?: Listing[];
  total?: number;
  page?: number;
};

function normalizeListingsResponse(data: BackendListingsResponse): ListingsResponse {
  const items = data.items ?? data.listings ?? [];
  return {
    items,
    total: data.total ?? items.length,
    page: data.page ?? 1,
  };
}

export async function parseListingText(text: string): Promise<ParsedListing> {
  const response = await apiClient.post<{
    parsed: ParsedListing;
    requiresCorrection: boolean;
    confidence: number;
  }>('/listings/parse-text', { text });
  return response.data.parsed;
}

export async function createListing(data: CreateListingRequest): Promise<Listing> {
  const response = await apiClient.post<Listing>('/listings', data);
  return response.data;
}

export async function getListing(id: string): Promise<Listing> {
  const response = await apiClient.get<Listing>(`/listings/${id}`);
  return response.data;
}

export async function getListings(params: SearchParams): Promise<ListingsResponse> {
  const response = await apiClient.get<BackendListingsResponse>('/listings', { params });
  return normalizeListingsResponse(response.data);
}

export async function updateListing(
  id: string,
  data: Partial<CreateListingRequest>
): Promise<Listing> {
  const response = await apiClient.put<Listing>(`/listings/${id}`, data);
  return response.data;
}

export async function deleteListing(id: string): Promise<void> {
  await apiClient.delete(`/listings/${id}`);
}

export async function getUploadUrl(
  listingId: string,
  filename: string
): Promise<{ uploadUrl: string; fileUrl: string }> {
  const response = await apiClient.post<{ uploadUrl: string; fileUrl: string }>(
    `/listings/${listingId}/upload-url`,
    { filename }
  );
  return response.data;
}

export async function featureListing(
  data: FeatureListingRequest
): Promise<PaymentResponse> {
  const response = await apiClient.post<PaymentResponse>('/listings/feature', data);
  return response.data;
}

export async function upgradePremium(): Promise<PaymentResponse> {
  const response = await apiClient.post<PaymentResponse>('/users/upgrade');
  return response.data;
}

export async function getProfile(): Promise<User> {
  const response = await apiClient.get<User>(getBackendProfilePath());
  return response.data;
}

export async function getSavedListings(): Promise<ListingsResponse> {
  const response = await apiClient.get<BackendListingsResponse>('/users/me/saved');
  return normalizeListingsResponse(response.data);
}

export async function saveListing(listingId: string): Promise<void> {
  await apiClient.post(`/listings/${listingId}/save`);
}

export async function unsaveListing(listingId: string): Promise<void> {
  await apiClient.delete(`/listings/${listingId}/save`);
}

export async function getMyListings(params?: SearchParams): Promise<ListingsResponse> {
  const response = await apiClient.get<BackendListingsResponse>('/users/me/listings', { params });
  return normalizeListingsResponse(response.data);
}

export async function searchNearby(
  lat: number,
  lng: number,
  radiusKm: number
): Promise<ListingsResponse> {
  const response = await apiClient.get<BackendListingsResponse>('/listings/nearby', {
    params: { lat, lng, radiusKm },
  });
  return normalizeListingsResponse(response.data);
}

export async function getLocationSuggestions(
  query: string
): Promise<LocationSuggestion[]> {
  const response = await apiClient.get<LocationSuggestion[]>('/locations/suggest', {
    params: { q: query },
  });
  return response.data;
}

export async function getProvinceSuggestions(q?: string): Promise<LocationOption[]> {
  const response = await apiClient.get<{ provinces: LocationOption[] }>('/locations/provinces', {
    params: { q },
  });
  return response.data.provinces;
}

export async function getCitySuggestions(
  provinceId: string,
  q?: string
): Promise<LocationOption[]> {
  const response = await apiClient.get<{ cities: LocationOption[] }>('/locations/cities', {
    params: { provinceId, q },
  });
  return response.data.cities;
}

export async function getDistrictSuggestions(
  cityId: string,
  q?: string
): Promise<LocationOption[]> {
  const response = await apiClient.get<{ districts: LocationOption[] }>('/locations/districts', {
    params: { cityId, q },
  });
  return response.data.districts;
}

export default apiClient;
