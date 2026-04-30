package handlers

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// StakeholderEventHandler handles domain events for stakeholder operations
type StakeholderEventHandler struct {
	stakeholderRepo outbound.StakeholderRepository
	logger          *zap.Logger
}

// NewStakeholderEventHandler creates a new stakeholder event handler
func NewStakeholderEventHandler(
	stakeholderRepo outbound.StakeholderRepository,
	logger *zap.Logger,
) *StakeholderEventHandler {
	return &StakeholderEventHandler{
		stakeholderRepo: stakeholderRepo,
		logger:          logger,
	}
}

// HandleDocumentCreated automatically adds the document creator as an owner stakeholder
func (h *StakeholderEventHandler) HandleDocumentCreated(envelope events.Envelope) {
	event, ok := envelope.Event.(events.DocumentCreated)
	if !ok {
		h.logger.Error("invalid event type for HandleDocumentCreated")
		return
	}

	// Run in background (async, non-blocking)
	go func() {
		ctx := context.Background()

		// Parse document ID
		documentID, err := uuid.Parse(event.DocumentID)
		if err != nil {
			h.logger.Error("failed to parse document ID in HandleDocumentCreated",
				zap.String("document_id", event.DocumentID),
				zap.Error(err),
			)
			return
		}

		// Parse creator user ID
		userID := event.CreatedBy
		if userID == "" {
			h.logger.Error("document created event has no creator user ID",
				zap.String("document_id", event.DocumentID),
			)
			return
		}

		// Check if user is already a stakeholder (idempotency)
		exists, err := h.stakeholderRepo.ExistsByDocumentIDAndUserID(ctx, documentID, userID)
		if err != nil {
			h.logger.Error("failed to check if stakeholder exists",
				zap.String("document_id", event.DocumentID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
			// Continue anyway - Save will fail if duplicate due to DB constraint
		}

		if exists {
			h.logger.Info("creator is already a stakeholder, skipping",
				zap.String("document_id", event.DocumentID),
				zap.String("user_id", userID),
			)
			return
		}

		// Create stakeholder with owner role
		sh, err := stakeholder.NewStakeholder(documentID, userID, stakeholder.RoleOwner)
		if err != nil {
			h.logger.Error("failed to create stakeholder entity",
				zap.String("document_id", event.DocumentID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
			return
		}

		// Save stakeholder
		if err := h.stakeholderRepo.Save(ctx, sh); err != nil {
			h.logger.Error("failed to save creator as stakeholder",
				zap.String("document_id", event.DocumentID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
			return
		}

		h.logger.Info("document creator automatically added as owner stakeholder",
			zap.String("document_id", event.DocumentID),
			zap.String("document_name", event.Name),
			zap.String("user_id", userID),
			zap.String("role", string(stakeholder.RoleOwner)),
		)
	}()
}
