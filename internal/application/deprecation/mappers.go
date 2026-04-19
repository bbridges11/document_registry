package deprecation

import "github.com/bbridges_11/document-registry/internal/domain/deprecation"

// toDeprecationOutput maps a deprecation entity to output view
func toDeprecationOutput(dep *deprecation.Deprecation) DeprecationOutput {
	return DeprecationOutput{
		ID:              dep.ID(),
		VersionID:       dep.VersionID(),
		DocumentID:      dep.DocumentID(),
		RequestedBy:     dep.RequestedBy(),
		RequestedAt:     dep.RequestedAt(),
		Reason:          dep.Reason(),
		DeprecationNote: dep.DeprecationNote(),
		AutoDeprecated:  dep.AutoDeprecated(),
		Status:          dep.Status().String(),
		DeprecatedBy:    dep.DeprecatedBy(),
		DeprecatedAt:    dep.DeprecatedAt(),
		SupersededBy:    dep.SupersededBy(),
		PreviousStatus:  dep.PreviousStatus(),
		CreatedAt:       dep.CreatedAt(),
		UpdatedAt:       dep.UpdatedAt(),
	}
}
