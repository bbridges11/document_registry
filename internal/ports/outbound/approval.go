package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/google/uuid"
)

type ApprovalRepository interface {
	Save(ctx context.Context, appr *approval.Approval) error
	GetByID(ctx context.Context, id uuid.UUID) (*approval.Approval, error)
	ListByVersionID(ctx context.Context, versionID uuid.UUID) ([]*approval.Approval, error)
	GetByVersionIDAndUserID(ctx context.Context, versionID uuid.UUID, userID string) (*approval.Approval, error)
	ExistsByVersionIDAndUserID(ctx context.Context, versionID uuid.UUID, userID string) (bool, error)
}
