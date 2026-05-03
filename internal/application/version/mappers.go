package version

import (
	domainApproval "github.com/bbridges_11/document-registry/internal/domain/approval"
	domainVersion "github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

func toVersionEntityView(ver *domainVersion.Version) VersionView {
	return VersionView{ID: ver.ID(), DocumentID: ver.DocumentID(), Version: ver.Version().String(), Status: ver.Status().String(), ContentKey: ver.ContentRef().Location(), ContentHash: ver.ContentHash(), Metadata: ver.Metadata(), CreatedBy: ver.CreatedBy(), CreatedAt: ver.CreatedAt(), UpdatedAt: ver.UpdatedAt()}
}
func toVersionDTOView(dto *VersionDTO) VersionView {
	return VersionView{
		ID:          dto.ID,
		DocumentID:  dto.DocumentID,
		Version:     dto.Version,
		Status:      dto.Status.String(),
		ContentKey:  dto.ContentKey,
		ContentHash: dto.ContentHash,
		Content:     dto.Content,
		ContentType: dto.ContentType,
		Metadata:    dto.Metadata,
		CreatedBy:   dto.CreatedBy,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}
func stringifyRoleMap(in map[domainApproval.ApprovalRole]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k.String()] = v
	}
	return out
}
func toApprovalSummaryView(dto ApprovalSummaryDTO) ApprovalSummaryView {
	return ApprovalSummaryView{Required: stringifyRoleMap(dto.Required), Received: stringifyRoleMap(dto.Received), Remaining: stringifyRoleMap(dto.Remaining), Complete: dto.Complete}
}
func toVersionStatusView(dto *VersionStatusDTO) VersionStatusView {
	actions := make([]string, len(dto.AllowedActions))
	for i, a := range dto.AllowedActions {
		actions[i] = string(a)
	}
	return VersionStatusView{ID: dto.ID, Version: dto.Version, Status: dto.Status.String(), Editable: dto.Editable, Terminal: dto.Terminal, Publishable: dto.Publishable, AllowedActions: actions, ApprovalSummary: toApprovalSummaryView(dto.ApprovalSummary)}
}
func toVersionApprovalsView(dto *VersionApprovalsDTO) VersionApprovalsView {
	approvals := make([]ApprovalView, 0, len(dto.Approvals))
	for _, appr := range dto.Approvals {
		approvals = append(approvals, ApprovalView{ID: appr.ID, UserID: appr.UserID, UserName: appr.UserName, Role: appr.Role.String(), Approved: appr.Approved, Comment: appr.Comment, CreatedAt: appr.CreatedAt})
	}
	audit := make([]ApprovalAuditView, 0, len(dto.AuditTrail))
	for _, entry := range dto.AuditTrail {
		audit = append(audit, ApprovalAuditView{UserID: entry.UserID, UserName: entry.UserName, Role: entry.Role.String(), Action: entry.Action, Comment: entry.Comment, Timestamp: entry.Timestamp})
	}
	return VersionApprovalsView{VersionID: dto.VersionID, Version: dto.Version, Status: dto.Status.String(), Approvals: approvals, Summary: toApprovalSummaryView(dto.Summary), AuditTrail: audit}
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
