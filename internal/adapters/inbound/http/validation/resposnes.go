package validation

// ValidateContentResponse contains the validation result
type ValidateContentResponse struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Field    string `json:"field"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// ErrorResponse for validation endpoint errors
type ErrorResponse struct {
	Error string `json:"error"`
}
