import type { ParsedListing, Location } from '@/types';

const CREATE_LISTING_DRAFT_KEY = 'propti:create-listing-draft';

export type CreateListingStep = 'choose' | 'parse' | 'form';

export type CreateListingDraft = {
  step: CreateListingStep;
  parseText?: string;
  parsedData?: ParsedListing;
  parsedLocation?: Partial<Location>;
  formValues?: Record<string, unknown>;
};

export function saveCreateListingDraft(storage: Storage, draft: CreateListingDraft) {
  storage.setItem(CREATE_LISTING_DRAFT_KEY, JSON.stringify(draft));
}

export function loadCreateListingDraft(storage: Storage): CreateListingDraft | null {
  const raw = storage.getItem(CREATE_LISTING_DRAFT_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as CreateListingDraft;
  } catch {
    return null;
  }
}

export function clearCreateListingDraft(storage: Storage) {
  storage.removeItem(CREATE_LISTING_DRAFT_KEY);
}
