package validation

import (
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
)

// toValidateContentInput maps form data to application input
func toValidateContentInput(actor appValidation.Actor, content []byte, documentType string) appValidation.ValidateContentInput {
	return appValidation.ValidateContentInput{
		Actor:        actor,
		Content:      content,
		DocumentType: documentType,
	}
}

// toValidateContentResponse maps application output to HTTP response
func toValidateContentResponse(output appValidation.ValidateContentOutput) ValidateContentResponse {
	issues := make([]ValidationIssue, 0, len(output.Issues))
	for _, issue := range output.Issues {
		issues = append(issues, ValidationIssue{
			Field:    issue.Field,
			Rule:     issue.Rule,
			Message:  issue.Message,
			Severity: issue.Severity,
		})
	}

	return ValidateContentResponse{
		Valid:  output.Valid,
		Issues: issues,
	}
}
