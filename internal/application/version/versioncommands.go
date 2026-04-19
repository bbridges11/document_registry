package version

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
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
	versionRepo      outbound.VersionRepository
	documentRepo     outbound.DocumentRepository
	approvalRepo     outbound.ApprovalRepository
	deprecationRepo  outbound.DeprecationRepository
	storageService   outbound.StorageService
	txManager        outbound.TransactionManager
	authz            outbound.AuthorizationService
	userRepo         outbound.UserRepository
	eventBus         outbound.EventBus
	workflowFactory  *workflow.Factory
	policyFactory    *approval.PolicyFactory
	contentValidator outbound.ContentValidator
}

func NewCommandService(
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	approvalRepo outbound.ApprovalRepository,
	deprecationRepo outbound.DeprecationRepository,
	storageService outbound.StorageService,
	txManager outbound.TransactionManager,
	authz outbound.AuthorizationService,
	userRepo outbound.UserRepository,
	eventBus outbound.EventBus,
	workflowFactory *workflow.Factory,
	policyFactory *approval.PolicyFactory,
	contentValidator outbound.ContentValidator,
) *CommandService {
	return &CommandService{
		versionRepo:      versionRepo,
		documentRepo:     documentRepo,
		approvalRepo:     approvalRepo,
		deprecationRepo:  deprecationRepo,
		storageService:   storageService,
		txManager:        txManager,
		authz:            authz,
		userRepo:         userRepo,
		eventBus:         eventBus,
		workflowFactory:  workflowFactory,
		policyFactory:    policyFactory,
		contentValidator: contentValidator,
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

	// Check authorization
	allowed := try.To1(s.authz.CanCreateVersion(ctx, cmd.CreatedBy, cmd.DocumentID.String()))
	if !allowed {
		return nil, nil, errors.ErrForbidden
	}

	// Step 3: VALIDATE CONTENT FIRST (before S3, before DB)
	var validationResult *outbound.ValidationResult
	if s.contentValidator.ShouldValidate(doc.DocumentType()) {
		validationResult = try.To1(s.contentValidator.Validate(ctx, outbound.ValidationRequest{
			DocumentType: doc.DocumentType(),
			Content:      contentBytes,
		}))

		// If validation failed, return immediately - NO S3 upload, NO DB transaction
		if !validationResult.Valid {
			return nil, validationResult, errors.New(
				errors.CodeValidationFailed,
				"content validation failed",
			)
		}
	}

	// Generate S3 key using new format: {documentType}/{documentID}/{version}/content
	s3Key := storage.GenerateS3Key(doc.DocumentType(), cmd.DocumentID.String(), cmd.Version)

	// Upload validated content to S3
	try.To(s.storageService.Upload(ctx, s3Key, bytes.NewReader(contentBytes)))

	// Check if version already exists - if so, fail immediately (no overwrites allowed)
	_, err = s.versionRepo.GetByDocumentIDAndVersion(ctx, cmd.DocumentID, cmd.Version)
	if err == nil {
		// Version exists - clean up S3 and fail
		_ = s.storageService.Delete(ctx, s3Key)
		return nil, nil, errors.New(errors.CodeVersionConflict, "version already exists")
	}

	// Validate version is strictly increasing
	latestVer, err := s.versionRepo.GetLatestByDocumentID(ctx, cmd.DocumentID)
	if err == nil {
		newSemVer := try.To1(version.NewSemanticVersion(cmd.Version))
		if !newSemVer.IsGreaterThan(latestVer.Version()) {
			// Clean up uploaded content since version validation failed
			_ = s.storageService.Delete(ctx, s3Key)
			return nil, nil, errors.New(errors.CodeVersionConflict, "version must be strictly greater than latest version")
		}
	}

	// Create new version with S3 key and calculated hash
	ver = try.To1(version.NewVersion(cmd.DocumentID, cmd.Version, s3Key, contentHash, cmd.CreatedBy, cmd.Metadata))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	// Publish validation audit event
	if validationResult != nil {
		s.publishValidationEvent(ctx, cmd.DocumentID.String(), ver.ID().String(), doc.DocumentType(), true, 0, cmd.CreatedBy)
	}

	return ver, validationResult, nil
}

func (s *CommandService) UpdateVersion(ctx context.Context, cmd UpdateVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Check authorization
	allowed := try.To1(s.authz.CanOverwriteVersion(ctx, cmd.UpdatedBy, cmd.ID.String()))
	if !allowed {
		return errors.ErrForbidden
	}

	// Update metadata only (content and hash remain unchanged)
	try.To(ver.Update(ver.ContentS3Key(), ver.ContentHash(), cmd.Metadata))
	try.To(s.versionRepo.Save(ctx, ver))

	return nil
}

func (s *CommandService) SubmitVersion(ctx context.Context, cmd SubmitVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Check authorization
	allowed := try.To1(s.authz.CanSubmitVersion(ctx, cmd.SubmittedBy, cmd.ID.String()))
	if !allowed {
		return errors.ErrForbidden
	}

	// Get document to determine workflow
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))

	try.To(ver.Submit(cmd.SubmittedBy, wf))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	return nil
}

func (s *CommandService) ReviewVersion(ctx context.Context, cmd ReviewVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Check authorization
	allowed := try.To1(s.authz.CanReviewVersion(ctx, cmd.ReviewedBy, cmd.ID.String()))
	if !allowed {
		return errors.ErrForbidden
	}

	// Get document to determine workflow
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))

	try.To(ver.Review(cmd.ReviewedBy, wf))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	return nil
}

func (s *CommandService) ApproveVersion(ctx context.Context, cmd ApproveVersionCommand) (err error) {
	defer err2.Handle(&err)

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

		// Check authorization
		allowed := try.To1(s.authz.CanApproveVersion(ctx, cmd.ApprovedBy, cmd.ID.String()))
		if !allowed {
			return errors.ErrForbidden
		}

		// Get document to determine policy
		doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
		policy := try.To1(s.policyFactory.GetPolicy(doc.DocumentType()))

		// Get user to determine approval roles
		// Parse user ID as UUID
		userID := try.To1(uuid.Parse(cmd.ApprovedBy))
		user := try.To1(s.userRepo.GetByID(ctx, userID))

		// Map user role to approval roles
		// Admin users can perform all approval roles
		// Contributors can perform technical approvals
		// Viewers cannot approve
		var userRoles []approval.ApprovalRole
		switch user.Role() {
		case "admin":
			userRoles = []approval.ApprovalRole{
				approval.ApprovalRoleTechnical,
				approval.ApprovalRoleArchitect,
				approval.ApprovalRoleProduct,
			}
		case "contributor":
			userRoles = []approval.ApprovalRole{
				approval.ApprovalRoleTechnical,
			}
		default:
			// Viewers have no approval roles
			userRoles = []approval.ApprovalRole{}
		}

		// Get existing approvals
		existingApprovals := try.To1(s.approvalRepo.ListByVersionID(ctx, ver.ID()))

		// Create approval
		newApproval := try.To1(approval.NewApproval(ver.ID(), cmd.ApprovedBy, cmd.Role, true, cmd.Comment))

		// Validate approval against policy
		try.To(policy.ValidateApproval(newApproval, ver.CreatedBy(), userRoles, existingApprovals))

		// Save approval
		try.To(s.approvalRepo.Save(ctx, newApproval))

		// Check if all approvals are satisfied
		allApprovals := append(existingApprovals, newApproval)
		if policy.IsSatisfied(allApprovals) {
			// Transition version to APPROVED
			wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))
			try.To(ver.Approve(cmd.ApprovedBy, cmd.Role.String(), wf))
			try.To(s.versionRepo.Save(ctx, ver))

			// Publish events after successful persistence
			s.publishEvents(ctx, ver, "")
		}

		return nil
	})
}

func (s *CommandService) RejectVersion(ctx context.Context, cmd RejectVersionCommand) (err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.ID))

	// Check authorization
	allowed := try.To1(s.authz.CanRejectVersion(ctx, cmd.RejectedBy, cmd.ID.String()))
	if !allowed {
		return errors.ErrForbidden
	}

	// Get document to determine workflow
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))

	try.To(ver.Reject(cmd.RejectedBy, cmd.Reason, wf))
	try.To(s.versionRepo.Save(ctx, ver))

	// Publish events after successful persistence
	s.publishEvents(ctx, ver, "")

	return nil
}

func (s *CommandService) PublishVersion(ctx context.Context, cmd PublishVersionCommand) (err error) {
	defer err2.Handle(&err)

	var eventsToPublish []events.Event

	try.To(s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		ver := try.To1(s.versionRepo.GetByID(txCtx, cmd.ID))

		// Check authorization
		allowed := try.To1(s.authz.CanPublishVersion(txCtx, cmd.PublishedBy, cmd.ID.String()))
		if !allowed {
			return errors.ErrForbidden
		}

		// Get document to determine workflow
		doc := try.To1(s.documentRepo.GetByID(txCtx, ver.DocumentID()))
		wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))

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
	}))

	// Publish all events after transaction commits
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
