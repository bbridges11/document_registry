package deprecation

import appdep "github.com/bbridges_11/document-registry/internal/application/deprecation"

// toDeprecationResponse maps application output to HTTP response
func toDeprecationResponse(output appdep.DeprecationOutput) DeprecationResponse {
	return DeprecationResponse{
		ID:              output.ID,
		VersionID:       output.VersionID,
		DocumentID:      output.DocumentID,
		RequestedBy:     output.RequestedBy,
		RequestedAt:     output.RequestedAt,
		Reason:          output.Reason,
		DeprecationNote: output.DeprecationNote,
		AutoDeprecated:  output.AutoDeprecated,
		Status:          output.Status,
		DeprecatedBy:    output.DeprecatedBy,
		DeprecatedAt:    output.DeprecatedAt,
		SupersededBy:    output.SupersededBy,
		PreviousStatus:  output.PreviousStatus,
		CreatedAt:       output.CreatedAt,
		UpdatedAt:       output.UpdatedAt,
	}
}

// toDeprecationStatusResponse maps application status output to HTTP response
func toDeprecationStatusResponse(output appdep.DeprecationStatusOutput) DeprecationStatusResponse {
	return DeprecationStatusResponse{
		Status:     output.Status,
		CanApprove: output.CanApprove,
		CanReject:  output.CanReject,
		CanCancel:  output.CanCancel,
	}
}
