package approval

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/google/uuid"
)

// GrantApprovalCommand represents the command to grant an approval
type GrantApprovalCommand struct {
	VersionID uuid.UUID
	UserID    string
	Role      approval.ApprovalRole
	Comment   string
}

// RevokeApprovalCommand represents the command to revoke an approval
type RevokeApprovalCommand struct {
	VersionID uuid.UUID
	UserID    string
}

// GetApprovalQuery represents the query to get a specific approval
type GetApprovalQuery struct {
	ID uuid.UUID
}

// ListApprovalsByVersionQuery represents the query to list approvals for a version
type ListApprovalsByVersionQuery struct {
	VersionID uuid.UUID
}

// ApprovalDTO represents an approval data transfer object
type ApprovalDTO struct {
	ID        uuid.UUID
	VersionID uuid.UUID
	UserID    string
	UserName  string
	Role      approval.ApprovalRole
	Approved  bool
	Comment   string
	CreatedAt time.Time
}

// ApprovalSummaryDTO represents approval summary for a version
type ApprovalSummaryDTO struct {
	VersionID uuid.UUID
	Required  map[approval.ApprovalRole]int
	Received  map[approval.ApprovalRole]int
	Remaining map[approval.ApprovalRole]int
	Complete  bool
}
