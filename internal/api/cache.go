package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nabilulilalbab/rclonestorage/internal/cache"
)

// handleClearCache handles clearing cache
// @Summary Clear cache
// @Description Clear system cache (admin only)
// @Tags system
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Cache cleared successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cache/clear [post]
func (a *API) handleClearCache(c *gin.Context) {
	// Clear temp cache
	tempDir := "./cache/temp"
	
	files, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list cache files",
		})
		return
	}
	
	var removedFiles []string
	var errors []string
	
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			errors = append(errors, filepath.Base(file))
		} else {
			removedFiles = append(removedFiles, filepath.Base(file))
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":       "Cache cleared",
		"removed_files": removedFiles,
		"removed_count": len(removedFiles),
		"errors":        errors,
		"cache_dir":     tempDir,
	})
}

// handleDeleteFile handles real file deletion from cloud storage
// @Summary Delete file
// @Description Delete a file from cloud storage (requires ownership or admin)
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security ApiKeyAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "File deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - not file owner"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /files/{id} [delete]
func (a *API) handleDeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	
	// First, find the file in cloud storage
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
	
	// Parse JSON output to find our file
	var files []map[string]interface{}
	if err := json.Unmarshal(output, &files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse file list",
			"details": err.Error(),
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
			"error": "File not found in cloud storage",
			"file_id": fileID,
		})
		return
	}
	
	filename := targetFile["Name"].(string)
	size := int64(targetFile["Size"].(float64))
	
	// Delete from cloud storage
	remotePath := fmt.Sprintf("union:uploads/%s", filename)
	deleteCmd := exec.Command("rclone", "delete", remotePath)
	if a.config.Rclone.ConfigPath != "" {
		deleteCmd.Env = append(deleteCmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	if err := deleteCmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file from cloud storage",
			"details": err.Error(),
			"file_id": fileID,
			"filename": filename,
		})
		return
	}
	
	// Also clear from cache if exists
	cacheManager, _ := cache.NewManager("./cache", 24*time.Hour, 10*1024*1024*1024)
	if cacheManager != nil {
		cacheKeys := []string{
			fmt.Sprintf("download_%s", fileID),
			fmt.Sprintf("stream_%s", fileID),
		}
		
		for _, key := range cacheKeys {
			cacheManager.Delete(context.Background(), key)
		}
	}
	
	// Also clean temp cache
	tempDir := "./cache/temp"
	pattern := filepath.Join(tempDir, fileID+"_*")
	tempFiles, _ := filepath.Glob(pattern)
	var deletedTempFiles []string
	for _, file := range tempFiles {
		if err := os.Remove(file); err == nil {
			deletedTempFiles = append(deletedTempFiles, filepath.Base(file))
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully from cloud storage",
		"file_id": fileID,
		"deleted_file": gin.H{
			"filename":    filename,
			"size":        size,
			"size_human":  formatBytes(size),
			"remote_path": remotePath,
		},
		"cache_cleared": gin.H{
			"download_cache": fmt.Sprintf("download_%s", fileID),
			"stream_cache":   fmt.Sprintf("stream_%s", fileID),
			"temp_files":     deletedTempFiles,
		},
		"status": "deleted_from_cloud",
	})
}