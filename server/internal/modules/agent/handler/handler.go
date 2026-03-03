// Package handler provides HTTP handlers for agent management.
package handler

import (
	"context"
	"strconv"

	"github.com/agentteams/server/internal/modules/agent/domain"
	"github.com/agentteams/server/internal/modules/agent/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// LatestMetricsService interface for fetching agent metrics
type LatestMetricsService interface {
	GetLatestMetricValues(ctx context.Context, agentID string) (cpuUsage, memoryPercent, diskPercent float64, err error)
}

// Handler handles agent HTTP requests.
type Handler struct {
	service              *service.Service
	latestMetricsService LatestMetricsService
}

// NewHandler creates a new agent handler.
func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// SetLatestMetricsService sets the metrics service for fetching latest metrics.
func (h *Handler) SetLatestMetricsService(svc LatestMetricsService) {
	h.latestMetricsService = svc
}

// CreateAgentRequest represents create agent request.
type CreateAgentRequest struct {
	Name     string          `json:"name" binding:"required,min=1,max=100"`
	Metadata domain.JSONB   `json:"metadata"`
}

// AgentResponse represents agent response.
type AgentResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	Version      string          `json:"version"`
	Hostname     string          `json:"hostname"`
	IPAddress    string          `json:"ip_address"`
	OSInfo       string          `json:"os_info"`
	Metadata     domain.JSONB    `json:"metadata"`
	LastSeenAt   *string         `json:"last_seen_at"`
	CreatedAt    string          `json:"created_at"`
	CPUUsage     *float64        `json:"cpu_usage,omitempty"`
	MemoryUsage  *float64        `json:"memory_usage,omitempty"`
	DiskUsage    *float64        `json:"disk_usage,omitempty"`
}

// CreateAgent handles agent creation.
func (h *Handler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	agent, token, err := h.service.CreateAgent(c.Request.Context(), req.Name, req.Metadata)
	if err != nil {
		if err == service.ErrAgentExists {
			response.Conflict(c, "agent already exists")
			return
		}
		response.InternalError(c, "failed to create agent")
		return
	}

	response.Created(c, gin.H{
		"id":         agent.ID,
		"name":       agent.Name,
		"token":      token, // Only returned on creation
		"status":     agent.Status,
		"created_at": agent.CreatedAt,
	})
}

// ListAgents handles listing agents.
func (h *Handler) ListAgents(c *gin.Context) {
	page := 1
	pageSize := 20
	status := c.Query("status")
	sort := c.Query("sort")
	order := c.Query("order")

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Validate sort field
	validSortFields := map[string]bool{
		"cpu_usage":      true,
		"memory_percent": true,
		"disk_percent":   true,
		"created_at":     true,
		"":               true,
	}
	if !validSortFields[sort] {
		sort = "" // Default to created_at
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var agents []domain.Agent
	var total int64
	var err error

	// Use sorting method if sort is specified
	if sort != "" {
		agents, total, err = h.service.ListAgentsWithSort(c.Request.Context(), service.ListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Sort:     sort,
			Order:    order,
		})
	} else {
		agents, total, err = h.service.ListAgents(c.Request.Context(), page, pageSize, status)
	}

	if err != nil {
		response.InternalError(c, "failed to list agents")
		return
	}

	// Map to response with metrics
	items := make([]AgentResponse, len(agents))
	for i, agent := range agents {
		items[i] = toAgentResponse(&agent)
		// Fetch latest metrics if metrics service is available
		if h.latestMetricsService != nil && agent.Status == domain.StatusOnline {
			cpu, mem, disk, _ := h.latestMetricsService.GetLatestMetricValues(c.Request.Context(), agent.ID)
			if cpu > 0 || mem > 0 || disk > 0 {
				items[i].CPUUsage = &cpu
				items[i].MemoryUsage = &mem
				items[i].DiskUsage = &disk
			}
		}
	}

	response.Paged(c, items, page, pageSize, total)
}

// GetAgent handles getting agent by ID.
func (h *Handler) GetAgent(c *gin.Context) {
	id := c.Param("id")

	agent, err := h.service.GetAgent(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAgentNotFound {
			response.NotFound(c, "agent not found")
			return
		}
		response.InternalError(c, "failed to get agent")
		return
	}

	resp := toAgentResponse(agent)
	// Fetch latest metrics if metrics service is available
	if h.latestMetricsService != nil && agent.Status == domain.StatusOnline {
		cpu, mem, disk, _ := h.latestMetricsService.GetLatestMetricValues(c.Request.Context(), agent.ID)
		if cpu > 0 || mem > 0 || disk > 0 {
			resp.CPUUsage = &cpu
			resp.MemoryUsage = &mem
			resp.DiskUsage = &disk
		}
	}

	response.Success(c, resp)
}

// DeleteAgent handles agent deletion.
func (h *Handler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAgent(c.Request.Context(), id); err != nil {
		if err == service.ErrAgentNotFound {
			response.NotFound(c, "agent not found")
			return
		}
		response.InternalError(c, "failed to delete agent")
		return
	}

	response.Success(c, gin.H{
		"message": "agent deleted",
	})
}

// UpdateStatusRequest represents update status request.
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=online offline maintenance"`
}

// UpdateStatus handles agent status update.
func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	if err := h.service.UpdateAgentStatus(c.Request.Context(), id, req.Status); err != nil {
		if err == service.ErrAgentNotFound {
			response.NotFound(c, "agent not found")
			return
		}
		response.InternalError(c, "failed to update status")
		return
	}

	response.Success(c, gin.H{
		"message": "status updated",
	})
}

// UpdateMetadataRequest represents update metadata request.
type UpdateMetadataRequest struct {
	Metadata domain.JSONB `json:"metadata" binding:"required"`
}

// UpdateMetadata handles agent metadata update.
func (h *Handler) UpdateMetadata(c *gin.Context) {
	id := c.Param("id")

	var req UpdateMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	if err := h.service.UpdateAgentMetadata(c.Request.Context(), id, req.Metadata); err != nil {
		if err == service.ErrAgentNotFound {
			response.NotFound(c, "agent not found")
			return
		}
		response.InternalError(c, "failed to update metadata")
		return
	}

	response.Success(c, gin.H{
		"message": "metadata updated",
	})
}

// RegisterRoutes registers agent routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	agents := r.Group("/agents")
	{
		agents.POST("", h.CreateAgent)
		agents.GET("", h.ListAgents)
		agents.GET("/:id", h.GetAgent)
		agents.DELETE("/:id", h.DeleteAgent)
		agents.PUT("/:id/status", h.UpdateStatus)
		agents.PUT("/:id/metadata", h.UpdateMetadata)
	}
}

// toAgentResponse converts agent to response.
func toAgentResponse(agent *domain.Agent) AgentResponse {
	var lastSeen *string
	if agent.LastSeenAt != nil {
		formatted := agent.LastSeenAt.Format("2006-01-02T15:04:05Z")
		lastSeen = &formatted
	}

	return AgentResponse{
		ID:         agent.ID,
		Name:       agent.Name,
		Status:     agent.Status,
		Version:    agent.Version,
		Hostname:   agent.Hostname,
		IPAddress:  agent.IPAddress,
		OSInfo:     agent.OSInfo,
		Metadata:   agent.Metadata,
		LastSeenAt: lastSeen,
		CreatedAt:  agent.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
