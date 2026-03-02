// Package service provides update business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/agentteams/server/internal/modules/update/domain"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/mq"
	"github.com/agentteams/server/internal/pkg/storage"
	"gorm.io/gorm"
)

var (
	// ErrVersionNotFound is returned when version is not found.
	ErrVersionNotFound = errors.New("version not found")
	// ErrVersionAlreadyExists is returned when version already exists.
	ErrVersionAlreadyExists = errors.New("version already exists")
	// ErrUpdateInProgress is returned when an update is already in progress.
	ErrUpdateInProgress = errors.New("update already in progress")
)

// Repository handles update data persistence.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new update repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateVersion creates a new version.
func (r *Repository) CreateVersion(ctx context.Context, version *domain.Version) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersion gets a version by ID.
func (r *Repository) GetVersion(ctx context.Context, id string) (*domain.Version, error) {
	var version domain.Version
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&version)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, result.Error
	}
	return &version, nil
}

// GetVersionByNumber gets a version by version number and platform.
func (r *Repository) GetVersionByNumber(ctx context.Context, version, platform string) (*domain.Version, error) {
	var v domain.Version
	result := r.db.WithContext(ctx).Where("version = ? AND platform = ?", version, platform).First(&v)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, result.Error
	}
	return &v, nil
}

// ListVersions lists all versions.
func (r *Repository) ListVersions(ctx context.Context, platform string, limit int) ([]domain.Version, error) {
	var versions []domain.Version
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	result := query.Order("created_at DESC").Find(&versions)
	if result.Error != nil {
		return nil, result.Error
	}
	return versions, nil
}

// GetLatestVersion gets the latest version for a platform.
func (r *Repository) GetLatestVersion(ctx context.Context, platform string) (*domain.Version, error) {
	var version domain.Version
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	result := query.Order("created_at DESC").First(&version)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, result.Error
	}
	return &version, nil
}

// DeleteVersion deletes a version.
func (r *Repository) DeleteVersion(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Version{}, "id = ?", id).Error
}

// CreateUpdateStatus creates an update status record.
func (r *Repository) CreateUpdateStatus(ctx context.Context, status *domain.UpdateStatus) error {
	return r.db.WithContext(ctx).Create(status).Error
}

// GetLatestUpdateStatus gets the latest update status for an agent.
func (r *Repository) GetLatestUpdateStatus(ctx context.Context, agentID string) (*domain.UpdateStatus, error) {
	var status domain.UpdateStatus
	result := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		First(&status)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No update status found
		}
		return nil, result.Error
	}
	return &status, nil
}

// UpdateUpdateStatus updates an update status.
func (r *Repository) UpdateUpdateStatus(ctx context.Context, status *domain.UpdateStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}

// Service provides update business logic.
type Service struct {
	repo    *Repository
	storage *storage.Client
	mq      *mq.Client
	logger  *logger.Logger
}

// NewService creates a new update service.
func NewService(repo *Repository, storage *storage.Client) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
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

// CreateVersion creates a new version.
func (s *Service) CreateVersion(ctx context.Context, version, platform, fileURL string, fileSize int64, fileHash, signature, releaseNotes string) (*domain.Version, error) {
	// Check if version already exists
	if _, err := s.repo.GetVersionByNumber(ctx, version, platform); err == nil {
		return nil, ErrVersionAlreadyExists
	}

	v := &domain.Version{
		Version:      version,
		Platform:     platform,
		FileURL:      fileURL,
		FileSize:     fileSize,
		FileHash:     fileHash,
		Signature:    signature,
		ReleaseNotes: releaseNotes,
		IsActive:     true,
	}

	if err := s.repo.CreateVersion(ctx, v); err != nil {
		return nil, err
	}

	// Publish update available event
	if s.mq != nil {
		_ = s.mq.PublishEvent(ctx, mq.EventUpdateAvailable, map[string]interface{}{
			"version":       version,
			"platform":      platform,
			"release_notes": releaseNotes,
			"timestamp":     time.Now().Unix(),
		})
	}

	return v, nil
}

// GetVersion gets a version by ID.
func (s *Service) GetVersion(ctx context.Context, id string) (*domain.Version, error) {
	return s.repo.GetVersion(ctx, id)
}

// GetVersionByNumber gets a version by version number and platform.
func (s *Service) GetVersionByNumber(ctx context.Context, version, platform string) (*domain.Version, error) {
	return s.repo.GetVersionByNumber(ctx, version, platform)
}

// ListVersions lists all versions.
func (s *Service) ListVersions(ctx context.Context, platform string, limit int) ([]domain.Version, error) {
	return s.repo.ListVersions(ctx, platform, limit)
}

// GetLatestVersion gets the latest version.
func (s *Service) GetLatestVersion(ctx context.Context, platform string) (*domain.Version, error) {
	return s.repo.GetLatestVersion(ctx, platform)
}

// DeleteVersion deletes a version.
func (s *Service) DeleteVersion(ctx context.Context, id string) error {
	// Get version to delete file from storage
	version, err := s.repo.GetVersion(ctx, id)
	if err != nil {
		return err
	}

	// Delete from storage if there's a file URL
	if s.storage != nil && version.FileURL != "" {
		if err := s.storage.Delete(ctx, version.FileURL); err != nil && s.logger != nil {
			s.logger.Warnw("Failed to delete version file from storage", "error", err, "path", version.FileURL)
		}
	}

	return s.repo.DeleteVersion(ctx, id)
}

// GetDownloadURL generates a signed URL for downloading a version.
func (s *Service) GetDownloadURL(ctx context.Context, versionID string, expiry time.Duration) (string, error) {
	version, err := s.repo.GetVersion(ctx, versionID)
	if err != nil {
		return "", err
	}

	// If there's already a URL, return it
	if version.FileURL != "" && len(version.FileURL) > 10 {
		// Check if it's a storage path or a full URL
		if version.FileURL[:4] == "http" {
			return version.FileURL, nil
		}
	}

	if s.storage == nil {
		return "", errors.New("storage not configured")
	}

	return s.storage.PresignedURL(ctx, version.FileURL, expiry)
}

// TriggerUpdate triggers an update for an agent.
func (s *Service) TriggerUpdate(ctx context.Context, agentID, versionID string) (*domain.UpdateStatus, error) {
	// Check if version exists
	version, err := s.repo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}

	// Check if there's already an update in progress
	latestStatus, err := s.repo.GetLatestUpdateStatus(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if latestStatus != nil && (latestStatus.Status == domain.UpdateStatusPending ||
		latestStatus.Status == domain.UpdateStatusDownloading ||
		latestStatus.Status == domain.UpdateStatusInstalling) {
		return nil, ErrUpdateInProgress
	}

	// Create update status
	status := &domain.UpdateStatus{
		AgentID:   agentID,
		VersionID: versionID,
		Status:    domain.UpdateStatusPending,
		StartedAt: time.Now(),
	}

	if err := s.repo.CreateUpdateStatus(ctx, status); err != nil {
		return nil, err
	}

	// Generate download URL
	downloadURL, err := s.GetDownloadURL(ctx, versionID, 1*time.Hour)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("Failed to generate download URL", "error", err)
		}
	}

	// Publish update command to agent
	if s.mq != nil {
		_ = s.mq.PublishEvent(ctx, "agent.update", map[string]interface{}{
			"agent_id":     agentID,
			"version_id":   versionID,
			"version":      version.Version,
			"download_url": downloadURL,
			"file_hash":    version.FileHash,
			"signature":    version.Signature,
			"timestamp":    time.Now().Unix(),
		})
	}

	if s.logger != nil {
		s.logger.Infow("Update triggered",
			"agent_id", agentID,
			"version_id", versionID,
			"version", version.Version,
		)
	}

	return status, nil
}

// UpdateStatus updates the status of an agent update.
func (s *Service) UpdateStatus(ctx context.Context, agentID, versionID, status, message string) error {
	// Get latest update status
	updateStatus, err := s.repo.GetLatestUpdateStatus(ctx, agentID)
	if err != nil {
		return err
	}
	if updateStatus == nil {
		return errors.New("no update status found")
	}

	// Update status
	updateStatus.Status = status
	updateStatus.Message = message

	if status == domain.UpdateStatusSuccess || status == domain.UpdateStatusFailed {
		now := time.Now()
		updateStatus.FinishedAt = &now
	}

	if err := s.repo.UpdateUpdateStatus(ctx, updateStatus); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.Infow("Update status updated",
			"agent_id", agentID,
			"version_id", versionID,
			"status", status,
			"message", message,
		)
	}

	return nil
}

// GetUpdateStatus gets the update status for an agent.
func (s *Service) GetUpdateStatus(ctx context.Context, agentID string) (*domain.UpdateStatus, error) {
	return s.repo.GetLatestUpdateStatus(ctx, agentID)
}

// CheckForUpdate checks if there's a newer version available for an agent.
func (s *Service) CheckForUpdate(ctx context.Context, currentVersion, platform string) (*domain.Version, error) {
	latest, err := s.repo.GetLatestVersion(ctx, platform)
	if err != nil {
		return nil, err
	}

	// Simple version comparison (in production, use proper semver comparison)
	if latest.Version != currentVersion {
		return latest, nil
	}

	return nil, nil // No update available
}
