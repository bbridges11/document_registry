package validation

import (
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

// Registry implements ValidatorRegistry interface
// Maps document type codes to their validators
type Registry struct {
	validators map[string]outbound.DocumentValidator
}

// NewRegistry creates a new validator registry with optional initial validators
func NewRegistry(validators ...outbound.DocumentValidator) *Registry {
	r := &Registry{
		validators: make(map[string]outbound.DocumentValidator),
	}

	// Register initial validators
	for _, v := range validators {
		r.Register(v)
	}

	return r
}

// GetValidator returns the validator for the given document type
// Returns nil if no validator is registered
func (r *Registry) GetValidator(documentType string) outbound.DocumentValidator {
	return r.validators[documentType]
}

// Register adds a validator to the registry
// If a validator for the same document type already exists, it will be replaced
func (r *Registry) Register(validator outbound.DocumentValidator) {
	r.validators[validator.SupportedType()] = validator
}

// HasValidator checks if a validator exists for the given document type
func (r *Registry) HasValidator(documentType string) bool {
	_, exists := r.validators[documentType]
	return exists
}
