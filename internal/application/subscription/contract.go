package subscription

import (
	"context"

	"github.com/google/uuid"
)

// Service defines subscription operations
type Service interface {
	// Subscribe subscribes a user's email to notifications
	Subscribe(ctx context.Context, input SubscribeInput) error

	// Unsubscribe removes a user's email subscription
	Unsubscribe(ctx context.Context, input UnsubscribeInput) error

	// ListSubscriptions returns all current subscriptions
	ListSubscriptions(ctx context.Context) ([]SubscriptionView, error)
}

// SubscribeInput is the input for subscribing a user
type SubscribeInput struct {
	UserID uuid.UUID
}

// UnsubscribeInput is the input for unsubscribing a user
type UnsubscribeInput struct {
	UserID uuid.UUID
}

// SubscriptionView is the output model for subscriptions
type SubscriptionView struct {
	Email  string `json:"email"`
	Status string `json:"status"` // "PendingConfirmation" or "Confirmed"
}
