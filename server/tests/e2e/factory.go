// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	agentdomain "github.com/agentteams/server/internal/modules/agent/domain"
	authdomain "github.com/agentteams/server/internal/modules/auth/domain"
	monitordomain "github.com/agentteams/server/internal/modules/monitor/domain"
	taskdomain "github.com/agentteams/server/internal/modules/task/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Factory provides functions to create test data on demand
type Factory struct {
	db *gorm.DB
}

// NewFactory creates a new factory
func NewFactory(db *gorm.DB) *Factory {
	return &Factory{db: db}
}

// CreateUserOption is a function that modifies a user during creation
type CreateUserOption func(*authdomain.User)

// CreateUser creates a test user with optional customization
func (f *Factory) CreateUser(ctx context.Context, opts ...CreateUserOption) (*authdomain.User, error) {
	user := &authdomain.User{
		ID:           uuid.NewString(),
		Username:     fmt.Sprintf("test-user-%s", uuid.NewString()[:8]),
		Email:        fmt.Sprintf("test-%s@test.com", uuid.NewString()[:8]),
		PasswordHash: hashToken("test-password"),
		Role:         authdomain.RoleViewer,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, opt := range opts {
		opt(user)
	}

	if err := f.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// WithUsername sets the username
func WithUsername(username string) CreateUserOption {
	return func(u *authdomain.User) {
		u.Username = username
	}
}

// WithEmail sets the email
func WithEmail(email string) CreateUserOption {
	return func(u *authdomain.User) {
		u.Email = email
	}
}

// WithRole sets the role
func WithRole(role string) CreateUserOption {
	return func(u *authdomain.User) {
		u.Role = role
	}
}

// CreateAgentOption is a function that modifies an agent during creation
type CreateAgentOption func(*agentdomain.Agent)

// CreateAgent creates a test agent with optional customization
func (f *Factory) CreateAgent(ctx context.Context, opts ...CreateAgentOption) (*agentdomain.Agent, error) {
	token := fmt.Sprintf("test-token-%s", uuid.NewString())
	agent := &agentdomain.Agent{
		ID:        uuid.NewString(),
		Name:      fmt.Sprintf("test-agent-%s", uuid.NewString()[:8]),
		TokenHash: hashToken(token),
		Status:    agentdomain.StatusOffline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(agent)
	}

	if err := f.db.WithContext(ctx).Create(agent).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Return agent with its plain token for testing
	agent.TokenHash = token // Store plain token for caller
	return agent, nil
}

// WithAgentName sets the agent name
func WithAgentName(name string) CreateAgentOption {
	return func(a *agentdomain.Agent) {
		a.Name = name
	}
}

// WithAgentStatus sets the agent status
func WithAgentStatus(status string) CreateAgentOption {
	return func(a *agentdomain.Agent) {
		a.Status = status
	}
}

// WithAgentIP sets the agent IP address
func WithAgentIP(ip string) CreateAgentOption {
	return func(a *agentdomain.Agent) {
		a.IPAddress = ip
	}
}

// CreateTaskOption is a function that modifies a task during creation
type CreateTaskOption func(*taskdomain.Task)

// CreateTask creates a test task with optional customization
func (f *Factory) CreateTask(ctx context.Context, agentID string, opts ...CreateTaskOption) (*taskdomain.Task, error) {
	task := &taskdomain.Task{
		ID:        uuid.NewString(),
		AgentID:   agentID,
		Type:      taskdomain.TypeExecShell,
		Params:    taskdomain.JSONB{"command": "echo hello"},
		Status:    taskdomain.StatusPending,
		Timeout:   60,
		CreatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(task)
	}

	if err := f.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// WithTaskType sets the task type
func WithTaskType(taskType string) CreateTaskOption {
	return func(t *taskdomain.Task) {
		t.Type = taskType
	}
}

// WithTaskParams sets the task params
func WithTaskParams(params taskdomain.JSONB) CreateTaskOption {
	return func(t *taskdomain.Task) {
		t.Params = params
	}
}

// WithTaskStatus sets the task status
func WithTaskStatus(status string) CreateTaskOption {
	return func(t *taskdomain.Task) {
		t.Status = status
	}
}

// WithTimeout sets the task timeout
func WithTimeout(timeout int) CreateTaskOption {
	return func(t *taskdomain.Task) {
		t.Timeout = timeout
	}
}

// CreateMetric creates a test metric record
func (f *Factory) CreateMetric(ctx context.Context, agentID string, cpu float64, memPercent float64) error {
	metric := &monitordomain.AgentMetric{
		ID:            uuid.NewString(),
		AgentID:       agentID,
		CPUUsage:      cpu,
		MemoryTotal:   16 * 1024 * 1024 * 1024, // 16GB
		MemoryUsed:    int64(float64(16*1024*1024*1024) * memPercent / 100),
		MemoryPercent: memPercent,
		DiskTotal:     500 * 1024 * 1024 * 1024, // 500GB
		DiskUsed:      250 * 1024 * 1024 * 1024, // 250GB
		DiskPercent:   50,
		Uptime:        3600,
		CollectedAt:   time.Now(),
	}

	return f.db.WithContext(ctx).Create(metric).Error
}

// ComputeTokenHash computes the token hash used by the system
func ComputeTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
