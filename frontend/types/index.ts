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
  images: string[];
  videos: string[];
  imageCount: number;
  premiumFeatures: PremiumFeatures;
  views: number;
  saves: number;
  moderationStatus: ModerationStatus;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  userId: string;
  googleId: string;
  email: string;
  name: string;
  profilePicture?: string;
  phone?: string;
  role: 'buyer' | 'seller' | 'both';
  subscription: {
    tier: 'free' | 'premium';
    monthlyListingsUsed: number;
    renewDate?: string;
  };
  createdAt: string;
  lastLoginAt: string;
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
  listingType?: ListingType;
  sortBy?: string;
  page?: number;
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
}

export interface FeatureListingRequest {
  listingId: string;
  durationDays: number;
}

export interface PaymentResponse {
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
