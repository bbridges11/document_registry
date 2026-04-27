package approval

import (
	"context"

	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/google/uuid"
)

type ServiceFacade struct {
	queries *QueryService
}

func NewService(queries *QueryService) *ServiceFacade {
	return &ServiceFacade{queries: queries}
}

func toView(dto *ApprovalDTO) View {
	return View{ID: dto.ID, VersionID: dto.VersionID, UserID: dto.UserID, UserName: dto.UserName, Role: dto.Role.String(), Approved: dto.Approved, Comment: dto.Comment, CreatedAt: dto.CreatedAt}
}
func (s *ServiceFacade) Get(ctx context.Context, input GetInput) (View, error) {
	dto, err := s.queries.GetApproval(ctx, GetApprovalQuery{ID: input.ID})
	if err != nil {
		return View{}, err
	}
	return toView(dto), nil
}
func (s *ServiceFacade) ListByVersion(ctx context.Context, input ListByVersionInput) ([]View, error) {
	dtos, err := s.queries.ListApprovalsByVersion(ctx, ListApprovalsByVersionQuery{VersionID: input.VersionID})
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(dtos))
	for _, dto := range dtos {
		out = append(out, toView(dto))
	}
	return out, nil
}
func (s *ServiceFacade) GetSummary(ctx context.Context, versionID uuid.UUID) (SummaryView, error) {
	dto, err := s.queries.GetApprovalSummary(ctx, versionID)
	if err != nil {
		return SummaryView{}, err
	}
	return SummaryView{VersionID: dto.VersionID, Required: stringifySummaryMap(dto.Required), Received: stringifySummaryMap(dto.Received), Remaining: stringifySummaryMap(dto.Remaining), Complete: dto.Complete}, nil
}
func stringifySummaryMap(in map[domainApproval.ApprovalRole]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k.String()] = v
	}
	return out
}
