package deprecation

import (
	"time"

	"github.com/google/uuid"
)

// DeprecationOutput represents a deprecation view
type DeprecationOutput struct {
	ID              uuid.UUID
	VersionID       uuid.UUID
	DocumentID      uuid.UUID
	RequestedBy     string
	RequestedAt     time.Time
	Reason          string
	DeprecationNote string
	AutoDeprecated  bool
	Status          string
	DeprecatedBy    string
	DeprecatedAt    *time.Time
	SupersededBy    *uuid.UUID
	PreviousStatus  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DeprecationStatusOutput represents deprecation status with permission info
type DeprecationStatusOutput struct {
	Status     string
	CanApprove bool
	CanReject  bool
	CanCancel  bool
}
