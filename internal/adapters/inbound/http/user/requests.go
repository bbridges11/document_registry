package user

// CreateUserRequest represents HTTP request for creating a user
type CreateUserRequest struct {
	ExternalID string `json:"external_id" validate:"required,len=7"`
	Email      string `json:"email" validate:"required,email"`
	Name       string `json:"name" validate:"required"`
	Role       string `json:"role" validate:"required,oneof=admin contributor viewer"`
}

// UpdateUserRequest represents HTTP request for updating a user
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required"`
	Role string `json:"role" validate:"required,oneof=admin contributor viewer"`
}
