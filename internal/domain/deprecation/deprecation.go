package deprecation

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

// Deprecation represents a version deprecation request
type Deprecation struct {
	shared.AggregateRoot
	id              uuid.UUID
	versionID       uuid.UUID
	documentID      uuid.UUID
	requestedBy     string
	requestedAt     time.Time
	reason          string
	deprecationNote string
	autoDeprecated  bool
	status          DeprecationStatus
	deprecatedBy    string
	deprecatedAt    *time.Time
	supersededBy    *uuid.UUID
	previousStatus  string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewDeprecation creates a new manual deprecation request
func NewDeprecation(
	versionID uuid.UUID,
	documentID uuid.UUID,
	requestedBy string,
	reason string,
	deprecationNote string,
	previousStatus string,
) (*Deprecation, error) {
	if versionID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "version ID is required")
	}
	if documentID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "document ID is required")
	}
	if requestedBy == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "requested by is required")
	}
	if reason == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "reason is required")
	}
	if previousStatus == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "previous status is required")
	}

	now := time.Now()
	dep := &Deprecation{
		id:              uuid.New(),
		versionID:       versionID,
		documentID:      documentID,
		requestedBy:     requestedBy,
		requestedAt:     now,
		reason:          reason,
		deprecationNote: deprecationNote,
		autoDeprecated:  false,
		status:          DeprecationStatusPending,
		previousStatus:  previousStatus,
		createdAt:       now,
		updatedAt:       now,
	}

	// Emit deprecation requested event
	dep.AggregateRoot.AddEvent(events.DeprecationRequested{
		DeprecationID: dep.id,
		VersionID:     dep.versionID,
		DocumentID:    dep.documentID,
		RequestedBy:   dep.requestedBy,
		Reason:        dep.reason,
		Timestamp:     now,
	})

	return dep, nil
}

// NewAutoDeprecation creates a new auto-deprecation (happens when newer version published)
func NewAutoDeprecation(
	versionID uuid.UUID,
	documentID uuid.UUID,
	supersededBy uuid.UUID,
	reason string,
	deprecatedBy string,
) (*Deprecation, error) {
	if versionID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "version ID is required")
	}
	if documentID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "document ID is required")
	}
	if supersededBy == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "superseded by version is required")
	}
	if deprecatedBy == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "deprecated by is required")
	}

	now := time.Now()
	dep := &Deprecation{
		id:             uuid.New(),
		versionID:      versionID,
		documentID:     documentID,
		requestedBy:    deprecatedBy,
		requestedAt:    now,
		reason:         reason,
		autoDeprecated: true,
		status:         DeprecationStatusApproved,
		deprecatedBy:   deprecatedBy,
		deprecatedAt:   &now,
		supersededBy:   &supersededBy,
		previousStatus: "PUBLISHED", // Auto-deprecation only happens to PUBLISHED versions
		createdAt:      now,
		updatedAt:      now,
	}

	// Emit auto-deprecation event
	dep.AggregateRoot.AddEvent(events.AutoDeprecationTriggered{
		VersionID:    dep.versionID,
		DocumentID:   dep.documentID,
		SupersededBy: supersededBy,
		Timestamp:    now,
	})

	return dep, nil
}

// Approve marks the deprecation as approved
func (d *Deprecation) Approve(approvedBy string) error {
	if !d.status.IsPending() {
		return errors.New(errors.CodeDeprecationNotPending, "deprecation is not pending")
	}
	if approvedBy == "" {
		return errors.New(errors.CodeInvalidArgument, "approved by is required")
	}

	d.status = DeprecationStatusApproved
	d.updatedAt = time.Now()

	d.AggregateRoot.AddEvent(events.DeprecationApproved{
		DeprecationID: d.id,
		VersionID:     d.versionID,
		DocumentID:    d.documentID,
		Timestamp:     d.updatedAt,
	})

	return nil
}

// Reject marks the deprecation as rejected
func (d *Deprecation) Reject(rejectedBy string, reason string) error {
	if !d.status.IsPending() {
		return errors.New(errors.CodeDeprecationNotPending, "deprecation is not pending")
	}
	if rejectedBy == "" {
		return errors.New(errors.CodeInvalidArgument, "rejected by is required")
	}

	d.status = DeprecationStatusRejected
	d.updatedAt = time.Now()

	d.AggregateRoot.AddEvent(events.DeprecationRejected{
		DeprecationID: d.id,
		VersionID:     d.versionID,
		DocumentID:    d.documentID,
		RejectedBy:    rejectedBy,
		Reason:        reason,
		Timestamp:     d.updatedAt,
	})

	return nil
}

// Cancel marks the deprecation as canceled
func (d *Deprecation) Cancel(canceledBy string) error {
	if !d.status.IsPending() {
		return errors.New(errors.CodeDeprecationNotPending, "deprecation is not pending")
	}
	if canceledBy == "" {
		return errors.New(errors.CodeInvalidArgument, "canceled by is required")
	}

	d.status = DeprecationStatusCanceled
	d.updatedAt = time.Now()

	d.AggregateRoot.AddEvent(events.DeprecationCanceled{
		DeprecationID: d.id,
		VersionID:     d.versionID,
		DocumentID:    d.documentID,
		CanceledBy:    canceledBy,
		Timestamp:     d.updatedAt,
	})

	return nil
}

// Complete finalizes the deprecation
func (d *Deprecation) Complete(completedBy string) error {
	if !d.status.IsApproved() {
		return errors.New(errors.CodeInvalidArgument, "can only complete approved deprecation")
	}
	if completedBy == "" {
		return errors.New(errors.CodeInvalidArgument, "completed by is required")
	}

	now := time.Now()
	d.deprecatedBy = completedBy
	d.deprecatedAt = &now
	d.updatedAt = now

	d.AggregateRoot.AddEvent(events.VersionDeprecated{
		VersionID:      d.versionID,
		DocumentID:     d.documentID,
		DeprecatedBy:   completedBy,
		Reason:         d.reason,
		AutoDeprecated: d.autoDeprecated,
		SupersededBy:   d.supersededBy,
		Timestamp:      now,
	})

	return nil
}

// Getters
func (d *Deprecation) ID() uuid.UUID             { return d.id }
func (d *Deprecation) VersionID() uuid.UUID      { return d.versionID }
func (d *Deprecation) DocumentID() uuid.UUID     { return d.documentID }
func (d *Deprecation) RequestedBy() string       { return d.requestedBy }
func (d *Deprecation) RequestedAt() time.Time    { return d.requestedAt }
func (d *Deprecation) Reason() string            { return d.reason }
func (d *Deprecation) DeprecationNote() string   { return d.deprecationNote }
func (d *Deprecation) AutoDeprecated() bool      { return d.autoDeprecated }
func (d *Deprecation) Status() DeprecationStatus { return d.status }
func (d *Deprecation) DeprecatedBy() string      { return d.deprecatedBy }
func (d *Deprecation) DeprecatedAt() *time.Time  { return d.deprecatedAt }
func (d *Deprecation) SupersededBy() *uuid.UUID  { return d.supersededBy }
func (d *Deprecation) PreviousStatus() string    { return d.previousStatus }
func (d *Deprecation) CreatedAt() time.Time      { return d.createdAt }
func (d *Deprecation) UpdatedAt() time.Time      { return d.updatedAt }

// RehydrateDeprecation reconstructs a deprecation from persistence
func RehydrateDeprecation(
	id uuid.UUID,
	versionID uuid.UUID,
	documentID uuid.UUID,
	requestedBy string,
	requestedAt time.Time,
	reason string,
	deprecationNote string,
	autoDeprecated bool,
	status DeprecationStatus,
	deprecatedBy string,
	deprecatedAt *time.Time,
	supersededBy *uuid.UUID,
	previousStatus string,
	createdAt time.Time,
	updatedAt time.Time,
) *Deprecation {
	return &Deprecation{
		id:              id,
		versionID:       versionID,
		documentID:      documentID,
		requestedBy:     requestedBy,
		requestedAt:     requestedAt,
		reason:          reason,
		deprecationNote: deprecationNote,
		autoDeprecated:  autoDeprecated,
		status:          status,
		deprecatedBy:    deprecatedBy,
		deprecatedAt:    deprecatedAt,
		supersededBy:    supersededBy,
		previousStatus:  previousStatus,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}
