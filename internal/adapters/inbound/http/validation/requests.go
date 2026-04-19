package validation

// ValidateDefinitionRequest is intentionally empty as we use multipart form
// The file content comes from the "file" field in the multipart form
// The document_type comes from the "document_type" field in the form
type ValidateDefinitionRequest struct {
	// Handled via multipart form parsing in handler
}
