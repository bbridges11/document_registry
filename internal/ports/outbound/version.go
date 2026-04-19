package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/google/uuid"
)

type VersionRepository interface {
	Save(ctx context.Context, ver *version.Version) error
	GetByID(ctx context.Context, id uuid.UUID) (*version.Version, error)
	GetByDocumentIDAndVersion(ctx context.Context, documentID uuid.UUID, version string) (*version.Version, error)
	ListByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*version.Version, error)
	GetLatestByDocumentID(ctx context.Context, documentID uuid.UUID) (*version.Version, error)
	GetPublishedByDocumentID(ctx context.Context, documentID uuid.UUID) (*version.Version, error)
	HasPublishedVersion(ctx context.Context, documentID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status workflow.Status) error
	ExistsByDocumentIDAndVersion(ctx context.Context, documentID uuid.UUID, version string) (bool, error)
}
