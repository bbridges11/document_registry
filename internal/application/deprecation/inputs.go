package deprecation

import "github.com/google/uuid"

// RequestDeprecationInput contains data for requesting version deprecation
type RequestDeprecationInput struct {
	Actor           Actor
	VersionID       uuid.UUID
	Reason          string
	DeprecationNote string
}

// ApproveDeprecationInput contains data for approving deprecation
type ApproveDeprecationInput struct {
	Actor     Actor
	VersionID uuid.UUID
	Role      string
	Comment   string
}

// RejectDeprecationInput contains data for rejecting deprecation
type RejectDeprecationInput struct {
	Actor     Actor
	VersionID uuid.UUID
	Reason    string
}

// CancelDeprecationInput contains data for canceling deprecation
type CancelDeprecationInput struct {
	Actor     Actor
	VersionID uuid.UUID
}

// GetDeprecationInput contains data for retrieving deprecation
type GetDeprecationInput struct {
	Actor     Actor
	VersionID uuid.UUID
}
