// Package task provides task management functionality.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Task represents a command execution task.
type Task struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID     string     `json:"agent_id" gorm:"type:uuid;not null;index"`
	Type        string     `json:"type" gorm:"size:50;not null"`
	Params      JSONB      `json:"params" gorm:"type:jsonb;default:'{}'"`
	Status      string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	Priority    int        `json:"priority" gorm:"not null;default:0"`
	Timeout     int        `json:"timeout" gorm:"not null;default:300"`
	Result      JSONB      `json:"result" gorm:"type:jsonb"`
	Output      string     `json:"output" gorm:"type:text"`
	ExitCode    *int       `json:"exit_code"`
	Duration    *float64   `json:"duration"`
	CreatedBy   string     `json:"created_by" gorm:"type:uuid"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Task type constants
const (
	TypeExecShell   = "exec_shell"
	TypeInitMachine = "init_machine"
	TypeCleanDisk   = "clean_disk"
)

// Task status constants
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusTimeout   = "timeout"
	StatusCancelled = "canceled"
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

// TableName returns the table name for Task model.
func (Task) TableName() string {
	return "tasks"
}

// IsPending checks if task is pending.
func (t *Task) IsPending() bool {
	return t.Status == StatusPending
}

// IsRunning checks if task is running.
func (t *Task) IsRunning() bool {
	return t.Status == StatusRunning
}

// IsCompleted checks if task is completed (success or failed).
func (t *Task) IsCompleted() bool {
	return t.Status == StatusSuccess || t.Status == StatusFailed ||
		t.Status == StatusTimeout || t.Status == StatusCancelled
}
