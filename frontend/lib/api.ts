import axios, { AxiosError } from 'axios';
import { getSession } from 'next-auth/react';
import type {
  Listing,
  ListingsResponse,
  ParsedListing,
  SmartSearchResponse,
  SearchParams,
  User,
  CreateListingRequest,
  UpdateListingRequest,
  FeatureListingRequest,
  PaymentResponse,
  LocationSuggestion,
  LocationOption,
  ContactRevealChannel,
  RevealListingContactResponse,
  UpdateProfileRequest,
  UploadPrepareRequest,
  UploadPrepareResponse,
  Lead,
  LeadListResponse,
  CreateLeadRequest,
  UpdateLeadStageRequest,
  AddLeadNoteRequest,
  CompleteFollowUpTaskRequest,
  AgentAnalytics,
} from '@/types';
import { getBackendAuthHeader, getBackendProfilePath } from '@/lib/backend-auth';
import { getApiErrorMessage } from '@/lib/api-error';
import type {
  WhatsAppLinkChallengeResponse,
  WhatsAppWriteEligibilityResponse,
} from '@/lib/whatsapp-linking';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'https://api.propti.id',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
  paramsSerializer: {
    serialize: (params: Record<string, unknown>) => {
      const qs = new URLSearchParams();
      for (const [key, value] of Object.entries(params)) {
        if (value === undefined || value === null || value === '') continue;
        if (Array.isArray(value)) {
          if (value.length > 0) qs.set(key, value.join(','));
        } else {
          qs.set(key, String(value));
        }
      }
      return qs.toString();
    },
  },
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
        const AUTH_PATHS = ['/login', '/callback', '/api/auth'];
        const currentPath = window.location.pathname;
        if (!AUTH_PATHS.some((p) => currentPath.startsWith(p))) {
          const callbackUrl = `${window.location.pathname}${window.location.search}`;
          window.location.href = `/login?callbackUrl=${encodeURIComponent(callbackUrl)}`;
        }
      }
    }
    const message = getApiErrorMessage(error.response?.data, error.message);
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

export async function parseSearchIntent(query: string): Promise<SmartSearchResponse> {
  const response = await apiClient.post<SmartSearchResponse>('/search/parse-intent', { query });
  return response.data;
}

export async function createListing(data: CreateListingRequest): Promise<Listing> {
  const response = await apiClient.post<Listing>('/listings', data);
  return response.data;
}

export async function getListing(id: string): Promise<Listing> {
  const response = await apiClient.get<Listing>(`/listings/${id}`);
  return response.data;
}

export async function getOwnerListing(id: string): Promise<Listing> {
  const response = await apiClient.get<Listing>(`/users/me/listings/${id}`);
  return response.data;
}

export async function trackListingView(id: string): Promise<Listing> {
  const response = await apiClient.post<Listing>(`/listings/${id}/view`);
  return response.data;
}

export async function revealListingContact(
  id: string,
  channel: ContactRevealChannel
): Promise<RevealListingContactResponse> {
  const response = await apiClient.post<RevealListingContactResponse>(`/listings/${id}/contact-reveal`, {
    channel,
  });
  return response.data;
}

export async function getListings(params: SearchParams): Promise<ListingsResponse> {
  const response = await apiClient.get<BackendListingsResponse>('/listings', { params });
  return normalizeListingsResponse(response.data);
}

export async function updateListing(
  id: string,
  data: UpdateListingRequest
): Promise<Listing> {
  const response = await apiClient.put<Listing>(`/listings/${id}`, data);
  return response.data;
}

export async function relistListing(id: string): Promise<Listing> {
  const response = await apiClient.post<Listing>(`/users/me/listings/${id}/relist`);
  return response.data;
}

export async function deleteListing(id: string): Promise<void> {
  await apiClient.delete(`/listings/${id}`);
}

export async function prepareListingUpload(
  data: UploadPrepareRequest
): Promise<UploadPrepareResponse> {
  const response = await apiClient.post<UploadPrepareResponse>('/listings/upload-prepare', data);
  return response.data;
}

export async function uploadListingImage(presignedUrl: string, file: File): Promise<void> {
  await axios.put(presignedUrl, file, {
    headers: {
      'Content-Type': file.type || 'application/octet-stream',
    },
  });
}

export async function featureListing(
  data: FeatureListingRequest
): Promise<PaymentResponse> {
  const response = await apiClient.post<PaymentResponse>('/premium/feature-listing', data);
  return response.data;
}

export async function upgradePremium(tier: 'basic' | 'premium' | 'pro' = 'premium'): Promise<PaymentResponse> {
  const response = await apiClient.post<PaymentResponse>('/premium/upgrade', { tier });
  return response.data;
}

export async function getProfile(): Promise<User> {
  const response = await apiClient.get<User>(getBackendProfilePath());
  return response.data;
}

export async function updateProfile(data: UpdateProfileRequest): Promise<User> {
  const response = await apiClient.put<User>(getBackendProfilePath(), data);
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

export async function getWhatsAppLinkStatus(): Promise<WhatsAppWriteEligibilityResponse> {
  const response = await apiClient.get<WhatsAppWriteEligibilityResponse>('/auth/whatsapp/link-status');
  return response.data;
}

export async function createWhatsAppLinkChallenge(phone: string): Promise<WhatsAppLinkChallengeResponse> {
  const response = await apiClient.post<WhatsAppLinkChallengeResponse>('/auth/whatsapp/link-challenge', {
    phone,
  });
  return response.data;
}

export async function verifyWhatsAppLink(
  challengeId: string,
  otpCode: string
): Promise<WhatsAppWriteEligibilityResponse> {
  const response = await apiClient.post<WhatsAppWriteEligibilityResponse>('/auth/whatsapp/link-verify', {
    challengeId,
    otpCode,
  });
  return response.data;
}

export async function disconnectWhatsAppLink(): Promise<WhatsAppWriteEligibilityResponse> {
  const response = await apiClient.delete<WhatsAppWriteEligibilityResponse>('/auth/whatsapp/link');
  return response.data;
}

export async function getLeads(params?: { stage?: string; dueOnly?: boolean }): Promise<LeadListResponse> {
  const response = await apiClient.get<LeadListResponse>('/leads', { params });
  return response.data;
}

export async function getLead(leadId: string): Promise<Lead> {
  const response = await apiClient.get<Lead>(`/leads/${leadId}`);
  return response.data;
}

export async function createLead(data: CreateLeadRequest): Promise<Lead> {
  const response = await apiClient.post<Lead>('/leads', data);
  return response.data;
}

export async function updateLeadStage(leadId: string, data: UpdateLeadStageRequest): Promise<Lead> {
  const response = await apiClient.post<Lead>(`/leads/${leadId}/stage`, data);
  return response.data;
}

export async function addLeadNote(leadId: string, data: AddLeadNoteRequest): Promise<Lead> {
  const response = await apiClient.post<Lead>(`/leads/${leadId}/notes`, data);
  return response.data;
}

export async function completeFollowUpTask(
  leadId: string,
  taskId: string,
  data: CompleteFollowUpTaskRequest
): Promise<Lead> {
  const response = await apiClient.post<Lead>(`/leads/${leadId}/followups/${taskId}/complete`, data);
  return response.data;
}

export async function getLeadAnalytics(): Promise<AgentAnalytics> {
  const response = await apiClient.get<AgentAnalytics>('/leads/analytics');
  return response.data;
}

export default apiClient;
