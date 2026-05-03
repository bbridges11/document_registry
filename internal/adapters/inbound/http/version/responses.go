package version

import (
	"time"

	"github.com/google/uuid"
)

type VersionResponse struct {
	ID           uuid.UUID      `json:"id"`
	DocumentID   uuid.UUID      `json:"document_id"`
	Version      string         `json:"version"`
	Status       string         `json:"status"`
	ContentS3Key string         `json:"content_s3_key"`
	ContentHash  string         `json:"content_hash"`
	Content      string         `json:"content,omitempty"`      // Base64 encoded content (only if include_content=true)
	ContentType  string         `json:"content_type,omitempty"` // MIME type (only if include_content=true)
	Metadata     map[string]any `json:"metadata"`
	CreatedBy    string         `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
type VersionStatusResponse struct {
	ID              uuid.UUID               `json:"id"`
	Version         string                  `json:"version"`
	Status          string                  `json:"status"`
	Editable        bool                    `json:"editable"`
	Terminal        bool                    `json:"terminal"`
	Publishable     bool                    `json:"publishable"`
	AllowedActions  []string                `json:"allowed_actions"`
	ApprovalSummary ApprovalSummaryResponse `json:"approval_summary"`
}
type ApprovalSummaryResponse struct {
	Required  map[string]int `json:"required"`
	Received  map[string]int `json:"received"`
	Remaining map[string]int `json:"remaining"`
	Complete  bool           `json:"complete"`
}
type VersionApprovalsResponse struct {
	VersionID  uuid.UUID               `json:"version_id"`
	Version    string                  `json:"version"`
	Status     string                  `json:"status"`
	Approvals  []ApprovalResponse      `json:"approvals"`
	Summary    ApprovalSummaryResponse `json:"summary"`
	AuditTrail []ApprovalAuditResponse `json:"audit_trail"`
}
type ApprovalResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Role      string    `json:"role"`
	Approved  bool      `json:"approved"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
type ApprovalAuditResponse struct {
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Role      string    `json:"role"`
	Action    string    `json:"action"`
	Comment   string    `json:"comment"`
	Timestamp time.Time `json:"timestamp"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateVersionResponse struct {
	Version          VersionResponse       `json:"version"`
	ValidationResult *ValidationResultView `json:"validation_result,omitempty"`
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
