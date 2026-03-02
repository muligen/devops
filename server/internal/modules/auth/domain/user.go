// Package auth provides authentication and authorization functionality.
package domain

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID           string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string     `json:"username" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	Email        string     `json:"email" gorm:"uniqueIndex;size:255"`
	Role         string     `json:"role" gorm:"size:20;not null;default:'viewer'"`
	Status       string     `json:"status" gorm:"size:20;not null;default:'active'"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`
}

// User role constants
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// User status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusLocked   = "locked"
)

// TableName returns the table name for User model.
func (User) TableName() string {
	return "users"
}

// IsAdmin checks if user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManageUsers checks if user can manage other users.
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

// CanCreateTasks checks if user can create tasks.
func (u *User) CanCreateTasks() bool {
	return u.Role == RoleAdmin || u.Role == RoleOperator
}

// CanViewAgents checks if user can view agents.
func (u *User) CanViewAgents() bool {
	return true // All authenticated users can view agents
}
