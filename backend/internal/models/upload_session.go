package models

import "time"

type UploadSession struct {
	SessionID           string     `json:"sessionId" dynamodbav:"sessionId"`
	UserID              string     `json:"userId" dynamodbav:"userId"`
	ListingID           string     `json:"listingId,omitempty" dynamodbav:"listingId,omitempty"`
	StagingKey          string     `json:"stagingKey" dynamodbav:"stagingKey"`
	ExpectedContentType string     `json:"expectedContentType" dynamodbav:"expectedContentType"`
	ExpectedMaxSize     int64      `json:"expectedMaxSize" dynamodbav:"expectedMaxSize"`
	ExpiresAt           time.Time  `json:"expiresAt" dynamodbav:"expiresAt"`
	ConsumedAt          *time.Time `json:"consumedAt,omitempty" dynamodbav:"consumedAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt" dynamodbav:"createdAt"`
}

type UploadPrepareRequest struct {
	ListingID          string         `json:"listingId,omitempty"`
	RetainedImageCount int            `json:"retainedImageCount"`
	FinalImageCount    int            `json:"finalImageCount"`
	NewImages          []NewImageSpec `json:"newImages"`
}

type NewImageSpec struct {
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type UploadSlot struct {
	SessionID    string `json:"sessionId"`
	PresignedURL string `json:"presignedUrl"`
	StagingKey   string `json:"stagingKey"`
	ExpiresAt    string `json:"expiresAt"`
}

type UploadPrepareResponse struct {
	Slots []UploadSlot `json:"slots"`
}
