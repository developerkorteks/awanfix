package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nabilulilalbab/rclonestorage/internal/auth"
	"github.com/nabilulilalbab/rclonestorage/internal/cache"
	"github.com/nabilulilalbab/rclonestorage/internal/config"
	"github.com/nabilulilalbab/rclonestorage/internal/storage"
)

var startTime = time.Now()

// API holds the API dependencies
type API struct {
	config      *config.Config
	storage     storage.UnionStorage
	authManager *auth.AuthManager
}

// NewAPI creates a new API instance
func NewAPI(cfg *config.Config, unionStorage storage.UnionStorage, authManager *auth.AuthManager) *API {
	return &API{
		config:      cfg,
		storage:     unionStorage,
		authManager: authManager,
	}
}

// SetupRoutes sets up all API routes with authentication
func SetupRoutes(r *gin.Engine, cfg *config.Config, authManager *auth.AuthManager) {
	// Initialize storage providers
	// TODO: Initialize actual storage providers
	
	api := NewAPI(cfg, nil, authManager) // Pass auth manager
	
	// Public API group (no authentication required)
	public := r.Group("/api/v1/public")
	{
		public.GET("/stats", api.handlePublicStats)
	}
	
	// Protected API group (authentication required)
	v1 := r.Group("/api/v1")
	v1.Use(authManager.Middleware.OptionalAuth()) // Allow both authenticated and API key access
	{
		// File management (requires authentication for upload/delete)
		v1.POST("/upload", authManager.Middleware.OptionalAuth(), authManager.Middleware.RequireAuth(), authManager.Middleware.AuditLog("upload"), api.handleUpload)
		v1.GET("/files", api.handleListFiles) // Can be public or user-specific
		v1.GET("/files/:id", api.handleGetFile)
		v1.DELETE("/files/:id", authManager.Middleware.OptionalAuth(), authManager.Middleware.RequireAuth(), authManager.Middleware.RequireFileOwnership(), authManager.Middleware.AuditLog("delete"), api.handleDeleteFile)
		
		// Download and streaming (can be public or authenticated)
		v1.GET("/download/:id", authManager.Middleware.AuditLog("download"), api.handleDownload)
		v1.GET("/stream/:id", authManager.Middleware.AuditLog("stream"), api.handleStream)
		v1.GET("/stream/:id/info", api.handleStreamInfo)
		
		// System endpoints (admin only) - Support both JWT and API key
		v1.GET("/stats", authManager.Middleware.OptionalAuth(), authManager.Middleware.RequireRole(auth.RoleAdmin), api.handleStats)
		v1.POST("/cache/clear", authManager.Middleware.OptionalAuth(), authManager.Middleware.RequireRole(auth.RoleAdmin), api.handleClearCache)
	}
}

// All handlers are now implemented in separate files:
// - handleUpload: upload.go
// - handleListFiles, handleGetFile, handleDownload: download.go  
// - handleStream, handleStreamInfo: stream.go
// - handleDeleteFile, handleClearCache: cache.go

// handleStats handles getting real system statistics
// @Summary Get system statistics
// @Description Get detailed system statistics including storage, cache, and system info (admin only)
// @Tags system
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "System statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Router /stats [get]
func (a *API) handleStats(c *gin.Context) {
	// Get real file count and size from cloud
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	var totalFiles int
	var totalSize int64
	
	if output, err := cmd.Output(); err == nil {
		var files []map[string]interface{}
		if json.Unmarshal(output, &files) == nil {
			totalFiles = len(files)
			for _, file := range files {
				if size, ok := file["Size"].(float64); ok {
					totalSize += int64(size)
				}
			}
		}
	}
	
	// Get cache statistics
	cacheManager, _ := cache.NewManager("./cache", 24*time.Hour, 10*1024*1024*1024)
	var cacheStats map[string]interface{}
	if cacheManager != nil {
		cacheStats = cacheManager.GetStats()
	} else {
		cacheStats = map[string]interface{}{
			"error": "Cache manager not available",
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"stats": gin.H{
			"storage": gin.H{
				"total_files":    totalFiles,
				"total_size":     totalSize,
				"size_human":     formatBytes(totalSize),
				"providers":      []string{"mega1", "mega2", "mega3", "gdrive"},
				"union_storage":  "active",
				"provider_count": 4,
			},
			"cache": cacheStats,
			"system": gin.H{
				"uptime":         time.Since(startTime),
				"cache_enabled":  true,
				"cache_ttl":      "24h",
				"max_cache_size": "10GB",
			},
		},
		"timestamp": time.Now(),
	})
}


// handlePublicStats handles getting public system statistics
// @Summary Get public statistics
// @Description Get public system statistics (no authentication required)
// @Tags public
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Public statistics"
// @Router /public/stats [get]
func (a *API) handlePublicStats(c *gin.Context) {
	// Get real file count and size from cloud
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	var totalFiles int
	var totalSize int64
	
	if output, err := cmd.Output(); err == nil {
		var files []map[string]interface{}
		if json.Unmarshal(output, &files) == nil {
			totalFiles = len(files)
			for _, file := range files {
				if size, ok := file["Size"].(float64); ok {
					totalSize += int64(size)
				}
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"public_stats": gin.H{
			"total_files":    totalFiles,
			"total_size":     totalSize,
			"size_human":     formatBytes(totalSize),
			"providers":      []string{"mega1", "mega2", "mega3", "gdrive"},
			"provider_count": 4,
			"features": []string{
				"multi-provider storage",
				"video streaming",
				"authentication",
				"api keys",
			},
		},
		"timestamp": time.Now(),
	})
}