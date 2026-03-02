// Package service provides monitoring business logic.
package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/agentteams/server/internal/modules/monitor/domain"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/mq"
	"gorm.io/gorm"
)

var (
	// ErrMetricNotFound is returned when metric is not found.
	ErrMetricNotFound = errors.New("metric not found")
	// ErrAlertRuleNotFound is returned when alert rule is not found.
	ErrAlertRuleNotFound = errors.New("alert rule not found")
)

// Time-series optimization constants
const (
	// DefaultMetricRetentionDays is the default retention period for raw metrics
	DefaultMetricRetentionDays = 7
	// DefaultAggregationInterval is the interval for aggregating old metrics
	DefaultAggregationInterval = 5 * time.Minute
)

// MetricBatch holds metrics for batch insertion.
type MetricBatch struct {
	metrics []*domain.AgentMetric
	mu      sync.Mutex
}

// NewMetricBatch creates a new metric batch.
func NewMetricBatch() *MetricBatch {
	return &MetricBatch{
		metrics: make([]*domain.AgentMetric, 0, 100),
	}
}

// Add adds a metric to the batch.
func (b *MetricBatch) Add(m *domain.AgentMetric) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metrics = append(b.metrics, m)
}

// Flush flushes the batch to the database and returns the count.
func (b *MetricBatch) Flush(ctx context.Context, db *gorm.DB) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.metrics) == 0 {
		return 0, nil
	}

	// Batch insert
	if err := db.WithContext(ctx).CreateInBatches(b.metrics, 100).Error; err != nil {
		return 0, err
	}

	count := len(b.metrics)
	b.metrics = make([]*domain.AgentMetric, 0, 100)
	return count, nil
}

// Repository handles monitor data persistence.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new monitor repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateMetric creates a new metric record.
func (r *Repository) CreateMetric(ctx context.Context, metric *domain.AgentMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// GetLatestMetric gets the latest metric for an agent.
func (r *Repository) GetLatestMetric(ctx context.Context, agentID string) (*domain.AgentMetric, error) {
	var metric domain.AgentMetric
	result := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("collected_at DESC").
		First(&metric)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, result.Error
	}
	return &metric, nil
}

// GetMetricsByAgent gets metrics for an agent within a time range.
func (r *Repository) GetMetricsByAgent(ctx context.Context, agentID string, start, end time.Time, limit int) ([]domain.AgentMetric, error) {
	var metrics []domain.AgentMetric
	query := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Where("collected_at >= ? AND collected_at <= ?", start, end).
		Order("collected_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&metrics)
	if result.Error != nil {
		return nil, result.Error
	}
	return metrics, nil
}

// DeleteOldMetrics deletes metrics older than the specified duration.
func (r *Repository) DeleteOldMetrics(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("collected_at < ?", olderThan).
		Delete(&domain.AgentMetric{}).Error
}

// CreateAlertRule creates an alert rule.
func (r *Repository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetAlertRule gets an alert rule by ID.
func (r *Repository) GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error) {
	var rule domain.AlertRule
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAlertRuleNotFound
		}
		return nil, result.Error
	}
	return &rule, nil
}

// ListAlertRules lists all alert rules.
func (r *Repository) ListAlertRules(ctx context.Context, enabledOnly bool) ([]domain.AlertRule, error) {
	var rules []domain.AlertRule
	query := r.db.WithContext(ctx)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	result := query.Order("created_at DESC").Find(&rules)
	if result.Error != nil {
		return nil, result.Error
	}
	return rules, nil
}

// UpdateAlertRule updates an alert rule.
func (r *Repository) UpdateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteAlertRule deletes an alert rule.
func (r *Repository) DeleteAlertRule(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.AlertRule{}, "id = ?", id).Error
}

// Service provides monitoring business logic.
type Service struct {
	repo          *Repository
	mq            *mq.Client
	logger        *logger.Logger
	batch         *MetricBatch
	agentStatusFn func(ctx context.Context, agentID string, status string) error
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewService creates a new monitor service.
func NewService(repo *Repository) *Service {
	return &Service{
		repo:  repo,
		batch: NewMetricBatch(),
		stopCh: make(chan struct{}),
	}
}

// SetMQ sets the message queue client.
func (s *Service) SetMQ(mq *mq.Client) {
	s.mq = mq
}

// SetLogger sets the logger.
func (s *Service) SetLogger(l *logger.Logger) {
	s.logger = l
}

// SetAgentStatusFn sets the function to update agent status.
func (s *Service) SetAgentStatusFn(fn func(ctx context.Context, agentID string, status string) error) {
	s.agentStatusFn = fn
}

// Start starts background workers.
func (s *Service) Start(ctx context.Context) {
	// Start batch flush worker
	s.wg.Add(1)
	go s.batchFlushWorker(ctx)

	// Start heartbeat timeout detector
	s.wg.Add(1)
	go s.heartbeatTimeoutWorker(ctx)

	// Start metric cleanup worker
	s.wg.Add(1)
	go s.metricCleanupWorker(ctx)

	// Start alert evaluation worker
	s.wg.Add(1)
	go s.alertEvaluationWorker(ctx)
}

// Stop stops background workers.
func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// batchFlushWorker periodically flushes the metric batch.
func (s *Service) batchFlushWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush on shutdown
			if _, err := s.batch.Flush(ctx, s.repo.db); err != nil && s.logger != nil {
				s.logger.Errorw("Failed to flush metrics on shutdown", "error", err)
			}
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if count, err := s.batch.Flush(ctx, s.repo.db); err != nil && s.logger != nil {
				s.logger.Errorw("Failed to flush metrics batch", "error", err)
			} else if count > 0 && s.logger != nil {
				s.logger.Debugw("Flushed metrics batch", "count", count)
			}
		}
	}
}

// heartbeatTimeoutWorker detects agents that haven't sent heartbeats.
func (s *Service) heartbeatTimeoutWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	timeout := 90 * time.Second // Consider offline after 90 seconds without heartbeat

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.detectTimedOutAgents(ctx, timeout)
		}
	}
}

// detectTimedOutAgents marks agents as offline if they haven't sent heartbeats.
func (s *Service) detectTimedOutAgents(ctx context.Context, timeout time.Duration) {
	// Query agents that are online but haven't been seen recently
	var agentIDs []string
	threshold := time.Now().Add(-timeout)

	result := s.repo.db.WithContext(ctx).
		Table("agents").
		Where("status = ? AND last_seen_at < ?", "online", threshold).
		Pluck("id", &agentIDs)

	if result.Error != nil {
		if s.logger != nil {
			s.logger.Errorw("Failed to query timed out agents", "error", result.Error)
		}
		return
	}

	for _, agentID := range agentIDs {
		if s.agentStatusFn != nil {
			if err := s.agentStatusFn(ctx, agentID, "offline"); err != nil && s.logger != nil {
				s.logger.Errorw("Failed to mark agent as offline", "error", err, "agent_id", agentID)
			} else if s.logger != nil {
				s.logger.Infow("Agent marked as offline due to heartbeat timeout", "agent_id", agentID)
			}
		}

		// Publish offline event
		if s.mq != nil {
			_ = s.mq.PublishEvent(ctx, mq.EventAgentOffline, map[string]interface{}{
				"agent_id":  agentID,
				"reason":    "heartbeat_timeout",
				"timestamp": time.Now().Unix(),
			})
		}
	}
}

// metricCleanupWorker periodically cleans up old metrics.
func (s *Service) metricCleanupWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.CleanupOldMetrics(ctx, DefaultMetricRetentionDays); err != nil && s.logger != nil {
				s.logger.Errorw("Failed to cleanup old metrics", "error", err)
			}
		}
	}
}

// alertEvaluationWorker periodically evaluates alert rules.
func (s *Service) alertEvaluationWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.evaluateAllRules(ctx)
		}
	}
}

// evaluateAllRules evaluates all enabled alert rules.
func (s *Service) evaluateAllRules(ctx context.Context) {
	rules, err := s.repo.ListAlertRules(ctx, true)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("Failed to list alert rules", "error", err)
		}
		return
	}

	for _, rule := range rules {
		if err := s.evaluateRule(ctx, &rule); err != nil && s.logger != nil {
			s.logger.Errorw("Failed to evaluate alert rule", "error", err, "rule_id", rule.ID)
		}
	}
}

// evaluateRule evaluates a single alert rule.
func (s *Service) evaluateRule(ctx context.Context, rule *domain.AlertRule) error {
	// Get recent metrics for the rule's metric type
	// For simplicity, we'll check the latest metrics for all agents
	var metrics []domain.AgentMetric

	// Get metrics from last duration window
	since := time.Now().Add(-time.Duration(rule.Duration) * time.Second)
	result := s.repo.db.WithContext(ctx).
		Where("collected_at >= ?", since).
		Order("collected_at DESC").
		Find(&metrics)

	if result.Error != nil {
		return result.Error
	}

	// Group by agent and check threshold
	agentMetrics := make(map[string][]domain.AgentMetric)
	for _, m := range metrics {
		agentMetrics[m.AgentID] = append(agentMetrics[m.AgentID], m)
	}

	for agentID, agentMetricList := range agentMetrics {
		if len(agentMetricList) == 0 {
			continue
		}

		// Get the value to check based on metric type
		var value float64
		latest := agentMetricList[0]

		switch rule.MetricType {
		case domain.MetricCPU:
			value = latest.CPUUsage
		case domain.MetricMemory:
			value = latest.MemoryPercent
		case domain.MetricDisk:
			value = latest.DiskPercent
		default:
			continue
		}

		// Evaluate the condition
		if s.EvaluateRule(rule, value) {
			// Alert triggered
			if s.logger != nil {
				s.logger.Warnw("Alert triggered",
					"rule_id", rule.ID,
					"rule_name", rule.Name,
					"agent_id", agentID,
					"metric_type", rule.MetricType,
					"value", value,
					"threshold", rule.Threshold,
				)
			}

			// Publish alert event
			if s.mq != nil {
				_ = s.mq.PublishEvent(ctx, "alert.triggered", map[string]interface{}{
					"rule_id":    rule.ID,
					"rule_name":  rule.Name,
					"agent_id":   agentID,
					"metric_type": rule.MetricType,
					"value":      value,
					"threshold":  rule.Threshold,
					"severity":   rule.Severity,
					"timestamp":  time.Now().Unix(),
				})
			}
		}
	}

	return nil
}

// StoreMetric stores a metric from an agent.
func (s *Service) StoreMetric(ctx context.Context, agentID string, cpuUsage float64, memTotal, memUsed int64, memPercent float64, diskTotal, diskUsed int64, diskPercent float64, uptime int64) error {
	metric := &domain.AgentMetric{
		AgentID:       agentID,
		CPUUsage:      cpuUsage,
		MemoryTotal:   memTotal,
		MemoryUsed:    memUsed,
		MemoryPercent: memPercent,
		DiskTotal:     diskTotal,
		DiskUsed:      diskUsed,
		DiskPercent:   diskPercent,
		Uptime:        uptime,
		CollectedAt:   time.Now(),
	}

	// Add to batch for time-series optimization
	s.batch.Add(metric)

	// Also do immediate insert for real-time queries
	return s.repo.CreateMetric(ctx, metric)
}

// StoreMetricBatch stores metrics in batch for better performance.
func (s *Service) StoreMetricBatch(ctx context.Context, metrics []*domain.AgentMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	return s.repo.db.WithContext(ctx).CreateInBatches(metrics, 100).Error
}

// GetLatestMetric gets the latest metric for an agent.
func (s *Service) GetLatestMetric(ctx context.Context, agentID string) (*domain.AgentMetric, error) {
	return s.repo.GetLatestMetric(ctx, agentID)
}

// GetAgentMetrics gets metrics for an agent within a time range.
func (s *Service) GetAgentMetrics(ctx context.Context, agentID string, start, end time.Time, limit int) ([]domain.AgentMetric, error) {
	return s.repo.GetMetricsByAgent(ctx, agentID, start, end, limit)
}

// CleanupOldMetrics deletes metrics older than retention period.
func (s *Service) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	olderThan := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOldMetrics(ctx, olderThan)
}

// CreateAlertRule creates a new alert rule.
func (s *Service) CreateAlertRule(ctx context.Context, name, description, metricType, condition string, threshold float64, duration int, severity string) (*domain.AlertRule, error) {
	rule := &domain.AlertRule{
		Name:        name,
		Description: description,
		MetricType:  metricType,
		Condition:   condition,
		Threshold:   threshold,
		Duration:    duration,
		Severity:    severity,
		Enabled:     true,
	}

	if err := s.repo.CreateAlertRule(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// GetAlertRule gets an alert rule by ID.
func (s *Service) GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error) {
	return s.repo.GetAlertRule(ctx, id)
}

// ListAlertRules lists all alert rules.
func (s *Service) ListAlertRules(ctx context.Context, enabledOnly bool) ([]domain.AlertRule, error) {
	return s.repo.ListAlertRules(ctx, enabledOnly)
}

// UpdateAlertRule updates an alert rule.
func (s *Service) UpdateAlertRule(ctx context.Context, id, name, description, metricType, condition string, threshold float64, duration int, severity string, enabled bool) (*domain.AlertRule, error) {
	rule, err := s.repo.GetAlertRule(ctx, id)
	if err != nil {
		return nil, err
	}

	rule.Name = name
	rule.Description = description
	rule.MetricType = metricType
	rule.Condition = condition
	rule.Threshold = threshold
	rule.Duration = duration
	rule.Severity = severity
	rule.Enabled = enabled

	if err := s.repo.UpdateAlertRule(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// DeleteAlertRule deletes an alert rule.
func (s *Service) DeleteAlertRule(ctx context.Context, id string) error {
	return s.repo.DeleteAlertRule(ctx, id)
}

// ToggleAlertRule enables or disables an alert rule.
func (s *Service) ToggleAlertRule(ctx context.Context, id string, enabled bool) error {
	rule, err := s.repo.GetAlertRule(ctx, id)
	if err != nil {
		return err
	}
	rule.Enabled = enabled
	return s.repo.UpdateAlertRule(ctx, rule)
}

// EvaluateRule evaluates an alert rule against a metric value.
func (s *Service) EvaluateRule(rule *domain.AlertRule, value float64) bool {
	switch rule.Condition {
	case domain.ConditionGreater:
		return value > rule.Threshold
	case domain.ConditionGreaterEqual:
		return value >= rule.Threshold
	case domain.ConditionLess:
		return value < rule.Threshold
	case domain.ConditionLessEqual:
		return value <= rule.Threshold
	case domain.ConditionEqual:
		return value == rule.Threshold
	case domain.ConditionNotEqual:
		return value != rule.Threshold
	default:
		return false
	}
}

// DashboardStats represents dashboard statistics.
type DashboardStats struct {
	TotalAgents      int64 `json:"total_agents"`
	OnlineAgents     int64 `json:"online_agents"`
	OfflineAgents    int64 `json:"offline_agents"`
	TotalTasks       int64 `json:"total_tasks"`
	PendingTasks     int64 `json:"pending_tasks"`
	RunningTasks     int64 `json:"running_tasks"`
	CompletedTasks   int64 `json:"completed_tasks"`
	FailedTasks      int64 `json:"failed_tasks"`
	AlertsTriggered  int64 `json:"alerts_triggered"`
}

// GetDashboardStats gets dashboard statistics.
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Agent stats
	if err := s.repo.db.WithContext(ctx).Model(&struct {
		TableName struct{} `gorm:"table:agents"`
	}{}).Where("deleted_at IS NULL").Count(&stats.TotalAgents).Error; err != nil {
		return nil, err
	}
	if err := s.repo.db.WithContext(ctx).Model(&struct {
		TableName struct{} `gorm:"table:agents"`
	}{}).Where("status = ? AND deleted_at IS NULL", "online").Count(&stats.OnlineAgents).Error; err != nil {
		return nil, err
	}
	stats.OfflineAgents = stats.TotalAgents - stats.OnlineAgents

	// Task stats
	taskModel := s.repo.db.WithContext(ctx).Table("tasks")
	if err := taskModel.Count(&stats.TotalTasks).Error; err != nil {
		return nil, err
	}
	if err := s.repo.db.WithContext(ctx).Table("tasks").Where("status = ?", "pending").Count(&stats.PendingTasks).Error; err != nil {
		return nil, err
	}
	if err := s.repo.db.WithContext(ctx).Table("tasks").Where("status = ?", "running").Count(&stats.RunningTasks).Error; err != nil {
		return nil, err
	}
	if err := s.repo.db.WithContext(ctx).Table("tasks").Where("status = ?", "success").Count(&stats.CompletedTasks).Error; err != nil {
		return nil, err
	}
	if err := s.repo.db.WithContext(ctx).Table("tasks").Where("status IN ?", []string{"failed", "timeout"}).Count(&stats.FailedTasks).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
