// Package domain provides alert event domain models.
package domain

import (
	"time"

	"github.com/lib/pq"
)

// AlertEvent status constants
const (
	AlertStatusPending      = "pending"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"
)

// AlertEvent represents an alert trigger event.
type AlertEvent struct {
	ID             string      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RuleID         *string     `json:"rule_id" gorm:"type:uuid"`
	AgentID        string      `json:"agent_id" gorm:"type:uuid;not null"`
	RuleName       string      `json:"rule_name" gorm:"-"`
	AgentName      string      `json:"agent_name" gorm:"-"`
	MetricValue    float64     `json:"metric_value" gorm:"not null"`
	Threshold      float64     `json:"threshold" gorm:"not null"`
	Status         string      `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	Message        string      `json:"message" gorm:"type:text"`
	TriggeredAt    time.Time   `json:"triggered_at" gorm:"not null;default:NOW()"`
	ResolvedAt     pq.NullTime `json:"resolved_at" gorm:"type:timestamp with time zone"`
	AcknowledgedBy *string     `json:"acknowledged_by" gorm:"type:uuid"`
	AcknowledgedAt pq.NullTime `json:"acknowledged_at" gorm:"type:timestamp with time zone"`
	CreatedAt      time.Time   `json:"created_at" gorm:"not null;default:NOW()"`

	// Relations
	Rule         *AlertRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
	Acknowledger *User      `json:"acknowledger,omitempty" gorm:"foreignKey:AcknowledgedBy"`
}

// TableName returns the table name for AlertEvent.
func (AlertEvent) TableName() string {
	return "alert_events"
}

// IsPending returns true if the alert is pending.
func (e *AlertEvent) IsPending() bool {
	return e.Status == AlertStatusPending
}

// IsAcknowledged returns true if the alert is acknowledged.
func (e *AlertEvent) IsAcknowledged() bool {
	return e.Status == AlertStatusAcknowledged
}

// IsResolved returns true if the alert is resolved.
func (e *AlertEvent) IsResolved() bool {
	return e.Status == AlertStatusResolved
}

// Acknowledge marks the alert as acknowledged.
func (e *AlertEvent) Acknowledge(userID string) {
	e.Status = AlertStatusAcknowledged
	e.AcknowledgedBy = &userID
	now := time.Now()
	e.AcknowledgedAt = pq.NullTime{Time: now, Valid: true}
}

// Resolve marks the alert as resolved.
func (e *AlertEvent) Resolve() {
	e.Status = AlertStatusResolved
	now := time.Now()
	e.ResolvedAt = pq.NullTime{Time: now, Valid: true}
}

// User represents a minimal user for the alert event relation.
type User struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Username string `json:"username" gorm:"type:varchar(100)"`
}
