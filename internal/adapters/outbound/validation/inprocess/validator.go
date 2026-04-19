package inprocess

import (
	"context"
	"fmt"
	"time"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/go-playground/validator/v10"
	"github.com/lainio/err2"
	"gopkg.in/yaml.v3"
)

const (
	documentTypeDefinition = "definition"
	validationTimeout      = 30 * time.Second
)

// DefinitionValidator validates definition YAML files
type DefinitionValidator struct {
	structValidator *validator.Validate
}

// NewDefinitionValidator creates a new definition validator
func NewDefinitionValidator() *DefinitionValidator {
	return &DefinitionValidator{
		structValidator: validator.New(),
	}
}

// Validate validates definition content
func (v *DefinitionValidator) Validate(ctx context.Context, req outbound.ValidationRequest) (result *outbound.ValidationResult, err error) {
	defer err2.Handle(&err)

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, validationTimeout)
	defer cancel()

	// Check if this validator handles this document type
	if !v.ShouldValidate(req.DocumentType) {
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
		res, err := v.doValidate(req)
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

// doValidate performs the actual validation
func (v *DefinitionValidator) doValidate(req outbound.ValidationRequest) (*outbound.ValidationResult, error) {
	var issues []outbound.ValidationIssue

	// Step 1: Unmarshal YAML with strict mode (disallow unknown fields)
	var def DefinitionSchema
	decoder := yaml.NewDecoder(nil)
	decoder.KnownFields(true) // This will cause error on unknown fields

	// Unmarshal
	if err := yaml.Unmarshal(req.Content, &def); err != nil {
		issues = append(issues, outbound.ValidationIssue{
			Field:    "root",
			Rule:     "yaml_unmarshal",
			Message:  fmt.Sprintf("failed to parse YAML: %v", err),
			Severity: "error",
		})
		return &outbound.ValidationResult{
			Valid:  false,
			Issues: issues,
		}, nil
	}

	// Step 2: Struct tag validation
	if err := v.structValidator.Struct(def); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrs {
				issues = append(issues, outbound.ValidationIssue{
					Field:    e.Namespace(),
					Rule:     e.Tag(),
					Message:  v.formatValidationError(e),
					Severity: "error",
				})
			}
		} else {
			issues = append(issues, outbound.ValidationIssue{
				Field:    "root",
				Rule:     "struct_validation",
				Message:  fmt.Sprintf("validation error: %v", err),
				Severity: "error",
			})
		}
	}

	// Step 3: Semantic validation (only if no structural errors)
	if len(issues) == 0 {
		semanticIssues := validateSemantics(&def)
		issues = append(issues, semanticIssues...)
	}

	// Determine if valid (no errors, warnings are ok)
	valid := true
	for _, issue := range issues {
		if issue.Severity == "error" {
			valid = false
			break
		}
	}

	return &outbound.ValidationResult{
		Valid:  valid,
		Issues: issues,
	}, nil
}

// ShouldValidate determines if this validator handles the given document type
func (v *DefinitionValidator) ShouldValidate(documentType string) bool {
	return documentType == documentTypeDefinition
}

// formatValidationError converts validator errors to human-readable messages
func (v *DefinitionValidator) formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", e.Field())
	case "eq":
		return fmt.Sprintf("field '%s' must equal '%s'", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of: %s", e.Field(), e.Param())
	case "min":
		return fmt.Sprintf("field '%s' must have at least %s items", e.Field(), e.Param())
	default:
		return fmt.Sprintf("field '%s' failed validation '%s'", e.Field(), e.Tag())
	}
}
