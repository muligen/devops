// Package service provides monitoring business logic.
package service

import (
	"context"
	"time"

	"github.com/agentteams/server/internal/modules/monitor/domain"
)

// AlertEventService provides alert event business logic.
type AlertEventService struct {
	repo      *AlertEventRepository
	agentRepo *Repository
}

// NewAlertEventService creates a new alert event service.
func NewAlertEventService(repo *AlertEventRepository, agentRepo *Repository) *AlertEventService {
	return &AlertEventService{
		repo:      repo,
		agentRepo: agentRepo,
	}
}

// CreateEvent creates a new alert event when an alert is triggered.
func (s *AlertEventService) CreateEvent(ctx context.Context, ruleID, agentID string, metricValue, threshold float64, message string) (*domain.AlertEvent, error) {
	event := &domain.AlertEvent{
		RuleID:      &ruleID,
		AgentID:     agentID,
		MetricValue: metricValue,
		Threshold:   threshold,
		Status:      domain.AlertStatusPending,
		Message:     message,
		TriggeredAt: time.Now(),
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// GetEvent gets an alert event by ID.
func (s *AlertEventService) GetEvent(ctx context.Context, id string) (*domain.AlertEvent, error) {
	return s.repo.GetByID(ctx, id)
}

// ListEvents lists alert events with filtering.
func (s *AlertEventService) ListEvents(ctx context.Context, opts AlertEventListOptions) ([]domain.AlertEvent, int64, error) {
	return s.repo.ListWithAgentInfo(ctx, opts)
}

// AcknowledgeEvent acknowledges an alert event.
func (s *AlertEventService) AcknowledgeEvent(ctx context.Context, id, userID string) (*domain.AlertEvent, error) {
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !event.IsPending() {
		return nil, ErrAlertEventNotFound
	}

	event.Acknowledge(userID)
	if err := s.repo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// ResolveEvent resolves an alert event.
func (s *AlertEventService) ResolveEvent(ctx context.Context, id string) (*domain.AlertEvent, error) {
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if event.IsResolved() {
		return event, nil // Already resolved
	}

	event.Resolve()
	if err := s.repo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// BatchAcknowledge acknowledges multiple alert events.
func (s *AlertEventService) BatchAcknowledge(ctx context.Context, ids []string, userID string) (int, error) {
	count := 0
	for _, id := range ids {
		event, err := s.repo.GetByID(ctx, id)
		if err != nil {
			continue
		}

		if event.IsPending() {
			event.Acknowledge(userID)
			if err := s.repo.Update(ctx, event); err == nil {
				count++
			}
		}
	}
	return count, nil
}

// GetPendingCount gets the count of pending alert events.
func (s *AlertEventService) GetPendingCount(ctx context.Context) (int64, error) {
	return s.repo.GetPendingCount(ctx)
}

// GetRecentCount gets the count of alert events in the last N hours.
func (s *AlertEventService) GetRecentCount(ctx context.Context, hours int) (int64, error) {
	return s.repo.GetRecentCount(ctx, hours)
}

// GetEventStats gets alert event statistics.
func (s *AlertEventService) GetEventStats(ctx context.Context) (*AlertEventStats, error) {
	pending, err := s.repo.GetPendingCount(ctx)
	if err != nil {
		return nil, err
	}

	recent24h, err := s.repo.GetRecentCount(ctx, 24)
	if err != nil {
		return nil, err
	}

	return &AlertEventStats{
		PendingCount:  pending,
		Recent24hCount: recent24h,
	}, nil
}

// AlertEventStats represents alert event statistics.
type AlertEventStats struct {
	PendingCount  int64 `json:"pending_count"`
	Recent24hCount int64 `json:"recent_24h_count"`
}
