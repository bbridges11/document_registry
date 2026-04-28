package postgres

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
)

// UserRoleResolverAdapter maps user roles to approval roles
type UserRoleResolverAdapter struct {
	userRepo outbound.UserRepository
}

// NewUserRoleResolverAdapter creates a new UserRoleResolverAdapter
func NewUserRoleResolverAdapter(userRepo outbound.UserRepository) *UserRoleResolverAdapter {
	return &UserRoleResolverAdapter{userRepo: userRepo}
}

// GetApprovalRoles returns the approval roles available to a user based on their user role
func (a *UserRoleResolverAdapter) GetApprovalRoles(ctx context.Context, userID uuid.UUID) ([]approval.ApprovalRole, error) {
	u, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Map user role to approval roles
	switch u.Role() {
	case user.UserRoleAdmin:
		// Admin users can perform all approval roles
		return []approval.ApprovalRole{
			approval.ApprovalRoleTechnical,
			approval.ApprovalRoleArchitect,
			approval.ApprovalRoleAdmin,
		}, nil
	case "architect":
		// Architects can perform architect approvals
		return []approval.ApprovalRole{
			approval.ApprovalRoleArchitect,
		}, nil
	case "engineer":
		// Engineers can perform technical approvals
		return []approval.ApprovalRole{
			approval.ApprovalRoleTechnical,
		}, nil
	default:
		// Product, viewers, or unknown roles have no approval permissions
		return []approval.ApprovalRole{}, nil
	}
}
