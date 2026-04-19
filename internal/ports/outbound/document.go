package outbound

import (
	"context"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/document"
	"github.com/google/uuid"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *document.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*document.Document, error)
	List(ctx context.Context, limit, offset int) ([]*document.Document, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	Search(ctx context.Context, filters DocumentSearchFilters) ([]DocumentSearchResult, int, error)
}

// DocumentSearchFilters encapsulates all search criteria
type DocumentSearchFilters struct {
	// Text search (name, description)
	Query string

	// Field filters
	Name         string
	Description  string
	Tags         []string
	DocumentType string
	CreatedBy    string

	// Date filters
	CreatedAfter  *time.Time
	CreatedBefore *time.Time

	// ID filters (for authorization scoping)
	DocumentIDs []uuid.UUID

	// Special filters (these will be implemented with subqueries)
	UserID            string // For my_documents, my_stakeholder_docs, etc.
	MyDocuments       bool
	MyStakeholderDocs bool
	PendingMyReview   bool
	PendingMyApproval bool
	RecentlyPublished bool

	// Sorting
	SortField string
	SortOrder string

	// Pagination
	Offset int
	Limit  int
}

// DocumentSearchResult is the repository-level result
type DocumentSearchResult struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	DocumentType        string
	Tags                []string
	CreatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LatestVersion       string
	LatestVersionStatus string
}
