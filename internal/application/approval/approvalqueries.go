package approval

import (
	"context"

	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type QueryService struct {
	approvalRepo  outbound.ApprovalRepository
	versionRepo   outbound.VersionRepository
	documentRepo  outbound.DocumentRepository
	userRepo      outbound.UserRepository
	policyFactory *domainApproval.PolicyFactory
}

func NewQueryService(
	approvalRepo outbound.ApprovalRepository,
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	userRepo outbound.UserRepository,
	policyFactory *domainApproval.PolicyFactory,
) *QueryService {
	return &QueryService{
		approvalRepo:  approvalRepo,
		versionRepo:   versionRepo,
		documentRepo:  documentRepo,
		userRepo:      userRepo,
		policyFactory: policyFactory,
	}
}

func (s *QueryService) GetApproval(ctx context.Context, query GetApprovalQuery) (dto *ApprovalDTO, err error) {
	defer err2.Handle(&err)

	appr := try.To1(s.approvalRepo.GetByID(ctx, query.ID))

	// Enrich with user information
	userID := try.To1(uuid.Parse(appr.UserID()))
	user := try.To1(s.userRepo.GetByID(ctx, userID))

	return &ApprovalDTO{
		ID:        appr.ID(),
		VersionID: appr.VersionID(),
		UserID:    appr.UserID(),
		UserName:  user.Name(),
		Role:      appr.Role(),
		Approved:  appr.Approved(),
		Comment:   appr.Comment(),
		CreatedAt: appr.CreatedAt(),
	}, nil
}

func (s *QueryService) ListApprovalsByVersion(ctx context.Context, query ListApprovalsByVersionQuery) (dtos []*ApprovalDTO, err error) {
	defer err2.Handle(&err)

	approvals := try.To1(s.approvalRepo.ListByVersionID(ctx, query.VersionID))

	// Enrich with user information
	userIDs := make([]string, 0, len(approvals))
	for _, appr := range approvals {
		userIDs = append(userIDs, appr.UserID())
	}

	// Fetch users individually and build map
	userMap := make(map[string]string) // map[userID]userName
	for _, userID := range userIDs {
		userUUID := try.To1(uuid.Parse(userID))
		user := try.To1(s.userRepo.GetByID(ctx, userUUID))
		userMap[userID] = user.Name()
	}

	dtos = make([]*ApprovalDTO, 0, len(approvals))
	for _, appr := range approvals {
		userName := ""
		if name, ok := userMap[appr.UserID()]; ok {
			userName = name
		}

		dtos = append(dtos, &ApprovalDTO{
			ID:        appr.ID(),
			VersionID: appr.VersionID(),
			UserID:    appr.UserID(),
			UserName:  userName,
			Role:      appr.Role(),
			Approved:  appr.Approved(),
			Comment:   appr.Comment(),
			CreatedAt: appr.CreatedAt(),
		})
	}

	return dtos, nil
}

func (s *QueryService) GetApprovalSummary(ctx context.Context, versionID uuid.UUID) (summary *ApprovalSummaryDTO, err error) {
	defer err2.Handle(&err)

	// Get version and document to determine policy
	ver := try.To1(s.versionRepo.GetByID(ctx, versionID))
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))

	policy := try.To1(s.policyFactory.GetPolicy(doc.DocumentType().ApprovalPolicy()))
	approvals := try.To1(s.approvalRepo.ListByVersionID(ctx, versionID))

	required := policy.RequiredApprovals()
	received := make(map[domainApproval.ApprovalRole]int)
	userRoles := make(map[string]map[domainApproval.ApprovalRole]bool)

	for _, appr := range approvals {
		if !appr.Approved() {
			continue
		}

		// Track which roles this user has already approved with
		if userRoles[appr.UserID()] == nil {
			userRoles[appr.UserID()] = make(map[domainApproval.ApprovalRole]bool)
		}

		// Same user cannot satisfy multiple roles
		if len(userRoles[appr.UserID()]) > 0 {
			continue
		}

		userRoles[appr.UserID()][appr.Role()] = true
		received[appr.Role()]++
	}

	remaining := make(map[domainApproval.ApprovalRole]int)
	for role, count := range required {
		diff := count - received[role]
		if diff > 0 {
			remaining[role] = diff
		}
	}

	complete := policy.IsSatisfied(approvals)

	return &ApprovalSummaryDTO{
		VersionID: versionID,
		Required:  required,
		Received:  received,
		Remaining: remaining,
		Complete:  complete,
	}, nil
}
