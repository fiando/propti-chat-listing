package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ImageEntry struct {
	ImageID      string    `json:"imageId" dynamodbav:"imageId"`
	S3Key        string    `json:"s3Key,omitempty" dynamodbav:"s3Key,omitempty"`
	ThumbnailKey string    `json:"thumbnailKey,omitempty" dynamodbav:"thumbnailKey,omitempty"`
	ContentType  string    `json:"contentType,omitempty" dynamodbav:"contentType,omitempty"`
	SizeBytes    int64     `json:"sizeBytes,omitempty" dynamodbav:"sizeBytes,omitempty"`
	IsFeatured   bool      `json:"isFeatured" dynamodbav:"isFeatured"`
	UploadedAt   time.Time `json:"uploadedAt" dynamodbav:"uploadedAt"`
	LegacyValue  string    `json:"-" dynamodbav:"-"`
}

func (e ImageEntry) IsLegacy() bool {
	return e.LegacyValue != ""
}

type ImageEntries []ImageEntry

func (entries *ImageEntries) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	if value == nil {
		*entries = ImageEntries{}
		return nil
	}

	if nullValue, ok := value.(*types.AttributeValueMemberNULL); ok && nullValue.Value {
		*entries = ImageEntries{}
		return nil
	}

	listValue, ok := value.(*types.AttributeValueMemberL)
	if !ok {
		return fmt.Errorf("expected list attribute for images, got %T", value)
	}

	result := make(ImageEntries, 0, len(listValue.Value))
	for _, item := range listValue.Value {
		switch typed := item.(type) {
		case *types.AttributeValueMemberS:
			result = append(result, ImageEntry{LegacyValue: typed.Value})
		case *types.AttributeValueMemberM:
			var entry ImageEntry
			if err := attributevalue.UnmarshalMap(typed.Value, &entry); err != nil {
				return err
			}
			result = append(result, entry)
		default:
			return fmt.Errorf("unsupported image entry attribute type %T", item)
		}
	}

	*entries = result
	return nil
}

func (entries ImageEntries) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	values := make([]types.AttributeValue, 0, len(entries))
	for _, entry := range entries {
		if entry.IsLegacy() {
			values = append(values, &types.AttributeValueMemberS{Value: entry.LegacyValue})
			continue
		}

		marshaled, err := attributevalue.MarshalMap(entry)
		if err != nil {
			return nil, err
		}
		values = append(values, &types.AttributeValueMemberM{Value: marshaled})
	}
	return &types.AttributeValueMemberL{Value: values}, nil
}

func (entries ImageEntries) LegacyValues() []string {
	legacy := make([]string, 0)
	for _, entry := range entries {
		if entry.IsLegacy() {
			legacy = append(legacy, entry.LegacyValue)
		}
	}
	return legacy
}

type ListingImageView struct {
	ImageID      string    `json:"imageId,omitempty"`
	URL          string    `json:"url,omitempty"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	ContentType  string    `json:"contentType,omitempty"`
	SizeBytes    int64     `json:"sizeBytes,omitempty"`
	IsFeatured   bool      `json:"isFeatured"`
	UploadedAt   time.Time `json:"uploadedAt,omitempty"`
	S3Key        string    `json:"-" dynamodbav:"-"`
	ThumbnailKey string    `json:"-" dynamodbav:"-"`
}

type ListingResponse struct {
	ListingID            string           `json:"listingId"`
	UserID               string           `json:"userId"`
	Title                string           `json:"title"`
	Description          string           `json:"description"`
	Price                float64          `json:"price"`
	PriceUnit            string           `json:"priceUnit"`
	ListingType          ListingType      `json:"listingType"`
	Status               ListingStatus    `json:"status"`
	PropertyDetails      PropertyDetails  `json:"propertyDetails"`
	Location             Location         `json:"location"`
	Images               any              `json:"images,omitempty"`
	Videos               []string         `json:"videos,omitempty"`
	ImageCount           int              `json:"imageCount"`
	PremiumFeatures      PremiumFeatures  `json:"premiumFeatures"`
	SellerName           string           `json:"sellerName,omitempty"`
	SellerPhone          string           `json:"sellerPhone,omitempty"`
	HasSellerPhone       bool             `json:"hasSellerPhone"`
	Views                int              `json:"views"`
	Saves                int              `json:"saves"`
	ModerationStatus     ModerationStatus `json:"moderationStatus"`
	ModerationReason     string           `json:"moderationReason,omitempty"`
	FeaturedThumbnailURL string           `json:"featuredThumbnailUrl,omitempty"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

func (r ListingResponse) MarshalJSON() ([]byte, error) {
	type alias ListingResponse
	return json.Marshal(alias(r))
}
