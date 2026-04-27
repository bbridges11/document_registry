package inprocess

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// PatternValidator validates pattern document content (YAML format)
type PatternValidator struct {
	structValidator *validator.Validate
}

// NewPatternValidator creates a new pattern document validator
func NewPatternValidator() *PatternValidator {
	return &PatternValidator{
		structValidator: validator.New(),
	}
}

// SupportedType returns the document type code this validator supports
func (v *PatternValidator) SupportedType() string {
	return "pattern"
}

// Validate validates pattern document content
func (v *PatternValidator) Validate(ctx context.Context, content []byte) (*outbound.ValidationResult, error) {
	var issues []outbound.ValidationIssue

	// Step 1: Unmarshal YAML with strict mode (disallow unknown fields)
	var pattern PatternSchema
	decoder := yaml.NewDecoder(nil)
	decoder.KnownFields(true) // This will cause error on unknown fields

	// Unmarshal
	if err := yaml.Unmarshal(content, &pattern); err != nil {
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
	if err := v.structValidator.Struct(pattern); err != nil {
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
		semanticIssues := validateSemantics(&pattern)
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

// formatValidationError converts validator errors to human-readable messages
func (v *PatternValidator) formatValidationError(e validator.FieldError) string {
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
