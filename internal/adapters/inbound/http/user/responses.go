package user

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse represents HTTP response for user operations
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UsersResponse represents HTTP response for list users
type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Count int            `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
