// Package service provides task business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/agentteams/server/internal/modules/task/domain"
	"gorm.io/gorm"
)

var (
	// ErrTaskNotFound is returned when task is not found.
	ErrTaskNotFound = errors.New("task not found")
	// ErrAgentOffline is returned when agent is offline.
	ErrAgentOffline = errors.New("agent offline")
	// ErrTaskAlreadyRunning is returned when trying to cancel a non-pending task.
	ErrTaskAlreadyRunning = errors.New("task already running")
)

// Repository handles task data persistence.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new task repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new task.
func (r *Repository) Create(ctx context.Context, task *domain.Task) error {
	result := r.db.WithContext(ctx).Create(task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID gets a task by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, result.Error
	}
	return &task, nil
}

// List lists tasks with pagination.
func (r *Repository) List(ctx context.Context, page, pageSize int, agentID, status string) ([]domain.Task, int64, error) {
	var tasks []domain.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Task{})

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return tasks, total, nil
}

// Update updates a task.
func (r *Repository) Update(ctx context.Context, task *domain.Task) error {
	result := r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateStatus updates task status.
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&domain.Task{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// ListPendingByAgent lists pending tasks for an agent.
func (r *Repository) ListPendingByAgent(ctx context.Context, agentID string) ([]domain.Task, error) {
	var tasks []domain.Task
	result := r.db.WithContext(ctx).
		Where("agent_id = ? AND status = ?", agentID, domain.StatusPending).
		Order("priority DESC, created_at ASC").
		Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// Service provides task business logic.
type Service struct {
	repo       *Repository
	mq         TaskQueue
	dispatcher TaskDispatcher
}

// TaskQueue defines the interface for task queue operations.
type TaskQueue interface {
	PublishTask(ctx context.Context, taskID string, taskData map[string]interface{}) error
}

// TaskDispatcher defines the interface for task dispatching.
type TaskDispatcher interface {
	SendCommand(agentID string, command map[string]interface{}) error
	IsAgentConnected(agentID string) bool
}

// NewService creates a new task service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SetQueue sets the message queue.
func (s *Service) SetQueue(mq TaskQueue) {
	s.mq = mq
}

// SetDispatcher sets the task dispatcher.
func (s *Service) SetDispatcher(dispatcher TaskDispatcher) {
	s.dispatcher = dispatcher
}

// CreateTask creates a new task.
func (s *Service) CreateTask(ctx context.Context, agentID, taskType string, params domain.JSONB, priority, timeout int, createdBy string) (*domain.Task, error) {
	task := &domain.Task{
		AgentID:   agentID,
		Type:      taskType,
		Params:    params,
		Status:    domain.StatusPending,
		Priority:  priority,
		Timeout:   timeout,
		CreatedBy: createdBy,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Publish task created event
	if s.mq != nil {
		_ = s.mq.PublishTask(ctx, task.ID, map[string]interface{}{
			"task_id":  task.ID,
			"agent_id": task.AgentID,
			"type":     task.Type,
			"status":   task.Status,
		})
	}

	// Dispatch task if agent is connected
	if s.dispatcher != nil && s.dispatcher.IsAgentConnected(agentID) {
		_ = s.dispatchTask(ctx, task)
	}

	return task, nil
}

// CreateBatchTasks creates multiple tasks.
func (s *Service) CreateBatchTasks(ctx context.Context, requests []CreateTaskRequest, createdBy string) ([]*domain.Task, error) {
	tasks := make([]*domain.Task, len(requests))

	for i, req := range requests {
		task, err := s.CreateTask(ctx, req.AgentID, req.Type, req.Params, req.Priority, req.Timeout, createdBy)
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}

	return tasks, nil
}

// CreateTaskRequest represents a task creation request.
type CreateTaskRequest struct {
	AgentID  string       `json:"agent_id"`
	Type     string       `json:"type"`
	Params   domain.JSONB `json:"params"`
	Priority int          `json:"priority"`
	Timeout  int          `json:"timeout"`
}

// GetTask gets a task by ID.
func (s *Service) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

// ListTasks lists tasks with pagination.
func (s *Service) ListTasks(ctx context.Context, page, pageSize int, agentID, status string) ([]domain.Task, int64, error) {
	return s.repo.List(ctx, page, pageSize, agentID, status)
}

// CancelTask cancels a pending task.
func (s *Service) CancelTask(ctx context.Context, id string) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !task.IsPending() {
		return ErrTaskAlreadyRunning
	}

	now := time.Now()
	task.Status = domain.StatusCancelled
	task.CompletedAt = &now

	return s.repo.Update(ctx, task)
}

// UpdateTaskResult updates task execution result.
func (s *Service) UpdateTaskResult(ctx context.Context, id string, status string, exitCode int, output string, duration float64) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	task.Status = status
	task.ExitCode = &exitCode
	task.Output = output
	task.Duration = &duration
	task.CompletedAt = &now

	return s.repo.Update(ctx, task)
}

// StartTask marks a task as running.
func (s *Service) StartTask(ctx context.Context, id string) error {
	now := time.Now()
	result := s.repo.db.WithContext(ctx).Model(&domain.Task{}).
		Where("id = ? AND status = ?", id, domain.StatusPending).
		Updates(map[string]interface{}{
			"status":     domain.StatusRunning,
			"started_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// dispatchTask dispatches a task to the connected agent.
func (s *Service) dispatchTask(ctx context.Context, task *domain.Task) error {
	if s.dispatcher == nil {
		return nil
	}

	command := map[string]interface{}{
		"command_id":   task.ID,
		"command_type": task.Type,
		"params":       task.Params,
		"timeout":      task.Timeout,
	}

	return s.dispatcher.SendCommand(task.AgentID, command)
}

// GetPendingTasksForAgent gets pending tasks for an agent.
func (s *Service) GetPendingTasksForAgent(ctx context.Context, agentID string) ([]domain.Task, error) {
	return s.repo.ListPendingByAgent(ctx, agentID)
}
