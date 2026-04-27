package document

// CreateDocumentRequest represents HTTP request for creating a document
type CreateDocumentRequest struct {
	Name         string         `json:"name" validate:"required"`
	Description  string         `json:"description"`
	Tags         []string       `json:"tags"`
	DocumentType string         `json:"document_type" validate:"required"`
	Version      string         `json:"version" validate:"required"`
	Content      string         `json:"content" validate:"required"` // Base64 encoded content
	Metadata     map[string]any `json:"metadata"`
}

type UpdateDocumentRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
