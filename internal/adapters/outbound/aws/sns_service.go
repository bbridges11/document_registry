package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// SNSServiceAdapter implements outbound.SNSService
type SNSServiceAdapter struct {
	client   *sns.Client
	topicArn string
}

// NewSNSServiceAdapter creates a new SNS service adapter
func NewSNSServiceAdapter(client *sns.Client, cfg config.SNSConfig) *SNSServiceAdapter {
	return &SNSServiceAdapter{
		client:   client,
		topicArn: cfg.TopicARN,
	}
}

// Subscribe subscribes an email to the SNS topic
func (s *SNSServiceAdapter) Subscribe(ctx context.Context, email string) (subscriptionArn string, err error) {
	defer err2.Handle(&err)

	if s.topicArn == "" {
		return "", errors.New(errors.CodeInternal, "SNS topic ARN not configured")
	}

	input := &sns.SubscribeInput{
		Protocol: aws.String("email"),
		TopicArn: aws.String(s.topicArn),
		Endpoint: aws.String(email),
	}

	result := try.To1(s.client.Subscribe(ctx, input))

	return aws.ToString(result.SubscriptionArn), nil
}

// Unsubscribe removes an email subscription
func (s *SNSServiceAdapter) Unsubscribe(ctx context.Context, email string) (err error) {
	defer err2.Handle(&err)

	if s.topicArn == "" {
		return errors.New(errors.CodeInternal, "SNS topic ARN not configured")
	}

	// Find subscription ARN by email
	subscriptions := try.To1(s.ListSubscriptions(ctx))

	var targetArn string
	for _, sub := range subscriptions {
		if sub.Endpoint == email {
			targetArn = sub.SubscriptionArn
			break
		}
	}

	if targetArn == "" {
		return errors.New(errors.CodeNotFound, fmt.Sprintf("subscription not found for email: %s", email))
	}

	// Unsubscribe
	input := &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(targetArn),
	}

	try.To1(s.client.Unsubscribe(ctx, input))

	return nil
}

// ListSubscriptions returns all subscriptions for the topic
func (s *SNSServiceAdapter) ListSubscriptions(ctx context.Context) (subs []outbound.SNSSubscription, err error) {
	defer err2.Handle(&err)

	if s.topicArn == "" {
		return nil, errors.New(errors.CodeInternal, "SNS topic ARN not configured")
	}

	input := &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(s.topicArn),
	}

	result := try.To1(s.client.ListSubscriptionsByTopic(ctx, input))

	subscriptions := []outbound.SNSSubscription{}
	for _, sub := range result.Subscriptions {
		// Only include email subscriptions
		if aws.ToString(sub.Protocol) == "email" {
			status := "Confirmed"
			if aws.ToString(sub.SubscriptionArn) == "PendingConfirmation" {
				status = "PendingConfirmation"
			}

			subscriptions = append(subscriptions, outbound.SNSSubscription{
				SubscriptionArn: aws.ToString(sub.SubscriptionArn),
				Endpoint:        aws.ToString(sub.Endpoint),
				Protocol:        aws.ToString(sub.Protocol),
				Status:          status,
			})
		}
	}

	// Handle pagination if needed
	for result.NextToken != nil {
		input.NextToken = result.NextToken
		result = try.To1(s.client.ListSubscriptionsByTopic(ctx, input))

		for _, sub := range result.Subscriptions {
			if aws.ToString(sub.Protocol) == "email" {
				status := "Confirmed"
				subArn := aws.ToString(sub.SubscriptionArn)
				if subArn == "PendingConfirmation" {
					status = "PendingConfirmation"
				}

				subscriptions = append(subscriptions, outbound.SNSSubscription{
					SubscriptionArn: subArn,
					Endpoint:        aws.ToString(sub.Endpoint),
					Protocol:        aws.ToString(sub.Protocol),
					Status:          status,
				})
			}
		}
	}

	return subscriptions, nil
}

// IsSubscribed checks if an email is already subscribed
func (s *SNSServiceAdapter) IsSubscribed(ctx context.Context, email string) (subscribed bool, err error) {
	defer err2.Handle(&err)

	subscriptions := try.To1(s.ListSubscriptions(ctx))

	for _, sub := range subscriptions {
		if sub.Endpoint == email {
			return true, nil
		}
	}

	return false, nil
}
