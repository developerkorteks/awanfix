package api

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nabilulilalbab/rclonestorage/internal/auth"
)

// handleUpload handles file upload with authentication and ownership tracking
// @Summary Upload file
// @Description Upload a file to cloud storage with authentication and ownership tracking
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Security ApiKeyAuth
// @Param file formData file true "File to upload"
// @Param description formData string false "File description"
// @Success 200 {object} map[string]interface{} "File uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - no file uploaded"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - upload permission denied or quota exceeded"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /upload [post]
func (a *API) handleUpload(c *gin.Context) {
	// Get current user
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Check if user can upload
	if !user.CanUpload() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Upload permission denied",
		})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// Check storage quota
	if !user.HasStorageSpace(file.Size) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Storage quota exceeded",
			"quota": user.StorageQuota,
			"used":  user.StorageUsed,
			"required": file.Size,
		})
		return
	}

	// Generate unique filename
	fileID := uuid.New().String()
	filename := fmt.Sprintf("%s_%s", fileID, file.Filename)
	
	// Create temp directory if not exists
	tempDir := "./cache/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create temp directory",
		})
		return
	}

	// Save file temporarily
	tempPath := filepath.Join(tempDir, filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}

	// Upload to union storage using rclone
	remotePath := fmt.Sprintf("union:uploads/%s", filename)
	
	// Execute rclone copy to upload file to cloud
	cmd := exec.Command("rclone", "copy", tempPath, "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	if err := cmd.Run(); err != nil {
		// Clean up temp file
		os.Remove(tempPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload to cloud storage",
			"details": err.Error(),
		})
		return
	}
	
	// Determine MIME type
	mimeType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".mp4":
		mimeType = "video/mp4"
	case ".mkv":
		mimeType = "video/x-matroska"
	case ".avi":
		mimeType = "video/x-msvideo"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".txt":
		mimeType = "text/plain"
	case ".pdf":
		mimeType = "application/pdf"
	}

	// Create file ownership record
	if err := a.authManager.DatabaseManager.CreateFileOwnership(
		user.ID,
		fileID,
		file.Filename,
		"union",
		file.Size,
		mimeType,
	); err != nil {
		// File uploaded but ownership tracking failed
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to create file ownership record: %v\n", err)
	}
	
	// Clean up temp file after successful upload
	os.Remove(tempPath)
	
	c.JSON(http.StatusOK, gin.H{
		"message":     "File uploaded successfully to cloud",
		"file_id":     fileID,
		"filename":    file.Filename,
		"size":        file.Size,
		"mime_type":   mimeType,
		"remote_path": remotePath,
		"status":      "uploaded_to_cloud",
		"uploaded_at": time.Now(),
		"owner":       user.Email,
	})
}