package publication

import (
	"time"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

// PublicationStatus represents the status of a publication
type PublicationStatus string

const (
	PublicationStatusSuccess PublicationStatus = "success"
	PublicationStatusFailed  PublicationStatus = "failed"
)

// IsValid checks if the publication status is valid
func (s PublicationStatus) IsValid() bool {
	return s == PublicationStatusSuccess || s == PublicationStatusFailed
}

// Publication represents a record of publishing a version to an external destination
type Publication struct {
	id           uuid.UUID
	versionID    uuid.UUID
	documentID   uuid.UUID
	publishedTo  string
	publishedBy  string
	publishedAt  time.Time
	status       PublicationStatus
	errorMessage string
}

// NewPublication creates a new publication record
func NewPublication(
	versionID uuid.UUID,
	documentID uuid.UUID,
	publishedTo string,
	publishedBy string,
) (*Publication, error) {
	if versionID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "version ID is required")
	}
	if documentID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "document ID is required")
	}
	if publishedTo == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "publishedTo is required")
	}
	if publishedBy == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "publishedBy is required")
	}

	return &Publication{
		id:          uuid.New(),
		versionID:   versionID,
		documentID:  documentID,
		publishedTo: publishedTo,
		publishedBy: publishedBy,
		publishedAt: time.Now().UTC(),
		status:      PublicationStatusSuccess, // Default to success, call MarkAsFailed if needed
	}, nil
}

// RehydratePublication reconstructs a publication from persistence
func RehydratePublication(
	id uuid.UUID,
	versionID uuid.UUID,
	documentID uuid.UUID,
	publishedTo string,
	publishedBy string,
	publishedAt time.Time,
	status PublicationStatus,
	errorMessage string,
) *Publication {
	return &Publication{
		id:           id,
		versionID:    versionID,
		documentID:   documentID,
		publishedTo:  publishedTo,
		publishedBy:  publishedBy,
		publishedAt:  publishedAt,
		status:       status,
		errorMessage: errorMessage,
	}
}

// MarkAsFailed marks the publication as failed with an error message
func (p *Publication) MarkAsFailed(errorMessage string) {
	p.status = PublicationStatusFailed
	p.errorMessage = errorMessage
}

// ID returns the publication ID
func (p *Publication) ID() uuid.UUID {
	return p.id
}

// VersionID returns the version ID
func (p *Publication) VersionID() uuid.UUID {
	return p.versionID
}

// DocumentID returns the document ID
func (p *Publication) DocumentID() uuid.UUID {
	return p.documentID
}

// PublishedTo returns the destination where this was published
func (p *Publication) PublishedTo() string {
	return p.publishedTo
}

// PublishedBy returns the user ID who triggered the publication
func (p *Publication) PublishedBy() string {
	return p.publishedBy
}

// PublishedAt returns when the publication occurred
func (p *Publication) PublishedAt() time.Time {
	return p.publishedAt
}

// Status returns the publication status
func (p *Publication) Status() PublicationStatus {
	return p.status
}

// ErrorMessage returns the error message if publication failed
func (p *Publication) ErrorMessage() string {
	return p.errorMessage
}
