// Package monitor provides monitoring functionality.
package domain

import (
	"time"
)

// AgentMetric represents system metrics collected from an agent.
type AgentMetric struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AgentID       string    `json:"agent_id" gorm:"type:uuid;not null;index"`
	CPUUsage      float64   `json:"cpu_usage" gorm:"not null;default:0"`
	MemoryTotal   int64     `json:"memory_total" gorm:"not null;default:0"`
	MemoryUsed    int64     `json:"memory_used" gorm:"not null;default:0"`
	MemoryPercent float64   `json:"memory_percent" gorm:"not null;default:0"`
	DiskTotal     int64     `json:"disk_total" gorm:"not null;default:0"`
	DiskUsed      int64     `json:"disk_used" gorm:"not null;default:0"`
	DiskPercent   float64   `json:"disk_percent" gorm:"not null;default:0"`
	Uptime        int64     `json:"uptime" gorm:"not null;default:0"`
	CollectedAt   time.Time `json:"collected_at" gorm:"not null;default:NOW();index"`
}

// TableName returns the table name for AgentMetric model.
func (AgentMetric) TableName() string {
	return "agent_metrics"
}

// Metric type constants
const (
	MetricCPU    = "cpu_usage"
	MetricMemory = "memory_percent"
	MetricDisk   = "disk_percent"
)

// AlertRule represents a monitoring alert rule.
type AlertRule struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	MetricType  string    `json:"metric_type" gorm:"size:50;not null"`
	Condition   string    `json:"condition" gorm:"size:10;not null"`
	Threshold   float64   `json:"threshold" gorm:"not null"`
	Duration    int       `json:"duration" gorm:"not null;default:60"` // seconds
	Severity    string    `json:"severity" gorm:"size:20;not null;default:'warning'"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for AlertRule model.
func (AlertRule) TableName() string {
	return "alert_rules"
}

// Alert condition constants
const (
	ConditionGreater      = ">"
	ConditionGreaterEqual = ">="
	ConditionLess         = "<"
	ConditionLessEqual    = "<="
	ConditionEqual        = "=="
	ConditionNotEqual     = "!="
)

// Alert severity constants
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)
