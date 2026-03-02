// Package domain provides user domain models.
package domain

import (
	"time"
)

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID           string                 `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string                `json:"user_id" gorm:"type:uuid;index"`
	Action       string                `json:"action" gorm:"size:50;not null;index"`
	ResourceType string                `json:"resource_type" gorm:"size:50;not null;index"`
	ResourceID   string                `json:"resource_id" gorm:"type:uuid;index"`
	Details      map[string]interface{} `json:"details" gorm:"type:jsonb;default:'{}'"`
	IPAddress    string                `json:"ip_address" gorm:"size:45"`
	UserAgent    string                `json:"user_agent" gorm:"type:text"`
	CreatedAt    time.Time             `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName returns the table name for AuditLog model.
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Audit action constants
const (
	ActionLogin       = "login"
	ActionLogout      = "logout"
	ActionCreate      = "create"
	ActionUpdate      = "update"
	ActionDelete      = "delete"
	ActionView        = "view"
	ActionExport      = "export"
	ActionExecute     = "execute"
)

// Resource type constants
const (
	ResourceUser   = "user"
	ResourceAgent  = "agent"
	ResourceTask   = "task"
	ResourceVersion = "version"
	ResourceAlert  = "alert"
	ResourceSystem = "system"
)
