package outbound

import "context"

// SNSSubscription represents an SNS subscription
type SNSSubscription struct {
	SubscriptionArn string
	Endpoint        string // email address
	Protocol        string // "email"
	Status          string // "PendingConfirmation", "Confirmed"
}

// SNSService defines operations for managing SNS subscriptions
type SNSService interface {
	// Subscribe subscribes an email to the configured SNS topic
	// Returns subscription ARN (pending confirmation)
	Subscribe(ctx context.Context, email string) (subscriptionArn string, err error)

	// Unsubscribe removes an email subscription from the SNS topic
	Unsubscribe(ctx context.Context, email string) error

	// ListSubscriptions returns all subscriptions for the configured SNS topic
	ListSubscriptions(ctx context.Context) ([]SNSSubscription, error)

	// IsSubscribed checks if an email is already subscribed to the topic
	IsSubscribed(ctx context.Context, email string) (bool, error)
}
