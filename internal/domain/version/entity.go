package version

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

type Version struct {
	shared.AggregateRoot
	id          uuid.UUID
	documentID  uuid.UUID
	version     *SemanticVersion
	status      workflow.Status
	contentRef  *shared.ContentReference
	contentHash string
	metadata    map[string]any
	createdBy   string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewVersion(documentID uuid.UUID, versionStr string, contentRef *shared.ContentReference, contentHash, createdBy string, metadata map[string]any) (*Version, error) {
	if documentID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "document ID is required")
	}
	if createdBy == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "created by is required")
	}
	if contentRef == nil {
		return nil, errors.New(errors.CodeInvalidArgument, "content reference is required")
	}
	if contentHash == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "content hash is required")
	}

	semVer, err := NewSemanticVersion(versionStr)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	v := &Version{
		id:          uuid.New(),
		documentID:  documentID,
		version:     semVer,
		status:      workflow.StatusDraft,
		contentRef:  contentRef,
		contentHash: contentHash,
		metadata:    metadata,
		createdBy:   createdBy,
		createdAt:   now,
		updatedAt:   now,
	}

	v.AggregateRoot.AddEvent(events.NewVersionCreated(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		v.createdBy,
	))

	return v, nil
}

func RehydrateVersion(id, documentID uuid.UUID, versionStr string, status workflow.Status, contentRef *shared.ContentReference, contentHash, createdBy string, metadata map[string]any, createdAt, updatedAt time.Time) (*Version, error) {
	semVer, err := NewSemanticVersion(versionStr)
	if err != nil {
		return nil, err
	}

	return &Version{
		id:          id,
		documentID:  documentID,
		version:     semVer,
		status:      status,
		contentRef:  contentRef,
		contentHash: contentHash,
		metadata:    metadata,
		createdBy:   createdBy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func (v *Version) Update(contentRef *shared.ContentReference, contentHash string, metadata map[string]any) error {
	if !v.status.IsEditable() {
		return errors.New(errors.CodePrecondition, "version is not editable")
	}

	v.contentRef = contentRef
	v.contentHash = contentHash
	v.metadata = metadata
	v.updatedAt = time.Now().UTC()

	return nil
}

func (v *Version) Submit(submittedBy string, wf workflow.Workflow) error {
	nextStatus, err := wf.NextStatus(v.status, workflow.ActionSubmit)
	if err != nil {
		return err
	}

	v.status = nextStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionSubmitted(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		submittedBy,
	))

	return nil
}

func (v *Version) Review(reviewedBy string, wf workflow.Workflow) error {
	nextStatus, err := wf.NextStatus(v.status, workflow.ActionReview)
	if err != nil {
		return err
	}

	v.status = nextStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionInReview(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		reviewedBy,
	))

	return nil
}

func (v *Version) Approve(approvedBy, role string, wf workflow.Workflow) error {
	// Note: approval validation happens in application layer
	// This just transitions state when all approvals are complete
	nextStatus, err := wf.NextStatus(v.status, workflow.ActionApprove)
	if err != nil {
		return err
	}

	v.status = nextStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionApproved(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		approvedBy,
		role,
	))

	return nil
}

func (v *Version) Reject(rejectedBy, reason string, wf workflow.Workflow) error {
	nextStatus, err := wf.NextStatus(v.status, workflow.ActionReject)
	if err != nil {
		return err
	}

	v.status = nextStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionRejected(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		rejectedBy,
		reason,
	))

	return nil
}

func (v *Version) Publish(publishedBy string, wf workflow.Workflow) error {
	nextStatus, err := wf.NextStatus(v.status, workflow.ActionPublish)
	if err != nil {
		return err
	}

	v.status = nextStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionPublished(
		v.id.String(),
		v.documentID.String(),
		v.version.String(),
		publishedBy,
	))

	return nil
}

func (v *Version) ID() uuid.UUID {
	return v.id
}

func (v *Version) DocumentID() uuid.UUID {
	return v.documentID
}

func (v *Version) Version() *SemanticVersion {
	return v.version
}

func (v *Version) Status() workflow.Status {
	return v.status
}

func (v *Version) ContentRef() *shared.ContentReference {
	return v.contentRef
}

func (v *Version) ContentHash() string {
	return v.contentHash
}

func (v *Version) Metadata() map[string]any {
	return v.metadata
}

func (v *Version) CreatedBy() string {
	return v.createdBy
}

func (v *Version) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Version) UpdatedAt() time.Time {
	return v.updatedAt
}

// Deprecate transitions the version to DEPRECATING status (locked, pending approval)
func (v *Version) Deprecate() error {
	if !v.status.CanBeDeprecated() {
		return errors.New(errors.CodeVersionNotPublished, "only PUBLISHED versions can be deprecated")
	}

	v.status = workflow.StatusDeprecating
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionLocked(
		v.id.String(),
		v.documentID.String(),
		workflow.StatusPublished.String(),
	))

	return nil
}

// CompleteDeprecation transitions the version to DEPRECATED status (finalized)
func (v *Version) CompleteDeprecation() error {
	if !v.status.IsDeprecating() && !v.status.CanBeDeprecated() {
		return errors.New(errors.CodeInvalidArgument, "version must be DEPRECATING or PUBLISHED to complete deprecation")
	}

	v.status = workflow.StatusDeprecated
	v.updatedAt = time.Now().UTC()

	// Event will be published by deprecation aggregate, not here
	return nil
}

// RejectDeprecation returns the version to its previous status (unlocks)
func (v *Version) RejectDeprecation(previousStatus workflow.Status) error {
	if !v.status.IsDeprecating() {
		return errors.New(errors.CodeInvalidArgument, "version is not in DEPRECATING status")
	}

	v.status = previousStatus
	v.updatedAt = time.Now().UTC()

	v.AggregateRoot.AddEvent(events.NewVersionUnlocked(
		v.id.String(),
		v.documentID.String(),
		previousStatus.String(),
	))

	return nil
}

// IsDeprecated returns true if version is deprecated
func (v *Version) IsDeprecated() bool {
	return v.status.IsDeprecated()
}

// IsDeprecating returns true if version is locked pending deprecation
func (v *Version) IsDeprecating() bool {
	return v.status.IsDeprecating()
}

// CanBeDeprecated returns true if version can be deprecated (is PUBLISHED)
func (v *Version) CanBeDeprecated() bool {
	return v.status.CanBeDeprecated()
}

// IsPolicySatisfied checks if the version's approvals satisfy the given policy
// This is used to determine when to emit VersionFullyApproved event
func (v *Version) IsPolicySatisfied(policy interface {
	IsSatisfied([]*approval.Approval) bool
}, approvals []*approval.Approval) bool {
	return policy.IsSatisfied(approvals)
}
