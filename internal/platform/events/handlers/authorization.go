package handlers

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"go.uber.org/zap"
)

// AuthorizationEventHandler handles domain events to create/delete authorization tuples
type AuthorizationEventHandler struct {
	authz  outbound.AuthorizationService
	logger *zap.Logger
}

func NewAuthorizationEventHandler(authz outbound.AuthorizationService, logger *zap.Logger) *AuthorizationEventHandler {
	return &AuthorizationEventHandler{
		authz:  authz,
		logger: logger,
	}
}

// HandleDocumentCreated creates owner tuple when document is created
func (h *AuthorizationEventHandler) HandleDocumentCreated(envelope events.Envelope) {
	event, ok := envelope.Event.(events.DocumentCreated)
	if !ok {
		h.logger.Error("invalid event type for HandleDocumentCreated")
		return
	}

	// Grant owner access to creator
	object := fmt.Sprintf("document:%s", event.DocumentID)
	userID := fmt.Sprintf("user:%s", event.CreatedBy)

	if err := h.authz.GrantAccess(context.Background(), userID, object, "owner"); err != nil {
		h.logger.Error("failed to grant owner access",
			zap.String("document_id", event.DocumentID),
			zap.String("user_id", event.CreatedBy),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("granted owner access for document",
		zap.String("document_id", event.DocumentID),
		zap.String("user_id", event.CreatedBy),
	)
}

// HandleStakeholderAdded creates role tuple when stakeholder is added
func (h *AuthorizationEventHandler) HandleStakeholderAdded(envelope events.Envelope) {
	event, ok := envelope.Event.(events.StakeholderAdded)
	if !ok {
		h.logger.Error("invalid event type for HandleStakeholderAdded")
		return
	}

	// Grant role access to stakeholder
	object := fmt.Sprintf("document:%s", event.DocumentID)
	userID := fmt.Sprintf("user:%s", event.UserID)
	relation := event.Role // owner, contributor, consumer

	if err := h.authz.GrantAccess(context.Background(), userID, object, relation); err != nil {
		h.logger.Error("failed to grant stakeholder access",
			zap.String("document_id", event.DocumentID),
			zap.String("user_id", event.UserID),
			zap.String("role", event.Role),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("granted stakeholder access",
		zap.String("document_id", event.DocumentID),
		zap.String("user_id", event.UserID),
		zap.String("role", event.Role),
	)
}

// HandleStakeholderRemoved deletes role tuple when stakeholder is removed
func (h *AuthorizationEventHandler) HandleStakeholderRemoved(envelope events.Envelope) {
	event, ok := envelope.Event.(events.StakeholderRemoved)
	if !ok {
		h.logger.Error("invalid event type for HandleStakeholderRemoved")
		return
	}

	// Revoke role access from stakeholder
	object := fmt.Sprintf("document:%s", event.DocumentID)
	userID := fmt.Sprintf("user:%s", event.UserID)
	relation := event.Role // Must be provided by event

	if err := h.authz.RevokeAccess(context.Background(), userID, object, relation); err != nil {
		h.logger.Error("failed to revoke stakeholder access",
			zap.String("document_id", event.DocumentID),
			zap.String("user_id", event.UserID),
			zap.String("role", event.Role),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("revoked stakeholder access",
		zap.String("document_id", event.DocumentID),
		zap.String("user_id", event.UserID),
		zap.String("role", event.Role),
	)
}

// HandleVersionPublished creates public tuple when version is published
func (h *AuthorizationEventHandler) HandleVersionPublished(envelope events.Envelope) {
	event, ok := envelope.Event.(events.VersionPublished)
	if !ok {
		h.logger.Error("invalid event type for HandleVersionPublished")
		return
	}

	// Grant public access when version is published
	object := fmt.Sprintf("document:%s", event.DocumentID)

	// Create tuple: user:* published document:X (public access)
	if err := h.authz.GrantAccess(context.Background(), "user:*", object, "published"); err != nil {
		h.logger.Error("failed to grant public access",
			zap.String("document_id", event.DocumentID),
			zap.String("version_id", event.VersionID),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("granted public access for published version",
		zap.String("document_id", event.DocumentID),
		zap.String("version_id", event.VersionID),
		zap.String("version", event.Version),
	)
}

// HandleVersionCreated creates parent_document relationship for version
func (h *AuthorizationEventHandler) HandleVersionCreated(envelope events.Envelope) {
	event, ok := envelope.Event.(events.VersionCreated)
	if !ok {
		h.logger.Error("invalid event type for HandleVersionCreated")
		return
	}

	// Create tuple: version:X parent_document document:Y
	// This allows version to inherit permissions from document
	versionObject := fmt.Sprintf("version:%s", event.VersionID)
	documentObject := fmt.Sprintf("document:%s", event.DocumentID)

	if err := h.authz.GrantAccess(context.Background(), documentObject, versionObject, "parent_document"); err != nil {
		h.logger.Error("failed to create version-document relationship",
			zap.String("version_id", event.VersionID),
			zap.String("document_id", event.DocumentID),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("created version-document relationship",
		zap.String("version_id", event.VersionID),
		zap.String("document_id", event.DocumentID),
	)
}
