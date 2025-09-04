package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo represents metadata about a file
type FileInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	IsDir    bool      `json:"is_dir"`
	MimeType string    `json:"mime_type"`
	Provider string    `json:"provider"`
	Path     string    `json:"path"`
}

// UploadOptions contains options for uploading files
type UploadOptions struct {
	Filename    string
	ContentType string
	Overwrite   bool
}

// DownloadOptions contains options for downloading files
type DownloadOptions struct {
	Range *RangeSpec
}

// RangeSpec represents HTTP range request
type RangeSpec struct {
	Start int64
	End   int64
}

// StorageProvider defines the interface for cloud storage providers
type StorageProvider interface {
	// Upload uploads a file to the storage provider
	Upload(ctx context.Context, reader io.Reader, path string, opts UploadOptions) (*FileInfo, error)
	
	// Download downloads a file from the storage provider
	Download(ctx context.Context, path string, opts DownloadOptions) (io.ReadCloser, error)
	
	// List lists files in the given directory
	List(ctx context.Context, path string) ([]*FileInfo, error)
	
	// Delete deletes a file from the storage provider
	Delete(ctx context.Context, path string) error
	
	// Stat gets file information
	Stat(ctx context.Context, path string) (*FileInfo, error)
	
	// GetURL gets a direct download URL (if supported)
	GetURL(ctx context.Context, path string, expires time.Duration) (string, error)
	
	// Name returns the provider name
	Name() string
	
	// IsAvailable checks if the provider is available
	IsAvailable(ctx context.Context) bool
}

// UnionStorage combines multiple storage providers
type UnionStorage interface {
	StorageProvider
	
	// AddProvider adds a storage provider to the union
	AddProvider(provider StorageProvider) error
	
	// RemoveProvider removes a storage provider from the union
	RemoveProvider(name string) error
	
	// GetProviders returns all providers
	GetProviders() []StorageProvider
	
	// GetProvider gets a specific provider by name
	GetProvider(name string) StorageProvider
}