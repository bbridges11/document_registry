package handlers

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationEventHandler handles domain events to send email notifications
type NotificationEventHandler struct {
	notificationService outbound.NotificationService
	userRepo            outbound.UserRepository
	versionRepo         outbound.VersionRepository
	documentRepo        outbound.DocumentRepository
	logger              *zap.Logger
}

// NewNotificationEventHandler creates a new notification event handler
func NewNotificationEventHandler(
	notificationService outbound.NotificationService,
	userRepo outbound.UserRepository,
	versionRepo outbound.VersionRepository,
	documentRepo outbound.DocumentRepository,
	logger *zap.Logger,
) *NotificationEventHandler {
	return &NotificationEventHandler{
		notificationService: notificationService,
		userRepo:            userRepo,
		versionRepo:         versionRepo,
		documentRepo:        documentRepo,
		logger:              logger,
	}
}

// HandleVersionApproved sends notification when a version is approved
func (h *NotificationEventHandler) HandleVersionApproved(envelope events.Envelope) {
	event, ok := envelope.Event.(events.VersionApproved)
	if !ok {
		h.logger.Error("invalid event type for HandleVersionApproved")
		return
	}

	// Run in background (non-blocking)
	go func() {
		ctx := context.Background()

		// Determine recipients - notify the version author
		recipients, err := h.getVersionApprovedRecipients(ctx, event)
		if err != nil {
			h.logger.Error("failed to determine notification recipients",
				zap.String("version_id", event.VersionID),
				zap.Error(err),
			)
			return
		}

		if len(recipients) == 0 {
			h.logger.Warn("no recipients found for version approved notification",
				zap.String("version_id", event.VersionID),
			)
			return
		}

		// Build email content
		subject, body, err := h.buildVersionApprovedEmail(ctx, event)
		if err != nil {
			h.logger.Error("failed to build notification email",
				zap.String("version_id", event.VersionID),
				zap.Error(err),
			)
			return
		}

		// Create notification
		notification := outbound.Notification{
			Recipients: recipients,
			Subject:    subject,
			Body:       body,
			Metadata: map[string]string{
				"event_type":  "version.approved",
				"version_id":  event.VersionID,
				"document_id": event.DocumentID,
			},
		}

		// Send notification (best effort - log errors but don't fail)
		if err := h.notificationService.SendNotification(ctx, notification); err != nil {
			h.logger.Error("failed to send version approved notification",
				zap.String("version_id", event.VersionID),
				zap.Strings("recipients", recipients),
				zap.Error(err),
			)
			return
		}

		h.logger.Info("version approved notification sent",
			zap.String("version_id", event.VersionID),
			zap.Strings("recipients", recipients),
		)
	}()
}

// getVersionApprovedRecipients determines who should be notified about version approval
func (h *NotificationEventHandler) getVersionApprovedRecipients(ctx context.Context, event events.VersionApproved) ([]string, error) {
	// Get version to find the author
	versionID, err := uuid.Parse(event.VersionID)
	if err != nil {
		return nil, fmt.Errorf("invalid version ID: %w", err)
	}

	version, err := h.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	// Get author user to get their email
	authorID, err := uuid.Parse(version.CreatedBy())
	if err != nil {
		return nil, fmt.Errorf("invalid author ID: %w", err)
	}

	author, err := h.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version author: %w", err)
	}

	// Return author's email
	return []string{author.Email()}, nil
}

// buildVersionApprovedEmail builds the email subject and body for version approval
func (h *NotificationEventHandler) buildVersionApprovedEmail(ctx context.Context, event events.VersionApproved) (string, string, error) {
	// Get version details
	versionID, err := uuid.Parse(event.VersionID)
	if err != nil {
		return "", "", fmt.Errorf("invalid version ID: %w", err)
	}

	version, err := h.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get version: %w", err)
	}

	// Get document details
	document, err := h.documentRepo.GetByID(ctx, version.DocumentID())
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %w", err)
	}

	// Get approver details (from event)
	approverID, err := uuid.Parse(event.ApprovedBy)
	if err != nil {
		return "", "", fmt.Errorf("invalid approver ID: %w", err)
	}

	approver, err := h.userRepo.GetByID(ctx, approverID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get approver: %w", err)
	}

	// Build subject
	subject := fmt.Sprintf("Version Approved: %s v%s", document.Name(), version.Version())

	// Build HTML body
	comment := ""
	if comment == "" {
		comment = "<em>No comment provided</em>"
	}

	body := fmt.Sprintf(`
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    h2 { color: #2c5282; }
    .info { background-color: #f7fafc; padding: 15px; border-radius: 5px; margin: 20px 0; }
    .info p { margin: 5px 0; }
    .button { 
      display: inline-block; 
      padding: 10px 20px; 
      background-color: #4299e1; 
      color: white; 
      text-decoration: none; 
      border-radius: 5px; 
      margin-top: 15px;
    }
    .footer { 
      margin-top: 30px; 
      padding-top: 20px; 
      border-top: 1px solid #e2e8f0; 
      color: #718096; 
      font-size: 12px; 
    }
  </style>
</head>
<body>
  <h2>🎉 Version Approved</h2>
  
  <p>Great news! Version <strong>%s</strong> of document <strong>%s</strong> has been approved.</p>
  
  <div class="info">
    <p><strong>Document:</strong> %s</p>
    <p><strong>Version:</strong> %s</p>
    <p><strong>Approved by:</strong> %s</p>
    <p><strong>Approval role:</strong> %s</p>
    <p><strong>Comment:</strong> %s</p>
  </div>
  
  <p>Your version is one step closer to being published!</p>
  
  <div class="footer">
    <p>This is an automated notification from Document Registry.</p>
    <p>Version ID: %s</p>
  </div>
</body>
</html>
`,
		version.Version(),
		document.Name(),
		document.Name(),
		version.Version(),
		approver.Name(),
		event.Role,
		comment,
		event.VersionID,
	)

	return subject, body, nil
}

// HandleVersionFullyApproved sends notification when a version is fully approved (policy satisfied)
func (h *NotificationEventHandler) HandleVersionFullyApproved(envelope events.Envelope) {
	event, ok := envelope.Event.(events.VersionFullyApproved)
	if !ok {
		h.logger.Error("invalid event type for HandleVersionFullyApproved")
		return
	}

	// Run in background (non-blocking)
	go func() {
		ctx := context.Background()

		// Determine recipients - notify product users
		recipients, err := h.getProductUserEmails(ctx)
		if err != nil {
			h.logger.Error("failed to determine notification recipients",
				zap.String("version_id", event.VersionID),
				zap.Error(err),
			)
			return
		}

		if len(recipients) == 0 {
			h.logger.Warn("no product users found for version fully approved notification",
				zap.String("version_id", event.VersionID),
			)
			return
		}

		// Build email content
		subject, body, err := h.buildVersionFullyApprovedEmail(ctx, event)
		if err != nil {
			h.logger.Error("failed to build notification email",
				zap.String("version_id", event.VersionID),
				zap.Error(err),
			)
			return
		}

		// Create notification
		notification := outbound.Notification{
			Recipients: recipients,
			Subject:    subject,
			Body:       body,
			Metadata: map[string]string{
				"event_type":  "version.fully_approved",
				"version_id":  event.VersionID,
				"document_id": event.DocumentID,
			},
		}

		// Send notification (best effort - log errors but don't fail)
		if err := h.notificationService.SendNotification(ctx, notification); err != nil {
			h.logger.Error("failed to send version fully approved notification",
				zap.String("version_id", event.VersionID),
				zap.Strings("recipients", recipients),
				zap.Error(err),
			)
			return
		}

		h.logger.Info("version fully approved notification sent",
			zap.String("version_id", event.VersionID),
			zap.Strings("recipients", recipients),
		)
	}()
}

// getProductUserEmails retrieves emails of all users with product role
func (h *NotificationEventHandler) getProductUserEmails(ctx context.Context) ([]string, error) {
	// For now, we'll need to implement a ListByRole method on UserRepository
	// As a workaround, we can return a placeholder
	// TODO: Implement UserRepository.ListByRole("product")

	// Placeholder - in production, query all users with role="product"
	// users, err := h.userRepo.ListByRole(ctx, "product")
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get product users: %w", err)
	// }
	//
	// emails := make([]string, 0, len(users))
	// for _, user := range users {
	//     emails = append(emails, user.Email())
	// }
	// return emails, nil

	// For now, log that we need this feature
	h.logger.Warn("getProductUserEmails not fully implemented - needs UserRepository.ListByRole")
	return []string{}, nil
}

// buildVersionFullyApprovedEmail builds the email for fully approved versions
func (h *NotificationEventHandler) buildVersionFullyApprovedEmail(ctx context.Context, event events.VersionFullyApproved) (string, string, error) {
	// Get version details
	versionID, err := uuid.Parse(event.VersionID)
	if err != nil {
		return "", "", fmt.Errorf("invalid version ID: %w", err)
	}

	version, err := h.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get version: %w", err)
	}

	// Get document details
	document, err := h.documentRepo.GetByID(ctx, version.DocumentID())
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %w", err)
	}

	// Build subject
	subject := fmt.Sprintf("Ready to Publish: %s v%s", document.Name(), version.Version().String())

	// Build HTML body
	body := fmt.Sprintf(`
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    h2 { color: #2c5282; }
    .info { background-color: #f7fafc; padding: 15px; border-radius: 5px; margin: 20px 0; }
    .info p { margin: 5px 0; }
    .action { 
      background-color: #48bb78; 
      color: white; 
      padding: 12px 24px; 
      text-align: center; 
      border-radius: 5px; 
      margin: 20px 0;
      font-weight: bold;
    }
    .footer { 
      margin-top: 30px; 
      padding-top: 20px; 
      border-top: 1px solid #e2e8f0; 
      color: #718096; 
      font-size: 12px; 
    }
  </style>
</head>
<body>
  <h2>✅ Version Ready to Publish</h2>
  
  <p>Version <strong>%s</strong> of document <strong>%s</strong> has received all required approvals and is ready to be published.</p>
  
  <div class="info">
    <p><strong>Document:</strong> %s</p>
    <p><strong>Version:</strong> %s</p>
    <p><strong>Status:</strong> APPROVED</p>
  </div>
  
  <div class="action">
    As a product team member, you can now publish this version.
  </div>
  
  <p>To publish this version, use the publish endpoint with your desired destination.</p>
  
  <div class="footer">
    <p>This is an automated notification from Document Registry.</p>
    <p>Version ID: %s</p>
  </div>
</body>
</html>
`,
		version.Version().String(),
		document.Name(),
		document.Name(),
		version.Version().String(),
		event.VersionID,
	)

	return subject, body, nil
}
