package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/google/uuid"
)

// Deprecated: Use UserRepository instead
type User struct {
	ID    string
	Name  string
	Email string
	Roles []approval.ApprovalRole
}

// Deprecated: External user service - being replaced by UserRepository
type UserService interface {
	GetByID(ctx context.Context, userID string) (*User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]*User, error)
	GetApprovalRoles(ctx context.Context, userID string) ([]approval.ApprovalRole, error)
	HealthCheck(ctx context.Context) error
}

// UserRepository defines the outbound port for user persistence
type UserRepository interface {
	Save(ctx context.Context, user *user.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	ListAll(ctx context.Context) ([]*user.User, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
