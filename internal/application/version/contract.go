package version

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type Actor struct{ UserID string }

type CreateInput struct {
	Actor      Actor
	DocumentID uuid.UUID
	Version    string
	Content    io.Reader
	Metadata   map[string]any
}

type UpdateInput struct {
	Actor    Actor
	ID       uuid.UUID
	Metadata map[string]any
}

type GetInput struct {
	Actor Actor
	ID    uuid.UUID
}
type ListByDocumentInput struct {
	Actor      Actor
	DocumentID uuid.UUID
}
type SubmitInput struct {
	Actor Actor
	ID    uuid.UUID
}
type ReviewInput struct {
	Actor Actor
	ID    uuid.UUID
}
type ApproveInput struct {
	Actor   Actor
	ID      uuid.UUID
	Comment string
}
type RejectInput struct {
	Actor  Actor
	ID     uuid.UUID
	Reason string
}
type PublishInput struct {
	Actor       Actor
	ID          uuid.UUID
	Destination string
}
type GetStatusInput struct {
	Actor Actor
	ID    uuid.UUID
}
type GetApprovalsInput struct {
	Actor Actor
	ID    uuid.UUID
}

type VersionView struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Version     string
	Status      string
	ContentKey  string
	ContentHash string
	Metadata    map[string]any
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ApprovalSummaryView struct {
	Required  map[string]int
	Received  map[string]int
	Remaining map[string]int
	Complete  bool
}

type ApprovalView struct {
	ID        uuid.UUID
	UserID    string
	UserName  string
	Role      string
	Approved  bool
	Comment   string
	CreatedAt time.Time
}

type ApprovalAuditView struct {
	UserID    string
	UserName  string
	Role      string
	Action    string
	Comment   string
	Timestamp time.Time
}

type VersionStatusView struct {
	ID              uuid.UUID
	Version         string
	Status          string
	Editable        bool
	Terminal        bool
	Publishable     bool
	AllowedActions  []string
	ApprovalSummary ApprovalSummaryView
}

type VersionApprovalsView struct {
	VersionID  uuid.UUID
	Version    string
	Status     string
	Approvals  []ApprovalView
	Summary    ApprovalSummaryView
	AuditTrail []ApprovalAuditView
}

type CreateOutput struct {
	Version          VersionView
	ValidationResult *ValidationResultView
}

type UpdateOutput struct {
	// Metadata-only update - no validation result needed
}

type ValidationResultView struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

type ValidationIssue struct {
	Field    string `json:"field"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (CreateOutput, error)
	Get(ctx context.Context, input GetInput) (VersionView, error)
	ListByDocument(ctx context.Context, input ListByDocumentInput) ([]VersionView, error)
	Update(ctx context.Context, input UpdateInput) error
	Submit(ctx context.Context, input SubmitInput) error
	Review(ctx context.Context, input ReviewInput) error
	Approve(ctx context.Context, input ApproveInput) error
	Reject(ctx context.Context, input RejectInput) error
	Publish(ctx context.Context, input PublishInput) error
	GetStatus(ctx context.Context, input GetStatusInput) (VersionStatusView, error)
	GetApprovals(ctx context.Context, input GetApprovalsInput) (VersionApprovalsView, error)
}
