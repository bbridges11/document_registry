package approval

import (
	"context"

	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type CommandService struct {
	approvalRepo  outbound.ApprovalRepository
	versionRepo   outbound.VersionRepository
	documentRepo  outbound.DocumentRepository
	userRepo      outbound.UserRepository
	policyFactory *domainApproval.PolicyFactory
}

func NewCommandService(
	approvalRepo outbound.ApprovalRepository,
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	userRepo outbound.UserRepository,
	policyFactory *domainApproval.PolicyFactory,
) *CommandService {
	return &CommandService{
		approvalRepo:  approvalRepo,
		versionRepo:   versionRepo,
		documentRepo:  documentRepo,
		userRepo:      userRepo,
		policyFactory: policyFactory,
	}
}

func (s *CommandService) GrantApproval(ctx context.Context, cmd GrantApprovalCommand) (err error) {
	defer err2.Handle(&err)

	// Get version to validate and get document type
	ver := try.To1(s.versionRepo.GetByID(ctx, cmd.VersionID))
	doc := try.To1(s.documentRepo.GetByID(ctx, ver.DocumentID()))

	// Get policy for document type
	policy := try.To1(s.policyFactory.GetPolicy(doc.DocumentType()))

	// Get user to determine approval roles
	userID := try.To1(uuid.Parse(cmd.UserID))
	user := try.To1(s.userRepo.GetByID(ctx, userID))

	// Map user role to approval roles
	var userRoles []domainApproval.ApprovalRole
	switch user.Role() {
	case "admin":
		userRoles = []domainApproval.ApprovalRole{
			domainApproval.ApprovalRoleTechnical,
			domainApproval.ApprovalRoleArchitect,
			domainApproval.ApprovalRoleProduct,
		}
	case "contributor":
		userRoles = []domainApproval.ApprovalRole{
			domainApproval.ApprovalRoleTechnical,
		}
	default:
		userRoles = []domainApproval.ApprovalRole{}
	}

	// Get existing approvals
	existingApprovals := try.To1(s.approvalRepo.ListByVersionID(ctx, cmd.VersionID))

	// Create new approval
	newApproval := try.To1(domainApproval.NewApproval(cmd.VersionID, cmd.UserID, cmd.Role, true, cmd.Comment))

	// Validate approval against policy
	try.To(policy.ValidateApproval(newApproval, ver.CreatedBy(), userRoles, existingApprovals))

	// Save approval
	try.To(s.approvalRepo.Save(ctx, newApproval))

	return nil
}

func (s *CommandService) RevokeApproval(ctx context.Context, cmd RevokeApprovalCommand) (err error) {
	defer err2.Handle(&err)

	// Get existing approval
	existingApproval := try.To1(s.approvalRepo.GetByVersionIDAndUserID(ctx, cmd.VersionID, cmd.UserID))

	if !existingApproval.Approved() {
		return errors.New(errors.CodeInvalidState, "approval already revoked or rejected")
	}

	// Note: In this implementation, we create a new approval record with approved=false
	// rather than deleting the existing one to maintain audit trail
	revokedApproval := try.To1(domainApproval.NewApproval(
		cmd.VersionID,
		cmd.UserID,
		existingApproval.Role(),
		false,
		"Approval revoked",
	))

	try.To(s.approvalRepo.Save(ctx, revokedApproval))

	return nil
}
