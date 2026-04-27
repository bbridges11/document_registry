package shared

import (
	"fmt"

	"github.com/bbridges_11/document-registry/pkg/errors"
)

// StorageBackend identifies the storage system
type StorageBackend string

const (
	StorageBackendS3 StorageBackend = "s3"
)

// ContentReference represents where content is stored
// This is a value object - immutable and self-validating
type ContentReference struct {
	backend  StorageBackend
	location string // backend-specific identifier
}

// NewContentReference creates a validated content reference
func NewContentReference(backend StorageBackend, location string) (*ContentReference, error) {
	if location == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "location is required")
	}

	if !isValidBackend(backend) {
		return nil, errors.New(errors.CodeInvalidArgument,
			fmt.Sprintf("invalid storage backend: %s", backend))
	}

	return &ContentReference{
		backend:  backend,
		location: location,
	}, nil
}

// MustNewContentReference creates a content reference or panics
// Use only when input is guaranteed valid (e.g., rehydration from DB)
func MustNewContentReference(backend StorageBackend, location string) *ContentReference {
	ref, err := NewContentReference(backend, location)
	if err != nil {
		panic(fmt.Sprintf("invalid content reference: %v", err))
	}
	return ref
}

// Backend returns the storage backend type
func (c *ContentReference) Backend() StorageBackend {
	return c.backend
}

// Location returns the storage-specific location identifier
func (c *ContentReference) Location() string {
	return c.location
}

// String returns a human-readable representation
func (c *ContentReference) String() string {
	return fmt.Sprintf("%s://%s", c.backend, c.location)
}

// Equals compares two content references
func (c *ContentReference) Equals(other *ContentReference) bool {
	if other == nil {
		return false
	}
	return c.backend == other.backend && c.location == other.location
}

func isValidBackend(backend StorageBackend) bool {
	switch backend {
	case StorageBackendS3:
		return true
	default:
		return false
	}
}
