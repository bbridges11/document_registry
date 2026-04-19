package validation

// ValidateDefinitionInput contains the input for definition validation
type ValidateDefinitionInput struct {
	Actor        Actor
	Content      []byte
	DocumentType string
}
