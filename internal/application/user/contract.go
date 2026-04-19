package user

import (
	"context"

	"github.com/google/uuid"
)

type CommandService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (UserDTO, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCommand) (UserDTO, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
	ActivateUser(ctx context.Context, id uuid.UUID) error
}

type QueryService interface {
	GetUser(ctx context.Context, query GetUserQuery) (UserDTO, error)
	GetUserByEmail(ctx context.Context, query GetUserByEmailQuery) (UserDTO, error)
	ListUsers(ctx context.Context, query ListUsersQuery) ([]UserDTO, error)
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
}
