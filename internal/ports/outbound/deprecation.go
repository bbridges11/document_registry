package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
	"github.com/google/uuid"
)

// DeprecationRepository defines the contract for deprecation persistence
type DeprecationRepository interface {
	// Save persists a deprecation entity
	Save(ctx context.Context, dep *deprecation.Deprecation) error

	// GetByID retrieves a deprecation by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*deprecation.Deprecation, error)

	// GetByVersionID retrieves the most recent deprecation for a version
	GetByVersionID(ctx context.Context, versionID uuid.UUID) (*deprecation.Deprecation, error)

	// GetPendingByVersionID retrieves a pending deprecation for a version (if exists)
	GetPendingByVersionID(ctx context.Context, versionID uuid.UUID) (*deprecation.Deprecation, error)

	// ListByDocumentID retrieves all deprecations for a document
	ListByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*deprecation.Deprecation, error)
}
