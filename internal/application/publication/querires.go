package publication

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// QueryServiceImpl implements QueryService
type QueryServiceImpl struct {
	publicationRepo outbound.PublicationRepository
}

// NewQueryService creates a new publication query service
func NewQueryService(publicationRepo outbound.PublicationRepository) *QueryServiceImpl {
	return &QueryServiceImpl{
		publicationRepo: publicationRepo,
	}
}

// GetPublication retrieves a single publication by ID
func (s *QueryServiceImpl) GetPublication(ctx context.Context, query GetPublicationQuery) (view PublicationView, err error) {
	defer err2.Handle(&err)

	pub := try.To1(s.publicationRepo.GetByID(ctx, query.ID))

	return PublicationView{
		ID:           pub.ID(),
		VersionID:    pub.VersionID(),
		DocumentID:   pub.DocumentID(),
		PublishedTo:  pub.PublishedTo(),
		PublishedBy:  pub.PublishedBy(),
		PublishedAt:  pub.PublishedAt(),
		Status:       string(pub.Status()),
		ErrorMessage: pub.ErrorMessage(),
	}, nil
}

// ListPublicationsByVersion retrieves all publications for a specific version
func (s *QueryServiceImpl) ListPublicationsByVersion(ctx context.Context, query ListPublicationsByVersionQuery) (views []PublicationView, err error) {
	defer err2.Handle(&err)

	pubs := try.To1(s.publicationRepo.ListByVersionID(ctx, query.VersionID))

	views = make([]PublicationView, len(pubs))
	for i, pub := range pubs {
		views[i] = PublicationView{
			ID:           pub.ID(),
			VersionID:    pub.VersionID(),
			DocumentID:   pub.DocumentID(),
			PublishedTo:  pub.PublishedTo(),
			PublishedBy:  pub.PublishedBy(),
			PublishedAt:  pub.PublishedAt(),
			Status:       string(pub.Status()),
			ErrorMessage: pub.ErrorMessage(),
		}
	}

	return views, nil
}

// ListPublicationsByDocument retrieves all publications for a specific document
func (s *QueryServiceImpl) ListPublicationsByDocument(ctx context.Context, query ListPublicationsByDocumentQuery) (views []PublicationView, err error) {
	defer err2.Handle(&err)

	pubs := try.To1(s.publicationRepo.ListByDocumentID(ctx, query.DocumentID))

	views = make([]PublicationView, len(pubs))
	for i, pub := range pubs {
		views[i] = PublicationView{
			ID:           pub.ID(),
			VersionID:    pub.VersionID(),
			DocumentID:   pub.DocumentID(),
			PublishedTo:  pub.PublishedTo(),
			PublishedBy:  pub.PublishedBy(),
			PublishedAt:  pub.PublishedAt(),
			Status:       string(pub.Status()),
			ErrorMessage: pub.ErrorMessage(),
		}
	}

	return views, nil
}

// ListAllPublications retrieves all publications with pagination
func (s *QueryServiceImpl) ListAllPublications(ctx context.Context, query ListAllPublicationsQuery) (result PublicationListResult, err error) {
	defer err2.Handle(&err)

	// Get total count
	total := try.To1(s.publicationRepo.Count(ctx))

	// Get paginated results
	pubs := try.To1(s.publicationRepo.ListAll(ctx, query.Limit, query.Offset))

	views := make([]PublicationView, len(pubs))
	for i, pub := range pubs {
		views[i] = PublicationView{
			ID:           pub.ID(),
			VersionID:    pub.VersionID(),
			DocumentID:   pub.DocumentID(),
			PublishedTo:  pub.PublishedTo(),
			PublishedBy:  pub.PublishedBy(),
			PublishedAt:  pub.PublishedAt(),
			Status:       string(pub.Status()),
			ErrorMessage: pub.ErrorMessage(),
		}
	}

	return PublicationListResult{
		Publications: views,
		Total:        total,
		Limit:        query.Limit,
		Offset:       query.Offset,
	}, nil
}
