// Package handler provides HTTP handlers for update management.
package handler

import (
	"io"
	"time"

	"github.com/agentteams/server/internal/modules/update/domain"
	"github.com/agentteams/server/internal/modules/update/service"
	"github.com/agentteams/server/internal/pkg/response"
	"github.com/agentteams/server/internal/pkg/storage"
	"github.com/gin-gonic/gin"
)

// Handler handles update HTTP requests.
type Handler struct {
	service *service.Service
	storage *storage.Client
}

// NewHandler creates a new update handler.
func NewHandler(svc *service.Service, storage *storage.Client) *Handler {
	return &Handler{
		service: svc,
		storage: storage,
	}
}

// VersionResponse represents a version response.
type VersionResponse struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	Platform     string `json:"platform"`
	FileURL      string `json:"file_url"`
	FileSize     int64  `json:"file_size"`
	FileHash     string `json:"file_hash"`
	ReleaseNotes string `json:"release_notes"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// CreateVersionRequest represents create version request.
type CreateVersionRequest struct {
	Version      string `json:"version" binding:"required"`
	Platform     string `json:"platform"`
	FileURL      string `json:"file_url"`
	FileHash     string `json:"file_hash"`
	ReleaseNotes string `json:"release_notes"`
}

// UploadVersion handles POST /api/v1/versions
func (h *Handler) UploadVersion(c *gin.Context) {
	// Parse form data
	version := c.PostForm("version")
	if version == "" {
		response.ValidationErr(c, "version is required")
		return
	}

	platform := c.PostForm("platform")
	if platform == "" {
		platform = "windows"
	}
	releaseNotes := c.PostForm("release_notes")

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ValidationErr(c, "file is required")
		return
	}
	defer file.Close()

	// Upload to storage
	objectName := "versions/" + platform + "/" + version + "/" + header.Filename
	contentType := "application/octet-stream"

	if err := h.storage.Upload(c.Request.Context(), objectName, file, header.Size, contentType); err != nil {
		response.InternalError(c, "failed to upload file")
		return
	}

	// Create version record
	v, err := h.service.CreateVersion(
		c.Request.Context(),
		version,
		platform,
		objectName,
		header.Size,
		"", // file hash - would be calculated in production
		"", // signature - would be generated in production
		releaseNotes,
	)
	if err != nil {
		response.InternalError(c, "failed to create version")
		return
	}

	response.Created(c, toVersionResponse(v))
}

// ListVersions handles GET /api/v1/versions
func (h *Handler) ListVersions(c *gin.Context) {
	platform := c.Query("platform")
	limit := 50

	versions, err := h.service.ListVersions(c.Request.Context(), platform, limit)
	if err != nil {
		response.InternalError(c, "failed to list versions")
		return
	}

	items := make([]VersionResponse, len(versions))
	for i, v := range versions {
		items[i] = toVersionResponse(&v)
	}

	response.Success(c, items)
}

// GetVersion handles GET /api/v1/versions/:id
func (h *Handler) GetVersion(c *gin.Context) {
	id := c.Param("id")

	v, err := h.service.GetVersion(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "version not found")
		return
	}

	response.Success(c, toVersionResponse(v))
}

// DeleteVersion handles DELETE /api/v1/versions/:id
func (h *Handler) DeleteVersion(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteVersion(c.Request.Context(), id); err != nil {
		response.NotFound(c, "version not found")
		return
	}

	response.Success(c, gin.H{"message": "version deleted"})
}

// GetDownloadURL handles GET /api/v1/versions/:id/download
func (h *Handler) GetDownloadURL(c *gin.Context) {
	id := c.Param("id")

	url, err := h.service.GetDownloadURL(c.Request.Context(), id, 1*time.Hour)
	if err != nil {
		response.NotFound(c, "version not found")
		return
	}

	response.Success(c, gin.H{
		"download_url": url,
		"expires_in":   3600,
	})
}

// CheckForUpdate handles GET /api/v1/versions/check
// This endpoint is for agents to check for updates
func (h *Handler) CheckForUpdate(c *gin.Context) {
	currentVersion := c.Query("current_version")
	platform := c.Query("platform")
	if platform == "" {
		platform = "windows"
	}

	v, err := h.service.CheckForUpdate(c.Request.Context(), currentVersion, platform)
	if err != nil {
		response.InternalError(c, "failed to check for updates")
		return
	}

	if v == nil {
		response.Success(c, gin.H{
			"update_available": false,
		})
		return
	}

	// Generate download URL
	downloadURL, err := h.service.GetDownloadURL(c.Request.Context(), v.ID, 1*time.Hour)
	if err != nil {
		response.InternalError(c, "failed to generate download URL")
		return
	}

	response.Success(c, gin.H{
		"update_available": true,
		"version":          v.Version,
		"platform":         v.Platform,
		"file_size":        v.FileSize,
		"file_hash":        v.FileHash,
		"signature":        v.Signature,
		"download_url":     downloadURL,
		"release_notes":    v.ReleaseNotes,
	})
}

// TriggerUpdateRequest represents trigger update request.
type TriggerUpdateRequest struct {
	VersionID string `json:"version_id" binding:"required"`
}

// TriggerUpdate handles POST /api/v1/agents/:id/update
func (h *Handler) TriggerUpdate(c *gin.Context) {
	agentID := c.Param("id")

	var req TriggerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	status, err := h.service.TriggerUpdate(c.Request.Context(), agentID, req.VersionID)
	if err != nil {
		response.InternalError(c, "failed to trigger update")
		return
	}

	response.Success(c, gin.H{
		"status_id": status.ID,
		"status":    status.Status,
		"message":   "update triggered",
	})
}

// UpdateStatusRequest represents update status request.
type UpdateStatusRequest struct {
	VersionID string `json:"version_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Message   string `json:"message"`
}

// UpdateStatus handles POST /api/v1/agents/:id/update/status
// This endpoint is for agents to report update status
func (h *Handler) UpdateStatus(c *gin.Context) {
	agentID := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErr(c, err.Error())
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), agentID, req.VersionID, req.Status, req.Message); err != nil {
		response.InternalError(c, "failed to update status")
		return
	}

	response.Success(c, gin.H{"message": "status updated"})
}

// GetUpdateStatus handles GET /api/v1/agents/:id/update/status
func (h *Handler) GetUpdateStatus(c *gin.Context) {
	agentID := c.Param("id")

	status, err := h.service.GetUpdateStatus(c.Request.Context(), agentID)
	if err != nil {
		response.InternalError(c, "failed to get update status")
		return
	}

	if status == nil {
		response.Success(c, gin.H{
			"status": "none",
		})
		return
	}

	response.Success(c, gin.H{
		"id":          status.ID,
		"version_id":  status.VersionID,
		"status":      status.Status,
		"message":     status.Message,
		"started_at":  status.StartedAt.Format(time.RFC3339),
		"finished_at": formatTimePtr(status.FinishedAt),
	})
}

// RegisterRoutes registers update routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	versions := r.Group("/versions")
	{
		versions.POST("", h.UploadVersion)
		versions.GET("", h.ListVersions)
		versions.GET("/check", h.CheckForUpdate)
		versions.GET("/:id", h.GetVersion)
		versions.DELETE("/:id", h.DeleteVersion)
		versions.GET("/:id/download", h.GetDownloadURL)
	}

	// Agent update routes
	agents := r.Group("/agents")
	{
		agents.POST("/:id/update", h.TriggerUpdate)
		agents.GET("/:id/update/status", h.GetUpdateStatus)
		agents.POST("/:id/update/status", h.UpdateStatus)
	}
}

// toVersionResponse converts version to response.
func toVersionResponse(v *domain.Version) VersionResponse {
	return VersionResponse{
		ID:           v.ID,
		Version:      v.Version,
		Platform:     v.Platform,
		FileURL:      v.FileURL,
		FileSize:     v.FileSize,
		FileHash:     v.FileHash,
		ReleaseNotes: v.ReleaseNotes,
		IsActive:     v.IsActive,
		CreatedAt:    v.CreatedAt.Format(time.RFC3339),
	}
}

// formatTimePtr formats a time pointer.
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// UploadHandler handles file uploads with streaming.
type UploadHandler struct {
	service *service.Service
	storage *storage.Client
}

// NewUploadHandler creates a new upload handler.
func NewUploadHandler(svc *service.Service, storage *storage.Client) *UploadHandler {
	return &UploadHandler{
		service: svc,
		storage: storage,
	}
}

// StreamUpload handles streaming file upload.
func (h *UploadHandler) StreamUpload(c *gin.Context) {
	// Get content length for streaming
	contentLength := c.Request.ContentLength
	if contentLength <= 0 {
		response.ValidationErr(c, "invalid content length")
		return
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "failed to read body")
		return
	}

	// Calculate hash (in production, use streaming hash calculation)
	// hash := sha256.Sum256(body)

	// Upload to storage
	objectName := "uploads/" + time.Now().Format("20060102-150405") + "-" + c.Param("filename")
	if err := h.storage.Upload(c.Request.Context(), objectName, nil, int64(len(body)), "application/octet-stream"); err != nil {
		response.InternalError(c, "failed to upload")
		return
	}

	response.Success(c, gin.H{
		"path": objectName,
		"size": len(body),
	})
}
