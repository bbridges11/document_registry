package outbound

import (
	"context"
	"io"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
)

// StorageBackend is the interface each storage adapter must implement
// This represents a specific storage implementation (S3, filesystem, database, etc.)
type StorageBackend interface {
	// Upload stores content to this specific backend
	Upload(ctx context.Context, key string, content io.Reader) error

	// Download retrieves content from this specific backend
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes content from this specific backend
	Delete(ctx context.Context, key string) error

	// Exists checks if content exists in this specific backend
	Exists(ctx context.Context, key string) (bool, error)

	// GenerateUploadURL creates a presigned URL (returns "" if unsupported)
	GenerateUploadURL(ctx context.Context, key string) (string, error)

	// Type returns the backend type identifier
	Type() shared.StorageBackend
}
