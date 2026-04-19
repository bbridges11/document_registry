package user

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"go.uber.org/zap"
)

type UserQueryService struct {
	repo   outbound.UserRepository
	logger *zap.Logger
}

func NewUserQueryService(
	repo outbound.UserRepository,
	logger *zap.Logger,
) *UserQueryService {
	return &UserQueryService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserQueryService) GetUser(ctx context.Context, query GetUserQuery) (dto UserDTO, err error) {
	defer err2.Handle(&err)

	usr := try.To1(s.repo.GetByID(ctx, query.ID))
	return ToDTO(usr), nil
}

func (s *UserQueryService) GetUserByEmail(ctx context.Context, query GetUserByEmailQuery) (dto UserDTO, err error) {
	defer err2.Handle(&err)

	usr := try.To1(s.repo.GetByEmail(ctx, query.Email))
	return ToDTO(usr), nil
}

func (s *UserQueryService) ListUsers(ctx context.Context, query ListUsersQuery) (dtos []UserDTO, err error) {
	defer err2.Handle(&err)

	users := try.To1(s.repo.ListAll(ctx))
	return ToDTOs(users), nil
}

func (s *UserQueryService) UserExists(ctx context.Context, id uuid.UUID) (exists bool, err error) {
	defer err2.Handle(&err)

	return s.repo.Exists(ctx, id)
}
