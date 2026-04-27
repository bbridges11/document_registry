package bootstrap

import (
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/platform/events/handlers"
	"go.uber.org/zap"
)

// registerNotificationHandlers registers event handlers for notifications
func registerNotificationHandlers(infra *infrastructure, log *zap.Logger) {
	log.Info("registering notification event handlers")

	// Create notification event handler
	log.Info("THIS IS WHERE NOTFICATIONS ARE REGISTERED wire_notification.go")
	notificationHandler := handlers.NewNotificationEventHandler(
		infra.NotificationService,
		infra.UserRepo,
		infra.VersionRepo,
		infra.DocumentRepo,
		log,
	)

	// Subscribe to events
	infra.EventBus.Subscribe(events.EventVersionApproved, notificationHandler.HandleVersionApproved)
	infra.EventBus.Subscribe(events.EventVersionFullyApproved, notificationHandler.HandleVersionFullyApproved)

	log.Info("notification event handlers registered successfully")
}
