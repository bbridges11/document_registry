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

	registerStakeholderHandlers(infra, log)

	log.Info("notification event handlers registered successfully")
}

// registerStakeholderHandlers registers event handlers for stakeholder operations
func registerStakeholderHandlers(infra *infrastructure, log *zap.Logger) {
	log.Info("registering stakeholder event handlers")

	// Create stakeholder event handler
	stakeholderHandler := handlers.NewStakeholderEventHandler(
		infra.StakeholderRepo,
		log,
	)

	// Subscribe to DocumentCreated event
	infra.EventBus.Subscribe(events.EventDocumentCreated, stakeholderHandler.HandleDocumentCreated)

	log.Info("stakeholder event handlers registered successfully")
}
