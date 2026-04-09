package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

type LeadStore interface {
	Put(ctx context.Context, lead *models.Lead) error
	GetByID(ctx context.Context, ownerUserID, leadID string) (*models.Lead, error)
	ListByOwner(ctx context.Context, ownerUserID string, limit int32) ([]models.Lead, error)
	ListByOwnerPaged(ctx context.Context, ownerUserID string, limit int32, cursor string) (*repository.LeadPage, error)
}

type LeadService struct {
	leadRepo LeadStore
	now      func() time.Time
	newID    func() string
}

func NewLeadService(leadRepo LeadStore) *LeadService {
	return &LeadService{
		leadRepo: leadRepo,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: uuid.NewString,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, ownerUserID string, req *models.CreateLeadRequest) (*models.Lead, error) {
	if strings.TrimSpace(ownerUserID) == "" {
		return nil, utils.ErrUnauthorized
	}
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, utils.NewAppError(400, "lead name is required")
	}
	now := s.now()
	leadID := s.newID()
	lead := &models.Lead{
		PK:          fmt.Sprintf("%s#%s", ownerUserID, leadID),
		SK:          leadID,
		LeadID:      leadID,
		OwnerUserID: ownerUserID,
		ListingID:   strings.TrimSpace(req.ListingID),
		Name:        strings.TrimSpace(req.Name),
		Phone:       strings.TrimSpace(req.Phone),
		Source:      strings.TrimSpace(req.Source),
		Stage:       models.LeadStageNew,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if note := strings.TrimSpace(req.Note); note != "" {
		lead.Notes = []string{note}
		lead.Activities = append(lead.Activities, models.LeadActivity{
			At:      now,
			Type:    "note_added",
			Message: note,
		})
	}
	lead.FollowUpTasks = s.ensureFollowUpTasks(nil, now, leadID)
	if err := s.leadRepo.Put(ctx, lead); err != nil {
		return nil, utils.ErrInternal
	}
	return lead, nil
}

func (s *LeadService) ListLeads(ctx context.Context, ownerUserID string, stage string, dueOnly bool) ([]models.Lead, error) {
	page, err := s.ListLeadsPaged(ctx, ownerUserID, stage, dueOnly, 200, "")
	if err != nil {
		return nil, err
	}
	return page.Leads, nil
}

func (s *LeadService) ListLeadsPaged(ctx context.Context, ownerUserID string, stage string, dueOnly bool, limit int32, cursor string) (*models.LeadListResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	page, err := s.leadRepo.ListByOwnerPaged(ctx, ownerUserID, limit, cursor)
	if err != nil {
		return nil, utils.ErrInternal
	}
	now := s.now()
	stageFilter := models.LeadStage(strings.TrimSpace(stage))

	filtered := make([]models.Lead, 0, len(page.Leads))
	for _, lead := range page.Leads {
		lead.FollowUpTasks = s.ensureFollowUpTasks(lead.FollowUpTasks, lead.CreatedAt, lead.LeadID)
		if stageFilter != "" && lead.Stage != stageFilter {
			continue
		}
		if dueOnly && !hasDuePendingTask(lead.FollowUpTasks, now) {
			continue
		}
		filtered = append(filtered, lead)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})
	return &models.LeadListResponse{
		Leads:      filtered,
		Total:      len(filtered),
		NextCursor: page.NextCursor,
	}, nil
}

func (s *LeadService) GetLead(ctx context.Context, ownerUserID, leadID string) (*models.Lead, error) {
	lead, err := s.leadRepo.GetByID(ctx, ownerUserID, leadID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	if lead == nil {
		return nil, utils.ErrNotFound
	}
	lead.FollowUpTasks = s.ensureFollowUpTasks(lead.FollowUpTasks, lead.CreatedAt, lead.LeadID)
	return lead, nil
}

func (s *LeadService) UpdateLeadStage(ctx context.Context, ownerUserID, leadID string, req *models.UpdateLeadStageRequest) (*models.Lead, error) {
	if req == nil || req.Stage == "" {
		return nil, utils.NewAppError(400, "stage is required")
	}
	if !isValidLeadStage(req.Stage) {
		return nil, utils.NewAppError(400, "invalid stage")
	}

	lead, err := s.GetLead(ctx, ownerUserID, leadID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	prev := lead.Stage
	lead.Stage = req.Stage
	lead.UpdatedAt = now
	if req.Stage == models.LeadStageViewing && lead.ViewedAt == nil {
		lead.ViewedAt = &now
	}
	if req.Stage == models.LeadStageDeal || req.Stage == models.LeadStageLost {
		lead.ClosedAt = &now
		for i := range lead.FollowUpTasks {
			if lead.FollowUpTasks[i].Status == models.FollowUpTaskStatusPending {
				lead.FollowUpTasks[i].Status = models.FollowUpTaskStatusSkipped
				lead.FollowUpTasks[i].UpdatedAt = now
			}
		}
	}

	message := fmt.Sprintf("stage changed: %s -> %s", prev, req.Stage)
	if reason := strings.TrimSpace(req.Reason); reason != "" {
		message = message + " (" + reason + ")"
	}
	lead.Activities = append(lead.Activities, models.LeadActivity{
		At:      now,
		Type:    "stage_changed",
		Message: message,
	})
	if err := s.leadRepo.Put(ctx, lead); err != nil {
		return nil, utils.ErrInternal
	}
	return lead, nil
}

func (s *LeadService) AddLeadNote(ctx context.Context, ownerUserID, leadID string, req *models.AddLeadNoteRequest) (*models.Lead, error) {
	if req == nil || strings.TrimSpace(req.Note) == "" {
		return nil, utils.NewAppError(400, "note is required")
	}
	lead, err := s.GetLead(ctx, ownerUserID, leadID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	note := strings.TrimSpace(req.Note)
	lead.Notes = append(lead.Notes, note)
	lead.UpdatedAt = now
	lead.LastContactAt = &now
	if lead.FirstResponseAt == nil {
		lead.FirstResponseAt = &now
	}
	lead.Activities = append(lead.Activities, models.LeadActivity{
		At:      now,
		Type:    "note_added",
		Message: note,
	})
	lead.FollowUpTasks = s.ensureFollowUpTasks(lead.FollowUpTasks, now, lead.LeadID)
	if err := s.leadRepo.Put(ctx, lead); err != nil {
		return nil, utils.ErrInternal
	}
	return lead, nil
}

func (s *LeadService) CompleteFollowUpTask(ctx context.Context, ownerUserID, leadID, taskID string, req *models.CompleteFollowUpTaskRequest) (*models.Lead, error) {
	lead, err := s.GetLead(ctx, ownerUserID, leadID)
	if err != nil {
		return nil, err
	}
	status := models.FollowUpTaskStatusCompleted
	if req != nil && req.Status != "" {
		status = req.Status
	}
	if status != models.FollowUpTaskStatusCompleted && status != models.FollowUpTaskStatusSkipped {
		return nil, utils.NewAppError(400, "invalid follow-up status")
	}
	now := s.now()
	found := false
	for i := range lead.FollowUpTasks {
		if lead.FollowUpTasks[i].TaskID != taskID {
			continue
		}
		lead.FollowUpTasks[i].Status = status
		lead.FollowUpTasks[i].UpdatedAt = now
		found = true
		break
	}
	if !found {
		return nil, utils.ErrNotFound
	}
	lead.UpdatedAt = now
	if status == models.FollowUpTaskStatusCompleted {
		lead.LastContactAt = &now
	}
	msg := "follow-up updated"
	if req != nil && strings.TrimSpace(req.Note) != "" {
		msg = strings.TrimSpace(req.Note)
	}
	lead.Activities = append(lead.Activities, models.LeadActivity{
		At:      now,
		Type:    "followup_updated",
		Message: msg,
	})
	if err := s.leadRepo.Put(ctx, lead); err != nil {
		return nil, utils.ErrInternal
	}
	return lead, nil
}

func (s *LeadService) Analytics(ctx context.Context, ownerUserID string) (*models.AgentAnalyticsResponse, error) {
	leads, err := s.ListLeads(ctx, ownerUserID, "", false)
	if err != nil {
		return nil, err
	}
	now := s.now()
	out := &models.AgentAnalyticsResponse{
		LeadCount: len(leads),
	}
	if len(leads) == 0 {
		return out, nil
	}

	viewingCount := 0
	dealCount := 0
	overdue := 0
	pending := 0

	for _, lead := range leads {
		switch lead.Stage {
		case models.LeadStageViewing, models.LeadStageNegotiation, models.LeadStageDeal:
			viewingCount++
		}
		if lead.Stage == models.LeadStageDeal {
			dealCount++
		}
		for _, task := range lead.FollowUpTasks {
			if task.Status != models.FollowUpTaskStatusPending {
				continue
			}
			pending++
			if task.DueAt.Before(now) {
				overdue++
			}
		}
	}

	out.PendingFollowUpCount = pending
	out.OverdueFollowUpCount = overdue
	out.LeadToViewingRate = ratio(viewingCount, len(leads))
	if viewingCount > 0 {
		out.ViewingToDealRate = ratio(dealCount, viewingCount)
	}
	if pending > 0 {
		out.OverdueFollowUpRate = ratio(overdue, pending)
	}
	return out, nil
}

func ratio(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func hasDuePendingTask(tasks []models.FollowUpTask, now time.Time) bool {
	for _, task := range tasks {
		if task.Status == models.FollowUpTaskStatusPending && !task.DueAt.After(now) {
			return true
		}
	}
	return false
}

func (s *LeadService) ensureFollowUpTasks(existing []models.FollowUpTask, anchor time.Time, leadID string) []models.FollowUpTask {
	offsets := []int{1, 3, 7}
	now := s.now()
	if len(existing) == 0 {
		tasks := make([]models.FollowUpTask, 0, len(offsets))
		for _, offset := range offsets {
			due := anchor.AddDate(0, 0, offset)
			tasks = append(tasks, models.FollowUpTask{
				TaskID:     s.newID(),
				LeadID:     leadID,
				OffsetDays: offset,
				DueAt:      due,
				Status:     models.FollowUpTaskStatusPending,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
		return tasks
	}

	present := make(map[int]bool, len(existing))
	for _, task := range existing {
		present[task.OffsetDays] = true
	}
	for _, offset := range offsets {
		if present[offset] {
			continue
		}
		due := anchor.AddDate(0, 0, offset)
		existing = append(existing, models.FollowUpTask{
			TaskID:     s.newID(),
			LeadID:     leadID,
			OffsetDays: offset,
			DueAt:      due,
			Status:     models.FollowUpTaskStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	sort.SliceStable(existing, func(i, j int) bool {
		return existing[i].DueAt.Before(existing[j].DueAt)
	})
	return existing
}

func isValidLeadStage(stage models.LeadStage) bool {
	switch stage {
	case models.LeadStageNew, models.LeadStageInterested, models.LeadStageViewing, models.LeadStageNegotiation, models.LeadStageDeal, models.LeadStageLost:
		return true
	default:
		return false
	}
}
