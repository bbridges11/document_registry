package stakeholder

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type CommandService struct {
	repo         outbound.StakeholderRepository
	documentRepo outbound.DocumentRepository

	userRepo outbound.UserRepository
	eventBus outbound.EventBus
}

func NewCommandService(
	repo outbound.StakeholderRepository,
	documentRepo outbound.DocumentRepository,

	userRepo outbound.UserRepository,
	eventBus outbound.EventBus,
) *CommandService {
	return &CommandService{
		repo:         repo,
		documentRepo: documentRepo,

		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (s *CommandService) AddStakeholder(ctx context.Context, cmd AddStakeholderCommand) (err error) {
	defer err2.Handle(&err)

	// Validate document exists
	try.To1(s.documentRepo.GetByID(ctx, cmd.DocumentID))

	// Validate user exists
	userID := try.To1(uuid.Parse(cmd.UserID))

	try.To1(s.userRepo.GetByID(ctx, userID))

	// Check if stakeholder already exists
	exists := try.To1(s.repo.ExistsByDocumentIDAndUserID(ctx, cmd.DocumentID, cmd.UserID))
	if exists {
		return errors.New(errors.CodeConflict, "stakeholder already exists")
	}

	// Create stakeholder
	sh := try.To1(stakeholder.NewStakeholder(cmd.DocumentID, cmd.UserID, cmd.Role))
	try.To(s.repo.Save(ctx, sh))

	// Publish event
	event := events.NewStakeholderAdded(
		cmd.DocumentID.String(),
		cmd.UserID,
		string(cmd.Role),
		cmd.AddedBy,
	)
	envelope := events.NewEnvelope(event, "")
	s.eventBus.Publish(ctx, envelope)

	return nil
}

func (s *CommandService) RemoveStakeholder(ctx context.Context, cmd RemoveStakeholderCommand) (err error) {
	defer err2.Handle(&err)

	// Validate document exists
	try.To1(s.documentRepo.GetByID(ctx, cmd.DocumentID))

	// Get stakeholder to find their role
	sh := try.To1(s.repo.GetByDocumentIDAndUserID(ctx, cmd.DocumentID, cmd.UserID))

	// Delete stakeholder
	try.To(s.repo.Delete(ctx, cmd.DocumentID, cmd.UserID))

	// Publish event with role information
	event := events.NewStakeholderRemoved(
		cmd.DocumentID.String(),
		cmd.UserID,
		string(sh.Role()),
		cmd.RemovedBy,
	)
	envelope := events.NewEnvelope(event, "")
	s.eventBus.Publish(ctx, envelope)

	return nil
}
