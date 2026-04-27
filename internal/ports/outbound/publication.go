package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/publication"
	"github.com/google/uuid"
)

// PublicationRepository defines operations for persisting publications
type PublicationRepository interface {
	// Save persists a publication
	Save(ctx context.Context, pub *publication.Publication) error

	// GetByID retrieves a publication by ID
	GetByID(ctx context.Context, id uuid.UUID) (*publication.Publication, error)

	// ListByVersionID retrieves all publications for a version
	ListByVersionID(ctx context.Context, versionID uuid.UUID) ([]*publication.Publication, error)

	// ListByDocumentID retrieves all publications for a document
	ListByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*publication.Publication, error)
}
