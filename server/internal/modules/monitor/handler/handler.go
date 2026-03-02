// Package handler provides HTTP handlers for monitoring.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/agentteams/server/internal/modules/monitor/domain"
	"github.com/agentteams/server/internal/modules/monitor/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles monitor HTTP requests.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new monitor handler.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// MetricResponse represents a metric response.
type MetricResponse struct {
	ID            string `json:"id"`
	AgentID       string  `json:"agent_id"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryTotal   int64   `json:"memory_total"`
	MemoryUsed    int64   `json:"memory_used"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskTotal     int64   `json:"disk_total"`
	DiskUsed      int64   `json:"disk_used"`
	DiskPercent   float64 `json:"disk_percent"`
	Uptime        int64   `json:"uptime"`
	CollectedAt   string  `json:"collected_at"`
}

// GetAgentMetrics handles GET /api/v1/agents/:id/metrics
func (h *Handler) GetAgentMetrics(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.ValidationErr(c, "agent_id is required")
		return
	}

	// Parse time range
	start, end := parseTimeRange(c)

	// Parse limit
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	metrics, err := h.service.GetAgentMetrics(c.Request.Context(), agentID, start, end, limit)
	if err != nil {
		response.InternalError(c, "failed to get metrics")
		return
	}

	// Map to response
	items := make([]MetricResponse, len(metrics))
	for i, m := range metrics {
		items[i] = MetricResponse{
			ID:            m.ID,
			AgentID:       m.AgentID,
			CPUUsage:      m.CPUUsage,
			MemoryTotal:   m.MemoryTotal,
			MemoryUsed:    m.MemoryUsed,
			MemoryPercent: m.MemoryPercent,
			DiskTotal:     m.DiskTotal,
			DiskUsed:      m.DiskUsed,
			DiskPercent:   m.DiskPercent,
			Uptime:        m.Uptime,
			CollectedAt:   m.CollectedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	response.Success(c, items)
}

// GetLatestMetric handles GET /api/v1/agents/:id/metrics/latest
func (h *Handler) GetLatestMetric(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.ValidationErr(c, "agent_id is required")
		return
	}

	metric, err := h.service.GetLatestMetric(c.Request.Context(), agentID)
	if err != nil {
		if err == service.ErrMetricNotFound {
			response.NotFound(c, "no metrics found for agent")
			return
		}
		response.InternalError(c, "failed to get metric")
		return
	}

	response.Success(c, MetricResponse{
		ID:            metric.ID,
		AgentID:       metric.AgentID,
		CPUUsage:      metric.CPUUsage,
		MemoryTotal:   metric.MemoryTotal,
		MemoryUsed:    metric.MemoryUsed,
		MemoryPercent: metric.MemoryPercent,
		DiskTotal:     metric.DiskTotal,
		DiskUsed:      metric.DiskUsed,
		DiskPercent:   metric.DiskPercent,
		Uptime:        metric.Uptime,
		CollectedAt:   metric.CollectedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GetDashboardStats handles GET /api/v1/dashboard/stats
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get dashboard stats")
		return
	}

	response.Success(c, stats)
}

// CreateAlertRuleRequest represents create alert rule request.
type CreateAlertRuleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description"`
	MetricType  string  `json:"metric_type" binding:"required,oneof=cpu_usage memory_percent disk_percent"`
	Condition   string  `json:"condition" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64 `json:"threshold" binding:"required"`
	Duration    int     `json:"duration" binding:"min=0"` // seconds
	Severity    string  `json:"severity" binding:"oneof=info warning critical"`
}

// CreateAlertRule handles POST /api/v1/alerts/rules
func (h *Handler) CreateAlertRule(c *gin.Context) {
	var req CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	// Set defaults
	if req.Duration == 0 {
		req.Duration = 60
	}
	if req.Severity == "" {
		req.Severity = domain.SeverityWarning
	}

	rule, err := h.service.CreateAlertRule(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.MetricType,
		req.Condition,
		req.Threshold,
		req.Duration,
		req.Severity,
	)
	if err != nil {
		response.InternalError(c, "failed to create alert rule")
		return
	}

	response.Created(c, toAlertRuleResponse(rule))
}

// AlertRuleResponse represents an alert rule response.
type AlertRuleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MetricType  string `json:"metric_type"`
	Condition   string `json:"condition"`
	Threshold   float64 `json:"threshold"`
	Duration    int    `json:"duration"`
	Severity    string `json:"severity"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListAlertRules handles GET /api/v1/alerts/rules
func (h *Handler) ListAlertRules(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	rules, err := h.service.ListAlertRules(c.Request.Context(), enabledOnly)
	if err != nil {
		response.InternalError(c, "failed to list alert rules")
		return
	}

	items := make([]AlertRuleResponse, len(rules))
	for i, r := range rules {
		items[i] = toAlertRuleResponse(&r)
	}

	response.Success(c, items)
}

// GetAlertRule handles GET /api/v1/alerts/rules/:id
func (h *Handler) GetAlertRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.service.GetAlertRule(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAlertRuleNotFound {
			response.NotFound(c, "alert rule not found")
			return
		}
		response.InternalError(c, "failed to get alert rule")
		return
	}

	response.Success(c, toAlertRuleResponse(rule))
}

// UpdateAlertRuleRequest represents update alert rule request.
type UpdateAlertRuleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description"`
	MetricType  string  `json:"metric_type" binding:"required,oneof=cpu_usage memory_percent disk_percent"`
	Condition   string  `json:"condition" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64 `json:"threshold" binding:"required"`
	Duration    int     `json:"duration" binding:"min=0"`
	Severity    string  `json:"severity" binding:"oneof=info warning critical"`
	Enabled     bool    `json:"enabled"`
}

// UpdateAlertRule handles PUT /api/v1/alerts/rules/:id
func (h *Handler) UpdateAlertRule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	rule, err := h.service.UpdateAlertRule(
		c.Request.Context(),
		id,
		req.Name,
		req.Description,
		req.MetricType,
		req.Condition,
		req.Threshold,
		req.Duration,
		req.Severity,
		req.Enabled,
	)
	if err != nil {
		if err == service.ErrAlertRuleNotFound {
			response.NotFound(c, "alert rule not found")
			return
		}
		response.InternalError(c, "failed to update alert rule")
		return
	}

	response.Success(c, toAlertRuleResponse(rule))
}

// DeleteAlertRule handles DELETE /api/v1/alerts/rules/:id
func (h *Handler) DeleteAlertRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAlertRule(c.Request.Context(), id); err != nil {
		if err == service.ErrAlertRuleNotFound {
			response.NotFound(c, "alert rule not found")
			return
		}
		response.InternalError(c, "failed to delete alert rule")
		return
	}

	response.Success(c, gin.H{"message": "alert rule deleted"})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// RegisterRoutes registers monitor routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Dashboard
	r.GET("/dashboard/stats", h.GetDashboardStats)

	// Agent metrics
	agents := r.Group("/agents")
	{
		agents.GET("/:id/metrics", h.GetAgentMetrics)
		agents.GET("/:id/metrics/latest", h.GetLatestMetric)
	}

	// Alert rules
	alerts := r.Group("/alerts")
	{
		alerts.GET("/rules", h.ListAlertRules)
		alerts.POST("/rules", h.CreateAlertRule)
		alerts.GET("/rules/:id", h.GetAlertRule)
		alerts.PUT("/rules/:id", h.UpdateAlertRule)
		alerts.DELETE("/rules/:id", h.DeleteAlertRule)
	}
}

// toAlertRuleResponse converts alert rule to response.
func toAlertRuleResponse(rule *domain.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		MetricType:  rule.MetricType,
		Condition:   rule.Condition,
		Threshold:   rule.Threshold,
		Duration:    rule.Duration,
		Severity:    rule.Severity,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   rule.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// parseTimeRange parses start and end time from query parameters.
func parseTimeRange(c *gin.Context) (start, end time.Time) {
	end = time.Now()
	start = end.Add(-1 * time.Hour) // Default: last hour

	if s := c.Query("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			end = t
		}
	}

	return start, end
}
