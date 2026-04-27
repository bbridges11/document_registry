package outbound

import "context"

// PublisherService defines operations for publishing version content to external systems
type PublisherService interface {
	// Publish sends version content to an external system
	Publish(ctx context.Context, request PublishRequest) error
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

	// Destination is where to publish ("dev-portal", "prod-gateway", etc.)
	Destination string
}
