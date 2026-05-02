package mock

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"go.uber.org/zap"
)

// PublisherAdapter is a mock implementation of PublisherService
// Always returns success and logs what would be published
type PublisherAdapter struct {
	logger *zap.Logger
}

// NewPublisherAdapter creates a new mock publisher adapter
func NewPublisherAdapter(logger *zap.Logger) *PublisherAdapter {
	return &PublisherAdapter{
		logger: logger,
	}
}

// Publish simulates publishing content to an external system
// Always succeeds and logs the request details
func (p *PublisherAdapter) Publish(ctx context.Context, request outbound.PublishRequest) error {
	p.logger.Info("mock publisher: would publish version",
		zap.String("document_id", request.DocumentID),
		zap.String("version_number", request.VersionNumber),
		zap.String("destination", request.Destination),
		zap.String("environment", request.Environment),
		zap.Int("content_size", len(request.Content)),
		zap.Any("metadata", request.Metadata),
	)

	// Simulate successful publish
	// In production, this would call an external API
	return nil
}

// SupportedEnvironments returns all standard environments for mock
func (p *PublisherAdapter) SupportedEnvironments() []string {
	return []string{"dev", "staging", "prod"}
}

// ValidateEnvironment always returns nil for mock (accepts all environments)
func (p *PublisherAdapter) ValidateEnvironment(environment string) error {
	return nil
}
