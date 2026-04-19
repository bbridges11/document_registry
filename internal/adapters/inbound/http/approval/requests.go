package approval

// GrantApprovalRequest represents HTTP request for granting an approval
type GrantApprovalRequest struct {
	Role    string `json:"role"`
	Comment string `json:"comment"`
}

// RevokeApprovalRequest represents HTTP request for revoking an approval
type RevokeApprovalRequest struct{}
