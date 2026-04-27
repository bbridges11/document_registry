package deprecation

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

// Service implements the deprecation UseCase interface
type Service struct {
	commands *CommandService
}

// NewService creates a new deprecation service
func NewService(commands *CommandService) UseCase {
	return &Service{
		commands: commands,
	}
}

// RequestDeprecation implements UseCase
func (s *Service) RequestDeprecation(ctx context.Context, input RequestDeprecationInput) (DeprecationOutput, error) {
	dep, err := s.commands.RequestDeprecation(ctx, RequestDeprecationCommand{
		VersionID:       input.VersionID,
		RequestedBy:     input.Actor.UserID,
		Reason:          input.Reason,
		DeprecationNote: input.DeprecationNote,
	})
	if err != nil {
		return DeprecationOutput{}, err
	}
	return toDeprecationOutput(dep), nil
}

// ApproveDeprecation implements UseCase
func (s *Service) ApproveDeprecation(ctx context.Context, input ApproveDeprecationInput) error {
	return s.commands.ApproveDeprecation(ctx, ApproveDeprecationCommand{
		VersionID:  input.VersionID,
		ApprovedBy: input.Actor.UserID,
		Role:       input.Role,
		Comment:    input.Comment,
	})
}

// RejectDeprecation implements UseCase
func (s *Service) RejectDeprecation(ctx context.Context, input RejectDeprecationInput) error {
	return s.commands.RejectDeprecation(ctx, RejectDeprecationCommand{
		VersionID:  input.VersionID,
		RejectedBy: input.Actor.UserID,
		Reason:     input.Reason,
	})
}

// CancelDeprecation implements UseCase
func (s *Service) CancelDeprecation(ctx context.Context, input CancelDeprecationInput) error {
	return s.commands.CancelDeprecation(ctx, CancelDeprecationCommand{
		VersionID:  input.VersionID,
		CanceledBy: input.Actor.UserID,
	})
}

// GetDeprecation implements UseCase
func (s *Service) GetDeprecation(ctx context.Context, input GetDeprecationInput) (DeprecationOutput, error) {
	dep, err := s.commands.GetDeprecation(ctx, input.VersionID)
	if err != nil {
		return DeprecationOutput{}, err
	}
	return toDeprecationOutput(dep), nil
}

// GetDeprecationStatus implements UseCase
func (s *Service) GetDeprecationStatus(ctx context.Context, input GetDeprecationInput) (DeprecationStatusOutput, error) {
	dep, err := s.commands.GetDeprecation(ctx, input.VersionID)
	if err != nil {
		return DeprecationStatusOutput{}, err
	}

	// TODO: Implement permission checks for CanApprove, CanReject, CanCancel
	// This will use AuthorizationService in Phase 3
	return DeprecationStatusOutput{
		Status:     dep.Status().String(),
		CanApprove: false, // Will be implemented with authz checks
		CanReject:  false,
		CanCancel:  false,
	}, nil
}

// CommandService holds dependencies for deprecation commands
type CommandService struct {
	deprecationRepo outbound.DeprecationRepository
	versionRepo     outbound.VersionRepository
	txManager       outbound.TransactionManager
	eventBus        outbound.EventBus
}

// NewCommandService creates a new command service
func NewCommandService(
	deprecationRepo outbound.DeprecationRepository,
	versionRepo outbound.VersionRepository,
	txManager outbound.TransactionManager,
	eventBus outbound.EventBus,
) *CommandService {
	return &CommandService{
		deprecationRepo: deprecationRepo,
		versionRepo:     versionRepo,
		txManager:       txManager,
		eventBus:        eventBus,
	}
}
