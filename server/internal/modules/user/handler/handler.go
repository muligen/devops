// Package handler provides HTTP handlers for user management.
package handler

import (
	"strconv"
	"time"

	"github.com/agentteams/server/internal/modules/auth/domain"
	authService "github.com/agentteams/server/internal/modules/auth/service"
	userDomain "github.com/agentteams/server/internal/modules/user/domain"
	"github.com/agentteams/server/internal/modules/user/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles user HTTP requests.
type Handler struct {
	authService *authService.Service
	userService *service.Service
}

// NewHandler creates a new user handler.
func NewHandler(authService *authService.Service, userService *service.Service) *Handler {
	return &Handler{
		authService: authService,
		userService: userService,
	}
}

// UserResponse represents a user response.
type UserResponse struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	LastLoginAt *string `json:"last_login_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateUserRequest represents create user request.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// CreateUser handles POST /api/v1/users
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	user, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email, req.Role)
	if err != nil {
		if err == authService.ErrUserExists {
			response.Conflict(c, "user already exists")
			return
		}
		response.InternalError(c, "failed to create user")
		return
	}

	// Log audit
	_ = h.userService.LogAudit(
		c.Request.Context(),
		getUserID(c),
		userDomain.ActionCreate,
		userDomain.ResourceUser,
		user.ID,
		c.ClientIP(),
		c.Request.UserAgent(),
		map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	)

	response.Created(c, toUserResponse(user))
}

// ListUsers handles GET /api/v1/users
func (h *Handler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	users, total, err := h.authService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}

	items := make([]UserResponse, len(users))
	for i, user := range users {
		items[i] = toUserResponse(&user)
	}

	response.Paged(c, items, page, pageSize, total)
}

// UpdateUserRequest represents update user request.
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Role   string `json:"role" binding:"omitempty,oneof=admin operator viewer"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive locked"`
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	user, err := h.authService.UpdateUser(c.Request.Context(), id, updates)
	if err != nil {
		if err == authService.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to update user")
		return
	}

	// Log audit
	_ = h.userService.LogAudit(
		c.Request.Context(),
		getUserID(c),
		userDomain.ActionUpdate,
		userDomain.ResourceUser,
		id,
		c.ClientIP(),
		c.Request.UserAgent(),
		updates,
	)

	response.Success(c, toUserResponse(user))
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.authService.DeleteUser(c.Request.Context(), id); err != nil {
		if err == authService.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to delete user")
		return
	}

	// Log audit
	_ = h.userService.LogAudit(
		c.Request.Context(),
		getUserID(c),
		userDomain.ActionDelete,
		userDomain.ResourceUser,
		id,
		c.ClientIP(),
		c.Request.UserAgent(),
		nil,
	)

	response.Success(c, gin.H{"message": "user deleted"})
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		response.Unauthorized(c, "not authenticated")
		return
	}

	user, err := h.authService.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to get user")
		return
	}

	response.Success(c, toUserResponse(user))
}

// ChangePasswordRequest represents change password request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword handles PUT /api/v1/users/me/password
func (h *Handler) ChangePassword(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		response.Unauthorized(c, "not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if err == authService.ErrInvalidCredentials {
			response.Unauthorized(c, "invalid old password")
			return
		}
		response.InternalError(c, "failed to change password")
		return
	}

	// Log audit
	_ = h.userService.LogAudit(
		c.Request.Context(),
		userID,
		userDomain.ActionUpdate,
		userDomain.ResourceUser,
		userID,
		c.ClientIP(),
		c.Request.UserAgent(),
		map[string]interface{}{"action": "password_change"},
	)

	response.Success(c, gin.H{"message": "password changed"})
}

// AuditLogResponse represents an audit log response.
type AuditLogResponse struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    string                 `json:"created_at"`
}

// GetAuditLogs handles GET /api/v1/audit-logs
func (h *Handler) GetAuditLogs(c *gin.Context) {
	userID := c.Query("user_id")
	action := c.Query("action")
	resourceType := c.Query("resource_type")

	page := 1
	pageSize := 50

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Parse time range
	start, end := parseTimeRange(c)

	logs, total, err := h.userService.GetAuditLogs(c.Request.Context(), userID, action, resourceType, start, end, page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get audit logs")
		return
	}

	items := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		items[i] = AuditLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Format(time.RFC3339),
		}
	}

	response.Paged(c, items, page, pageSize, total)
}

// RegisterRoutes registers user routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.ListUsers)
		users.GET("/me", h.GetCurrentUser)
		users.PUT("/me/password", h.ChangePassword)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	// Audit logs
	r.GET("/audit-logs", h.GetAuditLogs)
}

// GetUser handles GET /api/v1/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.authService.GetUser(c.Request.Context(), id)
	if err != nil {
		if err == authService.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to get user")
		return
	}

	response.Success(c, toUserResponse(user))
}

// toUserResponse converts user to response.
func toUserResponse(user *domain.User) UserResponse {
	resp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastLoginAt != nil {
		formatted := user.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &formatted
	}

	return resp
}

// getUserID extracts user ID from context.
func getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	if s, ok := userID.(string); ok {
		return s
	}
	return ""
}

// parseInt parses a string to int.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseTimeRange parses time range from query parameters.
func parseTimeRange(c *gin.Context) (start, end time.Time) {
	end = time.Now()
	start = end.Add(-24 * time.Hour) // Default: last 24 hours

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
