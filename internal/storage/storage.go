package storage

import (
	"context"
	"io"
)

// Storage defines the interface for file storage
type Storage interface {
	// Save saves a file and returns the storage key
	Save(ctx context.Context, reader io.Reader, filename string) (string, error)

	// Get retrieves a file by key
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete deletes a file by key
	Delete(ctx context.Context, key string) error

	// GetURL returns a public URL for the file (if applicable)
	GetURL(key string) string
}
