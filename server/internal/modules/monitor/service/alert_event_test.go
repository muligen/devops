package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentteams/server/internal/modules/monitor/domain"
	"github.com/agentteams/server/internal/modules/monitor/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables with SQLite-compatible schema
	err = db.Exec(`CREATE TABLE alert_events (
		id TEXT PRIMARY KEY,
		rule_id TEXT,
		agent_id TEXT NOT NULL,
		metric_value REAL NOT NULL,
		threshold REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		message TEXT,
		triggered_at DATETIME NOT NULL,
		resolved_at DATETIME,
		acknowledged_by TEXT,
		acknowledged_at DATETIME,
		created_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE alert_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		metric_type TEXT NOT NULL,
		condition TEXT NOT NULL,
		threshold REAL NOT NULL,
		duration INTEGER NOT NULL DEFAULT 60,
		severity TEXT NOT NULL DEFAULT 'warning',
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	return db
}

func TestAlertEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	event := &domain.AlertEvent{
		ID:          "test-event-1",
		AgentID:     "agent-1",
		MetricValue: 85.5,
		Threshold:   80.0,
		Status:      domain.AlertStatusPending,
		Message:     "CPU usage exceeded threshold",
		TriggeredAt: time.Now(),
	}

	err := repo.Create(ctx, event)
	require.NoError(t, err)
}

func TestAlertEventRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	// Create test event
	event := &domain.AlertEvent{
		ID:          "test-event-2",
		AgentID:     "agent-1",
		MetricValue: 85.5,
		Threshold:   80.0,
		Status:      domain.AlertStatusPending,
		TriggeredAt: time.Now(),
	}
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, retrieved.ID)
	assert.Equal(t, event.AgentID, retrieved.AgentID)
	assert.Equal(t, event.MetricValue, retrieved.MetricValue)

	// Not found
	_, err = repo.GetByID(ctx, "non-existent")
	assert.Equal(t, service.ErrAlertEventNotFound, err)
}

func TestAlertEventRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	// Create test events
	for i := 0; i < 5; i++ {
		status := domain.AlertStatusPending
		if i%2 == 0 {
			status = domain.AlertStatusAcknowledged
		}
		event := &domain.AlertEvent{
			ID:          string(rune('a' + i)),
			AgentID:     "agent-1",
			MetricValue: float64(80 + i),
			Threshold:   80.0,
			Status:      status,
			TriggeredAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	// List all
	events, total, err := repo.List(ctx, service.AlertEventListOptions{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, events, 5)

	// List by status
	events, total, err = repo.List(ctx, service.AlertEventListOptions{
		Status:   domain.AlertStatusPending,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total) // 2 pending events

	// Pagination
	events, total, err = repo.List(ctx, service.AlertEventListOptions{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, events, 2)
}

func TestAlertEventRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	// Create test event
	event := &domain.AlertEvent{
		ID:          "test-event-update",
		AgentID:     "agent-1",
		MetricValue: 85.5,
		Threshold:   80.0,
		Status:      domain.AlertStatusPending,
		TriggeredAt: time.Now(),
	}
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Update status
	event.Status = domain.AlertStatusAcknowledged
	err = repo.Update(ctx, event)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.AlertStatusAcknowledged, retrieved.Status)
}

func TestAlertEventRepository_GetPendingCount(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	// Create test events with different statuses
	for i := 0; i < 3; i++ {
		event := &domain.AlertEvent{
			ID:          string(rune('p' + i)),
			AgentID:     "agent-1",
			MetricValue: 85.0,
			Threshold:   80.0,
			Status:      domain.AlertStatusPending,
			TriggeredAt: time.Now(),
		}
		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		event := &domain.AlertEvent{
			ID:          string(rune('a' + i)),
			AgentID:     "agent-1",
			MetricValue: 85.0,
			Threshold:   80.0,
			Status:      domain.AlertStatusAcknowledged,
			TriggeredAt: time.Now(),
		}
		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	count, err := repo.GetPendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestAlertEventRepository_GetRecentCount(t *testing.T) {
	db := setupTestDB(t)
	repo := service.NewAlertEventRepository(db)
	ctx := context.Background()

	// Create recent event
	event := &domain.AlertEvent{
		ID:          "recent-event",
		AgentID:     "agent-1",
		MetricValue: 85.0,
		Threshold:   80.0,
		Status:      domain.AlertStatusPending,
		TriggeredAt: time.Now().Add(-1 * time.Hour),
	}
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Create old event (outside 24h)
	oldEvent := &domain.AlertEvent{
		ID:          "old-event",
		AgentID:     "agent-1",
		MetricValue: 85.0,
		Threshold:   80.0,
		Status:      domain.AlertStatusPending,
		TriggeredAt: time.Now().Add(-25 * time.Hour),
	}
	err = repo.Create(ctx, oldEvent)
	require.NoError(t, err)

	count, err := repo.GetRecentCount(ctx, 24)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only 1 recent event
}

func TestAlertEvent_Acknowledge(t *testing.T) {
	event := &domain.AlertEvent{
		Status: domain.AlertStatusPending,
	}

	assert.True(t, event.IsPending())

	userID := "user-123"
	event.Acknowledge(userID)

	assert.Equal(t, domain.AlertStatusAcknowledged, event.Status)
	assert.Equal(t, &userID, event.AcknowledgedBy)
	assert.True(t, event.AcknowledgedAt.Valid)
}

func TestAlertEvent_Resolve(t *testing.T) {
	event := &domain.AlertEvent{
		Status: domain.AlertStatusPending,
	}

	event.Resolve()

	assert.Equal(t, domain.AlertStatusResolved, event.Status)
	assert.True(t, event.ResolvedAt.Valid)
}
