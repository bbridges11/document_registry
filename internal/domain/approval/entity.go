package approval

import (
	"time"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

type ApprovalRole string

const (
	ApprovalRoleTechnical ApprovalRole = "technical"
	ApprovalRoleArchitect ApprovalRole = "architect"
	ApprovalRoleProduct   ApprovalRole = "product"
)

func (a ApprovalRole) String() string {
	return string(a)
}

func (r ApprovalRole) IsValid() bool {
	return r == ApprovalRoleTechnical || r == ApprovalRoleArchitect || r == ApprovalRoleProduct
}

type Approval struct {
	id        uuid.UUID
	versionID uuid.UUID
	userID    string
	role      ApprovalRole
	approved  bool
	comment   string
	createdAt time.Time
}

func NewApproval(versionID uuid.UUID, userID string, role ApprovalRole, approved bool, comment string) (*Approval, error) {
	if versionID == uuid.Nil {
		return nil, errors.New(errors.CodeInvalidArgument, "version ID is required")
	}
	if userID == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "user ID is required")
	}
	if !role.IsValid() {
		return nil, errors.New(errors.CodeInvalidArgument, "invalid approval role")
	}

	return &Approval{
		id:        uuid.New(),
		versionID: versionID,
		userID:    userID,
		role:      role,
		approved:  approved,
		comment:   comment,
		createdAt: time.Now().UTC(),
	}, nil
}

func RehydrateApproval(id, versionID uuid.UUID, userID string, role ApprovalRole, approved bool, comment string, createdAt time.Time) *Approval {
	return &Approval{
		id:        id,
		versionID: versionID,
		userID:    userID,
		role:      role,
		approved:  approved,
		comment:   comment,
		createdAt: createdAt,
	}
}

func (a *Approval) ID() uuid.UUID {
	return a.id
}

func (a *Approval) VersionID() uuid.UUID {
	return a.versionID
}

func (a *Approval) UserID() string {
	return a.userID
}

func (a *Approval) Role() ApprovalRole {
	return a.role
}

func (a *Approval) Approved() bool {
	return a.approved
}

func (a *Approval) Comment() string {
	return a.comment
}

func (a *Approval) CreatedAt() time.Time {
	return a.createdAt
}
