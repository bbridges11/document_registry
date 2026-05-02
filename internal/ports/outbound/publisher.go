package outbound

import "context"

// PublisherService defines operations for publishing version content to external systems
type PublisherService interface {
	// Publish sends version content to a specific environment of this destination
	Publish(ctx context.Context, request PublishRequest) error

	// SupportedEnvironments returns which environments are configured for this publisher
	SupportedEnvironments() []string

	// ValidateEnvironment checks if environment is valid and enabled for this publisher
	ValidateEnvironment(environment string) error
}

// PublishRequest contains data needed to publish a version
type PublishRequest struct {
	// DocumentID is the document identifier
	DocumentID string

	// VersionNumber is the semantic version number
	VersionNumber string

	// Content is the actual content to publish (YAML, JSON, etc.)
	Content []byte

	// Metadata contains additional context about the version
	Metadata map[string]any

	// Destination is the publisher type ("dev-portal", "api-gateway", etc.)
	Destination string

	// Environment is the target environment ("dev", "staging", "prod")
	Environment string
}
