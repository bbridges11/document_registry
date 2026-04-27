package document

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type QueryService struct {
	repo        outbound.DocumentRepository
	versionRepo outbound.VersionRepository
}

func NewQueryService(
	repo outbound.DocumentRepository,
	versionRepo outbound.VersionRepository,

) *QueryService {
	return &QueryService{
		repo:        repo,
		versionRepo: versionRepo,
	}
}

func (s *QueryService) GetDocument(ctx context.Context, query GetDocumentQuery, userID string) (dto *DocumentDTO, err error) {
	defer err2.Handle(&err)

	doc := try.To1(s.repo.GetByID(ctx, query.ID))

	// Check if document has published version (public access)
	hasPublished := try.To1(s.versionRepo.HasPublishedVersion(ctx, doc.ID()))
	if hasPublished {
		// Document is public - skip authorization
		return &DocumentDTO{
			ID:           doc.ID(),
			Name:         doc.Name(),
			Description:  doc.Description(),
			Tags:         doc.Tags(),
			DocumentType: doc.DocumentType().String(),
			CreatedAt:    doc.CreatedAt(),
			UpdatedAt:    doc.UpdatedAt(),
			CreatedBy:    doc.CreatedBy(),
		}, nil
	}

	return &DocumentDTO{
		ID:           doc.ID(),
		Name:         doc.Name(),
		Description:  doc.Description(),
		Tags:         doc.Tags(),
		DocumentType: doc.DocumentType().String(),
		CreatedAt:    doc.CreatedAt(),
		UpdatedAt:    doc.UpdatedAt(),
		CreatedBy:    doc.CreatedBy(),
	}, nil
}

func (s *QueryService) ListDocuments(ctx context.Context, query ListDocumentsQuery, userID string) (dtos []*DocumentDTO, err error) {
	defer err2.Handle(&err)

	docs := try.To1(s.repo.List(ctx, query.Limit, query.Offset))

	dtos = make([]*DocumentDTO, 0, len(docs))
	for _, doc := range docs {
		// Check if document has published version (public access)

		dtos = append(dtos, &DocumentDTO{
			ID:           doc.ID(),
			Name:         doc.Name(),
			Description:  doc.Description(),
			Tags:         doc.Tags(),
			DocumentType: doc.DocumentType().String(),
			CreatedAt:    doc.CreatedAt(),
			UpdatedAt:    doc.UpdatedAt(),
			CreatedBy:    doc.CreatedBy(),
		})
	}

	return dtos, nil
}

func (s *QueryService) SearchDocuments(ctx context.Context, input SearchInput) (output SearchOutput, err error) {
	defer err2.Handle(&err)

	// Build filters from input
	filters := outbound.DocumentSearchFilters{
		Query:             input.Query,
		Name:              input.Name,
		Description:       input.Description,
		Tags:              input.Tags,
		DocumentType:      input.DocumentType,
		CreatedBy:         input.CreatedBy,
		CreatedAfter:      input.CreatedAfter,
		CreatedBefore:     input.CreatedBefore,
		UserID:            input.Actor.UserID,
		MyDocuments:       input.MyDocuments,
		MyStakeholderDocs: input.MyStakeholderDocs,
		PendingMyReview:   input.PendingMyReview,
		PendingMyApproval: input.PendingMyApproval,
		RecentlyPublished: input.RecentlyPublished,
		SortField:         input.SortField,
		SortOrder:         input.SortOrder,
	}

	// Set default values
	if filters.SortField == "" {
		filters.SortField = "created_at"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}

	// Pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	size := input.Size
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	filters.Limit = size
	filters.Offset = (page - 1) * size

	// Execute search
	results, total := try.To2(s.repo.Search(ctx, filters))

	// Filter results by authorization
	// For each document, check if user can access it
	authorizedResults := make([]DocumentSearchResult, 0)

	for _, result := range results {
		// Check if document has published version (public access)
		hasPublished := result.LatestVersionStatus == "PUBLISHED"

		if hasPublished {
			// Published documents are visible to everyone
			authorizedResults = append(authorizedResults, DocumentSearchResult{
				ID:                  result.ID,
				Name:                result.Name,
				Description:         result.Description,
				DocumentType:        result.DocumentType,
				Tags:                result.Tags,
				CreatedBy:           result.CreatedBy,
				CreatedAt:           result.CreatedAt,
				UpdatedAt:           result.UpdatedAt,
				LatestVersion:       result.LatestVersion,
				LatestVersionStatus: result.LatestVersionStatus,
			})
			continue
		}

	}

	// Calculate pagination info
	totalPages := (total + size - 1) / size

	return SearchOutput{
		Results: authorizedResults,
		Pagination: PaginationInfo{
			Page:       page,
			Size:       size,
			Total:      len(authorizedResults), // Total after authorization filter
			TotalPages: totalPages,
		},
	}, nil
}
