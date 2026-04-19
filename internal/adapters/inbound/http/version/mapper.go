package version

import appver "github.com/bbridges_11/document-registry/internal/application/version"

func toVersionResponse(v appver.VersionView) VersionResponse {
	return VersionResponse{ID: v.ID, DocumentID: v.DocumentID, Version: v.Version, Status: v.Status, ContentS3Key: v.ContentKey, ContentHash: v.ContentHash, Metadata: v.Metadata, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func toCreateVersionResponse(output appver.CreateOutput) CreateVersionResponse {
	resp := CreateVersionResponse{
		Version: toVersionResponse(output.Version),
	}

	if output.ValidationResult != nil {
		issues := make([]ValidationIssue, 0, len(output.ValidationResult.Issues))
		for _, issue := range output.ValidationResult.Issues {
			issues = append(issues, ValidationIssue{
				Field:    issue.Field,
				Rule:     issue.Rule,
				Message:  issue.Message,
				Severity: issue.Severity,
			})
		}
		resp.ValidationResult = &ValidationResultView{
			Valid:  output.ValidationResult.Valid,
			Issues: issues,
		}
	}

	return resp
}

func toApprovalSummaryResponse(v appver.ApprovalSummaryView) ApprovalSummaryResponse {
	return ApprovalSummaryResponse{Required: v.Required, Received: v.Received, Remaining: v.Remaining, Complete: v.Complete}
}
func toVersionStatusResponse(v appver.VersionStatusView) VersionStatusResponse {
	return VersionStatusResponse{ID: v.ID, Version: v.Version, Status: v.Status, Editable: v.Editable, Terminal: v.Terminal, Publishable: v.Publishable, AllowedActions: v.AllowedActions, ApprovalSummary: toApprovalSummaryResponse(v.ApprovalSummary)}
}
func toVersionApprovalsResponse(v appver.VersionApprovalsView) VersionApprovalsResponse {
	approvals := make([]ApprovalResponse, 0, len(v.Approvals))
	for _, a := range v.Approvals {
		approvals = append(approvals, ApprovalResponse{ID: a.ID, UserID: a.UserID, UserName: a.UserName, Role: a.Role, Approved: a.Approved, Comment: a.Comment, CreatedAt: a.CreatedAt})
	}
	audit := make([]ApprovalAuditResponse, 0, len(v.AuditTrail))
	for _, a := range v.AuditTrail {
		audit = append(audit, ApprovalAuditResponse{UserID: a.UserID, UserName: a.UserName, Role: a.Role, Action: a.Action, Comment: a.Comment, Timestamp: a.Timestamp})
	}
	return VersionApprovalsResponse{VersionID: v.VersionID, Version: v.Version, Status: v.Status, Approvals: approvals, Summary: toApprovalSummaryResponse(v.Summary), AuditTrail: audit}
}
