const LISTING_VIEW_PREFIX = 'propti:listing-view:';
const VIEW_TRACK_COOLDOWN_MS = 1000 * 60 * 60 * 24;

export function getTrackedListingViewKey(listingId: string) {
  return `${LISTING_VIEW_PREFIX}${listingId}`;
}

export function shouldTrackListingView(
  storage: Storage,
  listingId: string,
  now = Date.now()
) {
  const raw = storage.getItem(getTrackedListingViewKey(listingId));
  if (!raw) {
    return true;
  }

  const trackedAt = Number(raw);
  if (!Number.isFinite(trackedAt)) {
    return true;
  }

  return now - trackedAt > VIEW_TRACK_COOLDOWN_MS;
}

export function markListingViewTracked(
  storage: Storage,
  listingId: string,
  now = Date.now()
) {
  storage.setItem(getTrackedListingViewKey(listingId), String(now));
}
