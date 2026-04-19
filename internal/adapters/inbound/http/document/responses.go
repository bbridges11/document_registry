package document

import (
	"time"

	"github.com/google/uuid"
)

type DocumentResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	DocumentType string    `json:"document_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    string    `json:"created_by"`
}

type DocumentWithVersionResponse struct {
	ID               uuid.UUID             `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Tags             []string              `json:"tags"`
	DocumentType     string                `json:"document_type"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	CreatedBy        string                `json:"created_by"`
	Version          VersionResponse       `json:"version"`
	ValidationResult *ValidationResultView `json:"validation_result,omitempty"`
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

type VersionResponse struct {
	ID           uuid.UUID      `json:"id"`
	DocumentID   uuid.UUID      `json:"document_id"`
	Version      string         `json:"version"`
	Status       string         `json:"status"`
	ContentS3Key string         `json:"content_s3_key"`
	ContentHash  string         `json:"content_hash"`
	Metadata     map[string]any `json:"metadata"`
	CreatedBy    string         `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// SearchDocumentResponse represents search results
type SearchDocumentResponse struct {
	Results    []DocumentSearchResultResponse `json:"results"`
	Pagination PaginationResponse             `json:"pagination"`
}

// DocumentSearchResultResponse represents a single search result
type DocumentSearchResultResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	DocumentType        string    `json:"document_type"`
	Tags                []string  `json:"tags"`
	CreatedBy           string    `json:"created_by"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	LatestVersion       string    `json:"latest_version"`
	LatestVersionStatus string    `json:"latest_version_status"`
}

// PaginationResponse contains pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
