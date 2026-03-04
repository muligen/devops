// Package response provides standardized HTTP response utilities.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	// Success
	CodeSuccess = 0

	// General errors (10000-10999)
	CodeBadRequest      = 10001
	CodeNotFound        = 10002
	CodeForbidden       = 10003
	CodeInternalError   = 10004
	CodeValidation      = 10005
	CodeConflict        = 10006
	CodeRateLimited     = 10007

	// Authentication errors (20000-20999)
	CodeUnauthorized    = 20001
	CodeTokenExpired    = 20002
	CodeInvalidCredentials = 20003
	CodeAccountLocked   = 20004

	// Agent errors (30000-30999)
	CodeAgentOffline    = 30001
	CodeAgentBusy       = 30002
	CodeAgentNotFound   = 30003
	CodeAgentExists     = 30004
	CodeAgentTokenInvalid = 30005

	// Task errors (40000-40999)
	CodeTaskTimeout     = 40001
	CodeTaskFailed      = 40002
	CodeTaskNotFound    = 40003
	CodeTaskCancelled   = 40004

	// System errors (50000-50999)
	CodeDatabaseError   = 50001
	CodeSystemError     = 50002
	CodeCacheError      = 50003
	CodeQueueError      = 50004
	CodeStorageError    = 50005
)

// Response is the standard API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Pagination holds pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PagedResponse is a response with pagination.
type PagedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination Pagination  `json:"pagination,omitempty"`
}

// Success returns a successful response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage returns a successful response with custom message.
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created returns a 201 Created response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// NoContent returns a 204 No Content response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paged returns a paginated response.
func Paged(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, PagedResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// BadRequest returns a 400 Bad Request response.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
	})
}

// ValidationErr returns a validation error response.
func ValidationErr(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeValidation,
		Message: message,
	})
}

// Unauthorized returns a 401 Unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// TokenExpired returns a token expired response.
func TokenExpired(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeTokenExpired,
		Message: "token expired",
	})
}

// Forbidden returns a 403 Forbidden response.
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound returns a 404 Not Found response.
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "resource not found"
	}
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Conflict returns a 409 Conflict response.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    CodeConflict,
		Message: message,
	})
}

// InternalError returns a 500 Internal Server Error response.
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
	})
}

// Error returns an error response with custom code and message.
func Error(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}
