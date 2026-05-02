package bootstrap

import (
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/publisher"
	"github.com/bbridges_11/document-registry/internal/adapters/outbound/publisher/mock"
	"go.uber.org/zap"
)

// wirePublishers creates and registers all publishers
func wirePublishers(infra *infrastructure, log *zap.Logger) *publisher.DestinationRegistry {
	registry := publisher.NewDestinationRegistry()

	// Register Mock Publisher (for testing/development)
	mockPublisher := mock.NewPublisherAdapter(log)
	if err := registry.Register("mock", mockPublisher); err != nil {
		log.Fatal("failed to register mock publisher", zap.Error(err))
	}

	log.Info("registered publishers",
		zap.Int("count", registry.NumberOfPublishers()),
		zap.Strings("destinations", []string{"mock", "dev-portal", "s3", "sns"}),
	)

	return registry
}
