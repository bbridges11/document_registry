package version

import (
	"io"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/google/uuid"
)

type CreateVersionCommand struct {
	DocumentID  uuid.UUID
	Version     string
	Content     io.Reader
	ContentType string
	ContentHash string
	Metadata    map[string]any
	CreatedBy   string
}

type UpdateVersionCommand struct {
	ID          uuid.UUID
	Content     io.Reader
	ContentType string
	ContentHash string
	Metadata    map[string]any
	UpdatedBy   string
}

type SubmitVersionCommand struct {
	ID          uuid.UUID
	SubmittedBy string
}

type ReviewVersionCommand struct {
	ID         uuid.UUID
	ReviewedBy string
	UserRole   string // User role from Actor (avoids DB query)
}

type ApproveVersionCommand struct {
	ID         uuid.UUID
	ApprovedBy string
	UserRole   string // User role from Actor (avoids DB query)
	Comment    string
}

type RejectVersionCommand struct {
	ID         uuid.UUID
	RejectedBy string
	UserRole   string // User role from Actor (avoids DB query)
	Reason     string
}

type PublishVersionCommand struct {
	ID          uuid.UUID
	PublishedBy string
	Destination string // Where to publish (e.g., "dev-portal", "s3", "sns")
	Environment string // Which environment (e.g., "dev", "staging", "prod")
}

type GetVersionQuery struct {
	ID uuid.UUID
}

type ListVersionsByDocumentQuery struct {
	DocumentID uuid.UUID
}

type GetVersionStatusQuery struct {
	ID uuid.UUID
}

type GetVersionApprovalsQuery struct {
	ID uuid.UUID
}

type VersionDTO struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Version     string
	Status      workflow.Status
	ContentKey  string // Generic key - abstracted from storage backend
	ContentHash string
	Metadata    map[string]any
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type VersionStatusDTO struct {
	ID              uuid.UUID
	Version         string
	Status          workflow.Status
	Editable        bool
	Terminal        bool
	Publishable     bool
	AllowedActions  []workflow.Action
	ApprovalSummary ApprovalSummaryDTO
}

type ApprovalSummaryDTO struct {
	Required  map[approval.ApprovalRole]int
	Received  map[approval.ApprovalRole]int
	Remaining map[approval.ApprovalRole]int
	Complete  bool
}

type VersionApprovalsDTO struct {
	VersionID  uuid.UUID
	Version    string
	Status     workflow.Status
	Approvals  []ApprovalDTO
	Summary    ApprovalSummaryDTO
	AuditTrail []ApprovalAuditDTO
}

type ApprovalDTO struct {
	ID        uuid.UUID
	UserID    string
	UserName  string
	Role      approval.ApprovalRole
	Approved  bool
	Comment   string
	CreatedAt time.Time
}

type ApprovalAuditDTO struct {
	UserID    string
	UserName  string
	Role      approval.ApprovalRole
	Action    string
	Comment   string
	Timestamp time.Time
}
