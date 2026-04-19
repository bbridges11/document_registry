package stakeholder

import (
	"context"
	"time"

	domainStakeholder "github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/google/uuid"
)

type Actor struct{ UserID string }

type AddInput struct {
	Actor      Actor
	DocumentID uuid.UUID
	UserID     string
	Role       domainStakeholder.Role
}
type RemoveInput struct {
	Actor      Actor
	DocumentID uuid.UUID
	UserID     string
}
type ListInput struct {
	Actor      Actor
	DocumentID uuid.UUID
}

type View struct {
	DocumentID uuid.UUID
	UserID     string
	UserName   string
	Role       string
	CreatedAt  time.Time
}

type Service interface {
	Add(ctx context.Context, input AddInput) error
	Remove(ctx context.Context, input RemoveInput) error
	List(ctx context.Context, input ListInput) ([]View, error)
}
