package handlers

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReviewerNotificationHandler handles ReviewersAssigned events and sends email notifications
type ReviewerNotificationHandler struct {
	userRepo            outbound.UserRepository
	notificationService outbound.NotificationService
	logger              *zap.Logger
}

func NewReviewerNotificationHandler(
	userRepo outbound.UserRepository,
	notificationService outbound.NotificationService,
	logger *zap.Logger,
) *ReviewerNotificationHandler {
	return &ReviewerNotificationHandler{
		userRepo:            userRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Handle processes ReviewersAssigned events
// Signature matches outbound.HandlerFunc: func(events.Envelope)
func (h *ReviewerNotificationHandler) Handle(envelope events.Envelope) {
	event, ok := envelope.Event.(events.ReviewersAssigned)
	if !ok {
		h.logger.Error("invalid event type for ReviewerNotificationHandler")
		return
	}

	h.logger.Info("handling reviewers assigned event",
		zap.String("version_id", event.VersionID),
		zap.Int("reviewer_count", len(event.ReviewerIDs)),
	)

	// Run in background (non-blocking)
	go func() {
		ctx := context.Background()

		// Build recipient list
		var recipients []string
		for _, userID := range event.ReviewerIDs {
			uid, err := uuid.Parse(userID)
			if err != nil {
				h.logger.Warn("invalid user ID in event", zap.String("user_id", userID), zap.Error(err))
				continue
			}

			user, err := h.userRepo.GetByID(ctx, uid)
			if err != nil {
				h.logger.Warn("user not found", zap.String("user_id", userID), zap.Error(err))
				continue
			}

			recipients = append(recipients, user.Email())
		}

		if len(recipients) == 0 {
			h.logger.Warn("no valid recipients for reviewer notification")
			return
		}

		// Build notification
		subject := fmt.Sprintf("Review Requested: %s v%s", event.DocumentName, event.VersionNum)
		body := fmt.Sprintf(
			`<html>
<body>
<p>You have been assigned to review a new version.</p>
<p><strong>Document:</strong> %s</p>
<p><strong>Version:</strong> %s</p>
<p><strong>Type:</strong> %s</p>
<p>Please review and approve when ready.</p>
</body>
</html>`,
			event.DocumentName,
			event.VersionNum,
			event.DocumentType,
		)

		notification := outbound.Notification{
			Recipients: recipients,
			Subject:    subject,
			Body:       body,
			Metadata: map[string]string{
				"version_id":  event.VersionID,
				"document_id": event.DocumentID,
				"event_type":  "reviewer_assigned",
			},
		}

		// Send notification
		if err := h.notificationService.SendNotification(ctx, notification); err != nil {
			h.logger.Error("failed to send reviewer notification",
				zap.Error(err),
				zap.String("version_id", event.VersionID),
			)
			return
		}

		h.logger.Info("reviewer notifications sent",
			zap.String("version_id", event.VersionID),
			zap.Int("recipient_count", len(recipients)),
		)
	}()
}
