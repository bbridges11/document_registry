package outbound

import "context"

// DocumentValidator validates content for a specific document type
// Each document type can have its own validator implementation
type DocumentValidator interface {
	// Validate validates document content and returns validation result
	Validate(ctx context.Context, content []byte) (*ValidationResult, error)

	// SupportedType returns the document type code this validator supports
	SupportedType() string
}

// ValidatorRegistry manages document validators by type
type ValidatorRegistry interface {
	// GetValidator returns a validator for the given document type
	// Returns nil if no validator is registered for the type
	GetValidator(documentType string) DocumentValidator

	// Register adds a validator to the registry
	Register(validator DocumentValidator)

	// HasValidator checks if a validator exists for the given document type
	HasValidator(documentType string) bool
}
