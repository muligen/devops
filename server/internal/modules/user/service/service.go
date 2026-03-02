// Package service provides user management business logic.
package service

import (
	"context"
	"time"

	"github.com/agentteams/server/internal/modules/user/domain"
	"github.com/agentteams/server/internal/pkg/logger"
	"gorm.io/gorm"
)

// AuditRepository handles audit log persistence.
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository.
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates an audit log entry.
func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List lists audit logs with filtering and pagination.
func (r *AuditRepository) List(ctx context.Context, userID, action, resourceType string, start, end time.Time, page, pageSize int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if !start.IsZero() {
		query = query.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("created_at <= ?", end)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return logs, total, nil
}

// Service provides user management business logic.
type Service struct {
	auditRepo *AuditRepository
	logger    *logger.Logger
}

// NewService creates a new user service.
func NewService(auditRepo *AuditRepository) *Service {
	return &Service{
		auditRepo: auditRepo,
	}
}

// SetLogger sets the logger.
func (s *Service) SetLogger(l *logger.Logger) {
	s.logger = l
}

// LogAudit logs an audit event.
func (s *Service) LogAudit(ctx context.Context, userID, action, resourceType, resourceID, ipAddress, userAgent string, details map[string]interface{}) error {
	log := &domain.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		if s.logger != nil {
			s.logger.Errorw("Failed to create audit log", "error", err)
		}
		return err
	}

	return nil
}

// GetAuditLogs retrieves audit logs with filtering.
func (s *Service) GetAuditLogs(ctx context.Context, userID, action, resourceType string, start, end time.Time, page, pageSize int) ([]domain.AuditLog, int64, error) {
	return s.auditRepo.List(ctx, userID, action, resourceType, start, end, page, pageSize)
}
