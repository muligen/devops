// Package service provides monitoring business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/agentteams/server/internal/modules/monitor/domain"
	"gorm.io/gorm"
)

// AlertEventRepository handles alert event data persistence.
type AlertEventRepository struct {
	db *gorm.DB
}

// NewAlertEventRepository creates a new alert event repository.
func NewAlertEventRepository(db *gorm.DB) *AlertEventRepository {
	return &AlertEventRepository{db: db}
}

// Create creates a new alert event.
func (r *AlertEventRepository) Create(ctx context.Context, event *domain.AlertEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetByID gets an alert event by ID.
func (r *AlertEventRepository) GetByID(ctx context.Context, id string) (*domain.AlertEvent, error) {
	var event domain.AlertEvent
	result := r.db.WithContext(ctx).
		Preload("Rule").
		Preload("Acknowledger").
		First(&event, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAlertEventNotFound
		}
		return nil, result.Error
	}
	return &event, nil
}

// ListOptions represents options for listing alert events.
type AlertEventListOptions struct {
	Status    string
	AgentID   string
	RuleID    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// List lists alert events with filtering and pagination.
func (r *AlertEventRepository) List(ctx context.Context, opts AlertEventListOptions) ([]domain.AlertEvent, int64, error) {
	var events []domain.AlertEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AlertEvent{}).
		Preload("Rule").
		Preload("Acknowledger")

	// Apply filters
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	if opts.AgentID != "" {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.RuleID != "" {
		query = query.Where("rule_id = ?", opts.RuleID)
	}
	if opts.StartTime != nil {
		query = query.Where("triggered_at >= ?", opts.StartTime)
	}
	if opts.EndTime != nil {
		query = query.Where("triggered_at <= ?", opts.EndTime)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}

	offset := (opts.Page - 1) * opts.PageSize
	result := query.Order("triggered_at DESC").
		Offset(offset).
		Limit(opts.PageSize).
		Find(&events)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return events, total, nil
}

// Update updates an alert event.
func (r *AlertEventRepository) Update(ctx context.Context, event *domain.AlertEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

// Delete deletes an alert event.
func (r *AlertEventRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.AlertEvent{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAlertEventNotFound
	}
	return nil
}

// GetPendingCount gets the count of pending alert events.
func (r *AlertEventRepository) GetPendingCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.AlertEvent{}).
		Where("status = ?", domain.AlertStatusPending).
		Count(&count).Error
	return count, err
}

// GetRecentCount gets the count of alert events in the last N hours.
func (r *AlertEventRepository) GetRecentCount(ctx context.Context, hours int) (int64, error) {
	var count int64
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	err := r.db.WithContext(ctx).Model(&domain.AlertEvent{}).
		Where("triggered_at >= ?", since).
		Count(&count).Error
	return count, err
}

// ListWithAgentInfo lists alert events with agent name populated.
func (r *AlertEventRepository) ListWithAgentInfo(ctx context.Context, opts AlertEventListOptions) ([]domain.AlertEvent, int64, error) {
	var events []domain.AlertEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AlertEvent{}).
		Select("alert_events.*, alert_rules.name as rule_name, agents.name as agent_name").
		Joins("LEFT JOIN alert_rules ON alert_events.rule_id = alert_rules.id").
		Joins("LEFT JOIN agents ON alert_events.agent_id = agents.id")

	// Apply filters
	if opts.Status != "" {
		query = query.Where("alert_events.status = ?", opts.Status)
	}
	if opts.AgentID != "" {
		query = query.Where("alert_events.agent_id = ?", opts.AgentID)
	}
	if opts.RuleID != "" {
		query = query.Where("alert_events.rule_id = ?", opts.RuleID)
	}
	if opts.StartTime != nil {
		query = query.Where("alert_events.triggered_at >= ?", opts.StartTime)
	}
	if opts.EndTime != nil {
		query = query.Where("alert_events.triggered_at <= ?", opts.EndTime)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}

	offset := (opts.Page - 1) * opts.PageSize
	result := query.Order("alert_events.triggered_at DESC").
		Offset(offset).
		Limit(opts.PageSize).
		Find(&events)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return events, total, nil
}
