package outbound

import "context"

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendNotification(ctx context.Context, notification Notification) error
}

// Notification represents a notification to be sent
type Notification struct {
	// Recipients is the list of email addresses to send the notification to
	Recipients []string

	// Subject is the email subject line
	Subject string

	// Body is the email body content (HTML formatted)
	Body string

	// Metadata contains optional metadata for tracking and logging
	Metadata map[string]string
}
