package events

import "time"

type EventType string

const (
	// Document events
	EventDocumentCreated EventType = "document.created"
	EventDocumentUpdated EventType = "document.updated"

	// Stakeholder events
	EventStakeholderAdded   EventType = "stakeholder.added"
	EventStakeholderRemoved EventType = "stakeholder.removed"

	// Version events
	EventVersionCreated    EventType = "version.created"
	EventVersionUpdated    EventType = "version.updated"
	EventVersionSubmitted  EventType = "version.submitted"
	EventVersionInReview   EventType = "version.in_review"
	EventVersionApproved   EventType = "version.approved"
	EventVersionRejected   EventType = "version.rejected"
	EventVersionPublished  EventType = "version.published"
	EventVersionDeprecated EventType = "version.deprecated"

	// Deprecation events
	EventDeprecationRequested     EventType = "deprecation.requested"
	EventVersionLocked            EventType = "version.locked"
	EventDeprecationApproved      EventType = "deprecation.approved"
	EventDeprecationRejected      EventType = "deprecation.rejected"
	EventVersionUnlocked          EventType = "version.unlocked"
	EventDeprecationCanceled      EventType = "deprecation.canceled"
	EventAutoDeprecationTriggered EventType = "deprecation.auto_triggered"

	// Approval events
	EventApprovalGranted EventType = "approval.granted"
	EventApprovalRevoked EventType = "approval.revoked"

	// User events
	EventUserCreated EventType = "user.created"
	EventUserUpdated EventType = "user.updated"
)

type Event interface {
	Type() EventType
	OccurredAt() time.Time
}
