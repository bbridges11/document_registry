package user

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"go.uber.org/zap"
)

type UserCommandService struct {
	repo     outbound.UserRepository
	eventBus outbound.EventBus
	logger   *zap.Logger
}

func NewUserCommandService(
	repo outbound.UserRepository,
	eventBus outbound.EventBus,
	logger *zap.Logger,
) *UserCommandService {
	return &UserCommandService{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *UserCommandService) CreateUser(ctx context.Context, cmd CreateUserCommand) (dto UserDTO, err error) {
	defer err2.Handle(&err)

	// Check if user already exists
	var existing *user.User
	existing, err = s.repo.GetByEmail(ctx, cmd.Email)
	if errors.IsCode(err, errors.CodeNotFound) {

		if existing != nil {
			return UserDTO{}, errors.New(errors.CodeConflict, "user with this email already exists")
		}

		// Parse role
		role := user.UserRole(cmd.Role)
		if !role.IsValid() {
			return UserDTO{}, errors.New(errors.CodeInvalidArgument, "invalid user role")
		}

		// Create domain aggregate
		usr := try.To1(user.NewUser(cmd.Email, cmd.Name, role))

		// Persist
		try.To(s.repo.Save(ctx, usr))

		// Publish events
		s.publishEvents(ctx, usr, "")

		s.logger.Info("user created",
			zap.String("user_id", usr.ID().String()),
			zap.String("email", usr.Email()),
		)

		return ToDTO(usr), nil
	}

	return UserDTO{}, err
}

func (s *UserCommandService) UpdateUser(ctx context.Context, cmd UpdateUserCommand) (dto UserDTO, err error) {
	defer err2.Handle(&err)

	// Load user
	usr := try.To1(s.repo.GetByID(ctx, cmd.ID))
	if usr == nil {
		return UserDTO{}, errors.ErrNotFound
	}

	// Parse role
	role := user.UserRole(cmd.Role)

	// Update
	try.To(usr.UpdateInfo(cmd.Name, role))

	// Persist
	try.To(s.repo.Save(ctx, usr))

	// Publish events
	s.publishEvents(ctx, usr, "")

	s.logger.Info("user updated",
		zap.String("user_id", usr.ID().String()),
	)

	return ToDTO(usr), nil
}

func (s *UserCommandService) DeactivateUser(ctx context.Context, id uuid.UUID) (err error) {
	defer err2.Handle(&err)

	usr := try.To1(s.repo.GetByID(ctx, id))
	if usr == nil {
		return errors.ErrNotFound
	}

	usr.Deactivate()
	try.To(s.repo.Save(ctx, usr))

	s.logger.Info("user deactivated",
		zap.String("user_id", id.String()),
	)

	return nil
}

func (s *UserCommandService) ActivateUser(ctx context.Context, id uuid.UUID) (err error) {
	defer err2.Handle(&err)

	usr := try.To1(s.repo.GetByID(ctx, id))
	if usr == nil {
		return errors.ErrNotFound
	}

	usr.Activate()
	try.To(s.repo.Save(ctx, usr))

	s.logger.Info("user activated",
		zap.String("user_id", id.String()),
	)

	return nil
}

func (s *UserCommandService) DeleteUser(ctx context.Context, id uuid.UUID) (err error) {
	defer err2.Handle(&err)

	usr := try.To1(s.repo.GetByID(ctx, id))
	if usr == nil {
		return errors.ErrNotFound
	}

	try.To(s.repo.Delete(ctx, id))

	s.logger.Info("user deleted",
		zap.String("user_id", id.String()),
	)

	return nil
}

func (s *UserCommandService) publishEvents(ctx context.Context, usr *user.User, traceID string) {
	domainEvents := usr.Events()
	if len(domainEvents) == 0 {
		return
	}

	envelopes := make([]events.Envelope, 0, len(domainEvents))
	for _, event := range domainEvents {
		envelopes = append(envelopes, events.NewEnvelope(event, traceID))
	}

	s.eventBus.Publish(ctx, envelopes...)
	usr.ClearEvents()
}
