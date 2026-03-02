// Package handler provides HTTP handlers for authentication.
package handler

import (
	"strconv"
	"time"

	"github.com/agentteams/server/internal/modules/auth/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles authentication HTTP requests.
type Handler struct {
	authService *service.Service
	jwtService  *service.JWTService
}

// NewHandler creates a new auth handler.
func NewHandler(authService *service.Service, jwtService *service.JWTService) *Handler {
	return &Handler{
		authService: authService,
		jwtService:  jwtService,
	}
}

// LoginRequest represents login request.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// Login handles user login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, "invalid request body")
		return
	}

	// Authenticate user
	user, err := h.authService.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			response.Unauthorized(c, "invalid username or password")
		case service.ErrAccountLocked:
			response.Forbidden(c, "account is locked")
		default:
			response.InternalError(c, "authentication failed")
		}
		return
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.InternalError(c, "failed to generate refresh token")
		return
	}

	// Build response
	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * time.Hour / time.Second), // 24 hours
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.Email = user.Email
	resp.User.Role = user.Role

	response.Success(c, resp)
}

// RefreshTokenRequest represents refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles token refresh.
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, "invalid request body")
		return
	}

	// Validate refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	// Generate new access token
	accessToken, err := h.jwtService.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}

	// Generate new refresh token
	refreshToken, err := h.jwtService.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		response.InternalError(c, "failed to generate refresh token")
		return
	}

	response.Success(c, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	})
}

// Logout handles user logout.
func (h *Handler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is handled client-side
	// by removing the token. Server-side token blacklisting could be
	// implemented with Redis if needed.
	response.Success(c, gin.H{
		"message": "logged out successfully",
	})
}

// CreateUserRequest represents create user request.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// CreateUser handles user creation (admin only).
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	user, err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email, req.Role)
	if err != nil {
		if err == service.ErrUserExists {
			response.Conflict(c, "user already exists")
			return
		}
		response.InternalError(c, "failed to create user")
		return
	}

	response.Created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// ListUsers handles listing users.
func (h *Handler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 20

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

	users, total, err := h.authService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}

	// Map to response
	items := make([]gin.H, len(users))
	for i, user := range users {
		items[i] = gin.H{
			"id":            user.ID,
			"username":      user.Username,
			"email":         user.Email,
			"role":          user.Role,
			"status":        user.Status,
			"last_login_at": user.LastLoginAt,
			"created_at":    user.CreatedAt,
		}
	}

	response.Paged(c, items, page, pageSize, total)
}

// RegisterRoutes registers auth routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/refresh", h.RefreshToken)
	}
}
