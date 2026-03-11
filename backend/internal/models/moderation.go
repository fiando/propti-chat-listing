package models

import "time"

type ModerationType string
type ModerationResult string
type ModeratorType string

const (
	ModerationTypeContent ModerationType = "content_check"
	ModerationTypeLegal   ModerationType = "legal_check"
	ModerationTypeSpam    ModerationType = "spam_check"

	ModerationResultApproved ModerationResult = "approved"
	ModerationResultRejected ModerationResult = "rejected"

	ModeratorAI    ModeratorType = "ai"
	ModeratorHuman ModeratorType = "human"
)

type Moderation struct {
	PK           string           `json:"pk" dynamodbav:"PK"`
	SK           string           `json:"sk" dynamodbav:"SK"`
	ModerationID string           `json:"moderationId" dynamodbav:"moderationId"`
	ListingID    string           `json:"listingId" dynamodbav:"listingId"`
	UserID       string           `json:"userId" dynamodbav:"userId"`
	Type         ModerationType   `json:"type" dynamodbav:"type"`
	Result       ModerationResult `json:"result" dynamodbav:"result"`
	Reason       string           `json:"reason,omitempty" dynamodbav:"reason,omitempty"`
	Moderator    ModeratorType    `json:"moderator" dynamodbav:"moderator"`
	Timestamp    time.Time        `json:"timestamp" dynamodbav:"timestamp"`
}

type ModerationActionRequest struct {
	Result ModerationResult `json:"result"`
	Reason string           `json:"reason"`
}
