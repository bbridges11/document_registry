package bootstrap

import (
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/platform/events/handlers"
	"go.uber.org/zap"
)

func registerEventHandlers(infra *infrastructure, log *zap.Logger) {
	log.Info("registering authorization event handlers")
	authzHandler := handlers.NewAuthorizationEventHandler(infra.AuthzService, log)
	infra.EventBus.Subscribe(events.EventDocumentCreated, authzHandler.HandleDocumentCreated)
	infra.EventBus.Subscribe(events.EventStakeholderAdded, authzHandler.HandleStakeholderAdded)
	infra.EventBus.Subscribe(events.EventStakeholderRemoved, authzHandler.HandleStakeholderRemoved)
	infra.EventBus.Subscribe(events.EventVersionCreated, authzHandler.HandleVersionCreated)
	infra.EventBus.Subscribe(events.EventVersionPublished, authzHandler.HandleVersionPublished)
}
