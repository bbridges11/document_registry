package document

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type CreateDocumentWithVersionCommand struct {
	// Document fields
	Name         string
	Description  string
	Tags         []string
	DocumentType string

	// Version fields
	Version     string
	Content     io.Reader
	ContentHash string
	Metadata    map[string]any

	CreatedBy string
}

type UpdateDocumentCommand struct {
	ID          uuid.UUID
	Name        string
	Description string
	Tags        []string
	UpdatedBy   string
}

type GetDocumentQuery struct {
	ID uuid.UUID
}

type ListDocumentsQuery struct {
	Limit  int
	Offset int
}

type DocumentDTO struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Tags         []string
	DocumentType string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
}

type DocumentWithVersionDTO struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Tags         []string
	DocumentType string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string

	// Embedded version
	Version VersionDTO
}

type VersionDTO struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Version     string
	Status      string
	ContentKey  string // Generic key - abstracted from storage backend
	ContentHash string
	Metadata    map[string]any
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
