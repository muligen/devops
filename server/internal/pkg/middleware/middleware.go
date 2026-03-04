// Package middleware provides HTTP middleware for the server.
package middleware

import (
	"strings"
	"time"

	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditLoggerConfig holds configuration for audit logging.
type AuditLoggerConfig struct {
	SkipPaths   []string
	SkipActions map[string]bool
}

// AuditLogger creates a middleware for audit logging.
func AuditLogger(auditService AuditService, cfg *AuditLoggerConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = &AuditLoggerConfig{}
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip if path should not be logged
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Only log mutations (POST, PUT, DELETE, PATCH)
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.Next()
			return
		}

		c.Next()

		// Get user info from context
		userID, _ := c.Get("user_id")
		userIDStr, _ := userID.(string)

		// Determine action from method
		action := actionFromMethod(method)

		// Determine resource type from path
		resourceType := resourceTypeFromPath(c.Request.URL.Path)

		// Get resource ID from path params
		resourceID := extractResourceID(c.Request.URL.Path)

		// Log the audit event
		if auditService != nil {
			_ = auditService.LogAudit(
				c.Request.Context(),
				userIDStr,
				action,
				resourceType,
				resourceID,
				c.ClientIP(),
				c.Request.UserAgent(),
				map[string]interface{}{
					"method":     method,
					"path":       c.Request.URL.Path,
					"status":     c.Writer.Status(),
					"user_agent": c.Request.UserAgent(),
				},
			)
		}
	}
}

// AuditService defines the interface for audit logging.
type AuditService interface {
	LogAudit(ctx interface{}, userID, action, resourceType, resourceID, ipAddress, userAgent string, details map[string]interface{}) error
}

// actionFromMethod returns an action type from HTTP method.
func actionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// resourceTypeFromPath extracts resource type from URL path.
func resourceTypeFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "api" {
		// /api/v1/users -> users
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return "unknown"
}

// extractResourceID extracts resource ID from URL path.
func extractResourceID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// /api/v1/users/123 -> 123
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

// RequestLogger returns a gin middleware for logging requests.
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fullPath := path
		if query != "" {
			fullPath = path + "?" + query
		}

		log.Info("HTTP request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", fullPath),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}

// Logger returns a gin middleware for logging requests.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		logger.Info("HTTP request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

// Recovery returns a gin middleware for recovering from panics.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// CORS returns a gin middleware for handling CORS.
func CORS(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
