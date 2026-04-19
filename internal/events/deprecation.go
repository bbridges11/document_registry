package events

import (
	"time"

	"github.com/google/uuid"
)

// DeprecationRequested is emitted when a manual deprecation request is created
type DeprecationRequested struct {
	DeprecationID uuid.UUID
	VersionID     uuid.UUID
	DocumentID    uuid.UUID
	RequestedBy   string
	Reason        string
	Timestamp     time.Time
}

func (e DeprecationRequested) Type() EventType {
	return EventDeprecationRequested
}

func (e DeprecationRequested) OccurredAt() time.Time {
	return e.Timestamp
}

// VersionLocked is emitted when a version enters DEPRECATING status
type VersionLocked struct {
	VersionID      uuid.UUID
	DocumentID     uuid.UUID
	PreviousStatus string
	Timestamp      time.Time
}

func NewVersionLocked(versionID, documentID, previousStatus string) VersionLocked {
	vid, _ := uuid.Parse(versionID)
	did, _ := uuid.Parse(documentID)
	return VersionLocked{
		VersionID:      vid,
		DocumentID:     did,
		PreviousStatus: previousStatus,
		Timestamp:      time.Now().UTC(),
	}
}

func (e VersionLocked) Type() EventType {
	return EventVersionLocked
}

func (e VersionLocked) OccurredAt() time.Time {
	return e.Timestamp
}

// DeprecationApproved is emitted when all required approvals are received
type DeprecationApproved struct {
	DeprecationID uuid.UUID
	VersionID     uuid.UUID
	DocumentID    uuid.UUID
	Timestamp     time.Time
}

func (e DeprecationApproved) Type() EventType {
	return EventDeprecationApproved
}

func (e DeprecationApproved) OccurredAt() time.Time {
	return e.Timestamp
}

// VersionDeprecated is emitted when a version is finalized as deprecated
type VersionDeprecated struct {
	VersionID      uuid.UUID
	DocumentID     uuid.UUID
	DeprecatedBy   string
	Reason         string
	AutoDeprecated bool
	SupersededBy   *uuid.UUID
	Timestamp      time.Time
}

func (e VersionDeprecated) Type() EventType {
	return EventVersionDeprecated
}

func (e VersionDeprecated) OccurredAt() time.Time {
	return e.Timestamp
}

// DeprecationRejected is emitted when a deprecation request is rejected
type DeprecationRejected struct {
	DeprecationID uuid.UUID
	VersionID     uuid.UUID
	DocumentID    uuid.UUID
	RejectedBy    string
	Reason        string
	Timestamp     time.Time
}

func (e DeprecationRejected) Type() EventType {
	return EventDeprecationRejected
}

func (e DeprecationRejected) OccurredAt() time.Time {
	return e.Timestamp
}

// VersionUnlocked is emitted when a version returns to its previous status
type VersionUnlocked struct {
	VersionID      uuid.UUID
	DocumentID     uuid.UUID
	RestoredStatus string
	Timestamp      time.Time
}

func NewVersionUnlocked(versionID, documentID, restoredStatus string) VersionUnlocked {
	vid, _ := uuid.Parse(versionID)
	did, _ := uuid.Parse(documentID)
	return VersionUnlocked{
		VersionID:      vid,
		DocumentID:     did,
		RestoredStatus: restoredStatus,
		Timestamp:      time.Now().UTC(),
	}
}

func (e VersionUnlocked) Type() EventType {
	return EventVersionUnlocked
}

func (e VersionUnlocked) OccurredAt() time.Time {
	return e.Timestamp
}

// DeprecationCanceled is emitted when a deprecation request is canceled
type DeprecationCanceled struct {
	DeprecationID uuid.UUID
	VersionID     uuid.UUID
	DocumentID    uuid.UUID
	CanceledBy    string
	Timestamp     time.Time
}

func (e DeprecationCanceled) Type() EventType {
	return EventDeprecationCanceled
}

func (e DeprecationCanceled) OccurredAt() time.Time {
	return e.Timestamp
}

// AutoDeprecationTriggered is emitted when auto-deprecation starts
type AutoDeprecationTriggered struct {
	VersionID    uuid.UUID
	DocumentID   uuid.UUID
	SupersededBy uuid.UUID
	Timestamp    time.Time
}

func (e AutoDeprecationTriggered) Type() EventType {
	return EventAutoDeprecationTriggered
}

func (e AutoDeprecationTriggered) OccurredAt() time.Time {
	return e.Timestamp
}
