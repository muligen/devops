// Package domain provides update domain models.
package domain

import (
	"time"
)

// Version represents an agent version.
type Version struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Version      string    `json:"version" gorm:"size:20;not null;uniqueIndex:uq_versions_version_platform"`
	Platform     string    `json:"platform" gorm:"size:50;not null;default:'windows';uniqueIndex:uq_versions_version_platform"`
	FileURL      string    `json:"file_url" gorm:"size:500"`
	FileHash     string    `json:"file_hash" gorm:"size:64"`
	FileSize     int64     `json:"file_size"`
	Signature    string    `json:"signature" gorm:"type:text"`
	ReleaseNotes string    `json:"release_notes" gorm:"type:text"`
	IsActive     bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	CreatedBy    string    `json:"created_by" gorm:"type:uuid"`
}

// TableName returns the table name for Version model.
func (Version) TableName() string {
	return "versions"
}

// UpdateStatus represents the status of an agent update.
type UpdateStatus struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID    string    `json:"agent_id" gorm:"type:uuid;not null;index"`
	VersionID  string    `json:"version_id" gorm:"type:uuid;not null;index"`
	Status     string    `json:"status" gorm:"size:20;not null"` // pending, downloading, installing, success, failed
	Message    string    `json:"message" gorm:"type:text"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for UpdateStatus model.
func (UpdateStatus) TableName() string {
	return "agent_update_status"
}

// Update status constants
const (
	UpdateStatusPending     = "pending"
	UpdateStatusDownloading = "downloading"
	UpdateStatusInstalling  = "installing"
	UpdateStatusSuccess     = "success"
	UpdateStatusFailed      = "failed"
)

// Update channel constants
const (
	ChannelStable = "stable"
	ChannelBeta   = "beta"
	ChannelAlpha  = "alpha"
)
