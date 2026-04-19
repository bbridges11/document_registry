package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/google/uuid"
)

type StakeholderRepository interface {
	Save(ctx context.Context, sh *stakeholder.Stakeholder) error
	GetByDocumentIDAndUserID(ctx context.Context, documentID uuid.UUID, userID string) (*stakeholder.Stakeholder, error)
	ListByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*stakeholder.Stakeholder, error)
	Delete(ctx context.Context, documentID uuid.UUID, userID string) error
	ExistsByDocumentIDAndUserID(ctx context.Context, documentID uuid.UUID, userID string) (bool, error)
}
