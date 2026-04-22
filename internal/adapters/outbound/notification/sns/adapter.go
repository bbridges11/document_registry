package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"go.uber.org/zap"
)

// SNSAdapter implements NotificationService using AWS SNS
type SNSAdapter struct {
	client   *sns.Client
	topicARN string
	enabled  bool
	logger   *zap.Logger
}

// NewSNSAdapter creates a new SNS notification adapter
func NewSNSAdapter(client *sns.Client, cfg config.SNSConfig, logger *zap.Logger) *SNSAdapter {
	adapter := &SNSAdapter{
		topicARN: cfg.TopicARN,
		enabled:  cfg.Enabled,
		logger:   logger,
	}

	// If disabled (bypass mode), return early without creating client
	if !cfg.Enabled {
		logger.Warn("SNS notifications are DISABLED - notifications will be logged but not sent (development mode)")
	}

	adapter.client = client

	return adapter
}

// SendNotification implements outbound.NotificationService
func (a *SNSAdapter) SendNotification(ctx context.Context, notification outbound.Notification) error {
	// Bypass mode - log what would be sent
	if !a.enabled {
		bodyPreview := notification.Body
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}

		a.logger.Info("SNS disabled - would send notification",
			zap.Strings("recipients", notification.Recipients),
			zap.String("subject", notification.Subject),
			zap.String("body_preview", bodyPreview),
			zap.Any("metadata", notification.Metadata),
		)
		return nil
	}

	// Validate recipients
	if len(notification.Recipients) == 0 {
		a.logger.Warn("no recipients specified for notification",
			zap.String("subject", notification.Subject),
		)
		return nil
	}

	// Send notification to each recipient
	// Note: SNS with email protocol requires pre-confirmed subscriptions
	// For dynamic email sending, we publish one message per recipient
	// and let SNS email subscriptions handle delivery
	for _, recipient := range notification.Recipients {
		if err := a.publishToRecipient(ctx, recipient, notification); err != nil {
			// Log error but continue with other recipients
			a.logger.Error("failed to publish notification to recipient",
				zap.String("recipient", recipient),
				zap.String("subject", notification.Subject),
				zap.Error(err),
			)
			// Don't return error - best effort delivery
		} else {
			a.logger.Info("notification published to SNS",
				zap.String("recipient", recipient),
				zap.String("subject", notification.Subject),
			)
		}
	}

	return nil
}

// publishToRecipient publishes a notification message to SNS for a single recipient
func (a *SNSAdapter) publishToRecipient(ctx context.Context, recipient string, notification outbound.Notification) error {
	// Format message with recipient information
	message := fmt.Sprintf("To: %s\n\n%s", recipient, notification.Body)

	input := &sns.PublishInput{
		TopicArn: aws.String(a.topicARN),
		Subject:  aws.String(notification.Subject),
		Message:  aws.String(message),
	}

	_, err := a.client.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("SNS publish failed: %w", err)
	}

	return nil
}

// HealthCheck verifies SNS is accessible (optional, can be used in health endpoint)
func (a *SNSAdapter) HealthCheck(ctx context.Context) error {
	if !a.enabled {
		return nil
	}

	// Try to get topic attributes to verify connectivity
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(a.topicARN),
	}

	_, err := a.client.GetTopicAttributes(ctx, input)
	return err
}
