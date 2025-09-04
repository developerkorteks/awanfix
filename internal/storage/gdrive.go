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

// GDriveProvider implements StorageProvider for Google Drive
type GDriveProvider struct {
	name       string
	remoteName string
	rcloneBin  string
	configPath string
	logger     *logrus.Logger
}

// NewGDriveProvider creates a new Google Drive storage provider
func NewGDriveProvider(name, remoteName, rcloneBin, configPath string) *GDriveProvider {
	return &GDriveProvider{
		name:       name,
		remoteName: remoteName,
		rcloneBin:  rcloneBin,
		configPath: configPath,
		logger:     logrus.New(),
	}
}

// Name returns the provider name
func (g *GDriveProvider) Name() string {
	return g.name
}

// Upload uploads a file to Google Drive
func (g *GDriveProvider) Upload(ctx context.Context, reader io.Reader, path string, opts UploadOptions) (*FileInfo, error) {
	// Create temporary file for upload
	tempFile := filepath.Join("/tmp", fmt.Sprintf("gdrive_upload_%s_%s", uuid.New().String(), opts.Filename))
	
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	// Execute rclone copy command
	cmd := g.buildRcloneCmd("copy", tempFile, remotePath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to upload to Google Drive: %w", err)
	}
	
	// Get file info after upload
	return g.Stat(ctx, path)
}

// Download downloads a file from Google Drive
func (g *GDriveProvider) Download(ctx context.Context, path string, opts DownloadOptions) (io.ReadCloser, error) {
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	// For range requests, handle differently
	if opts.Range != nil {
		return g.downloadWithRange(ctx, remotePath, opts.Range)
	}
	
	// Execute rclone cat command to stream file content
	cmd := g.buildRcloneCmd("cat", remotePath)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rclone cat: %w", err)
	}
	
	return &cmdReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}

// List lists files in Google Drive directory
func (g *GDriveProvider) List(ctx context.Context, path string) ([]*FileInfo, error) {
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	// Execute rclone lsjson command
	cmd := g.buildRcloneCmd("lsjson", remotePath)
	
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files from Google Drive: %w", err)
	}
	
	// Parse JSON output and convert to FileInfo
	var files []*FileInfo
	// TODO: Parse JSON output properly
	
	return files, nil
}

// Delete deletes a file from Google Drive
func (g *GDriveProvider) Delete(ctx context.Context, path string) error {
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	cmd := g.buildRcloneCmd("delete", remotePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete file from Google Drive: %w", err)
	}
	
	return nil
}

// Stat gets file information from Google Drive
func (g *GDriveProvider) Stat(ctx context.Context, path string) (*FileInfo, error) {
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	// Execute rclone lsjson for single file
	cmd := g.buildRcloneCmd("lsjson", remotePath)
	
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Parse output and return FileInfo
	return &FileInfo{
		ID:       uuid.New().String(),
		Name:     filepath.Base(path),
		Path:     path,
		Provider: g.name,
		ModTime:  time.Now(),
	}, nil
}

// GetURL gets a direct download URL from Google Drive
func (g *GDriveProvider) GetURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	// Google Drive supports direct links via rclone link command
	remotePath := fmt.Sprintf("%s:%s", g.remoteName, path)
	
	cmd := g.buildRcloneCmd("link", remotePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Google Drive link: %w", err)
	}
	
	return string(output), nil
}

// IsAvailable checks if Google Drive provider is available
func (g *GDriveProvider) IsAvailable(ctx context.Context) bool {
	// Test connection by listing root directory
	cmd := g.buildRcloneCmd("lsd", fmt.Sprintf("%s:", g.remoteName))
	err := cmd.Run()
	return err == nil
}

// buildRcloneCmd builds an rclone command with proper configuration
func (g *GDriveProvider) buildRcloneCmd(operation string, args ...string) *exec.Cmd {
	cmdArgs := []string{operation}
	cmdArgs = append(cmdArgs, args...)
	
	cmd := exec.Command(g.rcloneBin, cmdArgs...)
	
	// Set config path if provided
	if g.configPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RCLONE_CONFIG=%s", g.configPath))
	}
	
	return cmd
}

// downloadWithRange handles HTTP range requests for Google Drive
func (g *GDriveProvider) downloadWithRange(ctx context.Context, remotePath string, rangeSpec *RangeSpec) (io.ReadCloser, error) {
	// Google Drive supports range requests better than Mega
	cmd := g.buildRcloneCmd("cat", remotePath)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rclone cat: %w", err)
	}
	
	// TODO: Implement proper range handling for Google Drive
	
	return &cmdReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}