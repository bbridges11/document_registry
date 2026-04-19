package deprecation

import "github.com/google/uuid"

type RequestDeprecationCommand struct {
	VersionID       uuid.UUID
	RequestedBy     string
	Reason          string
	DeprecationNote string
}

type ApproveDeprecationCommand struct {
	VersionID  uuid.UUID
	ApprovedBy string
	Role       string
	Comment    string
}

type RejectDeprecationCommand struct {
	VersionID  uuid.UUID
	RejectedBy string
	Reason     string
}

type CancelDeprecationCommand struct {
	VersionID  uuid.UUID
	CanceledBy string
}
