package document

import appdoc "github.com/bbridges_11/document-registry/internal/application/document"

func toSearchResponse(output appdoc.SearchOutput) SearchDocumentResponse {
	results := make([]DocumentSearchResultResponse, 0, len(output.Results))
	for _, r := range output.Results {
		results = append(results, DocumentSearchResultResponse{
			ID:                  r.ID,
			Name:                r.Name,
			Description:         r.Description,
			DocumentType:        r.DocumentType,
			Tags:                r.Tags,
			CreatedBy:           r.CreatedBy,
			CreatedAt:           r.CreatedAt,
			UpdatedAt:           r.UpdatedAt,
			LatestVersion:       r.LatestVersion,
			LatestVersionStatus: r.LatestVersionStatus,
		})
	}

	return SearchDocumentResponse{
		Results: results,
		Pagination: PaginationResponse{
			Page:       output.Pagination.Page,
			Size:       output.Pagination.Size,
			Total:      output.Pagination.Total,
			TotalPages: output.Pagination.TotalPages,
		},
	}
}

func toDocumentResponse(v appdoc.DocumentView) DocumentResponse {
	return DocumentResponse{ID: v.ID, Name: v.Name, Description: v.Description, Tags: v.Tags, DocumentType: v.DocumentType, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, CreatedBy: v.CreatedBy}
}
func toCreateDocumentResponse(out appdoc.CreateOutput) DocumentWithVersionResponse {
	resp := DocumentWithVersionResponse{
		ID:           out.Document.ID,
		Name:         out.Document.Name,
		Description:  out.Document.Description,
		Tags:         out.Document.Tags,
		DocumentType: out.Document.DocumentType,
		CreatedAt:    out.Document.CreatedAt,
		UpdatedAt:    out.Document.UpdatedAt,
		CreatedBy:    out.Document.CreatedBy,
		Version: VersionResponse{
			ID:           out.Version.ID,
			DocumentID:   out.Version.DocumentID,
			Version:      out.Version.Version,
			Status:       out.Version.Status,
			ContentS3Key: out.Version.ContentKey,
			ContentHash:  out.Version.ContentHash,
			Metadata:     out.Version.Metadata,
			CreatedBy:    out.Version.CreatedBy,
			CreatedAt:    out.Version.CreatedAt,
			UpdatedAt:    out.Version.UpdatedAt,
		},
	}

	if out.ValidationResult != nil {
		issues := make([]ValidationIssue, 0, len(out.ValidationResult.Issues))
		for _, issue := range out.ValidationResult.Issues {
			issues = append(issues, ValidationIssue{
				Field:    issue.Field,
				Rule:     issue.Rule,
				Message:  issue.Message,
				Severity: issue.Severity,
			})
		}
		resp.ValidationResult = &ValidationResultView{
			Valid:  out.ValidationResult.Valid,
			Issues: issues,
		}
	}

	return resp
}
