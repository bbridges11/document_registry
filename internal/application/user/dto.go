package user

import (
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/google/uuid"
)

// CreateUserCommand represents the command to create a new user
type CreateUserCommand struct {
	Email string
	Name  string
	Role  string
}

// UpdateUserCommand represents the command to update a user
type UpdateUserCommand struct {
	ID   uuid.UUID
	Name string
	Role string
}

// GetUserQuery represents a query to get a user by ID
type GetUserQuery struct {
	ID uuid.UUID
}

// GetUserByEmailQuery represents a query to get a user by email
type GetUserByEmailQuery struct {
	Email string
}

// ListUsersQuery represents a query to list all users
type ListUsersQuery struct {
	// Could add pagination params here
}

// UserDTO represents user data transfer object
type UserDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDTO converts domain user to DTO
func ToDTO(u *user.User) UserDTO {
	return UserDTO{
		ID:        u.ID(),
		Email:     u.Email(),
		Name:      u.Name(),
		Role:      string(u.Role()),
		Active:    u.Active(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

// ToDTOs converts slice of domain users to DTOs
func ToDTOs(users []*user.User) []UserDTO {
	dtos := make([]UserDTO, len(users))
	for i, u := range users {
		dtos[i] = ToDTO(u)
	}
	return dtos
}
