// Package service provides agent business logic.
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/agentteams/server/internal/modules/agent/domain"
	"gorm.io/gorm"
)

var (
	// ErrAgentNotFound is returned when agent is not found.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrAgentExists is returned when agent already exists.
	ErrAgentExists = errors.New("agent already exists")
	// ErrInvalidToken is returned when token is invalid.
	ErrInvalidToken = errors.New("invalid token")
)

// Repository handles agent data persistence.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new agent repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new agent.
func (r *Repository) Create(ctx context.Context, agent *domain.Agent) error {
	result := r.db.WithContext(ctx).Create(agent)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAgentExists
		}
		return result.Error
	}
	return nil
}

// GetByID gets an agent by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	var agent domain.Agent
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&agent)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, result.Error
	}
	return &agent, nil
}

// GetByName gets an agent by name.
func (r *Repository) GetByName(ctx context.Context, name string) (*domain.Agent, error) {
	var agent domain.Agent
	result := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&agent)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, result.Error
	}
	return &agent, nil
}

// GetByTokenHash gets an agent by token hash.
func (r *Repository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Agent, error) {
	var agent domain.Agent
	result := r.db.WithContext(ctx).Where("token_hash = ? AND deleted_at IS NULL", tokenHash).First(&agent)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, result.Error
	}
	return &agent, nil
}

// List lists agents with pagination.
func (r *Repository) List(ctx context.Context, page, pageSize int, status string) ([]domain.Agent, int64, error) {
	var agents []domain.Agent
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Agent{}).Where("deleted_at IS NULL")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&agents)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return agents, total, nil
}

// Update updates an agent.
func (r *Repository) Update(ctx context.Context, agent *domain.Agent) error {
	result := r.db.WithContext(ctx).Save(agent)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft deletes an agent.
func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAgentNotFound
	}
	return nil
}

// UpdateStatus updates agent status.
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":      status,
			"last_seen_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAgentNotFound
	}
	return nil
}

// UpdateLastSeen updates the last seen time.
func (r *Repository) UpdateLastSeen(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("last_seen_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Service provides agent business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new agent service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateAgent creates a new agent with generated token.
func (s *Service) CreateAgent(ctx context.Context, name string, metadata domain.JSONB) (*domain.Agent, string, error) {
	// Check if name exists
	if _, err := s.repo.GetByName(ctx, name); err == nil {
		return nil, "", ErrAgentExists
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		return nil, "", err
	}

	// Hash token for storage
	tokenHash := hashToken(token)

	agent := &domain.Agent{
		Name:      name,
		TokenHash: tokenHash,
		Status:    domain.StatusOffline,
		Metadata:  metadata,
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, "", err
	}

	return agent, token, nil
}

// GetAgent gets an agent by ID.
func (s *Service) GetAgent(ctx context.Context, id string) (*domain.Agent, error) {
	return s.repo.GetByID(ctx, id)
}

// GetAgentByToken gets an agent by token.
func (s *Service) GetAgentByToken(ctx context.Context, token string) (*domain.Agent, error) {
	tokenHash := hashToken(token)
	return s.repo.GetByTokenHash(ctx, tokenHash)
}

// ListAgents lists agents with pagination.
func (s *Service) ListAgents(ctx context.Context, page, pageSize int, status string) ([]domain.Agent, int64, error) {
	return s.repo.List(ctx, page, pageSize, status)
}

// UpdateAgentStatus updates agent status.
func (s *Service) UpdateAgentStatus(ctx context.Context, id, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// UpdateAgentMetadata updates agent metadata.
func (s *Service) UpdateAgentMetadata(ctx context.Context, id string, metadata domain.JSONB) error {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	agent.Metadata = metadata
	return s.repo.Update(ctx, agent)
}

// UpdateAgentInfo updates agent info (hostname, IP, OS, version).
func (s *Service) UpdateAgentInfo(ctx context.Context, id, hostname, ipAddress, osInfo, version string) error {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if hostname != "" {
		agent.Hostname = hostname
	}
	if ipAddress != "" {
		agent.IPAddress = ipAddress
	}
	if osInfo != "" {
		agent.OSInfo = osInfo
	}
	if version != "" {
		agent.Version = version
	}

	return s.repo.Update(ctx, agent)
}

// DeleteAgent deletes an agent.
func (s *Service) DeleteAgent(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// generateToken generates a random token.
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken hashes a token using SHA256.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
