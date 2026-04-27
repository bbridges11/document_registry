package version

// CreateVersionRequest represents HTTP request for creating a version
type CreateVersionRequest struct {
	Version     string         `json:"version" validate:"required"`
	Content     string         `json:"content" validate:"required"` // Base64 encoded content
	ContentType string         `json:"content_type"`
	Metadata    map[string]any `json:"metadata"`
}

// UpdateVersionRequest represents HTTP request for updating a version (metadata only)
type UpdateVersionRequest struct {
	Metadata map[string]any `json:"metadata"`
}

// SubmitVersionRequest represents HTTP request for submitting a version
type SubmitVersionRequest struct{}

// ReviewVersionRequest represents HTTP request for reviewing a version
type ReviewVersionRequest struct{}

// ApproveVersionRequest represents HTTP request for approving a version
type ApproveVersionRequest struct {
	Role    string `json:"role" form:"role"`
	Comment string `json:"comment" form:"comment"`
}

// RejectVersionRequest represents HTTP request for rejecting a version
type RejectVersionRequest struct {
	Reason string `json:"reason" form:"reason"`
}

// PublishVersionRequest represents HTTP request for publishing a version
type PublishVersionRequest struct {
	Destination string `json:"destination" validate:"required"`
}
