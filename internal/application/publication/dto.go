package publication

import (
	"time"

	"github.com/google/uuid"
)

// GetPublicationQuery retrieves a single publication
type GetPublicationQuery struct {
	ID uuid.UUID
}

// ListPublicationsByVersionQuery retrieves publications for a version
type ListPublicationsByVersionQuery struct {
	VersionID uuid.UUID
}

// ListPublicationsByDocumentQuery retrieves publications for a document
type ListPublicationsByDocumentQuery struct {
	DocumentID uuid.UUID
}

// ListAllPublicationsQuery retrieves all publications with pagination
type ListAllPublicationsQuery struct {
	Limit  int
	Offset int
}

// PublicationView is a read model for publication data
type PublicationView struct {
	ID           uuid.UUID `json:"id"`
	VersionID    uuid.UUID `json:"version_id"`
	DocumentID   uuid.UUID `json:"document_id"`
	PublishedTo  string    `json:"published_to"`
	PublishedBy  string    `json:"published_by"`
	PublishedAt  time.Time `json:"published_at"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// PublicationListResult contains paginated publication results
type PublicationListResult struct {
	Publications []PublicationView `json:"publications"`
	Total        int               `json:"total"`
	Limit        int               `json:"limit"`
	Offset       int               `json:"offset"`
}
