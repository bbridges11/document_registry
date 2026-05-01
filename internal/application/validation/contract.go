package validation

import "context"

// Actor represents the user performing validation
type Actor struct {
	UserID string
	Role   string // User role for authorization
}

// UseCase defines the validation use case interface
type UseCase interface {
	ValidateContent(ctx context.Context, input ValidateContentInput) (ValidateContentOutput, error)
}
