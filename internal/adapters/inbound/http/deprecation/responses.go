package deprecation

import (
	"time"

	"github.com/google/uuid"
)

// DeprecationResponse represents a deprecation in HTTP responses
type DeprecationResponse struct {
	ID              uuid.UUID  `json:"id"`
	VersionID       uuid.UUID  `json:"version_id"`
	DocumentID      uuid.UUID  `json:"document_id"`
	RequestedBy     string     `json:"requested_by"`
	RequestedAt     time.Time  `json:"requested_at"`
	Reason          string     `json:"reason"`
	DeprecationNote string     `json:"deprecation_note,omitempty"`
	AutoDeprecated  bool       `json:"auto_deprecated"`
	Status          string     `json:"status"`
	DeprecatedBy    string     `json:"deprecated_by,omitempty"`
	DeprecatedAt    *time.Time `json:"deprecated_at,omitempty"`
	SupersededBy    *uuid.UUID `json:"superseded_by,omitempty"`
	PreviousStatus  string     `json:"previous_status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// DeprecationStatusResponse represents deprecation status with permission info
type DeprecationStatusResponse struct {
	Status     string `json:"status"`
	CanApprove bool   `json:"can_approve"`
	CanReject  bool   `json:"can_reject"`
	CanCancel  bool   `json:"can_cancel"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
