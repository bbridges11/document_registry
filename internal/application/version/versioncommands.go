package version

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bbridges_11/document-registry/internal/adapters/outbound/publisher"
	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
	"github.com/bbridges_11/document-registry/internal/domain/publication"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/platform/storage"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type CommandService struct {
	versionRepo       outbound.VersionRepository
	documentRepo      outbound.DocumentRepository
	approvalRepo      outbound.ApprovalRepository
	deprecationRepo   outbound.DeprecationRepository
	publicationRepo   outbound.PublicationRepository
	storageService    outbound.StorageService
	publisherRegistry *publisher.DestinationRegistry
	txManager         outbound.TransactionManager
	userRepo          outbound.UserRepository
	userRoleResolver  outbound.UserRoleResolver
	eventBus          outbound.EventBus
	workflowFactory   *workflow.Factory
	policyFactory     *approval.PolicyFactory
	contentValidator  outbound.ContentValidator
}

func NewCommandService(
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	approvalRepo outbound.ApprovalRepository,
	deprecationRepo outbound.DeprecationRepository,
	publicationRepo outbound.PublicationRepository,
	storageService outbound.StorageService,
	publisherRegistry *publisher.DestinationRegistry,
	txManager outbound.TransactionManager,
	userRepo outbound.UserRepository,
	userRoleResolver outbound.UserRoleResolver,
	eventBus outbound.EventBus,
	workflowFactory *workflow.Factory,
	policyFactory *approval.PolicyFactory,
	contentValidator outbound.ContentValidator,
) *CommandService {
	return &CommandService{
		versionRepo:       versionRepo,
		documentRepo:      documentRepo,
		approvalRepo:      approvalRepo,
		deprecationRepo:   deprecationRepo,
		publicationRepo:   publicationRepo,
		storageService:    storageService,
		publisherRegistry: publisherRegistry,
		txManager:         txManager,
		userRepo:          userRepo,
		userRoleResolver:  userRoleResolver,
		eventBus:          eventBus,
		workflowFactory:   workflowFactory,
		policyFactory:     policyFactory,
		contentValidator:  contentValidator,
	}
}

func (s *CommandService) CreateVersion(ctx context.Context, cmd CreateVersionCommand) (ver *version.Version, valResult *outbound.ValidationResult, err error) {
	defer err2.Handle(&err)

	// Step 1: Read content into memory with size limit
	const maxContentSize = 10 * 1024 * 1024 // 10MB
	limitReader := io.LimitReader(cmd.Content, maxContentSize+1)
	contentBytes := try.To1(io.ReadAll(limitReader))

	if len(contentBytes) > maxContentSize {
		return nil, nil, errors.New(
			errors.CodeInvalidArgument,
			fmt.Sprintf("content exceeds maximum size of %d bytes", maxContentSize),
		)
	}

	// Step 2: Calculate content hash
	contentHash := storage.CalculateContentHash(contentBytes)

	// Validate document exists and get document type
	doc := try.To1(s.documentRepo.GetByID(ctx, cmd.DocumentID))

	// Step 3: VALIDATE CONTENT FIRST (before storage, before DB)
	var validationResult *outbound.ValidationResult
	if doc.DocumentType().RequiresValidation() {
		validationResult = try.To1(s.contentValidator.Validate(ctx, outbound.ValidationRequest{
			DocumentType: doc.DocumentType().Code(),
			Content:      contentBytes,
		}))

		// If validation failed, return immediately - NO storage upload, NO DB transaction
		if !validationResult.Valid {
			return nil, validationResult, errors.New(
				errors.CodeValidationFailed,
				"content validation failed",
			)
		}
	}

	// Generate storage key using format: {documentType}/{documentID}/{version}/content
	storageKey := storage.GenerateS3Key(doc.DocumentType().Code(), cmd.DocumentID.String(), cmd.Version)

	// Upload validated content to storage (using S3 backend by default)
	contentRef := try.To1(s.storageService.Upload(ctx, outbound.StorageRequest{
		Backend: shared.StorageBackendS3,
		Key:     storageKey,
		Content: bytes.NewReader(contentBytes),
	}))

	// Check if version already exists - if so, fail immediately (no overwrites allowed)
	_, err = s.versionRepo.GetByDocumentIDAndVersion(ctx, cmd.DocumentID, cmd.Version)
	if err == nil {
		// Version exists - clean up storage and fail
		_ = s.storageService.Delete(ctx, contentRef)
		return nil, nil, errors.New(errors.CodeVersionConflict, "version already exists")
	}

	// Validate version is strictly increasing
	latestVer, err := s.versionRepo.GetLatestByDocumentID(ctx, cmd.DocumentID)
	if err == nil {
		newSemVer := try.To1(version.NewSemanticVersion(cmd.Version))
		if !newSemVer.IsGreaterThan(latestVer.Version()) {
			// Clean up uploaded content since version validation failed
			_ = s.storageService.Delete(ctx, contentRef)
			return nil, nil, errors.New(errors.CodeVersionConflict, "version must be strictly greater than latest version")
		}
	}

	// Create new version with content reference and calculated hash
	ver = try.To1(version.NewVersion(cmd.DocumentID, cmd.Version, contentRef, contentHash, cmd.CreatedBy, cmd.Metadata))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	// Publish validation audit event
	if validationResult != nil {
		s.publishValidationEvent(ctx, cmd.DocumentID.String(), ver.ID().String(), doc.DocumentType().Code(), true, 0, cmd.CreatedBy)
	}

	return ver, validationResult, nil
}

func (s *CommandService) UpdateVersion(ctx context.Context, cmd UpdateVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Update metadata only (content reference and hash remain unchanged)
	try.To(ver.Update(ver.ContentRef(), ver.ContentHash(), cmd.Metadata))
	try.To(s.versionRepo.Save(ctx, ver))

	return nil
}

func (s *CommandService) SubmitVersion(ctx context.Context, cmd SubmitVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Get document to determine workflow
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType().WorkflowType()))

	try.To(ver.Submit(cmd.SubmittedBy, wf))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	return nil
}

func (s *CommandService) ReviewVersion(ctx context.Context, cmd ReviewVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Validate version state - manual review only allowed from SUBMITTED
	if ver.Status() != workflow.StatusSubmitted {
		return errors.New(errors.CodePrecondition, "version must be SUBMITTED to manually transition to IN_REVIEW")
	}

	// Get document to determine workflow
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType().WorkflowType()))

	// Verify user has approval roles (only users who can approve can also review)
	userID := try.To1(uuid.Parse(cmd.ReviewedBy))
	availableRoles := try.To1(s.userRoleResolver.GetApprovalRoles(ctx, userID, cmd.UserRole))

	// Validate user has at least one approval role
	if len(availableRoles) == 0 {
		return errors.New(errors.CodeForbidden, "user has no approval roles and cannot review versions")
	}

	// Transition to IN_REVIEW
	try.To(ver.Review(cmd.ReviewedBy, wf))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	return nil
}

func (s *CommandService) ApproveVersion(ctx context.Context, cmd ApproveVersionCommand) (err error) {
	defer err2.Handle(&err)

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// 🔒 LOCK: Get version with FOR UPDATE to prevent concurrent modifications
		ver := try.To1(s.versionRepo.GetByIDForUpdate(ctx, cmd.ID))

		// Validate version state - must be SUBMITTED or IN_REVIEW
		if ver.Status() != workflow.StatusSubmitted && ver.Status() != workflow.StatusInReview {
			return errors.New(errors.CodePrecondition, "version must be SUBMITTED or IN_REVIEW to approve")
		}

		// Get document to determine policy and workflow
		doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
		policy := try.To1(s.policyFactory.GetPolicy(doc.DocumentType().ApprovalPolicy()))
		wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType().WorkflowType()))

		// Parse user ID and get approval roles from resolver (using role from Actor context)
		userID := try.To1(uuid.Parse(cmd.ApprovedBy))
		availableRoles := try.To1(s.userRoleResolver.GetApprovalRoles(ctx, userID, cmd.UserRole))

		// Validate user has at least one approval role
		if len(availableRoles) == 0 {
			return errors.New(errors.CodeForbidden, "user has no approval roles")
		}

		// Determine which role to use - use first available role
		// For users with multiple roles (e.g., admin), this will use technical (first in list)
		roleToUse := availableRoles[0]

		// Auto-transition to IN_REVIEW if version is SUBMITTED
		if ver.Status() == workflow.StatusSubmitted {
			try.To(ver.Review(cmd.ApprovedBy, wf))
			try.To(s.versionRepo.Save(ctx, ver))

			// Publish VersionInReview event
			reviewEvent := events.NewVersionInReview(
				ver.ID().String(),
				ver.DocumentID().String(),
				ver.Version().String(),
				cmd.ApprovedBy,
			)
			s.eventBus.Publish(ctx, events.NewEnvelope(reviewEvent, ""))
		}

		// Get existing approvals (read after lock, so data is fresh)
		existingApprovals := try.To1(s.approvalRepo.ListByVersionID(ctx, ver.ID()))

		// Create approval with server-determined role
		newApproval := try.To1(approval.NewApproval(ver.ID(), cmd.ApprovedBy, roleToUse, true, cmd.Comment))

		// Validate approval against policy
		try.To(policy.ValidateApproval(newApproval, ver.CreatedBy(), availableRoles, existingApprovals))

		// Save approval
		try.To(s.approvalRepo.Save(ctx, newApproval))

		// Publish individual VersionApproved event
		approvedEvent := events.NewVersionApproved(
			ver.ID().String(),
			ver.DocumentID().String(),
			ver.Version().String(),
			cmd.ApprovedBy,
			roleToUse.String(),
		)
		s.eventBus.Publish(ctx, events.NewEnvelope(approvedEvent, ""))

		// Check if all approvals are satisfied
		allApprovals := append(existingApprovals, newApproval)
		if policy.IsSatisfied(allApprovals) {
			// Transition version to APPROVED
			try.To(ver.Approve(cmd.ApprovedBy, roleToUse.String(), wf))
			try.To(s.versionRepo.Save(ctx, ver))

			// Publish events after successful persistence
			s.publishEvents(ctx, ver, "")

			// Emit VersionFullyApproved event - policy is satisfied
			fullyApprovedEvent := events.NewVersionFullyApproved(
				ver.ID().String(),
				ver.DocumentID().String(),
				ver.Version().String(),
			)
			s.eventBus.Publish(ctx, events.NewEnvelope(fullyApprovedEvent, ""))
		}

		return nil
	})
}

func (s *CommandService) RejectVersion(ctx context.Context, cmd RejectVersionCommand) (err error) {
	defer err2.Handle(&err)

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// 🔒 LOCK: Get version with FOR UPDATE to prevent concurrent modifications
		ver := try.To1(s.versionRepo.GetByIDForUpdate(ctx, cmd.ID))

		// Validate version state - must be SUBMITTED or IN_REVIEW
		if ver.Status() != workflow.StatusSubmitted && ver.Status() != workflow.StatusInReview {
			return errors.New(errors.CodePrecondition, "version must be SUBMITTED or IN_REVIEW to reject")
		}

		// Get document to determine workflow
		doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
		wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType().WorkflowType()))

		// Verify user has approval roles (only users who can approve can also reject)
		userID := try.To1(uuid.Parse(cmd.RejectedBy))
		availableRoles := try.To1(s.userRoleResolver.GetApprovalRoles(ctx, userID, cmd.UserRole))

		// Validate user has at least one approval role
		if len(availableRoles) == 0 {
			return errors.New(errors.CodeForbidden, "user has no approval roles and cannot reject versions")
		}

		// Auto-transition to IN_REVIEW if version is SUBMITTED
		if ver.Status() == workflow.StatusSubmitted {
			try.To(ver.Review(cmd.RejectedBy, wf))
			try.To(s.versionRepo.Save(ctx, ver))

			// Publish VersionInReview event
			reviewEvent := events.NewVersionInReview(
				ver.ID().String(),
				ver.DocumentID().String(),
				ver.Version().String(),
				cmd.RejectedBy,
			)
			s.eventBus.Publish(ctx, events.NewEnvelope(reviewEvent, ""))
		}

		// Now reject (transitions IN_REVIEW → REJECTED)
		try.To(ver.Reject(cmd.RejectedBy, cmd.Reason, wf))
		try.To(s.versionRepo.Save(ctx, ver))

		// Publish events after successful persistence
		s.publishEvents(ctx, ver, "")

		return nil
	})
}

func (s *CommandService) PublishVersion(ctx context.Context, cmd PublishVersionCommand) (err error) {
	defer err2.Handle(&err)

	// Validate destination
	if cmd.Destination == "" {
		return errors.New(errors.CodeInvalidArgument, "destination is required")
	}

	// Validate environment
	if cmd.Environment == "" {
		return errors.New(errors.CodeInvalidArgument, "environment is required")
	}

	// Get publisher from registry
	publisherService, err := s.publisherRegistry.GetPublisher(cmd.Destination)
	if err != nil {
		return errors.New(errors.CodeInvalidArgument, fmt.Sprintf("invalid destination: %v", err))
	}

	// Validate environment is supported
	if err := publisherService.ValidateEnvironment(cmd.Environment); err != nil {
		return errors.New(errors.CodeInvalidArgument, fmt.Sprintf("invalid environment: %v", err))
	}

	var eventsToPublish []events.Event
	var pub *publication.Publication

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		ver := try.To1(s.versionRepo.GetByID(txCtx, cmd.ID))

		// Check user has product role
		userID := try.To1(uuid.Parse(cmd.PublishedBy))
		user := try.To1(s.userRepo.GetByID(txCtx, userID))

		if user.Role() != "product" {
			return errors.New(errors.CodeForbidden, "only product role can publish versions")
		}

		// Check version is approved
		if ver.Status() != workflow.StatusApproved {
			return errors.New(errors.CodeInvalidState, "version must be approved before publishing")
		}

		// Get document
		doc := try.To1(s.documentRepo.GetByID(txCtx, ver.DocumentID()))
		wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType().WorkflowType()))

		// Get content from S3
		content := try.To1(s.storageService.Download(txCtx, ver.ContentRef()))
		defer content.Close()

		// Read content into bytes
		contentBytes := try.To1(io.ReadAll(content))

		// Call publisher with environment
		publishReq := outbound.PublishRequest{
			DocumentID:    ver.DocumentID().String(),
			VersionNumber: ver.Version().String(),
			Content:       contentBytes,
			Metadata:      ver.Metadata(),
			Destination:   cmd.Destination,
			Environment:   cmd.Environment,
		}

		err := publisherService.Publish(txCtx, publishReq)
		if err != nil {
			// Publication failed - create failed publication record
			failedPub := try.To1(publication.NewPublication(
				ver.ID(),
				ver.DocumentID(),
				cmd.Destination,
				cmd.PublishedBy,
			))
			failedPub.MarkAsFailed(err.Error())
			try.To(s.publicationRepo.Save(txCtx, failedPub))

			return fmt.Errorf("failed to publish to %s (%s): %w", cmd.Destination, cmd.Environment, err)
		}

		// Publication successful - create success publication record
		pub = try.To1(publication.NewPublication(
			ver.ID(),
			ver.DocumentID(),
			cmd.Destination,
			cmd.PublishedBy,
		))
		try.To(s.publicationRepo.Save(txCtx, pub))

		// Check for existing PUBLISHED version (enforce single-published rule)
		previousPublished, err := s.versionRepo.GetPublishedByDocumentID(txCtx, ver.DocumentID())
		if err == nil && previousPublished != nil && previousPublished.ID() != ver.ID() {
			// Auto-deprecate the previous PUBLISHED version
			dep := try.To1(deprecation.NewAutoDeprecation(
				previousPublished.ID(),
				previousPublished.DocumentID(),
				ver.ID(),
				fmt.Sprintf("Superseded by version %s", ver.Version().String()),
				cmd.PublishedBy,
			))

			// Complete deprecation immediately (auto-deprecation approved by default)
			try.To(previousPublished.CompleteDeprecation())

			// Save deprecation and old version
			try.To(s.deprecationRepo.Save(txCtx, dep))
			try.To(s.versionRepo.Save(txCtx, previousPublished))

			// Collect events to publish after transaction
			eventsToPublish = append(eventsToPublish, previousPublished.Events()...)
			eventsToPublish = append(eventsToPublish, dep.Events()...)
			previousPublished.ClearEvents()
			dep.ClearEvents()
		}

		// Publish the new version
		try.To(ver.Publish(cmd.PublishedBy, wf))
		try.To(s.versionRepo.Save(txCtx, ver))

		// Collect events to publish after transaction
		eventsToPublish = append(eventsToPublish, ver.Events()...)
		ver.ClearEvents()

		return nil
	})
	try.To(err)

	// Publish all events after transaction commits (only if successful)
	for _, evt := range eventsToPublish {
		s.eventBus.Publish(ctx, events.NewEnvelope(evt, ""))
	}

	return nil
}

func (s *CommandService) publishEvents(ctx context.Context, ver *version.Version, traceID string) {
	domainEvents := ver.Events()
	if len(domainEvents) == 0 {
		return
	}

	envelopes := make([]events.Envelope, 0, len(domainEvents))
	for _, event := range domainEvents {
		envelopes = append(envelopes, events.NewEnvelope(event, traceID))
	}

	s.eventBus.Publish(ctx, envelopes...)
	ver.ClearEvents()
}

func (s *CommandService) publishValidationEvent(ctx context.Context, documentID, versionID, documentType string, valid bool, issueCount int, validatedBy string) {
	envelope := events.NewEnvelope(
		events.NewValidationCompleted(documentID, versionID, documentType, valid, issueCount, validatedBy),
		"",
	)
	s.eventBus.Publish(ctx, envelope)
}
