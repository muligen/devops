// Package handler provides HTTP handlers for task management.
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/agentteams/server/internal/modules/task/domain"
	"github.com/agentteams/server/internal/modules/task/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles task HTTP requests.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new task handler.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// CreateTaskRequest represents create task request.
type CreateTaskRequest struct {
	AgentID  string       `json:"agent_id" binding:"required"`
	Type     string       `json:"type" binding:"required,oneof=exec_shell init_machine clean_disk"`
	Params   domain.JSONB `json:"params"`
	Priority int          `json:"priority"`
	Timeout  int          `json:"timeout"`
}

// BatchCreateTaskRequest represents batch create tasks request.
type BatchCreateTaskRequest struct {
	Tasks []CreateTaskRequest `json:"tasks" binding:"required,min=1,max=100"`
}

// TaskResponse represents task response.
type TaskResponse struct {
	ID          string       `json:"id"`
	AgentID     string       `json:"agent_id"`
	Type        string       `json:"type"`
	Params      domain.JSONB `json:"params"`
	Status      string       `json:"status"`
	Priority    int          `json:"priority"`
	Timeout     int          `json:"timeout"`
	Result      domain.JSONB `json:"result,omitempty"`
	Output      string       `json:"output,omitempty"`
	ExitCode    *int         `json:"exit_code,omitempty"`
	Duration    *float64     `json:"duration,omitempty"`
	CreatedBy   string       `json:"created_by"`
	CreatedAt   string       `json:"created_at"`
	StartedAt   *string      `json:"started_at,omitempty"`
	CompletedAt *string      `json:"completed_at,omitempty"`
}

// CreateTask handles task creation.
// POST /api/v1/tasks
func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	// Set defaults
	if req.Params == nil {
		req.Params = domain.JSONB{}
	}
	if req.Timeout == 0 {
		req.Timeout = 300 // Default 5 minutes
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	task, err := h.service.CreateTask(
		c.Request.Context(),
		req.AgentID,
		req.Type,
		req.Params,
		req.Priority,
		req.Timeout,
		userIDStr,
	)
	if err != nil {
		response.InternalError(c, "failed to create task")
		return
	}

	response.Created(c, toTaskResponse(task))
}

// BatchCreateTasks handles batch task creation.
// POST /api/v1/tasks/batch
func (h *Handler) BatchCreateTasks(c *gin.Context) {
	var req BatchCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Convert requests
	requests := make([]service.CreateTaskRequest, len(req.Tasks))
	for i, t := range req.Tasks {
		params := t.Params
		if params == nil {
			params = domain.JSONB{}
		}
		timeout := t.Timeout
		if timeout == 0 {
			timeout = 300
		}
		requests[i] = service.CreateTaskRequest{
			AgentID:  t.AgentID,
			Type:     t.Type,
			Params:   params,
			Priority: t.Priority,
			Timeout:  timeout,
		}
	}

	tasks, err := h.service.CreateBatchTasks(c.Request.Context(), requests, userIDStr)
	if err != nil {
		response.InternalError(c, "failed to create tasks")
		return
	}

	// Map to response
	items := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = toTaskResponse(task)
	}

	response.Created(c, gin.H{
		"tasks": items,
		"count": len(items),
	})
}

// ListTasks handles listing tasks.
// GET /api/v1/tasks
func (h *Handler) ListTasks(c *gin.Context) {
	page := 1
	pageSize := 20
	agentID := c.Query("agent_id")
	status := c.Query("status")
	taskType := c.Query("type")

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

	// Build filter
	tasks, total, err := h.service.ListTasks(c.Request.Context(), page, pageSize, agentID, status)
	if err != nil {
		response.InternalError(c, "failed to list tasks")
		return
	}

	// Filter by type if specified (done in memory since it's not indexed)
	if taskType != "" {
		var filtered []domain.Task
		for _, t := range tasks {
			if t.Type == taskType {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
		total = int64(len(filtered))
	}

	// Map to response
	items := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = toTaskResponse(&task)
	}

	response.Paged(c, items, page, pageSize, total)
}

// GetTask handles getting task by ID.
// GET /api/v1/tasks/:id
func (h *Handler) GetTask(c *gin.Context) {
	id := c.Param("id")

	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.InternalError(c, "failed to get task")
		return
	}

	response.Success(c, toTaskResponse(task))
}

// CancelTask handles task cancellation.
// DELETE /api/v1/tasks/:id
func (h *Handler) CancelTask(c *gin.Context) {
	id := c.Param("id")

	err := h.service.CancelTask(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrTaskNotFound:
			response.NotFound(c, "task not found")
		case service.ErrTaskAlreadyRunning:
			response.Error(c, http.StatusBadRequest, response.CodeTaskFailed, "task already running, cannot cancel")
		default:
			response.InternalError(c, "failed to cancel task")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "task cancelled",
		"id":      id,
	})
}

// StreamTaskOutput handles task output streaming via SSE.
// GET /api/v1/tasks/:id/output
func (h *Handler) StreamTaskOutput(c *gin.Context) {
	id := c.Param("id")

	// Verify task exists
	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrTaskNotFound {
			response.NotFound(c, "task not found")
			return
		}
		response.InternalError(c, "failed to get task")
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Create flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.InternalError(c, "streaming not supported")
		return
	}

	// Send initial status
	sendSSE(c, flusher, "status", gin.H{
		"task_id": id,
		"status":  task.Status,
	})

	// If task is already completed, send final output and close
	if task.IsCompleted() {
		sendSSE(c, flusher, "output", gin.H{
			"task_id": id,
			"output":  task.Output,
			"final":   true,
		})
		sendSSE(c, flusher, "complete", gin.H{
			"task_id":    id,
			"status":     task.Status,
			"exit_code":  task.ExitCode,
			"duration":   task.Duration,
			"completed":  true,
		})
		return
	}

	// Poll for updates (in a real implementation, this would use pub/sub)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastOutputLen := 0
	timeout := time.After(30 * time.Minute) // Max streaming time

	for {
		select {
		case <-c.Request.Context().Done():
			// Client disconnected
			return
		case <-timeout:
			// Timeout reached
			sendSSE(c, flusher, "error", gin.H{
				"task_id": id,
				"error":   "stream timeout",
			})
			return
		case <-ticker.C:
			// Get updated task
			currentTask, err := h.service.GetTask(c.Request.Context(), id)
			if err != nil {
				sendSSE(c, flusher, "error", gin.H{
					"task_id": id,
					"error":   "failed to get task",
				})
				return
			}

			// Send new output if any
			if len(currentTask.Output) > lastOutputLen {
				newOutput := currentTask.Output[lastOutputLen:]
				sendSSE(c, flusher, "output", gin.H{
					"task_id": id,
					"output":  newOutput,
					"final":   false,
				})
				lastOutputLen = len(currentTask.Output)
			}

			// Check if task completed
			if currentTask.IsCompleted() {
				// Send any remaining output
				if len(currentTask.Output) > lastOutputLen {
					sendSSE(c, flusher, "output", gin.H{
						"task_id": id,
						"output":  currentTask.Output[lastOutputLen:],
						"final":   true,
					})
				}
				sendSSE(c, flusher, "complete", gin.H{
					"task_id":    id,
					"status":     currentTask.Status,
					"exit_code":  currentTask.ExitCode,
					"duration":   currentTask.Duration,
					"completed":  true,
				})
				return
			}
		}
	}
}

// sendSSE sends a Server-Sent Event.
func sendSSE(c *gin.Context, flusher http.Flusher, event string, data interface{}) {
	_, _ = c.Writer.WriteString(fmt.Sprintf("event: %s\n", event))
	_, _ = c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", mustMarshal(data)))
	flusher.Flush()
}

// mustMarshal converts data to JSON string.
func mustMarshal(data interface{}) string {
	// Simple JSON marshal - in production use proper error handling
	// Using fmt.Sprintf with %v as fallback
	b, err := jsonMarshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// jsonMarshal is a wrapper for encoding/json.Marshal.
func jsonMarshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// RegisterRoutes registers task routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.POST("/batch", h.BatchCreateTasks)
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.GET("/:id/output", h.StreamTaskOutput)
		tasks.DELETE("/:id", h.CancelTask)
	}
}

// toTaskResponse converts task to response.
func toTaskResponse(task *domain.Task) TaskResponse {
	var startedAt, completedAt *string
	if task.StartedAt != nil {
		s := task.StartedAt.Format("2006-01-02T15:04:05Z")
		startedAt = &s
	}
	if task.CompletedAt != nil {
		s := task.CompletedAt.Format("2006-01-02T15:04:05Z")
		completedAt = &s
	}

	return TaskResponse{
		ID:          task.ID,
		AgentID:     task.AgentID,
		Type:        task.Type,
		Params:      task.Params,
		Status:      task.Status,
		Priority:    task.Priority,
		Timeout:     task.Timeout,
		Result:      task.Result,
		Output:      task.Output,
		ExitCode:    task.ExitCode,
		Duration:    task.Duration,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z"),
		StartedAt:   startedAt,
		CompletedAt: completedAt,
	}
}
