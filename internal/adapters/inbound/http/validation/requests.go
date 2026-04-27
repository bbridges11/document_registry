package validation

// ValidateContentRequest represents HTTP request for validating document content
type ValidateContentRequest struct {
	Content      string `json:"content" validate:"required"`       // Base64 encoded content
	DocumentType string `json:"document_type" validate:"required"` // Document type code (e.g., "pattern")
}
