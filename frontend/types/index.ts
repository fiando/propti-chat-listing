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
export type ModerationStatus = 'approved' | 'pending' | 'rejected';

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
    tier: 'free' | 'premium';
    monthlyListingsUsed: number;
    activeListingsCount?: number;
    renewDate?: string;
  };
  subscriptionStatus?: SubscriptionStatus;
  createdAt: string;
  lastLoginAt: string;
}

export type SubscriptionStatus = 'active' | 'expiring_soon' | 'expired' | 'free' | 'loading';

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
  free: 3,
  premium: 15,
} as const;

export type ImageLimitTier = keyof typeof ImageLimits;
