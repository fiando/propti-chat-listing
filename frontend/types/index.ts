export interface PropertyDetails {
  landArea: number;
  buildingArea: number;
  bedrooms: number;
  bathrooms: number;
  frontWidth?: number;
  orientation?: string;
  legalStatus?: string;
  powerConsumption?: string;
  amenities: string[];
}

export interface Location {
  address: string;
  googlePlaceId?: string;
  latitude?: number;
  longitude?: number;
  province?: string;
  city?: string;
  district?: string;
  nearbyPlaces?: string[];
}

export interface PremiumFeatures {
  isPremium: boolean;
  isFeatured: boolean;
  featuredUntil?: string;
  promotionUntil?: string;
}

export type ListingType = 'sell' | 'rent';
export type ListingStatus = 'active' | 'sold' | 'archived';
export type ModerationStatus = 'approved' | 'pending' | 'rejected' | 'draft';
export type SearchMode = 'manual' | 'smart';

export interface ListingImageView {
  imageId?: string;
  url?: string;
  thumbnailUrl?: string;
  contentType?: string;
  sizeBytes?: number;
  isFeatured?: boolean;
  uploadedAt?: string;
}

export type ListingImageValue = string | ListingImageView;

export interface ListingFormImage {
  id: string;
  kind: 'existing' | 'new';
  previewUrl: string;
  remoteUrl?: string;
  imageId?: string;
  file?: File;
  isFeatured: boolean;
}

export interface Listing {
  listingId: string;
  userId: string;
  title: string;
  description: string;
  price: number;
  priceUnit: 'total' | 'monthly' | 'yearly';
  listingType: ListingType;
  status: ListingStatus;
  propertyDetails: PropertyDetails;
  location: Location;
  images: ListingImageValue[];
  videos: string[];
  imageCount: number;
  premiumFeatures: PremiumFeatures;
  sellerName?: string;
  sellerPhone?: string;
  hasSellerPhone?: boolean;
  views: number;
  saves: number;
  contactReveals?: number;
  moderationStatus: ModerationStatus;
  featuredThumbnailUrl?: string;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface User {
  userId: string;
  googleId: string;
  email: string;
  name: string;
  profilePicture?: string;
  phone?: string;
  role: 'buyer' | 'seller' | 'both';
  preferences?: UserPreferences;
  savedListingIds?: string[];
  subscription: {
    tier: 'free' | 'basic' | 'premium' | 'pro';
    monthlyListingsUsed: number;
    activeListingsCount?: number;
    renewDate?: string;
  };
  subscriptionStatus?: SubscriptionStatus;
  createdAt: string;
  lastLoginAt: string;
}

export type SubscriptionStatus = 'active' | 'expiring_soon' | 'expired' | 'free' | 'loading';

export type SubscriptionTier = User['subscription']['tier'];

export interface SubscriptionPlanEntitlement {
  tier: SubscriptionTier;
  priceIdr: number;
  activeListingCap: number;
  photoCapPerListing: number;
  waReadAllowed: boolean;
  waCreateAllowed: boolean;
  voiceMinutesPerMonth: number;
}

export const SubscriptionPlans: Record<SubscriptionTier, SubscriptionPlanEntitlement> = {
  free: {
    tier: 'free',
    priceIdr: 0,
    activeListingCap: 5,
    photoCapPerListing: 5,
    waReadAllowed: false,
    waCreateAllowed: true,
    voiceMinutesPerMonth: 0,
  },
  // 'basic' is kept as a migration alias for existing users grandfathered into Premium.
  basic: {
    tier: 'basic',
    priceIdr: 99000,
    activeListingCap: 25,
    photoCapPerListing: 15,
    waReadAllowed: true,
    waCreateAllowed: true,
    voiceMinutesPerMonth: 90,
  },
  premium: {
    tier: 'premium',
    priceIdr: 99000,
    activeListingCap: 25,
    photoCapPerListing: 15,
    waReadAllowed: true,
    waCreateAllowed: true,
    voiceMinutesPerMonth: 90,
  },
  pro: {
    tier: 'pro',
    priceIdr: 179000,
    activeListingCap: 100,
    photoCapPerListing: 25,
    waReadAllowed: true,
    waCreateAllowed: true,
    voiceMinutesPerMonth: 180,
  },
} as const;

export type ContactRevealChannel = 'whatsapp' | 'phone';

export interface RevealListingContactResponse {
  sellerName: string;
  sellerPhone: string;
  channel: ContactRevealChannel;
}

export interface UserPreferences {
  favoriteLocations: string[];
  searchHistory: string[];
  notifications: boolean;
}

export interface UpdateProfileRequest {
  phone?: string;
  role?: 'buyer' | 'seller' | 'both';
  preferences?: UserPreferences;
}

export interface ParsedLocationSuggestion {
  province?: string;
  city?: string;
  district?: string;
  normalizedAddress?: string;
  confidence?: number;
}

export interface ParsedListing {
  title: string;
  description: string;
  price: number;
  priceUnit: string;
  propertyDetails: PropertyDetails;
  address: string;
  locationSuggestion?: ParsedLocationSuggestion;
  confidence: number;
  requiresManualReview: boolean;
  warnings: string[];
}

export interface LocationOption {
  id: string;
  name: string;
  parentId?: string;
}

export interface SearchParams {
  searchMode?: SearchMode;
  smartQuery?: string;
  q?: string;
  province?: string;
  city?: string;
  priceMin?: number;
  priceMax?: number;
  bedrooms?: number;
  bathrooms?: number;
  buildingAreaMin?: number;
  buildingAreaMax?: number;
  landAreaMin?: number;
  landAreaMax?: number;
  legalStatus?: string;
  amenities?: string[];
  listingType?: ListingType;
  sortBy?: string;
  page?: number;
  pageSize?: number;
}

export interface SmartSearchIntent {
  query: string;
  keywordQuery?: string;
  listingType?: ListingType | '';
  province?: string;
  city?: string;
  priceMin?: number;
  priceMax?: number;
  bedrooms?: number;
  bathrooms?: number;
  buildingAreaMin?: number;
  buildingAreaMax?: number;
  landAreaMin?: number;
  landAreaMax?: number;
  legalStatus?: string;
  amenities?: string[];
  sortBy?: string;
  confidence?: number;
}

export interface SmartSearchResponse {
  searchParams: SearchParams;
  normalized: SmartSearchIntent;
  metadata: {
    locationResolved: boolean;
  };
}

export interface ListingsResponse {
  items: Listing[];
  total: number;
  page: number;
}

export interface CreateListingRequest {
  title: string;
  description: string;
  price: number;
  priceUnit: 'total' | 'monthly' | 'yearly';
  listingType: ListingType;
  propertyDetails: PropertyDetails;
  location: Location;
  images?: string[];
  videos?: string[];
  newImageUploadSessionIds?: string[];
  featuredUploadSessionId?: string;
}

export interface UpdateListingRequest {
  title?: string;
  description?: string;
  price?: number;
  priceUnit?: 'total' | 'monthly' | 'yearly';
  listingType?: ListingType;
  status?: ListingStatus;
  propertyDetails?: PropertyDetails;
  location?: Location;
  images?: string[];
  videos?: string[];
  retainedImageIds?: string[];
  newImageUploadSessionIds?: string[];
  featuredImageId?: string;
  featuredUploadSessionId?: string;
}

export interface NewImageSpec {
  contentType: string;
  sizeBytes: number;
}

export interface UploadSlot {
  sessionId: string;
  presignedUrl: string;
  stagingKey: string;
  expiresAt: string;
}

export interface UploadPrepareRequest {
  listingId?: string;
  retainedImageCount: number;
  finalImageCount: number;
  newImages: NewImageSpec[];
}

export interface UploadPrepareResponse {
  slots: UploadSlot[];
}

export type LeadStage = 'new' | 'interested' | 'viewing' | 'negotiation' | 'deal' | 'lost';
export type FollowUpTaskStatus = 'pending' | 'completed' | 'skipped';

export interface LeadActivity {
  at: string;
  type: string;
  message: string;
}

export interface FollowUpTask {
  taskId: string;
  leadId: string;
  offsetDays: number;
  dueAt: string;
  status: FollowUpTaskStatus;
  createdAt: string;
  updatedAt: string;
}

export interface Lead {
  leadId: string;
  ownerUserId: string;
  listingId?: string;
  name: string;
  phone?: string;
  source?: string;
  stage: LeadStage;
  notes?: string[];
  activities?: LeadActivity[];
  followUpTasks?: FollowUpTask[];
  lastContactAt?: string;
  firstResponseAt?: string;
  viewedAt?: string;
  closedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface LeadListResponse {
  leads: Lead[];
  total: number;
  nextCursor?: string;
}

export interface CreateLeadRequest {
  listingId?: string;
  name: string;
  phone?: string;
  source?: string;
  note?: string;
}

export interface UpdateLeadStageRequest {
  stage: LeadStage;
  reason?: string;
}

export interface AddLeadNoteRequest {
  note: string;
}

export interface CompleteFollowUpTaskRequest {
  status: FollowUpTaskStatus;
  note?: string;
}

export interface AgentAnalytics {
  leadCount: number;
  leadToViewingRate: number;
  viewingToDealRate: number;
  overdueFollowUpRate: number;
  pendingFollowUpCount: number;
  overdueFollowUpCount: number;
}

export interface FeatureListingRequest {
  listingId: string;
  durationDays: number;
  type: 'featured' | 'promotion';
}

export interface PaymentResponse {
  transactionId: string;
  paymentUrl: string;
  orderId: string;
  amount: number;
}

export interface LocationSuggestion {
  placeId: string;
  description: string;
  city?: string;
  district?: string;
}

// ImageLimits are the canonical per-listing image caps per subscription tier.
// Must stay in sync with backend/internal/utils validator constants.
export const ImageLimits = {
  free: SubscriptionPlans.free.photoCapPerListing,
  basic: SubscriptionPlans.basic.photoCapPerListing,
  premium: SubscriptionPlans.premium.photoCapPerListing,
  pro: SubscriptionPlans.pro.photoCapPerListing,
} as const;

export type ImageLimitTier = keyof typeof ImageLimits;
