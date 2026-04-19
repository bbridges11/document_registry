package stakeholder

import (
	"time"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner       Role = "owner"
	RoleContributor Role = "contributor"
	RoleConsumer    Role = "consumer"
)

func (r Role) IsValid() bool {
	return r == RoleOwner || r == RoleContributor || r == RoleConsumer
}

type Stakeholder struct {
	documentID uuid.UUID
	userID     string
	role       Role
	createdAt  time.Time
}

func NewStakeholder(documentID uuid.UUID, userID string, role Role) (*Stakeholder, error) {
	if documentID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "document ID is required")
	}
	if userID == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "user ID is required")
	}
	if !role.IsValid() {
		return nil, errors.New(errors.CodeInvalidArgument, "invalid role")
	}

	return &Stakeholder{
		documentID: documentID,
		userID:     userID,
		role:       role,
		createdAt:  time.Now().UTC(),
	}, nil
}

func RehydrateStakeholder(documentID uuid.UUID, userID string, role Role, createdAt time.Time) *Stakeholder {
	return &Stakeholder{
		documentID: documentID,
		userID:     userID,
		role:       role,
		createdAt:  createdAt,
	}
}

func (s *Stakeholder) DocumentID() uuid.UUID {
	return s.documentID
}

func (s *Stakeholder) UserID() string {
	return s.userID
}

func (s *Stakeholder) Role() Role {
	return s.role
}

func (s *Stakeholder) CreatedAt() time.Time {
	return s.createdAt
}
