package document

import (
	domainDoc "github.com/bbridges_11/document-registry/internal/domain/document"
	domainVer "github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

func toDocumentView(doc *domainDoc.Document) DocumentView {
	return DocumentView{
		ID:           doc.ID(),
		Name:         doc.Name(),
		Description:  doc.Description(),
		Tags:         doc.Tags(),
		DocumentType: doc.DocumentType().Code(),
		CreatedAt:    doc.CreatedAt(),
		UpdatedAt:    doc.UpdatedAt(),
		CreatedBy:    doc.CreatedBy(),
	}
}

func toVersionView(ver *domainVer.Version) VersionView {
	return VersionView{
		ID:          ver.ID(),
		DocumentID:  ver.DocumentID(),
		Version:     ver.Version().String(),
		Status:      ver.Status().String(),
		ContentKey:  ver.ContentRef().Location(),
		ContentHash: ver.ContentHash(),
		Metadata:    ver.Metadata(),
		CreatedBy:   ver.CreatedBy(),
		CreatedAt:   ver.CreatedAt(),
		UpdatedAt:   ver.UpdatedAt(),
	}
}

func toDocumentDTOView(dto *DocumentDTO) DocumentView {
	return DocumentView{
		ID:           dto.ID,
		Name:         dto.Name,
		Description:  dto.Description,
		Tags:         dto.Tags,
		DocumentType: dto.DocumentType,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
		CreatedBy:    dto.CreatedBy,
	}
}

func toValidationResultView(result *outbound.ValidationResult) *ValidationResultView {
	if result == nil {
		return nil
	}

	issues := make([]ValidationIssue, 0, len(result.Issues))
	for _, issue := range result.Issues {
		issues = append(issues, ValidationIssue{
			Field:    issue.Field,
			Rule:     issue.Rule,
			Message:  issue.Message,
			Severity: issue.Severity,
		})
	}

	return &ValidationResultView{
		Valid:  result.Valid,
		Issues: issues,
	}
}
