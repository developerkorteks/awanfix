package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// UnionStorageImpl implements UnionStorage interface
type UnionStorageImpl struct {
	providers map[string]StorageProvider
	mu        sync.RWMutex
	logger    *logrus.Logger
}

// NewUnionStorage creates a new union storage
func NewUnionStorage() *UnionStorageImpl {
	return &UnionStorageImpl{
		providers: make(map[string]StorageProvider),
		logger:    logrus.New(),
	}
}

// AddProvider adds a storage provider to the union
func (u *UnionStorageImpl) AddProvider(provider StorageProvider) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	name := provider.Name()
	if _, exists := u.providers[name]; exists {
		return fmt.Errorf("provider %s already exists", name)
	}

	u.providers[name] = provider
	u.logger.Infof("Added storage provider: %s", name)
	
	return nil
}

// RemoveProvider removes a storage provider from the union
func (u *UnionStorageImpl) RemoveProvider(name string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, exists := u.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(u.providers, name)
	u.logger.Infof("Removed storage provider: %s", name)
	
	return nil
}

// GetProviders returns all providers
func (u *UnionStorageImpl) GetProviders() []StorageProvider {
	u.mu.RLock()
	defer u.mu.RUnlock()

	providers := make([]StorageProvider, 0, len(u.providers))
	for _, provider := range u.providers {
		providers = append(providers, provider)
	}
	
	return providers
}

// GetProvider gets a specific provider by name
func (u *UnionStorageImpl) GetProvider(name string) StorageProvider {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.providers[name]
}

// Name returns the union storage name
func (u *UnionStorageImpl) Name() string {
	return "union"
}

// Upload uploads a file to the best available provider
func (u *UnionStorageImpl) Upload(ctx context.Context, reader io.Reader, path string, opts UploadOptions) (*FileInfo, error) {
	provider := u.selectBestProvider(ctx)
	if provider == nil {
		return nil, fmt.Errorf("no available providers for upload")
	}

	u.logger.Infof("Uploading %s to provider %s", path, provider.Name())
	return provider.Upload(ctx, reader, path, opts)
}

// Download downloads a file from any available provider
func (u *UnionStorageImpl) Download(ctx context.Context, path string, opts DownloadOptions) (io.ReadCloser, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var lastErr error
	
	// Try each provider until we find the file
	for _, provider := range u.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		reader, err := provider.Download(ctx, path, opts)
		if err == nil {
			u.logger.Infof("Downloaded %s from provider %s", path, provider.Name())
			return reader, nil
		}
		
		lastErr = err
		u.logger.Debugf("Failed to download %s from provider %s: %v", path, provider.Name(), err)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to download from all providers: %w", lastErr)
	}
	
	return nil, fmt.Errorf("no available providers for download")
}

// List lists files from all providers
func (u *UnionStorageImpl) List(ctx context.Context, path string) ([]*FileInfo, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var allFiles []*FileInfo
	fileMap := make(map[string]*FileInfo) // Deduplicate by path

	for _, provider := range u.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		files, err := provider.List(ctx, path)
		if err != nil {
			u.logger.Warnf("Failed to list files from provider %s: %v", provider.Name(), err)
			continue
		}

		for _, file := range files {
			// Use the first occurrence of each file path
			if _, exists := fileMap[file.Path]; !exists {
				fileMap[file.Path] = file
				allFiles = append(allFiles, file)
			}
		}
	}

	return allFiles, nil
}

// Delete deletes a file from all providers that have it
func (u *UnionStorageImpl) Delete(ctx context.Context, path string) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var errors []error
	deleted := false

	for _, provider := range u.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		err := provider.Delete(ctx, path)
		if err == nil {
			deleted = true
			u.logger.Infof("Deleted %s from provider %s", path, provider.Name())
		} else {
			errors = append(errors, fmt.Errorf("provider %s: %w", provider.Name(), err))
		}
	}

	if !deleted && len(errors) > 0 {
		return fmt.Errorf("failed to delete from any provider: %v", errors)
	}

	return nil
}

// Stat gets file information from the first provider that has it
func (u *UnionStorageImpl) Stat(ctx context.Context, path string) (*FileInfo, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var lastErr error

	for _, provider := range u.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		info, err := provider.Stat(ctx, path)
		if err == nil {
			return info, nil
		}
		
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to stat from all providers: %w", lastErr)
	}
	
	return nil, fmt.Errorf("no available providers for stat")
}

// GetURL gets a direct download URL from the first provider that supports it
func (u *UnionStorageImpl) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var lastErr error

	for _, provider := range u.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		url, err := provider.GetURL(ctx, path, expires)
		if err == nil {
			return url, nil
		}
		
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to get URL from all providers: %w", lastErr)
	}
	
	return "", fmt.Errorf("no available providers support direct URLs")
}

// IsAvailable checks if any provider is available
func (u *UnionStorageImpl) IsAvailable(ctx context.Context) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, provider := range u.providers {
		if provider.IsAvailable(ctx) {
			return true
		}
	}
	
	return false
}

// selectBestProvider selects the best provider for upload based on availability and load
func (u *UnionStorageImpl) selectBestProvider(ctx context.Context) StorageProvider {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Simple round-robin selection for now
	// In production, you might want to consider:
	// - Provider availability
	// - Current load/usage
	// - Storage quotas
	// - Geographic location
	
	for _, provider := range u.providers {
		if provider.IsAvailable(ctx) {
			return provider
		}
	}
	
	return nil
}