package validation

// ValidateDefinitionOutput contains the validation result
type ValidateDefinitionOutput struct {
	Valid  bool
	Issues []ValidationIssue
}

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Field    string `json:"field"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}
