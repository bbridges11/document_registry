package openfga

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/zap"
)

type AuthorizationAdapter struct {
	client  *client.OpenFgaClient
	storeID string
	enabled bool
	logger  *zap.Logger
}

func NewAuthorizationAdapter(cfg config.OpenFGAConfig, logger *zap.Logger) (*AuthorizationAdapter, error) {
	adapter := &AuthorizationAdapter{
		storeID: cfg.StoreID,
		enabled: cfg.Enabled,
		logger:  logger,
	}

	// If disabled (bypass mode), return early
	if !cfg.Enabled {
		logger.Warn("OpenFGA authorization is DISABLED - all checks will pass (development mode)")
		return adapter, nil
	}

	// Initialize OpenFGA client using v0.8.0 SDK
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:  cfg.URL(),
		StoreId: cfg.StoreID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	adapter.client = fgaClient

	// Verify store exists
	if err := adapter.verifyStore(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify OpenFGA store: %w", err)
	}

	// Ensure authorization model exists
	if err := adapter.ensureModel(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure authorization model: %w", err)
	}

	logger.Info("OpenFGA authorization adapter initialized",
		zap.String("url", cfg.URL()),
		zap.String("store_id", cfg.StoreID),
	)

	return adapter, nil
}

func (a *AuthorizationAdapter) verifyStore(ctx context.Context) error {
	if !a.enabled {
		return nil
	}

	_, err := a.client.GetStore(ctx).Execute()
	if err != nil {
		return fmt.Errorf("store %s not found or inaccessible: %w", a.storeID, err)
	}

	return nil
}

func (a *AuthorizationAdapter) ensureModel(ctx context.Context, cfg config.OpenFGAConfig) error {
	if !a.enabled {
		return nil
	}

	loader := NewModelLoader(a.client, a.storeID, a.logger)

	// If ModelID is provided, verify it exists
	if cfg.ModelID != "" {
		if err := loader.VerifyModel(ctx, cfg.ModelID); err != nil {
			a.logger.Warn("configured model not found, will attempt to create new model",
				zap.String("model_id", cfg.ModelID),
				zap.Error(err),
			)
		} else {
			a.logger.Info("using configured authorization model",
				zap.String("model_id", cfg.ModelID),
			)
			return nil
		}
	}

	// Ensure model exists (create if necessary and auto-create is enabled)
	modelPath := "authorization-model.fga"
	modelID, err := loader.EnsureModel(ctx, modelPath, cfg.AutoCreate)
	if err != nil {
		return err
	}

	a.logger.Info("authorization model ready",
		zap.String("model_id", modelID),
		zap.Bool("auto_created", cfg.ModelID == ""),
	)

	return nil
}

// Check performs an authorization check
func (a *AuthorizationAdapter) Check(ctx context.Context, userID, object, relation string) (bool, error) {
	if !a.enabled {
		// Bypass mode - always return true
		return true, nil
	}

	body := client.ClientCheckRequest{
		User:     userID,
		Relation: relation,
		Object:   object,
	}

	response, err := a.client.Check(ctx).Body(body).Execute()
	if err != nil {
		a.logger.Error("OpenFGA check failed",
			zap.String("user", userID),
			zap.String("object", object),
			zap.String("relation", relation),
			zap.Error(err),
		)
		return false, fmt.Errorf("authorization check failed: %w", err)
	}

	allowed := response.GetAllowed()
	a.logger.Debug("authorization check completed",
		zap.String("user", userID),
		zap.String("object", object),
		zap.String("relation", relation),
		zap.Bool("allowed", allowed),
	)

	return allowed, nil
}

// GrantAccess creates a tuple granting access
func (a *AuthorizationAdapter) GrantAccess(ctx context.Context, userID, object, relation string) error {
	if !a.enabled {
		// Bypass mode - no-op
		a.logger.Debug("OpenFGA bypass mode - skipping grant",
			zap.String("user", userID),
			zap.String("object", object),
			zap.String("relation", relation),
		)
		return nil
	}

	body := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     userID,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := a.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("failed to grant access: %w", err)
	}

	a.logger.Debug("granted access",
		zap.String("user", userID),
		zap.String("object", object),
		zap.String("relation", relation),
	)

	return nil
}

// RevokeAccess deletes a tuple revoking access
func (a *AuthorizationAdapter) RevokeAccess(ctx context.Context, userID, object, relation string) error {
	if !a.enabled {
		// Bypass mode - no-op
		a.logger.Debug("OpenFGA bypass mode - skipping revoke",
			zap.String("user", userID),
			zap.String("object", object),
			zap.String("relation", relation),
		)
		return nil
	}

	body := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     userID,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := a.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("failed to revoke access: %w", err)
	}

	a.logger.Debug("revoked access",
		zap.String("user", userID),
		zap.String("object", object),
		zap.String("relation", relation),
	)

	return nil
}

// HealthCheck verifies OpenFGA is accessible
func (a *AuthorizationAdapter) HealthCheck(ctx context.Context) error {
	if !a.enabled {
		return nil
	}

	_, err := a.client.GetStore(ctx).Execute()
	return err
}

// CanAccessDocument checks if user can access a document
func (a *AuthorizationAdapter) CanAccessDocument(ctx context.Context, userID, documentID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("document:%s", documentID), "can_view")
}

// CanEditDocument checks if user can edit a document
func (a *AuthorizationAdapter) CanEditDocument(ctx context.Context, userID, documentID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("document:%s", documentID), "can_edit")
}

// CanDeleteDocument checks if user can delete a document
func (a *AuthorizationAdapter) CanDeleteDocument(ctx context.Context, userID, documentID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("document:%s", documentID), "can_delete")
}

// CanModifyStakeholders checks if user can manage stakeholders
func (a *AuthorizationAdapter) CanModifyStakeholders(ctx context.Context, userID, documentID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("document:%s", documentID), "can_manage_stakeholders")
}

// CanCreateVersion checks if user can create a version
func (a *AuthorizationAdapter) CanCreateVersion(ctx context.Context, userID, documentID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("document:%s", documentID), "can_create_version")
}

func (a *AuthorizationAdapter) CanOverwriteVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_edit")
}

func (a *AuthorizationAdapter) CanSubmitVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_submit")
}

func (a *AuthorizationAdapter) CanReviewVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_review")
}

func (a *AuthorizationAdapter) CanApproveVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_approve")
}

func (a *AuthorizationAdapter) CanRejectVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_approve")
}

func (a *AuthorizationAdapter) CanPublishVersion(ctx context.Context, userID, versionID string) (bool, error) {
	if !a.enabled {
		return true, nil
	}
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_publish")
}

// Deprecation authorization methods

// CanRequestDeprecation checks if user can request deprecation for a version
// Allowed: document creator, owners, admins
func (a *AuthorizationAdapter) CanRequestDeprecation(ctx context.Context, userID, versionID string) (bool, error) {
	// Bypass mode - allow all
	if !a.enabled {
		return true, nil
	}

	// Check if user can publish (same permissions for now)
	// In production, you might want a separate "can_deprecate" relation
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_publish")
}

// CanApproveDeprecation checks if user can approve/reject deprecation requests
// Allowed: users with "deprecation-approver" role, admins
func (a *AuthorizationAdapter) CanApproveDeprecation(ctx context.Context, userID, versionID string) (bool, error) {
	// Bypass mode - allow all
	if !a.enabled {
		return true, nil
	}

	// Check if user has approval rights
	// For now, use can_approve relation (same as version approval)
	// In production, you might want a separate "can_approve_deprecation" relation
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "can_approve")
}

// CanCancelDeprecation checks if user can cancel a pending deprecation request
// Allowed: original requestor, admins
func (a *AuthorizationAdapter) CanCancelDeprecation(ctx context.Context, userID, versionID, requestedBy string) (bool, error) {
	// Bypass mode - allow all
	if !a.enabled {
		return true, nil
	}

	// Allow if user is the original requestor
	if userID == requestedBy {
		return true, nil
	}

	// Otherwise check if user has admin/owner rights
	return a.Check(ctx, fmt.Sprintf("user:%s", userID), fmt.Sprintf("version:%s", versionID), "owner")
}
