package stakeholder

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/google/uuid"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type QueryService struct {
	repo         outbound.StakeholderRepository
	documentRepo outbound.DocumentRepository

	userRepo outbound.UserRepository
}

func NewQueryService(
	repo outbound.StakeholderRepository,
	documentRepo outbound.DocumentRepository,

	userRepo outbound.UserRepository,
) *QueryService {
	return &QueryService{
		repo:         repo,
		documentRepo: documentRepo,

		userRepo: userRepo,
	}
}

func (s *QueryService) ListStakeholders(ctx context.Context, query ListStakeholdersQuery, userID string) (dtos []*StakeholderDTO, err error) {
	defer err2.Handle(&err)

	// Validate document exists
	try.To1(s.documentRepo.GetByID(ctx, query.DocumentID))

	stakeholders := try.To1(s.repo.ListByDocumentID(ctx, query.DocumentID))

	// Enrich with user information
	userIDs := make([]string, 0, len(stakeholders))
	for _, sh := range stakeholders {
		userIDs = append(userIDs, sh.UserID())
	}

	// Fetch users individually and build map
	userMap := make(map[string]string) // map[userID]userName
	for _, userID := range userIDs {
		userUUID := try.To1(uuid.Parse(userID))
		user := try.To1(s.userRepo.GetByID(ctx, userUUID))
		userMap[userID] = user.Name()
	}

	dtos = make([]*StakeholderDTO, 0, len(stakeholders))
	for _, sh := range stakeholders {
		userName := ""
		if name, ok := userMap[sh.UserID()]; ok {
			userName = name
		}

		dtos = append(dtos, &StakeholderDTO{
			DocumentID: sh.DocumentID(),
			UserID:     sh.UserID(),
			UserName:   userName,
			Role:       sh.Role(),
			CreatedAt:  sh.CreatedAt(),
		})
	}

	return dtos, nil
}
