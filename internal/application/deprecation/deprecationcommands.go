package deprecation

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// Command types

// RequestDeprecation creates a new manual deprecation request
func (s *CommandService) RequestDeprecation(ctx context.Context, cmd RequestDeprecationCommand) (dep *deprecation.Deprecation, err error) {
	defer err2.Handle(&err)

	// Validate version exists and is PUBLISHED
	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.VersionID))
	if !ver.CanBeDeprecated() {
		return nil, errors.New(errors.CodeVersionNotPublished, "only PUBLISHED versions can be deprecated")
	}

	// Check for existing pending deprecation
	existingDep, err := s.deprecationRepo.GetPendingByVersionID(ctx, cmd.VersionID)
	if err == nil && existingDep != nil {
		return nil, errors.New(errors.CodeConflict, "deprecation request already pending for this version")
	}

	// Authorization check - verify user can request deprecation (creator/owner/admin)
	// Users who can request deprecation:
	// 1. Document creator
	// 2. Document owners (stakeholders with owner role)
	// 3. System admins
	canRequest := try.To1(s.authz.CanRequestDeprecation(ctx, cmd.RequestedBy, cmd.VersionID.String()))
	if !canRequest {
		return nil, errors.New(errors.CodeCannotDeprecate, "user not authorized to request deprecation for this version")
	}

	// Create entities and update in transaction
	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Create deprecation entity
		dep = try.To1(deprecation.NewDeprecation(
			cmd.VersionID,
			ver.DocumentID(),
			cmd.RequestedBy,
			cmd.Reason,
			cmd.DeprecationNote,
			ver.Status().String(),
		))

		// Lock version (PUBLISHED → DEPRECATING)
		try.To(ver.Deprecate())

		// Save both
		try.To(s.deprecationRepo.Save(txCtx, dep))
		try.To(s.versionRepo.Save(txCtx, ver))

		return nil
	}))

	// Publish events after transaction succeeds
	s.publishEvents(ctx, dep, "")
	s.publishEvents(ctx, ver, "")

	// Clear events
	dep.ClearEvents()
	ver.ClearEvents()

	return dep, nil
}

// ApproveDeprecation approves a pending deprecation
func (s *CommandService) ApproveDeprecation(ctx context.Context, cmd ApproveDeprecationCommand) (err error) {
	defer err2.Handle(&err)

	// Get pending deprecation
	dep := try.To1(s.deprecationRepo.GetPendingByVersionID(ctx, cmd.VersionID))
	if dep == nil {
		return errors.New(errors.CodeDeprecationNotFound, "no pending deprecation found for version")
	}

	// Get version
	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.VersionID))

	// Authorization check - verify user can approve deprecation (has approver role)
	// Users who can approve deprecation:
	// 1. Users with "deprecation-approver" role
	// 2. System admins
	canApprove := try.To1(s.authz.CanApproveDeprecation(ctx, cmd.ApprovedBy, cmd.VersionID.String()))
	if !canApprove {
		return errors.ErrForbidden
	}

	// Update entities in transaction
	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Approve deprecation
		try.To(dep.Approve(cmd.ApprovedBy))

		// Complete deprecation - finalize
		try.To(dep.Complete(cmd.ApprovedBy))

		// Complete version deprecation (DEPRECATING → DEPRECATED)
		try.To(ver.CompleteDeprecation())

		// Save both
		try.To(s.deprecationRepo.Save(txCtx, dep))
		try.To(s.versionRepo.Save(txCtx, ver))

		return nil
	}))

	// Publish events after transaction succeeds
	s.publishEvents(ctx, dep, "")
	s.publishEvents(ctx, ver, "")

	// Clear events
	dep.ClearEvents()
	ver.ClearEvents()

	return nil
}

// RejectDeprecation rejects a pending deprecation
func (s *CommandService) RejectDeprecation(ctx context.Context, cmd RejectDeprecationCommand) (err error) {
	defer err2.Handle(&err)

	// Get pending deprecation
	dep := try.To1(s.deprecationRepo.GetPendingByVersionID(ctx, cmd.VersionID))
	if dep == nil {
		return errors.New(errors.CodeDeprecationNotFound, "no pending deprecation found for version")
	}

	// Get version
	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.VersionID))

	// Authorization check - verify user can reject deprecation (has approver role)
	// Users who can reject deprecation:
	// 1. Users with "deprecation-approver" role
	// 2. System admins
	canReject := try.To1(s.authz.CanApproveDeprecation(ctx, cmd.RejectedBy, cmd.VersionID.String()))
	if !canReject {
		return errors.ErrForbidden
	}

	// Parse previous status
	prevStatus := workflow.Status(dep.PreviousStatus())

	// Update entities in transaction
	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Reject deprecation
		try.To(dep.Reject(cmd.RejectedBy, cmd.Reason))

		// Unlock version (DEPRECATING → previous status, typically PUBLISHED)
		try.To(ver.RejectDeprecation(prevStatus))

		// Save both
		try.To(s.deprecationRepo.Save(txCtx, dep))
		try.To(s.versionRepo.Save(txCtx, ver))

		return nil
	}))

	// Publish events after transaction succeeds
	s.publishEvents(ctx, dep, "")
	s.publishEvents(ctx, ver, "")

	// Clear events
	dep.ClearEvents()
	ver.ClearEvents()

	return nil
}

// CancelDeprecation cancels a pending deprecation
func (s *CommandService) CancelDeprecation(ctx context.Context, cmd CancelDeprecationCommand) (err error) {
	defer err2.Handle(&err)

	// Get pending deprecation
	dep := try.To1(s.deprecationRepo.GetPendingByVersionID(ctx, cmd.VersionID))
	if dep == nil {
		return errors.New(errors.CodeDeprecationNotFound, "no pending deprecation found for version")
	}

	// Get version
	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.VersionID))

	// Authorization check - verify user can cancel (original requestor or admin)
	// Users who can cancel deprecation:
	// 1. Original requestor (the user who requested the deprecation)
	// 2. System admins
	canCancel := try.To1(s.authz.CanCancelDeprecation(ctx, cmd.CanceledBy, cmd.VersionID.String(), dep.RequestedBy()))
	if !canCancel {
		return errors.ErrForbidden
	}

	// Parse previous status
	prevStatus := workflow.Status(dep.PreviousStatus())

	// Update entities in transaction
	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Cancel deprecation
		try.To(dep.Cancel(cmd.CanceledBy))

		// Unlock version (DEPRECATING → previous status, typically PUBLISHED)
		try.To(ver.RejectDeprecation(prevStatus))

		// Save both
		try.To(s.deprecationRepo.Save(txCtx, dep))
		try.To(s.versionRepo.Save(txCtx, ver))

		return nil
	}))

	// Publish events after transaction succeeds
	s.publishEvents(ctx, dep, "")
	s.publishEvents(ctx, ver, "")

	// Clear events
	dep.ClearEvents()
	ver.ClearEvents()

	return nil
}

// GetDeprecation retrieves a deprecation by version ID
func (s *CommandService) GetDeprecation(ctx context.Context, versionID uuid.UUID) (*deprecation.Deprecation, error) {
	// Get the most recent deprecation for this version
	dep, err := s.deprecationRepo.GetByVersionID(ctx, versionID)
	if err != nil {
		return nil, errors.New(errors.CodeDeprecationNotFound, "deprecation not found")
	}
	return dep, nil
}

func (s *CommandService) publishEvents(ctx context.Context, dep shared.EventEmitter, traceID string) {
	domainEvents := dep.Events()
	if len(domainEvents) == 0 {
		return
	}

	envelopes := make([]events.Envelope, 0, len(domainEvents))
	for _, event := range domainEvents {
		envelopes = append(envelopes, events.NewEnvelope(event, traceID))
	}

	s.eventBus.Publish(ctx, envelopes...)
	dep.ClearEvents()
}
