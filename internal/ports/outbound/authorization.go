package outbound

import "context"

type AuthorizationService interface {
	// Check methods
	Check(ctx context.Context, userID, object, relation string) (bool, error)
	CanAccessDocument(ctx context.Context, userID, documentID string) (bool, error)
	CanModifyStakeholders(ctx context.Context, userID, documentID string) (bool, error)
	CanCreateVersion(ctx context.Context, userID, documentID string) (bool, error)
	CanOverwriteVersion(ctx context.Context, userID, versionID string) (bool, error)
	CanSubmitVersion(ctx context.Context, userID, versionID string) (bool, error)
	CanReviewVersion(ctx context.Context, userID, versionID string) (bool, error)
	CanApproveVersion(ctx context.Context, userID, versionID string) (bool, error)
	CanRejectVersion(ctx context.Context, userID, versionID string) (bool, error)
	CanPublishVersion(ctx context.Context, userID, versionID string) (bool, error)

	// Deprecation authorization
	CanRequestDeprecation(ctx context.Context, userID, versionID string) (bool, error)
	CanApproveDeprecation(ctx context.Context, userID, versionID string) (bool, error)
	CanCancelDeprecation(ctx context.Context, userID, versionID, requestedBy string) (bool, error)

	// Tuple management
	GrantAccess(ctx context.Context, userID, object, relation string) error
	RevokeAccess(ctx context.Context, userID, object, relation string) error

	// Health check
	HealthCheck(ctx context.Context) error
}
