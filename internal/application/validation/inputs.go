package validation

// ValidateContentInput contains the input for content validation
type ValidateContentInput struct {
	Actor        Actor
	Content      []byte
	DocumentType string
}
