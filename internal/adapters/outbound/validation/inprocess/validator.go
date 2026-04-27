package inprocess

import (
	"context"
	"fmt"
	"time"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/lainio/err2"
)

const (
	validationTimeout = 30 * time.Second
)

// ContentValidator validates document content using a registry of type-specific validators
type ContentValidator struct {
	registry outbound.ValidatorRegistry
}

// NewContentValidator creates a new content validator with a validator registry
func NewContentValidator(registry outbound.ValidatorRegistry) *ContentValidator {
	return &ContentValidator{
		registry: registry,
	}
}

// Validate validates document content based on document type
func (v *ContentValidator) Validate(ctx context.Context, req outbound.ValidationRequest) (result *outbound.ValidationResult, err error) {
	defer err2.Handle(&err)

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, validationTimeout)
	defer cancel()

	// Check if validator exists for this document type
	validator := v.registry.GetValidator(req.DocumentType)
	if validator == nil {
		// No validator registered = valid by default (document type doesn't require validation)
		return &outbound.ValidationResult{
			Valid:  true,
			Issues: []outbound.ValidationIssue{},
		}, nil
	}

	// Channel for validation result
	resultChan := make(chan *outbound.ValidationResult, 1)
	errChan := make(chan error, 1)

	// Run validation in goroutine to respect timeout
	go func() {
		res, err := validator.Validate(ctx, req.Content)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("validation timeout after %v", validationTimeout)
	case err := <-errChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

// ShouldValidate determines if a validator is registered for the given document type
func (v *ContentValidator) ShouldValidate(documentType string) bool {
	return v.registry.HasValidator(documentType)
}
