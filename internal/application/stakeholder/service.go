package stakeholder

import "context"

type UseCaseService struct {
	commands *CommandService
	queries  *QueryService
}

func NewService(commands *CommandService, queries *QueryService) *UseCaseService {
	return &UseCaseService{commands: commands, queries: queries}
}
func (s *UseCaseService) Add(ctx context.Context, input AddInput) error {
	return s.commands.AddStakeholder(ctx, AddStakeholderCommand{DocumentID: input.DocumentID, UserID: input.UserID, Role: input.Role, AddedBy: input.Actor.UserID})
}
func (s *UseCaseService) Remove(ctx context.Context, input RemoveInput) error {
	return s.commands.RemoveStakeholder(ctx, RemoveStakeholderCommand{DocumentID: input.DocumentID, UserID: input.UserID, RemovedBy: input.Actor.UserID})
}
func (s *UseCaseService) List(ctx context.Context, input ListInput) ([]View, error) {
	dtos, err := s.queries.ListStakeholders(ctx, ListStakeholdersQuery{DocumentID: input.DocumentID}, input.Actor.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(dtos))
	for _, dto := range dtos {
		out = append(out, View{DocumentID: dto.DocumentID, UserID: dto.UserID, UserName: dto.UserName, Role: string(dto.Role), CreatedAt: dto.CreatedAt})
	}
	return out, nil
}
