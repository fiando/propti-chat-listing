import type { RevealListingContactResponse } from '@/types';

const revealedContactCache = new Map<string, RevealListingContactResponse>();
const pendingRevealRequests = new Map<string, Promise<RevealListingContactResponse>>();

export function getCachedRevealedContact(listingId: string): RevealListingContactResponse | null {
  return revealedContactCache.get(listingId) ?? null;
}

export function clearRevealedContactCache() {
  revealedContactCache.clear();
  pendingRevealRequests.clear();
}

export async function getOrLoadRevealedContact(
  listingId: string,
  loader: () => Promise<RevealListingContactResponse>
): Promise<RevealListingContactResponse> {
  const cached = getCachedRevealedContact(listingId);
  if (cached) {
    return cached;
  }

  const pending = pendingRevealRequests.get(listingId);
  if (pending) {
    return pending;
  }

  const request = loader()
    .then((contact) => {
      revealedContactCache.set(listingId, contact);
      return contact;
    })
    .finally(() => {
      pendingRevealRequests.delete(listingId);
    });

  pendingRevealRequests.set(listingId, request);
  return request;
}
