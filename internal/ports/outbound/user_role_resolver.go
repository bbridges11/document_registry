package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/google/uuid"
)

// UserRoleResolver resolves user roles to approval roles
type UserRoleResolver interface {
	// GetApprovalRoles returns the approval roles available to a user based on their user role
	// userRole should be the user's role (e.g., "admin", "contributor", "viewer")
	GetApprovalRoles(ctx context.Context, userID uuid.UUID, userRole string) ([]approval.ApprovalRole, error)
}
