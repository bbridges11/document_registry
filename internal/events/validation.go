package events

import "time"

// Validation event types
const (
	EventValidationCompleted EventType = "validation.completed"
)

// ValidationCompleted event is published when content validation runs (success or failure)
type ValidationCompleted struct {
	documentID   string
	versionID    string
	documentType string
	valid        bool
	issueCount   int
	validatedBy  string
	occurredAt   time.Time
}

func NewValidationCompleted(
	documentID, versionID, documentType string,
	valid bool,
	issueCount int,
	validatedBy string,
) *ValidationCompleted {
	return &ValidationCompleted{
		documentID:   documentID,
		versionID:    versionID,
		documentType: documentType,
		valid:        valid,
		issueCount:   issueCount,
		validatedBy:  validatedBy,
		occurredAt:   time.Now().UTC(),
	}
}

func (e *ValidationCompleted) Type() EventType       { return EventValidationCompleted }
func (e *ValidationCompleted) OccurredAt() time.Time { return e.occurredAt }
func (e *ValidationCompleted) DocumentID() string    { return e.documentID }
func (e *ValidationCompleted) VersionID() string     { return e.versionID }
func (e *ValidationCompleted) DocumentType() string  { return e.documentType }
func (e *ValidationCompleted) Valid() bool           { return e.valid }
func (e *ValidationCompleted) IssueCount() int       { return e.issueCount }
func (e *ValidationCompleted) ValidatedBy() string   { return e.validatedBy }
