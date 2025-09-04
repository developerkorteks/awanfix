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

// handleStream handles video streaming with HTTP range requests
// @Summary Stream video file
// @Description Stream video file with range support for progressive loading
// @Tags streaming
// @Produce video/*
// @Param id path string true "File ID"
// @Param Range header string false "Range header for partial content"
// @Success 200 {file} file "Video stream"
// @Success 206 {file} file "Partial content"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /stream/{id} [get]
func (a *API) handleStream(c *gin.Context) {
	fileID := c.Param("id")
	
	// Get file info first
	fileInfo, err := a.getFileInfo(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
			"file_id": fileID,
		})
		return
	}
	
	// Check if file is streamable
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	if !isStreamableFormat(ext) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File format not streamable",
			"format": ext,
			"file_id": fileID,
		})
		return
	}
	
	// Initialize cache
	cacheManager, err := cache.NewManager("./cache", 24*time.Hour, 10*1024*1024*1024)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize cache",
		})
		return
	}
	
	cacheKey := fmt.Sprintf("stream_%s", fileID)
	
	// Parse range header
	rangeHeader := c.GetHeader("Range")
	var start, end int64
	var isRangeRequest bool
	
	if rangeHeader != "" {
		isRangeRequest = true
		ranges := parseRangeHeader(rangeHeader, fileInfo.Size)
		if len(ranges) > 0 {
			start = ranges[0].Start
			end = ranges[0].End
		}
	} else {
		start = 0
		end = fileInfo.Size - 1
	}
	
	// Try cache first for full file
	if !isRangeRequest {
		if reader, entry, err := cacheManager.Get(context.Background(), cacheKey); err == nil {
			defer reader.Close()
			
			c.Header("Content-Type", getContentType(ext))
			c.Header("Content-Length", strconv.FormatInt(entry.Size, 10))
			c.Header("Accept-Ranges", "bytes")
			c.Header("X-Cache", "HIT")
			
			io.Copy(c.Writer, reader)
			return
		}
	}
	
	// Stream from cloud with range support
	if isRangeRequest {
		a.streamWithRange(c, fileInfo, start, end)
	} else {
		a.streamFullFile(c, fileInfo, cacheManager, cacheKey)
	}
}

// streamWithRange handles range requests for video streaming
func (a *API) streamWithRange(c *gin.Context, fileInfo *FileInfo, start, end int64) {
	// For range requests, we need to download the specific range
	// Since rclone doesn't support range directly, we'll stream and seek
	
	cmd := exec.Command("rclone", "cat", fmt.Sprintf("union:uploads/%s", fileInfo.Filename))
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create stream pipe",
		})
		return
	}
	
	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start stream",
		})
		return
	}
	
	// Skip to start position
	if start > 0 {
		io.CopyN(io.Discard, stdout, start)
	}
	
	// Calculate content length for range
	contentLength := end - start + 1
	
	// Set range response headers
	c.Header("Content-Type", getContentType(filepath.Ext(fileInfo.Name)))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size))
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-Cache", "MISS")
	c.Status(http.StatusPartialContent)
	
	// Stream the requested range
	io.CopyN(c.Writer, stdout, contentLength)
	cmd.Wait()
}

// streamFullFile handles full file streaming with caching
func (a *API) streamFullFile(c *gin.Context, fileInfo *FileInfo, cacheManager *cache.Manager, cacheKey string) {
	cmd := exec.Command("rclone", "cat", fmt.Sprintf("union:uploads/%s", fileInfo.Filename))
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create stream pipe",
		})
		return
	}
	
	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start stream",
		})
		return
	}
	
	// Set headers for full file
	c.Header("Content-Type", getContentType(filepath.Ext(fileInfo.Name)))
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-Cache", "MISS")
	
	// Create a tee reader to cache while streaming
	pr, pw := io.Pipe()
	teeReader := io.TeeReader(stdout, pw)
	
	// Cache in background
	go func() {
		defer pw.Close()
		defer cmd.Wait()
		cacheManager.Put(context.Background(), cacheKey, pr, fileInfo.Size)
	}()
	
	// Stream to client
	io.Copy(c.Writer, teeReader)
}

// handleStreamInfo handles getting real stream info
// @Summary Get stream info
// @Description Get video stream information including duration, bitrate, resolution
// @Tags streaming
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{} "Stream information"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Router /stream/{id}/info [get]
func (a *API) handleStreamInfo(c *gin.Context) {
	fileID := c.Param("id")
	
	// Get file info
	fileInfo, err := a.getFileInfo(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
			"file_id": fileID,
		})
		return
	}
	
	// Check if file is streamable
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	streamable := isStreamableFormat(ext)
	
	if !streamable {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is not streamable",
			"format": ext,
			"file_id": fileID,
		})
		return
	}
	
	// Get real file metadata
	fileType := getFileType(ext)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Stream info retrieved successfully",
		"file_id": fileID,
		"info": gin.H{
			"filename":    fileInfo.Name,
			"size":        fileInfo.Size,
			"size_human":  formatBytes(fileInfo.Size),
			"format":      strings.TrimPrefix(ext, "."),
			"type":        fileType,
			"modified":    fileInfo.ModTime,
			"streamable":  true,
			"provider":    "union",
		},
		"streaming_urls": gin.H{
			"direct":     fmt.Sprintf("/api/v1/stream/%s", fileID),
			"download":   fmt.Sprintf("/api/v1/download/%s", fileID),
		},
		"capabilities": gin.H{
			"range_requests": true,
			"progressive":    true,
			"cacheable":      true,
		},
		"source": "cloud_storage",
	})
}

// Helper functions
type FileInfo struct {
	ID       string
	Name     string
	Filename string
	Size     int64
	ModTime  string
}

type RangeSpec struct {
	Start int64
	End   int64
}

// getFileInfo retrieves file information from cloud
func (a *API) getFileInfo(fileID string) (*FileInfo, error) {
	cmd := exec.Command("rclone", "lsjson", "union:uploads/")
	if a.config.Rclone.ConfigPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", a.config.Rclone.ConfigPath))
	}
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var files []map[string]interface{}
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, err
	}
	
	for _, file := range files {
		if name, ok := file["Name"].(string); ok {
			if strings.HasPrefix(name, fileID+"_") {
				parts := strings.SplitN(name, "_", 2)
				originalName := name
				if len(parts) > 1 {
					originalName = parts[1]
				}
				
				return &FileInfo{
					ID:       fileID,
					Name:     originalName,
					Filename: name,
					Size:     int64(file["Size"].(float64)),
					ModTime:  file["ModTime"].(string),
				}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("file not found")
}

// isStreamableFormat checks if file format is streamable
func isStreamableFormat(ext string) bool {
	streamableFormats := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".aac":  true,
		".ogg":  true,
	}
	return streamableFormats[ext]
}

// getFileType returns file type based on extension
func getFileType(ext string) string {
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	default:
		return "unknown"
	}
}

// getContentType returns MIME type for file extension
func getContentType(ext string) string {
	contentTypes := map[string]string{
		".mp4":  "video/mp4",
		".mkv":  "video/x-matroska",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
	}
	
	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "application/octet-stream"
}

// parseRangeHeader parses HTTP Range header
func parseRangeHeader(rangeHeader string, fileSize int64) []RangeSpec {
	var ranges []RangeSpec
	
	// Remove "bytes=" prefix
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return ranges
	}
	
	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	rangeParts := strings.Split(rangeStr, ",")
	
	for _, part := range rangeParts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeBounds := strings.Split(part, "-")
			if len(rangeBounds) == 2 {
				var start, end int64
				var err error
				
				if rangeBounds[0] != "" {
					start, err = strconv.ParseInt(rangeBounds[0], 10, 64)
					if err != nil {
						continue
					}
				}
				
				if rangeBounds[1] != "" {
					end, err = strconv.ParseInt(rangeBounds[1], 10, 64)
					if err != nil {
						continue
					}
				} else {
					end = fileSize - 1
				}
				
				if start <= end && start < fileSize {
					if end >= fileSize {
						end = fileSize - 1
					}
					ranges = append(ranges, RangeSpec{Start: start, End: end})
				}
			}
		}
	}
	
	return ranges
}