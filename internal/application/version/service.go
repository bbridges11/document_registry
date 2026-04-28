package version

import (
	"context"
)

type Service struct {
	commands *CommandService
	queries  *QueryService
}

func NewService(commands *CommandService, queries *QueryService) *Service {
	return &Service{commands: commands, queries: queries}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	ver, valResult, err := s.commands.CreateVersion(ctx, CreateVersionCommand{
		DocumentID: input.DocumentID,
		Version:    input.Version,
		Content:    input.Content,
		Metadata:   input.Metadata,
		CreatedBy:  input.Actor.UserID,
	})
	if err != nil {
		// Check if it's a validation error with results
		if valResult != nil && !valResult.Valid {
			// Return partial output with validation errors
			return CreateOutput{
				ValidationResult: toValidationResultView(valResult),
			}, err
		}
		return CreateOutput{}, err
	}

	output := CreateOutput{
		Version: toVersionEntityView(ver),
	}

	if valResult != nil {
		output.ValidationResult = toValidationResultView(valResult)
	}

	return output, nil
}
func (s *Service) Get(ctx context.Context, input GetInput) (VersionView, error) {
	dto, err := s.queries.GetVersion(ctx, GetVersionQuery{ID: input.ID}, input.Actor.UserID)
	if err != nil {
		return VersionView{}, err
	}
	return toVersionDTOView(dto), nil
}
func (s *Service) ListByDocument(ctx context.Context, input ListByDocumentInput) ([]VersionView, error) {
	dtos, err := s.queries.ListVersionsByDocument(ctx, ListVersionsByDocumentQuery{DocumentID: input.DocumentID}, input.Actor.UserID)
	if err != nil {
		return nil, err
	}
	views := make([]VersionView, 0, len(dtos))
	for _, dto := range dtos {
		views = append(views, toVersionDTOView(dto))
	}
	return views, nil
}
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	return s.commands.UpdateVersion(ctx, UpdateVersionCommand{
		ID:        input.ID,
		Metadata:  input.Metadata,
		UpdatedBy: input.Actor.UserID,
	})
}
func (s *Service) Submit(ctx context.Context, input SubmitInput) error {
	return s.commands.SubmitVersion(ctx, SubmitVersionCommand{ID: input.ID, SubmittedBy: input.Actor.UserID})
}
func (s *Service) Review(ctx context.Context, input ReviewInput) error {
	return s.commands.ReviewVersion(ctx, ReviewVersionCommand{ID: input.ID, ReviewedBy: input.Actor.UserID})
}
func (s *Service) Approve(ctx context.Context, input ApproveInput) error {
	return s.commands.ApproveVersion(ctx, ApproveVersionCommand{ID: input.ID, ApprovedBy: input.Actor.UserID, Comment: input.Comment})
}
func (s *Service) Reject(ctx context.Context, input RejectInput) error {
	return s.commands.RejectVersion(ctx, RejectVersionCommand{ID: input.ID, RejectedBy: input.Actor.UserID, Reason: input.Reason})
}
func (s *Service) Publish(ctx context.Context, input PublishInput) error {
	return s.commands.PublishVersion(ctx, PublishVersionCommand{ID: input.ID, PublishedBy: input.Actor.UserID})
}
func (s *Service) GetStatus(ctx context.Context, input GetStatusInput) (VersionStatusView, error) {
	dto, err := s.queries.GetVersionStatus(ctx, GetVersionStatusQuery{ID: input.ID}, input.Actor.UserID)
	if err != nil {
		return VersionStatusView{}, err
	}
	return toVersionStatusView(dto), nil
}
func (s *Service) GetApprovals(ctx context.Context, input GetApprovalsInput) (VersionApprovalsView, error) {
	dto, err := s.queries.GetVersionApprovals(ctx, GetVersionApprovalsQuery{ID: input.ID}, input.Actor.UserID)
	if err != nil {
		return VersionApprovalsView{}, err
	}
	return toVersionApprovalsView(dto), nil
}
