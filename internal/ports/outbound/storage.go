package outbound

import (
	"context"
	"io"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
)

// StorageRequest encapsulates what to store and where
type StorageRequest struct {
	Backend     shared.StorageBackend
	Key         string // backend-specific identifier
	Content     io.Reader
	ContentType string            // optional MIME type
	Metadata    map[string]string // backend-specific metadata
}

// StorageService handles content storage across multiple backends
// Implementation is delegated to backend-specific adapters via a router
type StorageService interface {
	// Upload stores content and returns a ContentReference
	Upload(ctx context.Context, req StorageRequest) (*shared.ContentReference, error)

	// Download retrieves content using a ContentReference
	Download(ctx context.Context, ref *shared.ContentReference) (io.ReadCloser, error)

	// Delete removes content using a ContentReference
	Delete(ctx context.Context, ref *shared.ContentReference) error

	// Exists checks if content exists for a ContentReference
	Exists(ctx context.Context, ref *shared.ContentReference) (bool, error)

	// GenerateUploadURL creates a presigned URL (if backend supports it)
	// Returns empty string if backend doesn't support presigned URLs
	GenerateUploadURL(ctx context.Context, req StorageRequest) (string, error)
}
