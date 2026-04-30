package publication

import (
	appPublication "github.com/bbridges_11/document-registry/internal/application/publication"
)

// toPublicationResponse maps PublicationView to PublicationResponse
func toPublicationResponse(view appPublication.PublicationView) PublicationResponse {
	return PublicationResponse{
		ID:           view.ID.String(),
		VersionID:    view.VersionID.String(),
		DocumentID:   view.DocumentID.String(),
		PublishedTo:  view.PublishedTo,
		PublishedBy:  view.PublishedBy,
		PublishedAt:  view.PublishedAt,
		Status:       view.Status,
		ErrorMessage: view.ErrorMessage,
	}
}

// toPublicationListResponse maps PublicationListResult to PublicationListResponse
func toPublicationListResponse(result appPublication.PublicationListResult) PublicationListResponse {
	publications := make([]PublicationResponse, len(result.Publications))
	for i, view := range result.Publications {
		publications[i] = toPublicationResponse(view)
	}

	return PublicationListResponse{
		Publications: publications,
		Total:        result.Total,
		Limit:        result.Limit,
		Offset:       result.Offset,
	}
}
