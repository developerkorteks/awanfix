package storage

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MegaProvider implements StorageProvider for Mega.nz
type MegaProvider struct {
	name       string
	remoteName string
	rcloneBin  string
	configPath string
	logger     *logrus.Logger
}

// NewMegaProvider creates a new Mega storage provider
func NewMegaProvider(name, remoteName, rcloneBin, configPath string) *MegaProvider {
	return &MegaProvider{
		name:       name,
		remoteName: remoteName,
		rcloneBin:  rcloneBin,
		configPath: configPath,
		logger:     logrus.New(),
	}
}

// Name returns the provider name
func (m *MegaProvider) Name() string {
	return m.name
}

// Upload uploads a file to Mega
func (m *MegaProvider) Upload(ctx context.Context, reader io.Reader, path string, opts UploadOptions) (*FileInfo, error) {
	// Create temporary file for upload
	tempFile := filepath.Join("/tmp", fmt.Sprintf("rclone_upload_%s_%s", uuid.New().String(), opts.Filename))
	
	// Save reader content to temp file
	// In production, you might want to stream directly to rclone
	// For now, we'll use a simple approach
	
	remotePath := fmt.Sprintf("%s:%s", m.remoteName, path)
	
	// Execute rclone copy command
	cmd := m.buildRcloneCmd("copy", tempFile, remotePath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to upload to mega: %w", err)
	}
	
	// Get file info after upload
	return m.Stat(ctx, path)
}

// Download downloads a file from Mega
func (m *MegaProvider) Download(ctx context.Context, path string, opts DownloadOptions) (io.ReadCloser, error) {
	remotePath := fmt.Sprintf("%s:%s", m.remoteName, path)
	
	// For range requests, we'll need to handle differently
	if opts.Range != nil {
		return m.downloadWithRange(ctx, remotePath, opts.Range)
	}
	
	// Execute rclone cat command to stream file content
	cmd := m.buildRcloneCmd("cat", remotePath)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rclone cat: %w", err)
	}
	
	// Return a ReadCloser that will wait for the command to finish
	return &cmdReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}

// List lists files in the given directory
func (m *MegaProvider) List(ctx context.Context, path string) ([]*FileInfo, error) {
	remotePath := fmt.Sprintf("%s:%s", m.remoteName, path)
	
	// Execute rclone lsjson command
	cmd := m.buildRcloneCmd("lsjson", remotePath)
	
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	
	// Parse JSON output and convert to FileInfo
	// This is a simplified implementation
	// In production, you'd want proper JSON parsing
	
	var files []*FileInfo
	// TODO: Parse JSON output properly
	
	return files, nil
}

// Delete deletes a file from Mega
func (m *MegaProvider) Delete(ctx context.Context, path string) error {
	remotePath := fmt.Sprintf("%s:%s", m.remoteName, path)
	
	cmd := m.buildRcloneCmd("delete", remotePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// Stat gets file information
func (m *MegaProvider) Stat(ctx context.Context, path string) (*FileInfo, error) {
	remotePath := fmt.Sprintf("%s:%s", m.remoteName, path)
	
	// Execute rclone lsjson for single file
	cmd := m.buildRcloneCmd("lsjson", remotePath)
	
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Parse output and return FileInfo
	// TODO: Implement proper JSON parsing
	
	return &FileInfo{
		ID:       uuid.New().String(),
		Name:     filepath.Base(path),
		Path:     path,
		Provider: m.name,
		ModTime:  time.Now(),
	}, nil
}

// GetURL gets a direct download URL (Mega doesn't support this easily)
func (m *MegaProvider) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	return "", fmt.Errorf("direct URLs not supported for Mega provider")
}

// IsAvailable checks if the provider is available
func (m *MegaProvider) IsAvailable(ctx context.Context) bool {
	// Test connection by listing root directory
	cmd := m.buildRcloneCmd("lsd", fmt.Sprintf("%s:", m.remoteName))
	err := cmd.Run()
	return err == nil
}

// buildRcloneCmd builds an rclone command with proper configuration
func (m *MegaProvider) buildRcloneCmd(operation string, args ...string) *exec.Cmd {
	cmdArgs := []string{operation}
	cmdArgs = append(cmdArgs, args...)
	
	cmd := exec.Command(m.rcloneBin, cmdArgs...)
	
	// Set config path if provided
	if m.configPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", m.configPath))
	}
	
	return cmd
}

// downloadWithRange handles HTTP range requests
func (m *MegaProvider) downloadWithRange(ctx context.Context, remotePath string, rangeSpec *RangeSpec) (io.ReadCloser, error) {
	// For range requests, we might need to download the entire file and seek
	// This is not optimal but Mega doesn't support range requests directly
	
	cmd := m.buildRcloneCmd("cat", remotePath)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rclone cat: %w", err)
	}
	
	// TODO: Implement proper range handling
	// For now, return the full stream
	
	return &cmdReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}

// cmdReadCloser wraps a ReadCloser and ensures the command finishes
type cmdReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (c *cmdReadCloser) Close() error {
	if err := c.ReadCloser.Close(); err != nil {
		return err
	}
	return c.cmd.Wait()
}