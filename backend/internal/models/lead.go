package models

import "time"

type LeadStage string
type FollowUpTaskStatus string

const (
	LeadStageNew         LeadStage = "new"
	LeadStageInterested  LeadStage = "interested"
	LeadStageViewing     LeadStage = "viewing"
	LeadStageNegotiation LeadStage = "negotiation"
	LeadStageDeal        LeadStage = "deal"
	LeadStageLost        LeadStage = "lost"

	FollowUpTaskStatusPending   FollowUpTaskStatus = "pending"
	FollowUpTaskStatusCompleted FollowUpTaskStatus = "completed"
	FollowUpTaskStatusSkipped   FollowUpTaskStatus = "skipped"
)

type LeadActivity struct {
	At      time.Time `json:"at" dynamodbav:"at"`
	Type    string    `json:"type" dynamodbav:"type"`
	Message string    `json:"message" dynamodbav:"message"`
}

type FollowUpTask struct {
	TaskID     string             `json:"taskId" dynamodbav:"taskId"`
	LeadID     string             `json:"leadId" dynamodbav:"leadId"`
	OffsetDays int                `json:"offsetDays" dynamodbav:"offsetDays"`
	DueAt      time.Time          `json:"dueAt" dynamodbav:"dueAt"`
	Status     FollowUpTaskStatus `json:"status" dynamodbav:"status"`
	CreatedAt  time.Time          `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" dynamodbav:"updatedAt"`
}

type Lead struct {
	PK              string         `json:"pk" dynamodbav:"PK"`
	SK              string         `json:"sk" dynamodbav:"SK"`
	LeadID          string         `json:"leadId" dynamodbav:"leadId"`
	OwnerUserID     string         `json:"ownerUserId" dynamodbav:"ownerUserId"`
	ListingID       string         `json:"listingId,omitempty" dynamodbav:"listingId,omitempty"`
	Name            string         `json:"name" dynamodbav:"name"`
	Phone           string         `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	Source          string         `json:"source,omitempty" dynamodbav:"source,omitempty"`
	Stage           LeadStage      `json:"stage" dynamodbav:"stage"`
	Notes           []string       `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
	Activities      []LeadActivity `json:"activities,omitempty" dynamodbav:"activities,omitempty"`
	FollowUpTasks   []FollowUpTask `json:"followUpTasks,omitempty" dynamodbav:"followUpTasks,omitempty"`
	LastContactAt   *time.Time     `json:"lastContactAt,omitempty" dynamodbav:"lastContactAt,omitempty"`
	FirstResponseAt *time.Time     `json:"firstResponseAt,omitempty" dynamodbav:"firstResponseAt,omitempty"`
	ViewedAt        *time.Time     `json:"viewedAt,omitempty" dynamodbav:"viewedAt,omitempty"`
	ClosedAt        *time.Time     `json:"closedAt,omitempty" dynamodbav:"closedAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt" dynamodbav:"updatedAt"`
}

type CreateLeadRequest struct {
	ListingID string `json:"listingId,omitempty"`
	Name      string `json:"name"`
	Phone     string `json:"phone,omitempty"`
	Source    string `json:"source,omitempty"`
	Note      string `json:"note,omitempty"`
}

type UpdateLeadStageRequest struct {
	Stage  LeadStage `json:"stage"`
	Reason string    `json:"reason,omitempty"`
}

type AddLeadNoteRequest struct {
	Note string `json:"note"`
}

type CompleteFollowUpTaskRequest struct {
	Status FollowUpTaskStatus `json:"status"`
	Note   string             `json:"note,omitempty"`
}

type LeadListResponse struct {
	Leads []Lead `json:"leads"`
	Total int    `json:"total"`
}

type AgentAnalyticsResponse struct {
	LeadCount             int     `json:"leadCount"`
	MedianResponseMinutes int     `json:"medianResponseMinutes"`
	LeadToViewingRate     float64 `json:"leadToViewingRate"`
	ViewingToDealRate     float64 `json:"viewingToDealRate"`
	OverdueFollowUpRate   float64 `json:"overdueFollowUpRate"`
	PendingFollowUpCount  int     `json:"pendingFollowUpCount"`
	OverdueFollowUpCount  int     `json:"overdueFollowUpCount"`
}
