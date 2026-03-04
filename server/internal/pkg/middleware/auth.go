// Package middleware provides authentication middleware.
package middleware

import (
	"strings"

	"github.com/agentteams/server/internal/modules/auth/domain"
	"github.com/agentteams/server/internal/modules/auth/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides JWT authentication middleware.
func AuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			if err == service.ErrTokenExpired {
				response.TokenExpired(c)
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRoles creates a middleware that requires specific roles.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Forbidden(c, "invalid role type")
			c.Abort()
			return
		}
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts user ID from context.
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if s, ok := userID.(string); ok {
			return s
		}
	}
	return ""
}

// GetUsername extracts username from context.
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok {
			return s
		}
	}
	return ""
}

// GetRole extracts role from context.
func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if s, ok := role.(string); ok {
			return s
		}
	}
	return ""
}

// IsAdmin checks if current user is admin.
func IsAdmin(c *gin.Context) bool {
	return GetRole(c) == domain.RoleAdmin
}

// CanCreateTask checks if current user can create tasks.
func CanCreateTask(c *gin.Context) bool {
	role := GetRole(c)
	return role == domain.RoleAdmin || role == domain.RoleOperator
}
