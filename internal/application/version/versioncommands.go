package version

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bbridges_11/document-registry/internal/adapters/outbound/publisher"
	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
	"github.com/bbridges_11/document-registry/internal/domain/document"
	"github.com/bbridges_11/document-registry/internal/domain/publication"
	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
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
	stakeholderRepo   outbound.StakeholderRepository
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
	stakeholderRepo outbound.StakeholderRepository,
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
		stakeholderRepo:   stakeholderRepo,
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

	// Collect events to publish after transaction
	var reviewerAssignmentEvents []events.Event

	// Submit version and auto-assign reviewers in transaction
	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Submit version
		try.To(ver.Submit(cmd.SubmittedBy, wf))
		try.To(s.versionRepo.Save(txCtx, ver))

		// Auto-assign reviewers and collect assigned user IDs
		assignedUserIDs := try.To1(s.assignReviewers(txCtx, ver, doc))

		// Build ReviewersAssigned event if reviewers were assigned
		if len(assignedUserIDs) > 0 {
			event := events.NewReviewersAssigned(
				ver.ID().String(),
				doc.ID().String(),
				doc.Name(),
				ver.Version().String(),
				doc.DocumentType().String(),
				assignedUserIDs,
			)
			reviewerAssignmentEvents = append(reviewerAssignmentEvents, event)
		}

		return nil
	}))

	// Publish version events after transaction commits
	s.publishEvents(ctx, ver, "")

	// Publish reviewer assignment events after transaction commits
	for _, event := range reviewerAssignmentEvents {
		s.eventBus.Publish(ctx, events.NewEnvelope(event, ""))
	}

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

// assignReviewers automatically assigns reviewers based on the document's approval policy
// Returns the list of user IDs that were assigned as reviewers
func (s *CommandService) assignReviewers(ctx context.Context, ver *version.Version, doc *document.Document) ([]string, error) {
	// Get approval policy for document type
	policy := try.To1(s.policyFactory.GetPolicy(doc.DocumentType().ApprovalPolicy()))

	// Get required approval roles from policy
	requiredRoles := policy.RequiredApprovals()

	// Collect all users to assign (deduplicated)
	assignedUserIDs := make(map[string]bool)

	// For each required approval role, find users with that role
	for approvalRole := range requiredRoles {
		userIDs := try.To1(s.findUsersWithApprovalRole(ctx, approvalRole))

		for _, userID := range userIDs {
			// Ensure user is a stakeholder
			try.To(s.ensureStakeholder(ctx, doc.ID(), userID))

			// Track assigned user (deduplicate)
			assignedUserIDs[userID] = true
		}
	}

	// Return list of assigned user IDs
	userIDList := make([]string, 0, len(assignedUserIDs))
	for userID := range assignedUserIDs {
		userIDList = append(userIDList, userID)
	}

	return userIDList, nil
}

// findUsersWithApprovalRole finds all active users who have the specified approval role
func (s *CommandService) findUsersWithApprovalRole(ctx context.Context, approvalRole approval.ApprovalRole) ([]string, error) {
	// Get all active users from the system
	allUsers := try.To1(s.userRepo.ListAll(ctx))

	var matchingUserIDs []string

	// Filter users by approval role
	for _, user := range allUsers {
		// Skip inactive users
		if !user.Active() {
			continue
		}

		// Get approval roles for this user
		approvalRoles := try.To1(s.userRoleResolver.GetApprovalRoles(ctx, user.ID(), string(user.Role())))

		// Check if user has the required approval role
		for _, role := range approvalRoles {
			if role == approvalRole {
				matchingUserIDs = append(matchingUserIDs, user.ID().String())
				break
			}
		}
	}

	return matchingUserIDs, nil
}

// ensureStakeholder ensures a user is a stakeholder for the document
func (s *CommandService) ensureStakeholder(ctx context.Context, documentID uuid.UUID, userID string) error {
	// Check if user is already a stakeholder
	exists := try.To1(s.stakeholderRepo.ExistsByDocumentIDAndUserID(ctx, documentID, userID))

	if exists {
		return nil // Already a stakeholder, nothing to do
	}

	// Create new stakeholder with contributor role (needed for review/approval)
	sh := try.To1(stakeholder.NewStakeholder(documentID, userID, stakeholder.RoleContributor))

	// Save stakeholder
	return s.stakeholderRepo.Save(ctx, sh)
}
