package stakeholder

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/google/uuid"
)

type AddStakeholderCommand struct {
	DocumentID uuid.UUID
	UserID     string
	Role       stakeholder.Role
	AddedBy    string
}

type RemoveStakeholderCommand struct {
	DocumentID uuid.UUID
	UserID     string
	RemovedBy  string
}

type ListStakeholdersQuery struct {
	DocumentID uuid.UUID
}

type StakeholderDTO struct {
	DocumentID uuid.UUID
	UserID     string
	UserName   string
	Role       stakeholder.Role
	CreatedAt  time.Time
}
