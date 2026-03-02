// Package agent provides agent management functionality.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Agent represents a registered agent machine.
type Agent struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name       string     `json:"name" gorm:"uniqueIndex;size:100;not null"`
	TokenHash  string     `json:"-" gorm:"size:255;not null"`
	Status     string     `json:"status" gorm:"size:20;not null;default:'offline'"`
	Version    string     `json:"version" gorm:"size:20;not null;default:'0.0.0'"`
	Hostname   string     `json:"hostname" gorm:"size:255"`
	IPAddress  string     `json:"ip_address" gorm:"size:45"`
	OSInfo     string     `json:"os_info" gorm:"size:255"`
	Metadata   JSONB      `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"index"`
}

// Agent status constants
const (
	StatusOnline      = "online"
	StatusOffline     = "offline"
	StatusMaintenance = "maintenance"
)

// JSONB is a type for JSONB fields in PostgreSQL.
type JSONB map[string]interface{}

// Value implements driver.Valuer interface for JSONB.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for JSONB.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// TableName returns the table name for Agent model.
func (Agent) TableName() string {
	return "agents"
}

// IsOnline checks if agent is online.
func (a *Agent) IsOnline() bool {
	return a.Status == StatusOnline
}
