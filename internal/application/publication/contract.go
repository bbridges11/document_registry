package publication

import (
	"context"
)

// QueryService defines read operations for publications
type QueryService interface {
	// GetPublication retrieves a single publication by ID
	GetPublication(ctx context.Context, query GetPublicationQuery) (PublicationView, error)

	// ListPublicationsByVersion retrieves all publications for a specific version
	ListPublicationsByVersion(ctx context.Context, query ListPublicationsByVersionQuery) ([]PublicationView, error)

	// ListPublicationsByDocument retrieves all publications for a specific document
	ListPublicationsByDocument(ctx context.Context, query ListPublicationsByDocumentQuery) ([]PublicationView, error)

	// ListAllPublications retrieves all publications with pagination
	ListAllPublications(ctx context.Context, query ListAllPublicationsQuery) (PublicationListResult, error)
}
