package stakeholder

import "github.com/bbridges_11/document-registry/internal/domain/stakeholder"

// AddStakeholderRequest represents HTTP request for adding a stakeholder
type AddStakeholderRequest struct {
	UserID string           `json:"user_id"`
	Role   stakeholder.Role `json:"role"`
}
