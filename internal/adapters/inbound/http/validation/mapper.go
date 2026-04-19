package validation

import (
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
)

// toValidateDefinitionInput maps form data to application input
func toValidateDefinitionInput(actor appValidation.Actor, content []byte, documentType string) appValidation.ValidateDefinitionInput {
	return appValidation.ValidateDefinitionInput{
		Actor:        actor,
		Content:      content,
		DocumentType: documentType,
	}
}

// toValidateDefinitionResponse maps application output to HTTP response
func toValidateDefinitionResponse(output appValidation.ValidateDefinitionOutput) ValidateDefinitionResponse {
	issues := make([]ValidationIssue, 0, len(output.Issues))
	for _, issue := range output.Issues {
		issues = append(issues, ValidationIssue{
			Field:    issue.Field,
			Rule:     issue.Rule,
			Message:  issue.Message,
			Severity: issue.Severity,
		})
	}

	return ValidateDefinitionResponse{
		Valid:  output.Valid,
		Issues: issues,
	}
}
