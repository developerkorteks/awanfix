package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileMetadata represents file metadata stored locally
type FileMetadata struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	RemotePath string    `json:"remote_path"`
	UploadedAt time.Time `json:"uploaded_at"`
	Provider   string    `json:"provider"`
	Status     string    `json:"status"`
}

// MetadataManager manages file metadata
type MetadataManager struct {
	metadataDir string
	mu          sync.RWMutex
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(metadataDir string) (*MetadataManager, error) {
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, err
	}
	
	return &MetadataManager{
		metadataDir: metadataDir,
	}, nil
}

// Store stores file metadata
func (m *MetadataManager) Store(metadata *FileMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	filePath := filepath.Join(m.metadataDir, metadata.ID+".json")
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filePath, data, 0644)
}

// Get retrieves file metadata
func (m *MetadataManager) Get(fileID string) (*FileMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	filePath := filepath.Join(m.metadataDir, fileID+".json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var metadata FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	
	return &metadata, nil
}

// List lists all file metadata
func (m *MetadataManager) List() ([]*FileMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	files, err := filepath.Glob(filepath.Join(m.metadataDir, "*.json"))
	if err != nil {
		return nil, err
	}
	
	var metadataList []*FileMetadata
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		var metadata FileMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}
		
		metadataList = append(metadataList, &metadata)
	}
	
	return metadataList, nil
}

// Delete removes file metadata
func (m *MetadataManager) Delete(fileID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	filePath := filepath.Join(m.metadataDir, fileID+".json")
	return os.Remove(filePath)
}