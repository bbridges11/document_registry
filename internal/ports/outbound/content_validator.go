package outbound

import "context"

// ContentValidator validates document content based on document type
type ContentValidator interface {
	// Validate validates content and returns structured validation result
	Validate(ctx context.Context, req ValidationRequest) (*ValidationResult, error)

	// ShouldValidate determines if validation is required for the given document type
	ShouldValidate(documentType string) bool
}

// ValidationRequest contains content to be validated
type ValidationRequest struct {
	DocumentType string
	Content      []byte
}

// ValidationResult contains the outcome of validation
type ValidationResult struct {
	Valid  bool
	Issues []ValidationIssue
}

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Field    string
	Rule     string
	Message  string
	Severity string // "error" or "warning"
}
