package subscription

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// ServiceImpl implements Service
type ServiceImpl struct {
	userRepo   outbound.UserRepository
	snsService outbound.SNSService
}

// NewService creates a new subscription service
func NewService(userRepo outbound.UserRepository, snsService outbound.SNSService) *ServiceImpl {
	return &ServiceImpl{
		userRepo:   userRepo,
		snsService: snsService,
	}
}

// Subscribe subscribes a user's email to notifications
func (s *ServiceImpl) Subscribe(ctx context.Context, input SubscribeInput) (err error) {
	defer err2.Handle(&err)

	// Get user to extract email
	user := try.To1(s.userRepo.GetByID(ctx, input.UserID))

	if user.Email() == "" {
		return errors.New(errors.CodeInvalidArgument, "user has no email address")
	}

	// Check if already subscribed
	isSubscribed := try.To1(s.snsService.IsSubscribed(ctx, user.Email()))
	if isSubscribed {
		// Idempotent - already subscribed is success
		return nil
	}

	// Subscribe to SNS topic
	try.To1(s.snsService.Subscribe(ctx, user.Email()))

	return nil
}

// Unsubscribe removes a user's email subscription
func (s *ServiceImpl) Unsubscribe(ctx context.Context, input UnsubscribeInput) (err error) {
	defer err2.Handle(&err)

	// Get user to extract email
	user := try.To1(s.userRepo.GetByID(ctx, input.UserID))

	if user.Email() == "" {
		return errors.New(errors.CodeInvalidArgument, "user has no email address")
	}

	// Unsubscribe from SNS topic
	try.To(s.snsService.Unsubscribe(ctx, user.Email()))

	return nil
}

// ListSubscriptions returns all current subscriptions
func (s *ServiceImpl) ListSubscriptions(ctx context.Context) (views []SubscriptionView, err error) {
	defer err2.Handle(&err)

	subscriptions := try.To1(s.snsService.ListSubscriptions(ctx))

	views = make([]SubscriptionView, len(subscriptions))
	for i, sub := range subscriptions {
		views[i] = SubscriptionView{
			Email:  sub.Endpoint,
			Status: sub.Status,
		}
	}

	return views, nil
}
