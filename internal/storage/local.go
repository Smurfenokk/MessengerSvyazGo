package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements Storage for local filesystem
type LocalStorage struct {
	basePath string
	publicURL string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(basePath, publicURL string) *LocalStorage {
	return &LocalStorage{
		basePath:  basePath,
		publicURL: strings.TrimSuffix(publicURL, "/"),
	}
}

// Save saves a file to local filesystem
func (ls *LocalStorage) Save(ctx context.Context, reader io.Reader, filename string) (string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(ls.basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique key (simple version - use UUID in production)
	key := filename
	if _, err := os.Stat(filepath.Join(ls.basePath, key)); err == nil {
		// File exists, append timestamp
		key = fmt.Sprintf("%d_%s", ctx.Value("timestamp"), filename)
	}

	// Create file
	filePath := filepath.Join(ls.basePath, key)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return key, nil
}

// Get retrieves a file from local filesystem
func (ls *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	filePath := filepath.Join(ls.basePath, key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete deletes a file from local filesystem
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(ls.basePath, key)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL returns a public URL for the file
func (ls *LocalStorage) GetURL(key string) string {
	return fmt.Sprintf("%s/%s", ls.publicURL, key)
}
