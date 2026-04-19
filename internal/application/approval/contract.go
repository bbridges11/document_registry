package approval

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GrantInput struct {
	VersionID uuid.UUID
	UserID    string
	Role      string
	Comment   string
}
type RevokeInput struct {
	VersionID uuid.UUID
	UserID    string
}
type GetInput struct{ ID uuid.UUID }
type ListByVersionInput struct{ VersionID uuid.UUID }

type View struct {
	ID        uuid.UUID
	VersionID uuid.UUID
	UserID    string
	UserName  string
	Role      string
	Approved  bool
	Comment   string
	CreatedAt time.Time
}
type SummaryView struct {
	VersionID uuid.UUID
	Required  map[string]int
	Received  map[string]int
	Remaining map[string]int
	Complete  bool
}

type Service interface {
	Grant(ctx context.Context, input GrantInput) error
	Revoke(ctx context.Context, input RevokeInput) error
	Get(ctx context.Context, input GetInput) (View, error)
	ListByVersion(ctx context.Context, input ListByVersionInput) ([]View, error)
	GetSummary(ctx context.Context, versionID uuid.UUID) (SummaryView, error)
}
