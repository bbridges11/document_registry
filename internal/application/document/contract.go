package document

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type Actor struct {
	UserID string
}

type CreateInput struct {
	Actor        Actor
	Name         string
	Description  string
	Tags         []string
	DocumentType string
	Version      string
	Content      io.Reader
	Metadata     map[string]any
}

type GetInput struct {
	Actor Actor
	ID    uuid.UUID
}

type ListInput struct {
	Actor  Actor
	Limit  int
	Offset int
}

type UpdateInput struct {
	Actor       Actor
	ID          uuid.UUID
	Name        string
	Description string
	Tags        []string
}

// SearchInput represents search criteria
type SearchInput struct {
	Actor Actor

	// Text search (searches name + description)
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

	// Special filters
	MyDocuments       bool // Documents created by actor
	MyStakeholderDocs bool // Documents where actor is stakeholder
	PendingMyReview   bool // Documents with versions pending actor's review
	PendingMyApproval bool // Documents with versions pending actor's approval
	RecentlyPublished bool // Documents with recently published versions

	// Sorting
	SortField string // "created_at", "updated_at", "name", "document_type"
	SortOrder string // "asc", "desc"

	// Pagination
	Page int
	Size int
}

// SearchOutput represents search results
type SearchOutput struct {
	Results    []DocumentSearchResult
	Pagination PaginationInfo
}

// DocumentSearchResult represents a single search result
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

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int
	Size       int
	Total      int
	TotalPages int
}

type DocumentView struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Tags         []string
	DocumentType string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
}

type VersionView struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Version     string
	Status      string
	ContentKey  string
	ContentHash string
	Metadata    map[string]any
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateOutput struct {
	Document         DocumentView
	Version          VersionView
	ValidationResult *ValidationResultView
}

type ValidationResultView struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

type ValidationIssue struct {
	Field    string `json:"field"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (CreateOutput, error)
	Get(ctx context.Context, input GetInput) (DocumentView, error)
	List(ctx context.Context, input ListInput) ([]DocumentView, error)
	Update(ctx context.Context, input UpdateInput) error
	Search(ctx context.Context, input SearchInput) (SearchOutput, error)
}
