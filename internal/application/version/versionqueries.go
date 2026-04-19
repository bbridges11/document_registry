package version

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type QueryService struct {
	versionRepo     outbound.VersionRepository
	documentRepo    outbound.DocumentRepository
	approvalRepo    outbound.ApprovalRepository
	authz           outbound.AuthorizationService
	userRepo        outbound.UserRepository
	workflowFactory *workflow.Factory
	policyFactory   *approval.PolicyFactory
}

func NewQueryService(
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	approvalRepo outbound.ApprovalRepository,
	authz outbound.AuthorizationService,
	userRepo outbound.UserRepository,
	workflowFactory *workflow.Factory,
	policyFactory *approval.PolicyFactory,
) *QueryService {
	return &QueryService{
		versionRepo:     versionRepo,
		documentRepo:    documentRepo,
		approvalRepo:    approvalRepo,
		authz:           authz,
		userRepo:        userRepo,
		workflowFactory: workflowFactory,
		policyFactory:   policyFactory,
	}
}

func (s *QueryService) GetVersion(ctx context.Context, query GetVersionQuery, userID string) (dto *VersionDTO, err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, query.ID))

	// Check if document has published version (public access)
	hasPublished := try.To1(s.versionRepo.HasPublishedVersion(ctx, ver.DocumentID()))

	if !hasPublished {
		// Not published - check authorization
		doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
		allowed := try.To1(s.authz.CanAccessDocument(ctx, userID, doc.ID().String()))
		if !allowed {
			return nil, errors.ErrForbidden
		}
	}

	return &VersionDTO{
		ID:           ver.ID(),
		DocumentID:   ver.DocumentID(),
		Version:      ver.Version().String(),
		Status:       ver.Status(),
		ContentS3Key: ver.ContentS3Key(),
		ContentHash:  ver.ContentHash(),
		Metadata:     ver.Metadata(),
		CreatedBy:    ver.CreatedBy(),
		CreatedAt:    ver.CreatedAt(),
		UpdatedAt:    ver.UpdatedAt(),
	}, nil
}

func (s *QueryService) ListVersionsByDocument(ctx context.Context, query ListVersionsByDocumentQuery, userID string) (dtos []*VersionDTO, err error) {
	defer err2.Handle(&err)

	// Check if document has published version (public access)
	hasPublished := try.To1(s.versionRepo.HasPublishedVersion(ctx, query.DocumentID))

	if !hasPublished {
		// Not published - check authorization
		allowed := try.To1(s.authz.CanAccessDocument(ctx, userID, query.DocumentID.String()))
		if !allowed {
			return nil, errors.ErrForbidden
		}
	}

	versions := try.To1(s.versionRepo.ListByDocumentID(ctx, query.DocumentID))

	dtos = make([]*VersionDTO, 0, len(versions))
	for _, ver := range versions {
		dtos = append(dtos, &VersionDTO{
			ID:           ver.ID(),
			DocumentID:   ver.DocumentID(),
			Version:      ver.Version().String(),
			Status:       ver.Status(),
			ContentS3Key: ver.ContentS3Key(),
			ContentHash:  ver.ContentHash(),
			Metadata:     ver.Metadata(),
			CreatedBy:    ver.CreatedBy(),
			CreatedAt:    ver.CreatedAt(),
			UpdatedAt:    ver.UpdatedAt(),
		})
	}

	return dtos, nil
}

func (s *QueryService) GetVersionStatus(ctx context.Context, query GetVersionStatusQuery, userID string) (dto *VersionStatusDTO, err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, query.ID))

	// Check authorization
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	allowed := try.To1(s.authz.CanAccessDocument(ctx, userID, doc.ID().String()))
	if !allowed {
		return nil, errors.ErrForbidden
	}

	// Get workflow to determine allowed actions
	wf := try.To1(s.workflowFactory.GetWorkflow(doc.DocumentType()))
	allowedActions := wf.AllowedActions(ver.Status())

	// Get approval summary
	approvalSummary := try.To1(s.getApprovalSummary(ctx, ver.ID(), doc.DocumentType()))

	return &VersionStatusDTO{
		ID:              ver.ID(),
		Version:         ver.Version().String(),
		Status:          ver.Status(),
		Editable:        ver.Status().IsEditable(),
		Terminal:        ver.Status().IsTerminal(),
		Publishable:     ver.Status().IsPublishable(),
		AllowedActions:  allowedActions,
		ApprovalSummary: approvalSummary,
	}, nil
}

func (s *QueryService) GetVersionApprovals(ctx context.Context, query GetVersionApprovalsQuery, userID string) (dto *VersionApprovalsDTO, err error) {
	defer err2.Handle(&err)

	ver := try.To1(s.versionRepo.GetByID(ctx, query.ID))

	// Check authorization
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))
	allowed := try.To1(s.authz.CanAccessDocument(ctx, userID, doc.ID().String()))
	if !allowed {
		return nil, errors.ErrForbidden
	}

	// Get approvals
	approvals := try.To1(s.approvalRepo.ListByVersionID(ctx, ver.ID()))

	// Enrich with user information
	userIDs := make([]string, 0, len(approvals))
	for _, appr := range approvals {
		userIDs = append(userIDs, appr.UserID())
	}

	userMap := make(map[string]string) // map[userID]userName
	for _, userID := range userIDs {
		userUUID := try.To1(uuid.Parse(userID))
		user := try.To1(s.userRepo.GetByID(ctx, userUUID))
		userMap[userID] = user.Name()
	}

	// Build approval DTOs
	approvalDTOs := make([]ApprovalDTO, 0, len(approvals))
	auditTrail := make([]ApprovalAuditDTO, 0, len(approvals))

	for _, appr := range approvals {
		userName := ""
		if user, ok := userMap[appr.UserID()]; ok {
			userName = user
		}

		approvalDTOs = append(approvalDTOs, ApprovalDTO{
			ID:        appr.ID(),
			UserID:    appr.UserID(),
			UserName:  userName,
			Role:      appr.Role(),
			Approved:  appr.Approved(),
			Comment:   appr.Comment(),
			CreatedAt: appr.CreatedAt(),
		})

		action := "rejected"
		if appr.Approved() {
			action = "approved"
		}

		auditTrail = append(auditTrail, ApprovalAuditDTO{
			UserID:    appr.UserID(),
			UserName:  userName,
			Role:      appr.Role(),
			Action:    action,
			Comment:   appr.Comment(),
			Timestamp: appr.CreatedAt(),
		})
	}

	// Get approval summary
	approvalSummary := try.To1(s.getApprovalSummary(ctx, ver.ID(), doc.DocumentType()))

	return &VersionApprovalsDTO{
		VersionID:  ver.ID(),
		Version:    ver.Version().String(),
		Status:     ver.Status(),
		Approvals:  approvalDTOs,
		Summary:    approvalSummary,
		AuditTrail: auditTrail,
	}, nil
}

func (s *QueryService) getApprovalSummary(ctx context.Context, versionID uuid.UUID, documentType string) (ApprovalSummaryDTO, error) {
	policy, err := s.policyFactory.GetPolicy(documentType)
	if err != nil {
		return ApprovalSummaryDTO{}, err
	}

	required := policy.RequiredApprovals()
	approvals, err := s.approvalRepo.ListByVersionID(ctx, versionID)
	if err != nil {
		return ApprovalSummaryDTO{}, err
	}

	received := make(map[approval.ApprovalRole]int)
	userRoles := make(map[string]map[approval.ApprovalRole]bool)

	for _, appr := range approvals {
		if !appr.Approved() {
			continue
		}

		// Track which roles this user has already approved with
		if userRoles[appr.UserID()] == nil {
			userRoles[appr.UserID()] = make(map[approval.ApprovalRole]bool)
		}

		// Same user cannot satisfy multiple roles
		if len(userRoles[appr.UserID()]) > 0 {
			continue
		}

		userRoles[appr.UserID()][appr.Role()] = true
		received[appr.Role()]++
	}

	remaining := make(map[approval.ApprovalRole]int)
	for role, count := range required {
		diff := count - received[role]
		if diff > 0 {
			remaining[role] = diff
		}
	}

	complete := policy.IsSatisfied(approvals)

	return ApprovalSummaryDTO{
		Required:  required,
		Received:  received,
		Remaining: remaining,
		Complete:  complete,
	}, nil
}
