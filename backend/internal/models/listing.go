package models

import "time"

type ListingType string
type ListingStatus string
type ModerationStatus string

const (
	ListingTypeSell ListingType = "sell"
	ListingTypeRent ListingType = "rent"

	ListingStatusActive   ListingStatus = "active"
	ListingStatusSold     ListingStatus = "sold"
	ListingStatusArchived ListingStatus = "archived"

	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusRejected ModerationStatus = "rejected"
	ModerationStatusDraft    ModerationStatus = "draft"
)

type PropertyDetails struct {
	LandArea         float64  `json:"landArea" dynamodbav:"landArea"`
	BuildingArea     float64  `json:"buildingArea" dynamodbav:"buildingArea"`
	Bedrooms         int      `json:"bedrooms" dynamodbav:"bedrooms"`
	Bathrooms        int      `json:"bathrooms" dynamodbav:"bathrooms"`
	FrontWidth       float64  `json:"frontWidth" dynamodbav:"frontWidth"`
	Orientation      string   `json:"orientation" dynamodbav:"orientation"`
	LegalStatus      string   `json:"legalStatus" dynamodbav:"legalStatus"`
	PowerConsumption string   `json:"powerConsumption" dynamodbav:"powerConsumption"`
	Amenities        []string `json:"amenities" dynamodbav:"amenities"`
}

type Location struct {
	Address       string   `json:"address" dynamodbav:"address"`
	GooglePlaceID string   `json:"googlePlaceId" dynamodbav:"googlePlaceId"`
	Latitude      float64  `json:"latitude" dynamodbav:"latitude"`
	Longitude     float64  `json:"longitude" dynamodbav:"longitude"`
	Province      string   `json:"province" dynamodbav:"province"`
	City          string   `json:"city" dynamodbav:"city"`
	District      string   `json:"district" dynamodbav:"district"`
	NearbyPlaces  []string `json:"nearbyPlaces" dynamodbav:"nearbyPlaces"`
}

type PremiumFeatures struct {
	IsPremium      bool       `json:"isPremium" dynamodbav:"isPremium"`
	IsFeatured     bool       `json:"isFeatured" dynamodbav:"isFeatured"`
	FeaturedUntil  *time.Time `json:"featuredUntil,omitempty" dynamodbav:"featuredUntil,omitempty"`
	PromotionUntil *time.Time `json:"promotionUntil,omitempty" dynamodbav:"promotionUntil,omitempty"`
}

type Listing struct {
	PK               string           `json:"pk" dynamodbav:"PK"`
	SK               string           `json:"sk" dynamodbav:"SK"`
	ListingID        string           `json:"listingId" dynamodbav:"listingId"`
	UserID           string           `json:"userId" dynamodbav:"userId"`
	Title            string           `json:"title" dynamodbav:"title"`
	Description      string           `json:"description" dynamodbav:"description"`
	Price            float64          `json:"price" dynamodbav:"price"`
	PriceUnit        string           `json:"priceUnit" dynamodbav:"priceUnit"`
	ListingType      ListingType      `json:"listingType" dynamodbav:"listingType"`
	Status           ListingStatus    `json:"status" dynamodbav:"status"`
	PropertyDetails  PropertyDetails  `json:"propertyDetails" dynamodbav:"propertyDetails"`
	Location         Location         `json:"location" dynamodbav:"location"`
	Images           ImageEntries     `json:"-" dynamodbav:"images"`
	Videos           []string         `json:"videos" dynamodbav:"videos"`
	ImageCount       int              `json:"imageCount" dynamodbav:"imageCount"`
	PremiumFeatures  PremiumFeatures  `json:"premiumFeatures" dynamodbav:"premiumFeatures"`
	SellerName       string           `json:"sellerName,omitempty" dynamodbav:"-"`
	SellerPhone      string           `json:"sellerPhone,omitempty" dynamodbav:"-"`
	HasSellerPhone   bool             `json:"hasSellerPhone" dynamodbav:"-"`
	Views            int              `json:"views" dynamodbav:"views"`
	Saves            int              `json:"saves" dynamodbav:"saves"`
	ContactReveals   int              `json:"contactReveals" dynamodbav:"contactReveals"`
	ModerationStatus ModerationStatus `json:"moderationStatus" dynamodbav:"moderationStatus"`
	ModerationReason string           `json:"moderationReason,omitempty" dynamodbav:"moderationReason,omitempty"`
	CreatedAt        time.Time        `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt" dynamodbav:"updatedAt"`
	ExpiresAt        *time.Time       `json:"expiresAt,omitempty" dynamodbav:"expiresAt,omitempty"`
}

type CreateListingRequest struct {
	Title                    string          `json:"title"`
	Description              string          `json:"description"`
	Price                    float64         `json:"price"`
	PriceUnit                string          `json:"priceUnit"`
	ListingType              ListingType     `json:"listingType"`
	PropertyDetails          PropertyDetails `json:"propertyDetails"`
	Location                 Location        `json:"location"`
	Images                   []string        `json:"images,omitempty"`
	Videos                   []string        `json:"videos"`
	NewImageUploadSessionIDs []string        `json:"newImageUploadSessionIds,omitempty"`
	FeaturedUploadSessionID  string          `json:"featuredUploadSessionId,omitempty"`
	IsDraft                  bool            `json:"isDraft,omitempty"`
}

type UpdateListingRequest struct {
	Title                    *string          `json:"title,omitempty"`
	Description              *string          `json:"description,omitempty"`
	Price                    *float64         `json:"price,omitempty"`
	PriceUnit                *string          `json:"priceUnit,omitempty"`
	Status                   *ListingStatus   `json:"status,omitempty"`
	PropertyDetails          *PropertyDetails `json:"propertyDetails,omitempty"`
	Location                 *Location        `json:"location,omitempty"`
	Images                   []string         `json:"images,omitempty"`
	Videos                   []string         `json:"videos,omitempty"`
	RetainedImageIDs         []string         `json:"retainedImageIds,omitempty"`
	NewImageUploadSessionIDs []string         `json:"newImageUploadSessionIds,omitempty"`
	FeaturedImageID          *string          `json:"featuredImageId,omitempty"`
	FeaturedUploadSessionID  *string          `json:"featuredUploadSessionId,omitempty"`
}

type ListingSearchParams struct {
	Query           string      `json:"q"`
	Province        string      `json:"province"`
	City            string      `json:"city"`
	ListingType     ListingType `json:"listingType"`
	PriceMin        float64     `json:"priceMin"`
	PriceMax        float64     `json:"priceMax"`
	Bedrooms        int         `json:"bedrooms"`
	Bathrooms       int         `json:"bathrooms"`
	BuildingAreaMin float64     `json:"buildingAreaMin"`
	BuildingAreaMax float64     `json:"buildingAreaMax"`
	LandAreaMin     float64     `json:"landAreaMin"`
	LandAreaMax     float64     `json:"landAreaMax"`
	LegalStatus     string      `json:"legalStatus"`
	Amenities       []string    `json:"amenities"`
	SortBy          string      `json:"sortBy"`
	Page            int         `json:"page"`
	PageSize        int         `json:"pageSize"`
}

type ContactRevealChannel string

const (
	ContactRevealChannelWhatsApp ContactRevealChannel = "whatsapp"
	ContactRevealChannelPhone    ContactRevealChannel = "phone"
)

type RevealListingContactRequest struct {
	Channel ContactRevealChannel `json:"channel"`
}

type ListingContactReveal struct {
	SellerName  string               `json:"sellerName"`
	SellerPhone string               `json:"sellerPhone"`
	Channel     ContactRevealChannel `json:"channel"`
}

type ParseTextRequest struct {
	Text string `json:"text"`
}

type ParseTextResponse struct {
	Parsed             ParsedListing `json:"parsed"`
	RequiresCorrection bool          `json:"requiresCorrection"`
	Confidence         float64       `json:"confidence"`
}

type ParsedLocationSuggestion struct {
	Province          string  `json:"province"`
	City              string  `json:"city"`
	District          string  `json:"district"`
	NormalizedAddress string  `json:"normalizedAddress"`
	Confidence        float64 `json:"confidence"`
}

type ParsedListing struct {
	Title                string                   `json:"title"`
	Description          string                   `json:"description"`
	Price                float64                  `json:"price"`
	PriceUnit            string                   `json:"priceUnit"`
	PropertyDetails      PropertyDetails          `json:"propertyDetails"`
	Address              string                   `json:"address"`
	LocationSuggestion   ParsedLocationSuggestion `json:"locationSuggestion"`
	Confidence           float64                  `json:"confidence"`
	RequiresManualReview bool                     `json:"requiresManualReview"`
	Warnings             []string                 `json:"warnings"`
}
