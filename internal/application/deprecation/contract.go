package deprecation

import (
	"context"
)

// UseCase defines the application boundary for deprecation operations
type UseCase interface {
	// RequestDeprecation initiates a manual deprecation request
	RequestDeprecation(ctx context.Context, input RequestDeprecationInput) (DeprecationOutput, error)

	// ApproveDeprecation approves a pending deprecation
	ApproveDeprecation(ctx context.Context, input ApproveDeprecationInput) error

	// RejectDeprecation rejects a pending deprecation
	RejectDeprecation(ctx context.Context, input RejectDeprecationInput) error

	// CancelDeprecation cancels a pending deprecation (by requestor or admin)
	CancelDeprecation(ctx context.Context, input CancelDeprecationInput) error

	// GetDeprecation retrieves deprecation details
	GetDeprecation(ctx context.Context, input GetDeprecationInput) (DeprecationOutput, error)

	// GetDeprecationStatus retrieves deprecation status with permission info
	GetDeprecationStatus(ctx context.Context, input GetDeprecationInput) (DeprecationStatusOutput, error)
}

// Actor represents the user performing an action
type Actor struct {
	UserID string
	Role   string // User role for authorization
}
