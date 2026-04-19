package validation

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// Service implements the validation use case
type Service struct {
	contentValidator outbound.ContentValidator
}

// NewService creates a new validation service
func NewService(contentValidator outbound.ContentValidator) UseCase {
	return &Service{
		contentValidator: contentValidator,
	}
}

// ValidateDefinition validates definition file content
func (s *Service) ValidateDefinition(ctx context.Context, input ValidateDefinitionInput) (output ValidateDefinitionOutput, err error) {
	defer err2.Handle(&err)

	// Call validator
	result := try.To1(s.contentValidator.Validate(ctx, outbound.ValidationRequest{
		DocumentType: input.DocumentType,
		Content:      input.Content,
	}))

	// Map to output
	output = ValidateDefinitionOutput{
		Valid:  result.Valid,
		Issues: make([]ValidationIssue, 0, len(result.Issues)),
	}

	for _, issue := range result.Issues {
		output.Issues = append(output.Issues, ValidationIssue{
			Field:    issue.Field,
			Rule:     issue.Rule,
			Message:  issue.Message,
			Severity: issue.Severity,
		})
	}

	return output, nil
}
