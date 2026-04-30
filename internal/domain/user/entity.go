package user

import (
	"regexp"
	"strings"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	UserRoleAdmin       UserRole = "admin"
	UserRoleContributor UserRole = "contributor"
	UserRoleViewer      UserRole = "viewer"
)

func (r UserRole) IsValid() bool {
	return r == UserRoleAdmin || r == UserRoleContributor || r == UserRoleViewer
}

// externalIDPattern validates external ID format (7 alphanumeric characters)
var externalIDPattern = regexp.MustCompile(`^[a-z0-9]{7}$`)

// ValidateExternalID validates external ID format
func ValidateExternalID(externalID string) error {
	if externalID == "" {
		return errors.New(errors.CodeInvalidArgument, "external ID is required")
	}

	// Normalize to lowercase
	normalized := strings.ToLower(externalID)

	// Check length
	if len(normalized) != 7 {
		return errors.New(errors.CodeInvalidArgument, "external ID must be exactly 7 characters")
	}

	// Check format (alphanumeric only)
	if !externalIDPattern.MatchString(normalized) {
		return errors.New(errors.CodeInvalidArgument, "external ID must contain only alphanumeric characters")
	}

	return nil
}

// NormalizeExternalID normalizes external ID to lowercase
func NormalizeExternalID(externalID string) string {
	return strings.ToLower(externalID)
}

// User is the aggregate root for user management
type User struct {
	shared.AggregateRoot
	id         uuid.UUID
	externalID string // Client-provided unique identifier (e.g., "a123456")
	email      string
	name       string
	role       UserRole
	active     bool
	createdAt  time.Time
	updatedAt  time.Time
}

// NewUser creates a new user aggregate
func NewUser(externalID, email, name string, role UserRole) (*User, error) {
	// Validate external ID
	if err := ValidateExternalID(externalID); err != nil {
		return nil, err
	}

	if email == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "email is required")
	}
	if name == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "name is required")
	}
	if !role.IsValid() {
		return nil, errors.New(errors.CodeInvalidArgument, "invalid user role")
	}

	// Normalize external ID
	normalizedExternalID := NormalizeExternalID(externalID)

	user := &User{
		id:         uuid.New(),
		externalID: normalizedExternalID,
		email:      email,
		name:       name,
		role:       role,
		active:     true,
		createdAt:  time.Now().UTC(),
		updatedAt:  time.Now().UTC(),
	}

	// Record domain event
	user.AddEvent(events.NewUserCreated(
		user.id.String(),
		user.email,
		user.name,
		string(user.role),
	))

	return user, nil
}

// RehydrateUser recreates a user from persistence
func RehydrateUser(id uuid.UUID, externalID, email, name string, role UserRole, active bool, createdAt, updatedAt time.Time) *User {
	return &User{
		id:         id,
		externalID: externalID,
		email:      email,
		name:       name,
		role:       role,
		active:     active,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// UpdateInfo updates user information
func (u *User) UpdateInfo(name string, role UserRole) error {
	if name == "" {
		return errors.New(errors.CodeInvalidArgument, "name is required")
	}
	if !role.IsValid() {
		return errors.New(errors.CodeInvalidArgument, "invalid user role")
	}

	u.name = name
	u.role = role
	u.updatedAt = time.Now().UTC()

	// Record domain event
	u.AddEvent(events.NewUserUpdated(
		u.id.String(),
		u.name,
		string(u.role),
	))

	return nil
}

// Deactivate marks the user as inactive
func (u *User) Deactivate() {
	u.active = false
	u.updatedAt = time.Now().UTC()
}

// Activate marks the user as active
func (u *User) Activate() {
	u.active = true
	u.updatedAt = time.Now().UTC()
}

// Getters
func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) ExternalID() string   { return u.externalID }
func (u *User) Email() string        { return u.email }
func (u *User) Name() string         { return u.name }
func (u *User) Role() UserRole       { return u.role }
func (u *User) Active() bool         { return u.active }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
