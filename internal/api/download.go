package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nabilulilalbab/rclonestorage/internal/cache"
)

// handleDownload handles file download with caching
// @Summary Download file
// @Description Download a file from storage with caching support
// @Tags files
// @Produce application/octet-stream
// @Param id path string true "File ID"
// @Success 200 {file} file "File content"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /download/{id} [get]
func (a *API) handleDownload(c *gin.Context) {
	fileID := c.Param("id")
	
	// Try to get from cache first
	cacheManager, err := cache.NewManager("./cache", 24*time.Hour, 10*1024*1024*1024) // 10GB
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize cache",
		})
		return
	}
	
	cacheKey := fmt.Sprintf("download_%s", fileID)
	
	// Check cache first
	if reader, entry, err := cacheManager.Get(context.Background(), cacheKey); err == nil {
		defer reader.Close()
		
		// Serve from cache
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", strconv.FormatInt(entry.Size, 10))
		c.Header("X-Cache", "HIT")
		
		io.Copy(c.Writer, reader)
		return
	}
	
	// Cache miss - download from cloud
	c.Header("X-Cache", "MISS")
	
	// List files to find the actual filename
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files from cloud",
			"details": err.Error(),
		})
		return
	}
	
	// Parse JSON output to find our file
	var files []map[string]interface{}
	if err := json.Unmarshal(output, &files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse file list",
		})
		return
	}
	
	var targetFile map[string]interface{}
	for _, file := range files {
		if name, ok := file["Name"].(string); ok {
			if strings.HasPrefix(name, fileID+"_") {
				targetFile = file
				break
			}
		}
	}
	
	if targetFile == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
			"file_id": fileID,
		})
		return
	}
	
	filename := targetFile["Name"].(string)
	// size := int64(targetFile["Size"].(float64)) // Not needed anymore
	
	// Download from cloud using rclone cat
	cmd = exec.Command("rclone", "cat", fmt.Sprintf("union:uploads/%s", filename))
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	// Get the file content
	fileContent, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to download file from cloud",
			"details": err.Error(),
		})
		return
	}
	
	// Cache the file content
	go func() {
		// Create a reader from the content for caching
		contentReader := strings.NewReader(string(fileContent))
		if _, err := cacheManager.Put(context.Background(), cacheKey, contentReader, int64(len(fileContent))); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to cache file %s: %v\n", fileID, err)
		}
	}()
	
	// Serve the file
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.Itoa(len(fileContent)))
	c.Header("X-Cache", "MISS")
	
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// handleListFiles handles listing files from cloud storage
// @Summary List files
// @Description Get list of files with optional filtering and pagination
// @Tags files
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} map[string]interface{} "List of files"
// @Router /files [get]
func (a *API) handleListFiles(c *gin.Context) {
	// List files from union storage using rclone
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files from cloud storage",
			"details": err.Error(),
		})
		return
	}
	
	// Parse JSON output
	var rcloneFiles []map[string]interface{}
	if err := json.Unmarshal(output, &rcloneFiles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse file list",
			"details": err.Error(),
		})
		return
	}
	
	// Convert to our format
	var files []gin.H
	var totalSize int64
	
	for _, file := range rcloneFiles {
		name := file["Name"].(string)
		size := int64(file["Size"].(float64))
		modTime := file["ModTime"].(string)
		
		// Extract file ID from filename (format: fileID_originalname)
		parts := strings.SplitN(name, "_", 2)
		fileID := parts[0]
		originalName := name
		if len(parts) > 1 {
			originalName = parts[1]
		}
		
		files = append(files, gin.H{
			"id":           fileID,
			"name":         originalName,
			"filename":     name,
			"size":         size,
			"modified":     modTime,
			"provider":     "union",
			"downloadable": true,
		})
		
		totalSize += size
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Files listed successfully",
		"files":      files,
		"total":      len(files),
		"total_size": totalSize,
		"provider":   "union (mega1 + mega2 + mega3 + gdrive)",
		"source":     "cloud_storage",
	})
}

// handleGetFile handles getting file info from cloud storage
// @Summary Get file info
// @Description Get detailed information about a specific file
// @Tags files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "File information"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Router /files/{id} [get]
func (a *API) handleGetFile(c *gin.Context) {
	fileID := c.Param("id")
	
	// List files from union storage to find our file
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access cloud storage",
			"details": err.Error(),
		})
		return
	}
	
	// Parse JSON output
	var rcloneFiles []map[string]interface{}
	if err := json.Unmarshal(output, &rcloneFiles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse file list",
			"details": err.Error(),
		})
		return
	}
	
	// Find our file
	var targetFile map[string]interface{}
	for _, file := range rcloneFiles {
		if name, ok := file["Name"].(string); ok {
			if strings.HasPrefix(name, fileID+"_") {
				targetFile = file
				break
			}
		}
	}
	
	if targetFile == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
			"file_id": fileID,
		})
		return
	}
	
	// Extract file information
	filename := targetFile["Name"].(string)
	size := int64(targetFile["Size"].(float64))
	modTime := targetFile["ModTime"].(string)
	isDir := targetFile["IsDir"].(bool)
	
	// Extract original name
	parts := strings.SplitN(filename, "_", 2)
	originalName := filename
	if len(parts) > 1 {
		originalName = parts[1]
	}
	
	// Determine file type
	ext := strings.ToLower(filepath.Ext(originalName))
	fileType := "unknown"
	streamable := false
	
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm":
		fileType = "video"
		streamable = true
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		fileType = "audio"
		streamable = true
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		fileType = "image"
	case ".pdf":
		fileType = "document"
	case ".txt", ".md", ".log":
		fileType = "text"
	default:
		fileType = "file"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "File info retrieved successfully",
		"file_id": fileID,
		"file": gin.H{
			"id":           fileID,
			"name":         originalName,
			"filename":     filename,
			"size":         size,
			"size_human":   formatBytes(size),
			"modified":     modTime,
			"is_dir":       isDir,
			"type":         fileType,
			"extension":    ext,
			"provider":     "union",
			"streamable":   streamable,
			"downloadable": true,
		},
		"actions": gin.H{
			"download": fmt.Sprintf("/api/v1/download/%s", fileID),
			"stream":   fmt.Sprintf("/api/v1/stream/%s", fileID),
			"delete":   fmt.Sprintf("/api/v1/files/%s", fileID),
		},
		"source": "cloud_storage",
	})
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}