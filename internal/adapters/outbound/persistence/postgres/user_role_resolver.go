package postgres

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/google/uuid"
)

// UserRoleResolverAdapter maps user roles to approval roles
type UserRoleResolverAdapter struct{}

// NewUserRoleResolverAdapter creates a new UserRoleResolverAdapter
func NewUserRoleResolverAdapter() *UserRoleResolverAdapter {
	return &UserRoleResolverAdapter{}
}

// GetApprovalRoles returns the approval roles available to a user based on their user role
// No database query needed - role is passed from request context
func (a *UserRoleResolverAdapter) GetApprovalRoles(ctx context.Context, userID uuid.UUID, userRole string) ([]approval.ApprovalRole, error) {
	// Map user role to approval roles
	switch userRole {
	case "admin":
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
	case "engineer", "contributor":
		// Engineers/Contributors can perform technical approvals
		return []approval.ApprovalRole{
			approval.ApprovalRoleTechnical,
		}, nil
	default:
		// Viewers or unknown roles have no approval permissions
		return []approval.ApprovalRole{}, nil
	}
}
