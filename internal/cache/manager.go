package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Manager handles file caching with TTL
type Manager struct {
	cacheDir    string
	ttl         time.Duration
	maxSize     int64
	currentSize int64
	metadata    *cache.Cache
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// CacheEntry represents a cached file entry
type CacheEntry struct {
	FilePath    string    `json:"file_path"`
	OriginalKey string    `json:"original_key"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	AccessedAt  time.Time `json:"accessed_at"`
	AccessCount int64     `json:"access_count"`
}

// NewManager creates a new cache manager
func NewManager(cacheDir string, ttl time.Duration, maxSize int64) (*Manager, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create subdirectories
	for _, subdir := range []string{"files", "metadata", "temp"} {
		if err := os.MkdirAll(filepath.Join(cacheDir, subdir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache subdirectory %s: %w", subdir, err)
		}
	}

	manager := &Manager{
		cacheDir: cacheDir,
		ttl:      ttl,
		maxSize:  maxSize,
		metadata: cache.New(ttl, ttl/2), // Cleanup every TTL/2
		logger:   logrus.New(),
	}

	// Calculate current cache size
	if err := manager.calculateCurrentSize(); err != nil {
		manager.logger.Warnf("Failed to calculate current cache size: %v", err)
	}

	// Start cleanup goroutine
	go manager.startCleanupRoutine()

	return manager, nil
}

// Get retrieves a file from cache
func (m *Manager) Get(ctx context.Context, key string) (io.ReadCloser, *CacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheKey := m.generateCacheKey(key)
	
	// Check if entry exists in metadata
	if item, found := m.metadata.Get(cacheKey); found {
		entry := item.(*CacheEntry)
		
		// Check if file still exists on disk
		if _, err := os.Stat(entry.FilePath); err == nil {
			// Update access time and count
			entry.AccessedAt = time.Now()
			entry.AccessCount++
			m.metadata.Set(cacheKey, entry, m.ttl)
			
			// Open file for reading
			file, err := os.Open(entry.FilePath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open cached file: %w", err)
			}
			
			return file, entry, nil
		} else {
			// File doesn't exist, remove from metadata
			m.metadata.Delete(cacheKey)
		}
	}
	
	return nil, nil, fmt.Errorf("cache miss for key: %s", key)
}

// Put stores a file in cache
func (m *Manager) Put(ctx context.Context, key string, reader io.Reader, size int64) (*CacheEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we need to free up space
	if err := m.ensureSpace(size); err != nil {
		return nil, fmt.Errorf("failed to ensure cache space: %w", err)
	}

	cacheKey := m.generateCacheKey(key)
	filePath := filepath.Join(m.cacheDir, "files", cacheKey)
	
	// Create temporary file first
	tempPath := filepath.Join(m.cacheDir, "temp", cacheKey+".tmp")
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy data to temp file
	written, err := io.Copy(tempFile, reader)
	if err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to move temp file to cache: %w", err)
	}

	// Create cache entry
	entry := &CacheEntry{
		FilePath:    filePath,
		OriginalKey: key,
		Size:        written,
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 1,
	}

	// Store in metadata
	m.metadata.Set(cacheKey, entry, m.ttl)
	m.currentSize += written

	m.logger.Infof("Cached file: %s (size: %d bytes)", key, written)
	
	return entry, nil
}

// Delete removes a file from cache
func (m *Manager) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cacheKey := m.generateCacheKey(key)
	
	if item, found := m.metadata.Get(cacheKey); found {
		entry := item.(*CacheEntry)
		
		// Remove file from disk
		if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove cached file: %w", err)
		}
		
		// Remove from metadata
		m.metadata.Delete(cacheKey)
		m.currentSize -= entry.Size
		
		m.logger.Infof("Removed cached file: %s", key)
	}
	
	return nil
}

// Clear removes all files from cache
func (m *Manager) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all files
	filesDir := filepath.Join(m.cacheDir, "files")
	if err := os.RemoveAll(filesDir); err != nil {
		return fmt.Errorf("failed to remove cache files: %w", err)
	}

	// Recreate files directory
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate cache files directory: %w", err)
	}

	// Clear metadata
	m.metadata.Flush()
	m.currentSize = 0

	m.logger.Info("Cache cleared")
	
	return nil
}

// Stats returns cache statistics
func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"current_size": m.currentSize,
		"max_size":     m.maxSize,
		"item_count":   m.metadata.ItemCount(),
		"hit_rate":     m.calculateHitRate(),
	}
}

// generateCacheKey generates a cache key from the original key
func (m *Manager) generateCacheKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// ensureSpace ensures there's enough space for a new file
func (m *Manager) ensureSpace(requiredSize int64) error {
	if m.currentSize+requiredSize <= m.maxSize {
		return nil
	}

	// Need to free up space - implement LRU eviction
	return m.evictLRU(requiredSize)
}

// evictLRU evicts least recently used files
func (m *Manager) evictLRU(requiredSize int64) error {
	// This is a simplified LRU implementation
	// In production, you'd want a more sophisticated approach
	
	items := m.metadata.Items()
	if len(items) == 0 {
		return fmt.Errorf("cache is full and no items to evict")
	}

	// Sort by access time and evict oldest
	// TODO: Implement proper LRU sorting
	
	return nil
}

// calculateCurrentSize calculates the current cache size
func (m *Manager) calculateCurrentSize() error {
	var totalSize int64
	
	filesDir := filepath.Join(m.cacheDir, "files")
	err := filepath.Walk(filesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	m.currentSize = totalSize
	return nil
}

// calculateHitRate calculates cache hit rate
func (m *Manager) calculateHitRate() float64 {
	// This is a placeholder - implement proper hit rate calculation
	return 0.0
}

// startCleanupRoutine starts the background cleanup routine
func (m *Manager) startCleanupRoutine() {
	ticker := time.NewTicker(m.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupExpired()
	}
}

// GetStats returns detailed cache statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := m.metadata.Items()
	var totalAccess int64
	var oldestEntry, newestEntry time.Time
	var hitCount, totalCount int64

	for _, item := range items {
		entry := item.Object.(*CacheEntry)
		totalAccess += entry.AccessCount
		totalCount++
		
		if oldestEntry.IsZero() || entry.CreatedAt.Before(oldestEntry) {
			oldestEntry = entry.CreatedAt
		}
		if newestEntry.IsZero() || entry.CreatedAt.After(newestEntry) {
			newestEntry = entry.CreatedAt
		}
		
		if entry.AccessCount > 1 {
			hitCount++
		}
	}

	hitRate := 0.0
	if totalCount > 0 {
		hitRate = float64(hitCount) / float64(totalCount)
	}

	return map[string]interface{}{
		"current_size":    m.currentSize,
		"max_size":        m.maxSize,
		"usage_percent":   float64(m.currentSize) / float64(m.maxSize) * 100,
		"item_count":      totalCount,
		"total_access":    totalAccess,
		"hit_rate":        hitRate,
		"oldest_entry":    oldestEntry,
		"newest_entry":    newestEntry,
		"ttl_hours":       m.ttl.Hours(),
		"cache_dir":       m.cacheDir,
	}
}

// cleanupExpired removes expired cache entries
func (m *Manager) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	items := m.metadata.Items()
	for key, item := range items {
		entry := item.Object.(*CacheEntry)
		
		// Check if file is expired
		if time.Since(entry.CreatedAt) > m.ttl {
			// Remove file
			if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
				m.logger.Warnf("Failed to remove expired cache file %s: %v", entry.FilePath, err)
			}
			
			// Remove from metadata
			m.metadata.Delete(key)
			m.currentSize -= entry.Size
			
			m.logger.Infof("Removed expired cache file: %s", entry.OriginalKey)
		}
	}
}